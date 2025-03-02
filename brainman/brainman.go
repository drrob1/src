package main

import (
	"fmt"
	"github.com/alexbrainman/printer"
	"log"
)

func main() {

	// Get the default printer name
	name, err := printer.Default()
	if err != nil {
		log.Fatalf("Failed to get default printer: %v", err)
	}
	fmt.Printf("Default printer: %s\n", name)

	// Open the printer
	p, err := printer.Open(name)
	if err != nil {
		log.Fatalf("Failed to open printer: %v", err)
	}
	defer p.Close()

	// Start a new document
	err = p.StartDocument("Test Document", "RAW")
	if err != nil {
		log.Fatalf("Failed to start document: %v", err)
	}

	// Start a new page
	err = p.StartPage()
	if err != nil {
		log.Fatalf("Failed to start page: %v", err)
	}

	// Write some content to the printer
	content := []byte("Hello, Printer!")
	_, err = p.Write(content)
	if err != nil {
		log.Fatalf("Failed to write to printer: %v", err)
	}

	// End the page
	err = p.EndPage()
	if err != nil {
		log.Fatalf("Failed to end page: %v", err)
	}

	// End the document
	err = p.EndDocument()
	if err != nil {
		log.Fatalf("Failed to end document: %v", err)
	}

	fmt.Println("Document sent to printer successfully")
}
