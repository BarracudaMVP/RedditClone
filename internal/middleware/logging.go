package middleware

import (
	"log"
	"net/http"
	"time"
)

type ResponseWriteWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *ResponseWriteWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapped := &ResponseWriteWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		start := time.Now()
		next.ServeHTTP(wrapped, r)
		log.Printf("%s %s %d %s", r.Method, r.RequestURI, wrapped.statusCode, time.Since(start))
	})
}
