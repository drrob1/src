package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
)

/*
  29 Aug 26 -- Seeing if I can get this to work using AI.  I'll refactor to, open the printer connection, convert the image to []byte and then print the image.
				Or maybe, convert the image to []byte and then open the printer connection and print the image.  Looks like all 3 steps occur in printImage.
  30 Aug 26 -- Codex reports that this is windows only code.  I forgot that.
               I asked Codex to fix the bugs.  It essentially rewrote the code.
The program now has three distinct printing paths, chosen by -type: Windows, raw TCP, or IPP.

Startup and flags
In netprint.go:56, main defines the command-line options.
Key behavior:
• -type defaults to windows.
• -address specifies a network printer.
• -adr remains as a shorter alias.
• TCP defaults to port 9100.
• IPP automatically switches to port 631 unless -port was explicitly supplied.
• An omitted Windows printer name means “use the default printer.”

The image is loaded and decoded, then printImage converts it to JPEG bytes.

Selecting a printing method

printImage dispatches based on the connection type:
switch strings.ToLower(config.ConnectionType) {
case "windows": return printWithWindowsAPI(...)
case "tcp": return printWithTCP(...)
case "ipp": return printWithIPP(...)
}

Raw TCP printing
printWithTCP connects directly to the printer, normally on port 9100.
address := net.JoinHostPort(config.Address, strconv.Itoa(config.Port))
net.JoinHostPort correctly formats both IPv4 and IPv6 addresses. The data is sent using:
io.Copy(conn, bytes.NewReader(data))
Unlike one conn.Write call, io.Copy continues until the entire JPEG has been written or an error occurs.
This mode still requires a printer that directly understands JPEG data.

Windows printing
printWithWindowsAPI no longer submits the JPEG as a RAW spooler job.
Instead, it:
1. Writes the JPEG to a temporary file.
2. Starts a noninteractive PowerShell process.
3. Loads the image through System.Drawing.
4. Creates a normal Windows PrintDocument.
5. Scales and centers the image inside the printable page margins.
6. Sends the rendered page through the selected Windows printer driver.
7. Disposes the Windows objects and deletes the temporary file.

The image path and printer name are passed through environment variables:
cmd.Env = append(
    os.Environ(),
    "NETPRINT_IMAGE="+tempName,
    "NETPRINT_PRINTER="+config.Name,
)

This avoids command-line quoting problems and prevents printer names from being interpreted as PowerShell code.

IPP printing
printWithIPP sends a proper Internet Printing Protocol request over HTTP. makeIPPPrintJob constructs the binary message containing:
• IPP version 2.0
• Print-Job operation
• Request ID
• UTF-8 character set
• Language
• Printer URI
• Username, when supplied
• image/jpeg document format
• The JPEG itself writeIPPAttribute encodes every attribute as: value tag → name length → name → value length → value

The HTTP request uses the required media type: req.Header.Set("Content-Type", "application/ipp")

After receiving the response, the code checks both levels of status:
• HTTP must report a successful 2xx result.
• The binary IPP status must be in its success range.

That distinction matters because an IPP server can return HTTP 200 while the actual print operation failed.

Overall flow
image file
    ↓
decode image
    ↓
encode as JPEG
    ↓
Windows renderer / raw TCP / IPP Print-Job
    ↓
printer

The Windows mode performs page rendering locally. TCP and IPP send JPEG, so those modes depend on the printer advertising JPEG support.
*/

const lastModified = "30 Aug 2026"

// Constants for printer connection types
const (
	ConnectionTypeWindows = "windows" // Use Windows printer API
	ConnectionTypeTCP     = "tcp"     // Direct TCP/IP connection
	ConnectionTypeIPP     = "ipp"     // Internet Printing Protocol
)
const HP8028e = "192.168.1.197"
const HP8028ePort = 9100
const HP8028 = "192.168.1.197"
const HP8620 = "192.168.1.208"
const HP8620Port = 9100

var formatNames = []string{"JPEG", "PNG", "GIF", "TIFF", "BMP"}

// PrinterConfig holds the configuration for connecting to a printer
type PrinterConfig struct {
	Name           string // Printer name (for Windows printers)
	Address        string // IP address or hostname for network printers
	Port           int    // Port number for network printers (default 9100 for raw TCP)
	ConnectionType string // Type of connection (windows, tcp, ipp)
	Username       string // Username for authenticated printers
	Password       string // Password for authenticated printers
}

