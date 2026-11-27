package main

import (
	"RedditClone/internal/handlers"
	"RedditClone/internal/middleware"
	"RedditClone/internal/post"
	"RedditClone/internal/user"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestCreatePostWithValidToken(t *testing.T) {
	secret := []byte("super-secret-jwt-for-testing-super-secret")
	postRepo := post.NewMemoryPostRepository()
	userRepo := user.NewMemoryUserRepository()
	postHandler := handlers.NewPostHandler(postRepo, userRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/posts", postHandler.CreatePost)
	handler := middleware.Auth(userRepo, secret, mux)

	u, err := userRepo.Create("alice1", "$2a$10$HJR14d5BhJZqA2OE3cYfY.lpyu2MSwAjKHXMHuIORZNG63zEzZ01u")
	if err != nil {
		t.Fatalf("error user create: %v", err)
	}
	tokenString, err := user.CreateToken(u, secret, time.Hour)
	if err != nil {
		t.Fatalf("error CreateToken: %v", err)
	}
	w := httptest.NewRecorder()
	body := strings.NewReader(`{"category":"music","type":"text","text":"adasd","title":"asdasd"}`)
	r := httptest.NewRequest("POST", "/api/posts", body)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+tokenString)
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusCreated, w.Body.String())
	}
	var p post.Post
	if err := json.NewDecoder(w.Body).Decode(&p); err != nil {
		t.Fatalf("decode created post: %v", err)
	}
	if p.Author.ID != u.ID {
		t.Errorf("author ID= %q, want %q", p.Author.ID, u.ID)
	}
}
func TestCreatePostWithExpiredToken(t *testing.T) {
	secret := []byte("super-secret-jwt-for-testing-super-secret")
	postRepo := post.NewMemoryPostRepository()
	userRepo := user.NewMemoryUserRepository()
	postHandler := handlers.NewPostHandler(postRepo, userRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/posts", postHandler.CreatePost)
	handler := middleware.Auth(userRepo, secret, mux)

	u, err := userRepo.Create("alice1", "$2a$10$HJR14d5BhJZqA2OE3cYfY.lpyu2MSwAjKHXMHuIORZNG63zEzZ01u")
	if err != nil {
		t.Fatalf("error user create: %v", err)
	}
	now := time.Now()
	claims := &user.JwtClaims{
		User: u,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
		},
	}
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		t.Fatalf("error CreateToken: %v", err)
	}
	w := httptest.NewRecorder()
	body := strings.NewReader(`{"category":"music","type":"text","text":"adasd","title":"asdasd"}`)
	r := httptest.NewRequest("POST", "/api/posts", body)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+tokenString)
	before := len(postRepo.GetAllPosts())
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusUnauthorized, w.Body.String())
	}
	if after := len(postRepo.GetAllPosts()); after != before {
		t.Errorf("post count after rejected request = %d, want = %d", after, before)
	}
}

func TestCreatePostWithInvalidToken(t *testing.T) {
	secret := []byte("super-secret-jwt-for-testing-super-secret")
	wrongSecret := []byte("anti-super-secret-jwt-for-testing-super-secret")
	postRepo := post.NewMemoryPostRepository()
	userRepo := user.NewMemoryUserRepository()
	postHandler := handlers.NewPostHandler(postRepo, userRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/posts", postHandler.CreatePost)
	handler := middleware.Auth(userRepo, secret, mux)

	u, err := userRepo.Create("alice1", "$2a$10$HJR14d5BhJZqA2OE3cYfY.lpyu2MSwAjKHXMHuIORZNG63zEzZ01u")
	if err != nil {
		t.Fatalf("error user create: %v", err)
	}
	tokenString, err := user.CreateToken(u, wrongSecret, time.Hour)
	if err != nil {
		t.Fatalf("error CreateToken: %v", err)
	}
	w := httptest.NewRecorder()
	body := strings.NewReader(`{"category":"music","type":"text","text":"adasd","title":"asdasd"}`)
	r := httptest.NewRequest("POST", "/api/posts", body)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+tokenString)
	before := len(postRepo.GetAllPosts())
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusUnauthorized, w.Body.String())
	}
	if after := len(postRepo.GetAllPosts()); after != before {
		t.Errorf("post count after rejected request = %d, want = %d", after, before)
	}
}
func TestCreateLinkPostWithValidToken(t *testing.T) {
	link := "https://spotify.com"
	secret := []byte("super-secret-jwt-for-testing-super-secret")
	postRepo := post.NewMemoryPostRepository()
	userRepo := user.NewMemoryUserRepository()
	postHandler := handlers.NewPostHandler(postRepo, userRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/posts", postHandler.CreatePost)
	handler := middleware.Auth(userRepo, secret, mux)

	u, err := userRepo.Create("alice1", "$2a$10$HJR14d5BhJZqA2OE3cYfY.lpyu2MSwAjKHXMHuIORZNG63zEzZ01u")
	if err != nil {
		t.Fatalf("error user create: %v", err)
	}
	tokenString, err := user.CreateToken(u, secret, time.Hour)
	if err != nil {
		t.Fatalf("error CreateToken: %v", err)
	}
	w := httptest.NewRecorder()
	body := strings.NewReader(`{"category":"music","type":"link","url":"` + link + `","title":"asdasd"}`)
	r := httptest.NewRequest("POST", "/api/posts", body)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+tokenString)
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusCreated, w.Body.String())
	}
	var p post.Post
	if err := json.NewDecoder(w.Body).Decode(&p); err != nil {
		t.Fatalf("decode created post: %v", err)
	}
	if p.Type != post.LinkPostType {
		t.Errorf("type = %q, want %q", p.Type, post.LinkPostType)
	}
	if p.URL != link {
		t.Errorf("URL = %q, want %q", p.URL, link)
	}
	if p.Author.ID != u.ID {
		t.Errorf("author ID= %q, want %q", p.Author.ID, u.ID)
	}
}

