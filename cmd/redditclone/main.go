package main

import (
	"RedditClone/internal/handlers"
	"RedditClone/internal/middleware"
	"RedditClone/internal/post"
	"RedditClone/internal/user"
	"log"
	"net/http"
	"os"
)

var (
	addr      = ":8080"
	staticDir = "./static"
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
)

func newMux(staticDir string) http.Handler {

	postRepo := post.NewMemoryPostRepository()
	userRepo := user.NewMemoryUserRepository()

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir(staticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, staticDir+"/html/favicon.ico")
	})
	mux.HandleFunc("GET /manifest.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, staticDir+"/html/manifest.json")
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		http.ServeFile(w, r, staticDir+"/html/index.html")
	})
	mux.Handle("/api", http.NotFoundHandler())
	mux.Handle("/api/", http.NotFoundHandler())

	postHandler := handlers.NewPostHandler(postRepo)
	userHandler := handlers.NewUserHandler(userRepo, jwtSecret)

	mux.HandleFunc("GET /api/posts", postHandler.List)
	mux.HandleFunc("GET /api/posts/", postHandler.List)
	mux.HandleFunc("GET /api/post/{postID}", postHandler.GetPost)
	mux.HandleFunc("GET /api/posts/{category}", postHandler.ListByCategory)
	mux.HandleFunc("POST /api/register", userHandler.Register)
	mux.HandleFunc("POST /api/login", userHandler.Login)
	return mux
}

func main() {
	mux := newMux(staticDir)
	mux = middleware.Logging(mux)
	server := http.Server{
		Addr:    addr,
		Handler: mux,
	}
	log.Printf("starting server at %s", addr)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("http failed: %v", err)
	}
}