func main() {
	var printerName, printerAddr string
	var printerPort int
	var connType, username, password string

	fmt.Printf(" Netprint Last modified: %s\n", lastModified)
	// Parse command line flags
	flag.StringVar(&printerName, "printer", "", "Printer name (default Windows printer if omitted)")
	flag.StringVar(&printerAddr, "address", HP8620, "Printer IP address or hostname")
	flag.StringVar(&printerAddr, "adr", HP8620, "Alias for -address")
	flag.IntVar(&printerPort, "port", HP8620Port, "Printer port (default 9100 for raw TCP)") // 9100 is for raw TCP access
	flag.StringVar(&connType, "type", "windows", "Connection type (windows, tcp, ipp)")
	flag.StringVar(&username, "user", "", "Username for authenticated printers")
	flag.StringVar(&password, "pass", "", "Password for authenticated printers")
	flag.Parse()
	portSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "port" {
			portSet = true
		}
	})
	if strings.EqualFold(connType, ConnectionTypeIPP) && !portSet {
		printerPort = 631
	}

	// Check if an image file was provided
	if flag.NArg() == 0 {
		fmt.Println("Usage: netprint [options] <image_file>")
		flag.PrintDefaults()
		return
	}

	// Get the image file path
	imagePath := flag.Arg(0)
	fullImagePath, err := filepath.Abs(imagePath)
	if err != nil {
		log.Fatalf("Error getting absolute path: %v", err)
	}

	// Create printer configuration
	config := PrinterConfig{
		Name:           printerName,
		Address:        printerAddr,
		Port:           printerPort,
		ConnectionType: connType,
		Username:       username,
		Password:       password,
	}

	// Load the image
	img, format, err := loadImage(fullImagePath)
	if err != nil {
		log.Fatalf("Error loading image: %v", err)
	}

	fmt.Printf("Image format: %s\n", formatNames[format])

	// Print the image
	err = printImage(img, config)
	if err != nil {
		log.Fatalf("Error printing image: %v", err)
	}

	fmt.Printf("Image %s sent to printer successfully\n", filepath.Base(fullImagePath))
}

// loadImage loads an image from the specified path
func loadImage(path string) (image.Image, imaging.Format, error) {
	// Check if file exists
	_, err := os.Stat(path)
	if err != nil {
		return nil, 0, fmt.Errorf("error accessing file: %v", err)
	}

	format, err := imaging.FormatFromFilename(path)
	if err != nil {
		return nil, 0, fmt.Errorf("error getting image format: %v", err)
	}

	// Load the image using imaging library
	img, err := imaging.Open(path)
	if err != nil {
		return nil, 0, fmt.Errorf("error opening image: %v", err)
	}

	return img, format, nil
}

// printImage sends the image to the printer based on the configuration
func printImage(img image.Image, config PrinterConfig) error {

	// Encode the image as JPEG
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, nil)
	//err := png.Encode(&buf, img)
	if err != nil {
		return fmt.Errorf("error encoding image: %v", err)
	}

	// Print based on connection type
	switch strings.ToLower(config.ConnectionType) {
	case ConnectionTypeWindows:
		return printWithWindowsAPI(buf.Bytes(), config)
	case ConnectionTypeTCP:
		return printWithTCP(buf.Bytes(), config)
	case ConnectionTypeIPP:
		return printWithIPP(buf.Bytes(), config)
	default:
		return fmt.Errorf("unsupported connection type: %s", config.ConnectionType)
	}
}

// printWithTCP prints directly to a network printer using TCP/IP
func printWithTCP(data []byte, config PrinterConfig) error {
	if config.Address == "" {
		return fmt.Errorf("printer address is required for TCP printing")
	}

	// Connect to the printer
	address := net.JoinHostPort(config.Address, strconv.Itoa(config.Port))
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return fmt.Errorf("error connecting to printer at %s: %v", address, err)
	}
	defer conn.Close()

	// Write all the data; net.Conn.Write is allowed to perform a short write.
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error sending data to printer: %v", err)
	}

	return nil
}

