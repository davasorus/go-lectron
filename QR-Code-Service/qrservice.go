package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/skip2/go-qrcode"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// QRService handles QR Code generation
type QRService struct {
	route string
}

// NewQRService creates a new instance of QRService
func NewQRService() *QRService {
	return &QRService{}
}

func (s *QRService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.route = options.Route
	return nil
}

// Generate generates a QR code for the given text and size.
func (s *QRService) Generate(text string, size int) ([]byte, error) {
	qr, err := qrcode.New(text, qrcode.Medium)
	if err != nil {
		return nil, err
	}

	png, err := qr.PNG(size)
	if err != nil {
		return nil, err
	}

	return png, nil
}

func (s *QRService) URL(text string, size int) (string, error) {
	if s.route == "" {
		return "", errors.New("http handler unavailable")
	}

	return fmt.Sprintf("%s?text=%s&size=%d", s.route, url.QueryEscape(text), size), nil
}

func (s *QRService) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	text := r.URL.Query().Get("text")
	if text == "" {
		http.Error(w, "Missing 'text' query parameter", http.StatusBadRequest)
		return
	}

	sizeStr := r.URL.Query().Get("size")
	if sizeStr == "" {
		sizeStr = "256" // default size
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 {
		http.Error(w, "Invalid 'size' query parameter", http.StatusBadRequest)
		return
	}

	qrCodeData, err := s.Generate(text, size)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	w.Write(qrCodeData)
}
