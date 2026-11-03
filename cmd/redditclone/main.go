package main

import (
	"log"
	"net/http"
)

var (
	addr      = ":8080"
	staticDir = "./static"
)

func newMux(staticDir string) *http.ServeMux {
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

	//mux.HandleFunc("GET /posts", ))
	//mux.HandleFunc("GET /posts/{$}", _)
	return mux
}

func main() {
	mux := newMux(staticDir)
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
