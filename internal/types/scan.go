package types

import "github.com/danielgtaylor/huma/v2" // Ensure huma is imported for FormFile

// Define constants for supported OCR languages.
const (
	LangEnglish    = "eng"
	LangNepali     = "nep"
	LangHindi      = "hin"
	LangDevanagari = "dev" // Script-specific, not a language
)

// SupportedOCRLanguages is a map for quick lookup of valid language/script codes.
var SupportedOCRLanguages = map[string]bool{
	LangEnglish:    true,
	LangNepali:     true,
	LangHindi:      true,
	LangDevanagari: true,
}

// ScanInput is the input structure for the /scan endpoint using multipart/form-data.
// It expects an image file and an optional language hint.
type ScanInput struct {
	RawBody huma.MultipartFormFiles[struct {
		Image huma.FormFile `form:"image" contentType:"image/*" required:"true" doc:"Image file for OCR scanning (e.g., JPEG, PNG)"`
		// Language field now uses `enum` tag for documentation and hints at allowed values.
		// Runtime validation will be added in the handler.
		Language string `form:"lang" huma:"example:eng,default:eng,enum:eng,nep,hin,dev" doc:"Tesseract language code (e.g., 'eng', 'nep', 'hin', 'dev' for Devanagari script). Use '+' to combine (e.g., 'eng+hin'). Default is 'eng'."`
	}]
}

// ScanOutput is the output structure for the /scan endpoint.
// It returns the extracted text.
type ScanOutput struct {
	Body struct {
		Text string `json:"text" huma:"example:Extracted text from the image"`
	}
}
