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

func WriteJSONError(w http.ResponseWriter, status int, value string) {
	WriteJSON(w, status, Error{
		Message: value,
	})
}
func WriteJSONValidationError(w http.ResponseWriter, errors ValidationError) {
	WriteJSON(w, http.StatusUnprocessableEntity, errors)
}
func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode response: %v", err)
	}
}
