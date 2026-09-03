package httpx

import (
	"encoding/json"
	"net/http"
)

type  Code string

const (
	CodeInvalidID Code = "invalid_id"
	CodeInternalError Code = "internal_error"
)

type errorPlayload struct {
	Code Code `json:"code"`
	Message string `json:"message"`
	Field string `json:"field,omitempty"`
}

type errorEnvelope struct {
	Error errorPlayload `json:"error"`
}

func Error(w http.ResponseWriter, status int, message string, code Code) {
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(errorEnvelope{Error: errorPlayload{
		Code: code,
		Message: message,
	}})
}

func ValidationError(w http.ResponseWriter, status int, message string, code Code,field string) {
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(errorEnvelope{Error: errorPlayload{
		Code: code,
		Message: message,
		Field: field,
	}})
}