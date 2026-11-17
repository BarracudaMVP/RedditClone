package middleware

import (
	"RedditClone/internal/handlers"
	"RedditClone/internal/user"
	"net/http"
	"strings"
)

var (
	publicURLs = map[string]struct{}{
		"/":                         {},
		"/static/":                  {},
		"GET /favicon.ico":          {},
		"GET /manifest.json":        {},
		"POST /api/register":        {},
		"POST /api/login":           {},
		"GET /api/posts":            {},
		"GET /api/posts/{$}":        {},
		"GET /api/posts/{category}": {},
		"GET /api/post/{postId}":    {},
		"GET /api/user/{username}":  {},
	}
)

func Auth(repo user.UserRepository, secret []byte, next *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pattern := next.Handler(r)
		if _, ok := publicURLs[pattern]; ok {
			next.ServeHTTP(w, r)
			return
		}
		credits := strings.Fields(r.Header.Get("Authorization"))
		if len(credits) != 2 || strings.ToLower(credits[0]) != "bearer" {
			handlers.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		tokenUser, err := user.ParseToken(credits[1], secret)
		if err != nil {
			handlers.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		u, err := repo.GetByUsername(tokenUser.Username)
		if err != nil {
			handlers.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if u.ID != tokenUser.ID {
			handlers.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		ctx := user.ContextWithUser(r.Context(), &u.User)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
