package httphelpers

import (
	"encoding/json"
	"log"
	"net/http"
)

// ErrorResponse represents a standard error response structure
type ErrorResponse struct {
	Error string `json:"error"`
}

// RespondWithError sends an HTTP error response with the specified status code and error message
func RespondWithError(w http.ResponseWriter, statusCode int, errorMsg string) {
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error: errorMsg,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// If we can't even encode the error response, log it and send a plain text response
		log.Printf("Failed to encode error response: %v", err)
		http.Error(w, errorMsg, statusCode)
	}
}

// RespondWithJSON sends a successful HTTP response with the provided data
func RespondWithJSON(w http.ResponseWriter, statusCode int, data any) error {
	w.WriteHeader(statusCode)

	if data != nil {
		return json.NewEncoder(w).Encode(data)
	}

	return nil
}
