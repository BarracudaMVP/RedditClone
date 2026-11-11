package handlers

import (
	"RedditClone/internal/user"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
)

type UserHandler struct {
	Repository user.UserRepository
	JWTSecret  []byte
	DummyHash  string
}

func NewUserHandler(repo user.UserRepository, secret []byte, dummyHash string) *UserHandler {
	return &UserHandler{
		Repository: repo,
		JWTSecret:  secret,
		DummyHash:  dummyHash,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	validateErrors := ValidationError{Errors: make([]ValidationErrorItem, 0)}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeJSONError(w, http.StatusBadRequest, "unknown payload")
		return
	}

	var form user.Credentials
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := decoder.Decode(&form); err != nil {
		writeJSONError(w, http.StatusBadRequest, "cannot decode payload")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeJSONError(w, http.StatusBadRequest, "cannot decode payload")
		return
	}
	if err := user.ValidateUsername(form.Username); err != nil {
		validateErrors.Add(BodyErrorLocation, "username", err.Error())
	}
	if err := user.ValidatePassword(form.Password); err != nil {
		validateErrors.Add(BodyErrorLocation, "password", err.Error())
	}
	if len(validateErrors.Errors) > 0 {
		writeJSONValidationError(w, validateErrors)
		return
	}
	passwordHash, err := user.HashPassword(form.Password)
	if err != nil {
		log.Printf("hash password: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	u, err := h.Repository.Create(form.Username, passwordHash)
	if err != nil {
		if errors.Is(err, user.ErrUserExists) {
			validateErrors.Add(BodyErrorLocation, "username", err.Error())
			writeJSONValidationError(w, validateErrors)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	tokenString, err := user.CreateToken(u, h.JWTSecret)
	if err != nil {
		log.Printf("sign token: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "cannot sign token")
		return
	}

	writeJSON(w, http.StatusCreated, user.AuthenticationResponse{
		Token: tokenString,
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	validateErrors := ValidationError{Errors: make([]ValidationErrorItem, 0)}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeJSONError(w, http.StatusBadRequest, "unknown payload")
		return
	}

	var form user.Credentials
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := decoder.Decode(&form); err != nil {
		writeJSONError(w, http.StatusBadRequest, "cannot decode payload")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeJSONError(w, http.StatusBadRequest, "cannot decode payload")
		return
	}
	if err := user.ValidateUsername(form.Username); err != nil {
		validateErrors.Add(BodyErrorLocation, "username", err.Error())
	}
	if err := user.ValidatePassword(form.Password); err != nil {
		validateErrors.Add(BodyErrorLocation, "password", err.Error())
	}
	if len(validateErrors.Errors) > 0 {
		writeJSONValidationError(w, validateErrors)
		return
	}
	storedUser, err := h.Repository.GetByUsername(form.Username)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			user.CheckPasswordHash(h.DummyHash, form.Password)
			writeJSONError(w, http.StatusUnauthorized, "invalid username or password")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !user.CheckPasswordHash(storedUser.PasswordHash, form.Password) {
		writeJSONError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}
	tokenString, err := user.CreateToken(storedUser.User, h.JWTSecret)
	if err != nil {
		log.Printf("sign token: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "cannot sign token")
		return
	}

	writeJSON(w, http.StatusOK, user.AuthenticationResponse{
		Token: tokenString,
	})
}
