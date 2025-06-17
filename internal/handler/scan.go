package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"github.com/axyut/niyamAPI/internal/types"
)

// RegisterScanHandlers registers the OCR scanning endpoint with the API.
// It's a method on the Handlers struct, giving it access to services.
func (h *Handlers) RegisterScanHandlers(api huma.API) {

	huma.Post(api, "/scan", func(ctx context.Context, input *types.ScanInput) (*types.ScanOutput, error) {
		log.Println("INFO: Received request for OCR scan (multipart form).")

		formData := input.RawBody.Data()

		// Validate if an image file was provided.
		if !formData.Image.IsSet {
			log.Println("ERROR: No image file provided in the form ('image' field is missing or empty).")
			return nil, huma.Error400BadRequest("No image file provided. Please upload a file with the 'image' field.", nil)
		}

		// Read the content of the uploaded image file.
		imageData, err := io.ReadAll(formData.Image)
		if err != nil {
			log.Printf("ERROR: Failed to read uploaded image file: %v", err)
			return nil, huma.Error400BadRequest(fmt.Sprintf("Failed to read image file: %v", err), nil)
		}

		// --- Language Validation and Normalization ---
		requestedLanguage := formData.Language
		if requestedLanguage == "" {
			requestedLanguage = types.LangEnglish // Default language if not explicitly provided
		}

		// Split the input language string by '+' or ',' to handle multiple languages.
		// Trim spaces and filter out empty parts.
		var rawLangCodes []string
		if strings.Contains(requestedLanguage, "+") {
			rawLangCodes = strings.Split(requestedLanguage, "+")
		} else if strings.Contains(requestedLanguage, ",") {
			rawLangCodes = strings.Split(requestedLanguage, ",")
		} else {
			rawLangCodes = []string{requestedLanguage}
		}

		validatedLangCodes := []string{}
		invalidLanguages := []string{}

		for _, code := range rawLangCodes {
			trimmedCode := strings.TrimSpace(code)
			if trimmedCode == "" {
				continue // Skip empty strings that might result from split (e.g., "eng++hin")
			}
			if _, ok := types.SupportedOCRLanguages[trimmedCode]; ok {
				validatedLangCodes = append(validatedLangCodes, trimmedCode)
			} else {
				invalidLanguages = append(invalidLanguages, trimmedCode)
			}
		}

		// Handle cases where all provided languages are invalid, or none were provided at all.
		if len(invalidLanguages) > 0 {
			errorMessage := fmt.Sprintf("Unsupported language code(s) found: '%s'. Supported codes are: %s.",
				strings.Join(invalidLanguages, "', '"), strings.Join(getSortedSupportedLangCodes(), ", "))
			log.Printf("ERROR: %s", errorMessage)
			return nil, huma.Error400BadRequest(errorMessage, nil)
		}

		// If no valid languages remain after filtering (e.g., input was "++"), fallback to default.
		if len(validatedLangCodes) == 0 {
			validatedLangCodes = append(validatedLangCodes, types.LangEnglish)
			log.Println("INFO: No valid language codes provided or all were invalid; defaulting to 'eng'.")
		}

		// Join validated language codes with '+' for Tesseract.
		finalLanguage := strings.Join(validatedLangCodes, "+")
		log.Printf("INFO: OCR language(s) finalized: %s", finalLanguage)
		// --- End Language Validation ---

		// Call the OCRService with both image data and the finalized language string.
		text, err := h.Services.OCRService.ExtractTextFromImage(ctx, imageData, finalLanguage)
		if err != nil {
			log.Printf("ERROR: Failed to process image for OCR: %v", err)
			return nil, huma.Error400BadRequest(fmt.Sprintf("Failed to process image for OCR: %v", err), nil)
		}

		log.Println("INFO: Text extracted successfully from image.")
		return &types.ScanOutput{
			Body: struct {
				Text string `json:"text" huma:"example:Extracted text from the image"`
			}{
				Text: text,
			},
		}, nil
	})
}

// getSortedSupportedLangCodes is a helper function to get a sorted list of supported language codes.
// This is used for consistent and readable error messages.
func getSortedSupportedLangCodes() []string {
	codes := []string{}
	for code := range types.SupportedOCRLanguages {
		codes = append(codes, code)
	}
	sort.Strings(codes) // Sort the codes alphabetically
	return codes
}