func TestDeletePostErrors(t *testing.T) {
	secret := []byte("super-secret-jwt-for-testing-super-secret")
	postRepo := post.NewMemoryPostRepository()
	userRepo := user.NewMemoryUserRepository()
	postHandler := handlers.NewPostHandler(postRepo, userRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/post/{postId}", postHandler.DeletePost)
	handler := middleware.Auth(userRepo, secret, mux)
	bob, err := userRepo.Create("bob", "$2a$10$HJR14d5BhJZqA2OE3cYfY.lpyu2MSwAjKHXMHuIORZNG63zEzZ01u")
	if err != nil {
		t.Fatalf("error user create: %v", err)
	}
	tokenString, err := user.CreateToken(bob, secret, time.Hour)
	if err != nil {
		t.Fatalf("error CreateToken: %v", err)
	}
	alice, err := userRepo.GetByUsername("alice")
	if err != nil {
		t.Fatalf("error GetByUsername: %v", err)
	}
	posts := postRepo.GetPostByAuthorID(alice.ID)
	if len(posts) == 0 {
		t.Fatalf("error not enough posts: %v", err)
	}
	alicePostId := posts[0].ID
	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/api/post/"+string(alicePostId), nil)
	r.Header.Set("Authorization", "Bearer "+tokenString)
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusUnauthorized, w.Body.String())
	}
	want := `{"message":"unauthorized"}`
	if got := strings.TrimSpace(w.Body.String()); got != want {
		t.Fatalf("body = %q, want = %q", got, want)
	}
	if !postRepo.PostExistsByID(alicePostId) {
		t.Fatalf("post exists, but was deleted")
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest("DELETE", "/api/post/kek23423d", nil)
	r.Header.Set("Authorization", "Bearer "+tokenString)
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusNotFound, w.Body.String())
	}
}
func TestCreatePostWithInvalidJSON(t *testing.T) {
	secret := []byte("super-secret-jwt-for-testing-super-secret")
	postRepo := post.NewMemoryPostRepository()
	userRepo := user.NewMemoryUserRepository()
	postHandler := handlers.NewPostHandler(postRepo, userRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/posts", postHandler.CreatePost)
	handler := middleware.Auth(userRepo, secret, mux)
	alice, err := userRepo.GetByUsername("alice")
	if err != nil {
		t.Fatalf("error GetByUsername: %v", err)
	}
	tokenString, err := user.CreateToken(alice.User, secret, time.Hour)
	if err != nil {
		t.Fatalf("error CreateToken: %v", err)
	}
	for _, body := range []string{"{", "{} {}", "null"} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/posts", strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+tokenString)
		r.Header.Set("Content-type", "application/json")
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("JSON %q: status = %d, want %d, body %s", body, w.Code, http.StatusBadRequest, w.Body.String())
		}
	}
}
func TestRepositoriesConcurrent(t *testing.T) {
	postRepo := post.NewMemoryPostRepository()
	userRepo := user.NewMemoryUserRepository()

	p := postRepo.GetAllPosts()[0]

	var wg sync.WaitGroup

	wg.Go(func() {
		for range 100 {
			if _, err := json.Marshal(postRepo.GetAllPosts()); err != nil {
				t.Error(err)
			}
			if _, err := userRepo.GetByUsername("alice"); err != nil {
				t.Error(err)
			}
		}
	})
	wg.Go(func() {
		for i := range 100 {
			vote := post.Upvote
			if i%2 == 0 {
				vote = post.Downvote
			}
			if _, err := postRepo.VotePost(p.ID, p.Author, vote); err != nil {
				t.Error(err)
			}
			name := user.Username("user" + strconv.Itoa(i))
			if _, err := userRepo.Create(name, ""); err != nil {
				t.Error(err)
			}
		}
	})
	wg.Wait()
}
