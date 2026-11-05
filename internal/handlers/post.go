package handlers

import (
	"RedditClone/internal/post"
	"encoding/json"
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
	posts := h.Repository.GetAll()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}
