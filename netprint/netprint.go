package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/godoes/printers"
)

/*
  29 Aug 26 -- Seeing if I can get this to work using AI.  I'll refactor to, open the printer connection, convert the image to []byte and then print the image.
				Or maybe, convert the image to []byte and then open the printer connection and print the image.  Looks like all 3 steps occur in printImage.
  30 Aug 26 -- Codex reports that this is windows only code.  I forgot that.
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
	flag.StringVar(&printerName, "printer", "HP8620", "Printer name")
	flag.StringVar(&printerAddr, "adr", HP8620, "Printer IP address or hostname")            // default is 192.168.1.208 for HP8620
	flag.IntVar(&printerPort, "port", HP8620Port, "Printer port (default 9100 for raw TCP)") // 9100 is for raw TCP access
	flag.StringVar(&connType, "type", "tcp", "Connection type (windows, tcp, ipp)")
	flag.StringVar(&username, "user", "", "Username for authenticated printers")
	flag.StringVar(&password, "pass", "", "Password for authenticated printers")
	flag.Parse()

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
	address := fmt.Sprintf("%s:%d", config.Address, config.Port)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return fmt.Errorf("error connecting to printer at %s: %v", address, err)
	}
	defer conn.Close()

	// Write the data
	_, err = conn.Write(data)
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

	// Get printer name
	printerName := config.Name
	if printerName == "" {
		// If no printer name specified, use default printer
		var err error
		printerName, err = printers.GetDefault()
		if err != nil {
			return fmt.Errorf("error getting default printer: %v", err)
		}
	}

	// Open the printer
	p, err := printers.Open(printerName)
	if err != nil {
		return fmt.Errorf("error opening printer %s: %v", printerName, err)
	}
	defer p.Close()

	// Start a document
	err = p.StartDocument("Image Print", "RAW")
	if err != nil {
		return fmt.Errorf("error starting document: %v", err)
	}

	// Start a page
	err = p.StartPage()
	if err != nil {
		return fmt.Errorf("error starting page: %v", err)
	}

	// Write the data
	_, err = p.Write(data)
	if err != nil {
		return fmt.Errorf("error writing to printer: %v", err)
	}

	// End the page
	err = p.EndPage()
	if err != nil {
		return fmt.Errorf("error ending page: %v", err)
	}

	// End the document
	err = p.EndDocument()
	if err != nil {
		return fmt.Errorf("error ending document: %v", err)
	}

	return nil
}

// printWithIPP prints using the Internet Printing Protocol (IPP)
func printWithIPP(data []byte, config PrinterConfig) error {
	if config.Address == "" {
		return fmt.Errorf("printer address is required for IPP printing")
	}

	// Construct the IPP URL
	ippURL := fmt.Sprintf("http://%s:%d/ipp/print", config.Address, config.Port)

	// Create a new HTTP request
	req, err := http.NewRequest("POST", ippURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/octet-stream")

	// Add basic authentication if credentials are provided
	if config.Username != "" && config.Password != "" {
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
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("printer returned error: %s - %s", resp.Status, string(body))
	}

	return nil
}
