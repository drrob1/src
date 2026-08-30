package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMakeIPPPrintJob(t *testing.T) {
	document := []byte("jpeg data")
	request, err := makeIPPPrintJob("http://printer:631/ipp/print", "rob", document)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(request)
	if err != nil {
		t.Fatal(err)
	}

	if got := body[:8]; string(got[:4]) != string([]byte{2, 0, 0, 2}) || binary.BigEndian.Uint32(got[4:]) != 1 {
		t.Fatalf("invalid IPP header: %v", got)
	}
	for _, value := range []string{"attributes-charset", "printer-uri", "ipp://printer:631/ipp/print", "requesting-user-name", "image/jpeg"} {
		if !strings.Contains(string(body), value) {
			t.Errorf("IPP request does not contain %q", value)
		}
	}
	if !strings.HasSuffix(string(body), string(document)) {
		t.Error("IPP request does not end with the document data")
	}
}

func TestPrintWithIPP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "application/ipp" {
			t.Errorf("Content-Type = %q, want application/ipp", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		if len(body) < 8 || body[0] != 2 || body[3] != 2 {
			t.Errorf("invalid IPP request header: %v", body)
		}
		w.Header().Set("Content-Type", "application/ipp")
		_, _ = w.Write([]byte{2, 0, 0, 0, 0, 0, 0, 1})
	}))
	defer server.Close()

	address := strings.TrimPrefix(server.URL, "http://")
	host, port, ok := strings.Cut(address, ":")
	if !ok {
		t.Fatalf("unexpected server address %q", address)
	}
	var portNumber int
	if _, err := fmt.Sscanf(port, "%d", &portNumber); err != nil {
		t.Fatal(err)
	}
	if err := printWithIPP([]byte("jpeg"), PrinterConfig{Address: host, Port: portNumber}); err != nil {
		t.Fatal(err)
	}
}

func TestPrintWithIPPRejectsIPPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte{2, 0, 4, 0, 0, 0, 0, 1})
	}))
	defer server.Close()

	address := strings.TrimPrefix(server.URL, "http://")
	host, port, _ := strings.Cut(address, ":")
	var portNumber int
	_, _ = fmt.Sscanf(port, "%d", &portNumber)
	err := printWithIPP([]byte("jpeg"), PrinterConfig{Address: host, Port: portNumber})
	if err == nil || !strings.Contains(err.Error(), "IPP status 0x0400") {
		t.Fatalf("got %v, want IPP status error", err)
	}
}
