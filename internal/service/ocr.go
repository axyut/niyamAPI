package service

import (
	"context"
	"fmt"
	"log"

	"github.com/otiai10/gosseract/v2"
)

// OCRService defines the interface for OCR-related business logic.
type OCRService interface {
	// ExtractTextFromImage now accepts raw image data as a byte slice and a language string.
	ExtractTextFromImage(ctx context.Context, imageData []byte, language string) (string, error)
}

// ocrService implements the OCRService interface.
type ocrService struct {
	// Add any dependencies for the OCR service here, e.g., configuration for Tesseract
}

// NewOCRService creates a new instance of OCRService.
func NewOCRService() OCRService {
	return &ocrService{}
}

// ExtractTextFromImage performs OCR on raw image data using the specified language(s).
func (s *ocrService) ExtractTextFromImage(ctx context.Context, imageData []byte, language string) (string, error) {
	// Ensure that image data is not empty to avoid errors with gosseract.
	if len(imageData) == 0 {
		return "", fmt.Errorf("empty image data provided")
	}

	client := gosseract.NewClient()
	defer client.Close() // Crucial: Ensure the Tesseract client is closed after use to release resources.

	// Set the OCR language(s). Tesseract accepts comma or plus-separated codes (e.g., "eng", "nep", "hin", "eng+nep").
	// This must be done before setting the image.
	if err := client.SetLanguage(language); err != nil {
		log.Printf("ERROR: Failed to set OCR language '%s': %v", language, err)
		return "", fmt.Errorf("unsupported OCR language or missing language data for '%s'", language)
	}

	// Set the image for OCR directly from the byte slice.
	if err := client.SetImageFromBytes(imageData); err != nil {
		log.Printf("ERROR: Failed to set image for OCR from bytes: %v", err)
		return "", fmt.Errorf("failed to prepare image for OCR processing")
	}

	// Perform the Optical Character Recognition.
	text, err := client.Text()
	if err != nil {
		log.Printf("ERROR: Failed to extract text from image using Tesseract (language: %s): %v", language, err)
		return "", fmt.Errorf("failed to extract text from image")
	}

	return text, nil
}
