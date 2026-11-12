package handlers

import (
	"RedditClone/internal/config"
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
	Cfg        *config.Config
}

func NewUserHandler(repo user.UserRepository, cfg *config.Config) *UserHandler {
	return &UserHandler{
		Repository: repo,
		Cfg:        cfg,
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
	passwordHash, err := user.HashPassword(form.Password, h.Cfg.BcryptConfig.Cost)
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
	tokenString, err := user.CreateToken(u, h.Cfg.JWT.Secret, h.Cfg.JWT.TTL)
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
			user.CheckPasswordHash(h.Cfg.BcryptConfig.DummyHash, form.Password)
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
	tokenString, err := user.CreateToken(storedUser.User, h.Cfg.JWT.Secret, h.Cfg.JWT.TTL)
	if err != nil {
		log.Printf("sign token: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "cannot sign token")
		return
	}

	writeJSON(w, http.StatusOK, user.AuthenticationResponse{
		Token: tokenString,
	})
}
