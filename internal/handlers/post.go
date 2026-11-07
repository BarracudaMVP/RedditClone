package handlers

import (
	"RedditClone/internal/id"
	"RedditClone/internal/post"
	"encoding/json"
	"errors"
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(p)
}

func (h *PostHandler) ListByCategory(w http.ResponseWriter, r *http.Request) {
	category, err := post.ParseCategory(r.PathValue("category"))
	if errors.Is(err, post.ErrInvalidCategory) {
		writeJSONError(w, "invalid category", http.StatusBadRequest)
		return
	}
	p := h.Repository.GetCategory(category)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(p)
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	postID, err := id.ParseID(r.PathValue("postID"))
	if err != nil {
		writeJSONError(w, "invalid post id", http.StatusBadRequest)
		return
	}
	p, err := h.Repository.GetByID(postID)
	if errors.Is(err, post.ErrPostNotFound) {
		writeJSONError(w, "post not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(p)
}

func writeJSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{
		"message": message,
	})
}
