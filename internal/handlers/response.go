package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorLocation string

const (
	BodyErrorLocation = "body"
)

type Error struct {
	Message string `json:"message"`
}

type ValidationError struct {
	Errors []ValidationErrorItem `json:"errors"`
}

func (v *ValidationError) Add(location ErrorLocation, param string, msg string) {
	err := ValidationErrorItem{
		Location: location,
		Param:    param,
		Msg:      msg,
	}
	v.Errors = append(v.Errors, err)
}

type ValidationErrorItem struct {
	Location ErrorLocation `json:"location,omitempty"`
	Param    string        `json:"param"`
	Msg      string        `json:"msg"`
}

func writeJSONError(w http.ResponseWriter, status int, value string) {
	writeJSON(w, status, Error{
		Message: value,
	})
}
func writeJSONValidationError(w http.ResponseWriter, errors ValidationError) {
	writeJSON(w, http.StatusUnprocessableEntity, errors)
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode response: %v", err)
	}
}
