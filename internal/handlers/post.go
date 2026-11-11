package handlers

import (
	"RedditClone/internal/id"
	"RedditClone/internal/post"
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
	writeJSON(w, http.StatusOK, p)
}

func (h *PostHandler) ListByCategory(w http.ResponseWriter, r *http.Request) {
	category, err := post.ParseCategory(r.PathValue("category"))
	if errors.Is(err, post.ErrInvalidCategory) {
		writeJSONError(w, http.StatusBadRequest, "invalid category")
		return
	}
	p := h.Repository.GetCategory(category)
	writeJSON(w, http.StatusOK, p)
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	postID, err := id.ParseID(r.PathValue("postID"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid post id")
		return
	}
	p, err := h.Repository.GetByID(postID)
	if errors.Is(err, post.ErrPostNotFound) {
		writeJSONError(w, http.StatusNotFound, "post not found")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, p)
}
