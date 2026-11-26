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

func newMux(cfg *config.Config, postRepo post.PostRepository, userRepo user.UserRepository) *http.ServeMux {
	staticDir := cfg.StaticDir

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

	postHandler := handlers.NewPostHandler(postRepo, userRepo)
	userHandler := handlers.NewUserHandler(userRepo, cfg)

	mux.HandleFunc("GET /api/posts", postHandler.List)
	mux.HandleFunc("GET /api/posts/{$}", postHandler.List)
	mux.HandleFunc("GET /api/posts/{category}", postHandler.ListByCategory)
	mux.HandleFunc("GET /api/post/{postId}", postHandler.GetPost)

	mux.HandleFunc("DELETE /api/post/{postId}", postHandler.DeletePost)

	mux.HandleFunc("GET /api/user/{username}", postHandler.ListByUser)

	mux.HandleFunc("POST /api/register", userHandler.Register)
	mux.HandleFunc("POST /api/login", userHandler.Login)

	mux.HandleFunc("POST /api/posts", postHandler.CreatePost)
	mux.HandleFunc("POST /api/posts/{$}", postHandler.CreatePost)

	mux.HandleFunc("POST /api/post/{postId}", postHandler.CreateComment)
	mux.HandleFunc("DELETE /api/post/{postId}/{commentId}", postHandler.DeleteComment)

	//mux.HandleFunc("GET /api/post/{postId}/upvote", postHandler.UpvotePost)
	//mux.HandleFunc("GET /api/post/{postId}/downvote", postHandler.DownvotePost)
	//mux.HandleFunc("GET /api/post/{postId}/unvote", postHandler.UnvotePost)

	return mux
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	postRepo := post.NewMemoryPostRepository()
	userRepo := user.NewMemoryUserRepository()

	mux := newMux(cfg, postRepo, userRepo)
	handler := middleware.Auth(userRepo, cfg.JWT.Secret, mux)
	handler = middleware.Logging(handler)
	server := http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: handler,
	}
	log.Printf("starting server at %s", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("http failed: %v", err)
	}
}
