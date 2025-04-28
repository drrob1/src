# Network Printer Image Sender

This Go program allows you to send images to network attached printers using various connection methods.

## Features

- Support for multiple connection types:
  - Windows printer API (for printers installed in Windows)
  - Direct TCP/IP connection (for raw network printers)
  - Internet Printing Protocol (IPP)
- Authentication support for printers that require it
- Simple command-line interface

## Requirements

- Go 1.16 or higher
- Required Go packages:
  - github.com/disintegration/imaging
  - github.com/godoes/printers (for Windows printing)

## Installation

1. Clone or download this repository
2. Install the required dependencies:

```bash
go get github.com/disintegration/imaging
go get github.com/godoes/printers
```

3. Build the program:

```bash
go build -o netprint.exe
```

## Usage

```
netprint [options] <image_file>
```

### Options

- `-printer` - Printer name (for Windows printers)
- `-address` - Printer IP address or hostname (required for TCP and IPP)
- `-port` - Printer port (default: 9100 for raw TCP)
- `-type` - Connection type: "windows", "tcp", or "ipp" (default: "windows")
- `-user` - Username for authenticated printers
- `-pass` - Password for authenticated printers

### Examples

#### Print to a Windows printer (including network printers installed in Windows)

```bash
netprint -printer "HP LaserJet Pro" image.jpg
```

If no printer is specified, the default Windows printer will be used:

```bash
netprint -type windows image.jpg
```

#### Print to a raw network printer using TCP/IP

```bash
netprint -type tcp -address 192.168.1.100 -port 9100 image.jpg
```

#### Print to an IPP-enabled printer

```bash
netprint -type ipp -address 192.168.1.100 -port 631 image.jpg
```

#### Print to a printer requiring authentication

```bash
netprint -type ipp -address 192.168.1.100 -user admin -pass secret image.jpg
```

## How It Works

### Windows Printing

When using the Windows connection type, the program uses the Windows printer API through the `github.com/godoes/printers` package. This works with any printer installed in Windows, including network printers that have been added to Windows.

### TCP/IP Printing

For direct TCP/IP printing, the program establishes a direct socket connection to the printer (typically on port 9100) and sends the image data directly. This is the most basic form of network printing and works with most network printers.

### IPP Printing

For IPP printing, the program sends an HTTP POST request to the printer's IPP endpoint (typically `/ipp/print`) with the image data. This is a more modern approach that supports additional features like authentication.

## Troubleshooting

- If you get an error connecting to the printer, verify that the printer is turned on and connected to the network
- For Windows printing, make sure the printer is properly installed in Windows
- For TCP/IP printing, try pinging the printer to ensure it's reachable
- For IPP printing, check if the printer supports IPP and verify the correct port (usually 631)
- Some printers may require specific data formats; this program sends JPEG-encoded images which should work with most printers

## License

This software is provided as-is under the MIT License.