// printWithWindowsAPI prints using the Windows printer API
func printWithWindowsAPI(data []byte, config PrinterConfig) error {
	// Check if running on Windows
	if runtime.GOOS != "windows" {
		return fmt.Errorf("windows printing is only supported on Windows OS")
	}

	temp, err := os.CreateTemp("", "netprint-*.jpg")
	if err != nil {
		return fmt.Errorf("error creating temporary image: %w", err)
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return fmt.Errorf("error writing temporary image: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("error closing temporary image: %w", err)
	}

	// Render through GDI so the Windows printer driver receives printer-ready data,
	// rather than incorrectly submitting the JPEG as a RAW spool job.
	const script = `Add-Type -AssemblyName System.Drawing
$image = [System.Drawing.Image]::FromFile($env:NETPRINT_IMAGE)
$document = New-Object System.Drawing.Printing.PrintDocument
if ($env:NETPRINT_PRINTER) { $document.PrinterSettings.PrinterName = $env:NETPRINT_PRINTER }
if (-not $document.PrinterSettings.IsValid) { throw "Printer '$env:NETPRINT_PRINTER' is not valid" }
$handler = [System.Drawing.Printing.PrintPageEventHandler]{
    param($sender,$event)
    $scale = [Math]::Min($event.MarginBounds.Width / $image.Width, $event.MarginBounds.Height / $image.Height)
    $width = [int]($image.Width * $scale)
    $height = [int]($image.Height * $scale)
    $x = $event.MarginBounds.Left + [int](($event.MarginBounds.Width - $width) / 2)
    $y = $event.MarginBounds.Top + [int](($event.MarginBounds.Height - $height) / 2)
    $event.Graphics.DrawImage($image,$x,$y,$width,$height)
}
$document.add_PrintPage($handler)
try { $document.Print() } finally { $document.remove_PrintPage($handler); $document.Dispose(); $image.Dispose() }`
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.Env = append(os.Environ(), "NETPRINT_IMAGE="+tempName, "NETPRINT_PRINTER="+config.Name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error printing through Windows: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// printWithIPP prints using the Internet Printing Protocol (IPP)
func printWithIPP(data []byte, config PrinterConfig) error {
	if config.Address == "" {
		return fmt.Errorf("printer address is required for IPP printing")
	}

	// Construct the IPP URL
	ippURL := "http://" + net.JoinHostPort(config.Address, strconv.Itoa(config.Port)) + "/ipp/print"

	body, err := makeIPPPrintJob(ippURL, config.Username, data)
	if err != nil {
		return err
	}

	// IPP uses HTTP as its transport, but requires an IPP binary message body.
	req, err := http.NewRequest("POST", ippURL, body)
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/ipp")

	// Add basic authentication if credentials are provided
	if config.Username != "" || config.Password != "" {
		req.SetBasicAuth(config.Username, config.Password)
	}

	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("printer returned error: %s - %s", resp.Status, string(body))
	}
	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading IPP response: %w", err)
	}
	if len(response) < 8 {
		return fmt.Errorf("printer returned an invalid IPP response")
	}
	status := binary.BigEndian.Uint16(response[2:4])
	if status > 0x00ff {
		return fmt.Errorf("printer returned IPP status 0x%04x", status)
	}

	return nil
}

func makeIPPPrintJob(printerURL, username string, document []byte) (*bytes.Reader, error) {
	var message bytes.Buffer
	message.Write([]byte{2, 0, 0, 2}) // IPP 2.0, Print-Job operation
	if err := binary.Write(&message, binary.BigEndian, uint32(1)); err != nil {
		return nil, fmt.Errorf("error encoding IPP request: %w", err)
	}
	message.WriteByte(0x01) // operation-attributes-tag
	writeIPPAttribute(&message, 0x47, "attributes-charset", "utf-8")
	writeIPPAttribute(&message, 0x48, "attributes-natural-language", "en")
	writeIPPAttribute(&message, 0x45, "printer-uri", strings.Replace(printerURL, "http://", "ipp://", 1))
	if username != "" {
		writeIPPAttribute(&message, 0x42, "requesting-user-name", username)
	}
	writeIPPAttribute(&message, 0x49, "document-format", "image/jpeg")
	message.WriteByte(0x03) // end-of-attributes-tag
	message.Write(document)
	return bytes.NewReader(message.Bytes()), nil
}

func writeIPPAttribute(dst *bytes.Buffer, tag byte, name, value string) {
	dst.WriteByte(tag)
	_ = binary.Write(dst, binary.BigEndian, uint16(len(name)))
	dst.WriteString(name)
	_ = binary.Write(dst, binary.BigEndian, uint16(len(value)))
	dst.WriteString(value)
}
