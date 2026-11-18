package handlers

import (
	"RedditClone/internal/id"
	"RedditClone/internal/post"
	"RedditClone/internal/user"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
)

type PostHandler struct {
	Repository post.PostRepository
}

func NewPostHandler(repo post.PostRepository) *PostHandler {
	return &PostHandler{
		Repository: repo,
	}
}
func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) {
	p := h.Repository.GetAll()
	WriteJSON(w, http.StatusOK, p)
}

func (h *PostHandler) ListByCategory(w http.ResponseWriter, r *http.Request) {
	category, err := post.ParseCategory(r.PathValue("category"))
	if errors.Is(err, post.ErrInvalidCategoryLength) {
		WriteJSONError(w, http.StatusBadRequest, "invalid category")
		return
	}
	p := h.Repository.GetCategory(category)
	WriteJSON(w, http.StatusOK, p)
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	postID, err := id.ParseID(r.PathValue("postId"))
	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, "invalid post id")
		return
	}
	p, err := h.Repository.GetByID(postID)
	if errors.Is(err, post.ErrPostNotFound) {
		WriteJSONError(w, http.StatusNotFound, "post not found")
		return
	}
	if err != nil {
		WriteJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	WriteJSON(w, http.StatusOK, p)
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	u, err := user.UserFromContext(r.Context())
	if err != nil {
		WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	validateErrors := ValidationError{Errors: make([]ValidationErrorItem, 0)}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		WriteJSONError(w, http.StatusBadRequest, "unknown payload")
		return
	}

	var form post.CreatePostRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := decoder.Decode(&form); err != nil {
		WriteJSONError(w, http.StatusBadRequest, "cannot decode payload")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		WriteJSONError(w, http.StatusBadRequest, "cannot decode payload")
		return
	}
}
