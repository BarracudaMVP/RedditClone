package main

import (
	"RedditClone/internal/config"
	"RedditClone/internal/handlers"
	"RedditClone/internal/middleware"
	"RedditClone/internal/post"
	"RedditClone/internal/user"
	"log"
	"net/http"
)

func newMux(cfg *config.Config) http.Handler {
	staticDir := cfg.StaticDir

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
	userHandler := handlers.NewUserHandler(userRepo, cfg)

	mux.HandleFunc("GET /api/posts", postHandler.List)
	mux.HandleFunc("GET /api/posts/{$}", postHandler.List)
	mux.HandleFunc("GET /api/posts/{category}", postHandler.ListByCategory)
	mux.HandleFunc("GET /api/post/{postID}", postHandler.GetPost)
	mux.HandleFunc("POST /api/register", userHandler.Register)
	mux.HandleFunc("POST /api/login", userHandler.Login)
	return mux
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	mux := newMux(cfg)
	mux = middleware.Logging(mux)
	server := http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: mux,
	}
	log.Printf("starting server at %s", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("http failed: %v", err)
	}
}
