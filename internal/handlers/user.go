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
		WriteJSONError(w, http.StatusBadRequest, "bad request")
		return
	}

	var form user.Credentials
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := decoder.Decode(&form); err != nil {
		WriteJSONError(w, http.StatusBadRequest, "bad request")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		WriteJSONError(w, http.StatusBadRequest, "bad request")
		return
	}
	if err := user.ValidateUsername(form.Username); err != nil {
		validateErrors.Add(BodyErrorLocation, "username", err.Error())
	}
	if err := user.ValidatePassword(form.Password); err != nil {
		validateErrors.Add(BodyErrorLocation, "password", err.Error())
	}
	if len(validateErrors.Errors) > 0 {
		WriteJSONValidationError(w, validateErrors)
		return
	}
	passwordHash, err := user.HashPassword(form.Password, h.Cfg.BcryptConfig.Cost)
	if err != nil {
		log.Printf("hash password: %v", err)
		WriteJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	u, err := h.Repository.Create(form.Username, passwordHash)
	if err != nil {
		if errors.Is(err, user.ErrUserExists) {
			validateErrors.Add(BodyErrorLocation, "username", err.Error())
			WriteJSONValidationError(w, validateErrors)
			return
		}
		WriteJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	tokenString, err := user.CreateToken(u, h.Cfg.JWT.Secret, h.Cfg.JWT.TTL)
	if err != nil {
		log.Printf("sign token: %v", err)
		WriteJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	WriteJSON(w, http.StatusCreated, user.AuthenticationResponse{
		Token: tokenString,
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	validateErrors := ValidationError{Errors: make([]ValidationErrorItem, 0)}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		WriteJSONError(w, http.StatusBadRequest, "bad request")
		return
	}

	var form user.Credentials
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := decoder.Decode(&form); err != nil {
		WriteJSONError(w, http.StatusBadRequest, "bad request")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		WriteJSONError(w, http.StatusBadRequest, "bad request")
		return
	}
	if err := user.ValidateUsername(form.Username); err != nil {
		validateErrors.Add(BodyErrorLocation, "username", err.Error())
	}
	if err := user.ValidatePassword(form.Password); err != nil {
		validateErrors.Add(BodyErrorLocation, "password", err.Error())
	}
	if len(validateErrors.Errors) > 0 {
		WriteJSONValidationError(w, validateErrors)
		return
	}
	storedUser, err := h.Repository.GetByUsername(form.Username)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			user.CheckPasswordHash(h.Cfg.BcryptConfig.DummyHash, form.Password)
			WriteJSONError(w, http.StatusUnauthorized, user.ErrUserNotFound.Error())
			return
		}
		WriteJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !user.CheckPasswordHash(storedUser.PasswordHash, form.Password) {
		WriteJSONError(w, http.StatusUnauthorized, user.ErrUserNotFound.Error())
		return
	}
	tokenString, err := user.CreateToken(storedUser.User, h.Cfg.JWT.Secret, h.Cfg.JWT.TTL)
	if err != nil {
		log.Printf("sign token: %v", err)
		WriteJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	WriteJSON(w, http.StatusOK, user.AuthenticationResponse{
		Token: tokenString,
	})
}
