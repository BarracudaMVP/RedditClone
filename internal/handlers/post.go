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
	postRepo post.PostRepository
	userRepo user.UserRepository
}

func NewPostHandler(postRepo post.PostRepository, userRepo user.UserRepository) *PostHandler {
	return &PostHandler{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}
func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) {
	p := h.postRepo.GetAllPosts()
	WriteJSON(w, http.StatusOK, p)
}

func (h *PostHandler) ListByCategory(w http.ResponseWriter, r *http.Request) {
	category, err := post.ParseCategory(r.PathValue("category"))
	if errors.Is(err, post.ErrInvalidCategoryLength) {
		WriteJSONError(w, http.StatusBadRequest, "invalid category")
		return
	}
	p := h.postRepo.GetPostsByCategory(category)
	WriteJSON(w, http.StatusOK, p)
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	postID, err := id.ParseID(r.PathValue("postId"))
	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, "invalid post id")
		return
	}
	p, err := h.postRepo.GetPostByIDAndIncrementViews(postID)
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
		WriteJSONError(w, http.StatusBadRequest, "bad request")
		return
	}

	var form post.CreatePostRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := decoder.Decode(&form); err != nil {
		WriteJSONError(w, http.StatusBadRequest, "bad request")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		WriteJSONError(w, http.StatusBadRequest, "bad request")
		return
	}
	if err := post.ValidateTitle(form.Title); err != nil {
		validateErrors.Add(BodyErrorLocation, "title", err.Error())
	}
	if err := post.ValidateCategory(string(form.Category)); err != nil {
		validateErrors.Add(BodyErrorLocation, "category", err.Error())
	}
	switch form.Type {
	case post.LinkPostType:
		if err := post.ValidatePostURL(form.URL); err != nil {
			validateErrors.Add(BodyErrorLocation, "url", err.Error())
		}
	case post.TextPostType:
		if err := post.ValidatePostText(form.Text); err != nil {
			validateErrors.Add(BodyErrorLocation, "text", err.Error())
		}
	default:
		validateErrors.Add(BodyErrorLocation, "type", post.ErrInvalidPostType.Error())

	}
	if len(validateErrors.Errors) > 0 {
		WriteJSONValidationError(w, validateErrors)
		return
	}
	post, err := h.postRepo.CreatePost(form, *u)
	if err != nil {
		WriteJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	WriteJSON(w, http.StatusCreated, post)

}

func (h *PostHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	username, err := user.ParseUsername(r.PathValue("username"))
	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, "bad request")
		return
	}

	u, err := h.userRepo.GetByUsername(username)
	if err != nil {
		WriteJSONError(w, http.StatusNotFound, user.ErrUserNotFound.Error())
		return
	}
	p := h.postRepo.GetPostByAuthorID(u.ID)
	WriteJSON(w, http.StatusOK, p)
}
func (h *PostHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	u, err := user.UserFromContext(r.Context())
	if err != nil {
		WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	pid, err := id.ParseID(r.PathValue("postId"))
	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, "bad request")
		return
	}
	if ok := h.postRepo.PostExistsByID(pid); !ok {
		WriteJSONError(w, http.StatusNotFound, post.ErrPostNotFound.Error())
		return
	}
	validateErrors := ValidationError{Errors: make([]ValidationErrorItem, 0)}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		WriteJSONError(w, http.StatusBadRequest, "bad request")
		return
	}

	var form post.CreateCommentRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := decoder.Decode(&form); err != nil {
		WriteJSONError(w, http.StatusBadRequest, "bad request")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		WriteJSONError(w, http.StatusBadRequest, "bad request")
		return
	}
	if err := post.ValidateComment(form.Comment); err != nil {
		validateErrors.Add(BodyErrorLocation, "comment", err.Error())
	}
	if len(validateErrors.Errors) > 0 {
		WriteJSONValidationError(w, validateErrors)
		return
	}
	p, err := h.postRepo.CreateComment(form, pid, *u)
	if err != nil {
		if errors.Is(err, post.ErrPostNotFound) {
			WriteJSONError(w, http.StatusNotFound, post.ErrPostNotFound.Error())
			return
		}
		WriteJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	WriteJSON(w, http.StatusCreated, p)

}
