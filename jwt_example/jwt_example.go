package main

import (
	"encoding/json"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

/*

curl -X POST -H "Content-Type: application/json" -d '{"login": "rvasily", "password": "love"}' http://localhost:8080/login

# Copy data.token from the login response into the following request:
curl -H "Authorization: Bearer <token-from-login-response>" http://localhost:8080/profile

*/

type User struct {
	ID       int
	FullName string
	Role     string
}

type Claims struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	Role  string `json:"role"`
	ID    int    `json:"id"`
	jwt.RegisteredClaims
}

type LoginForm struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

var (
	users = map[string]User{
		"rvasily":        {450, "Vasily Romanov", "user"}, // a tribute to the original course author
		"romanov.vasily": {42, "Василий Романов", "admin"},
	}

	ExamplePassword    = "love"
	ExampleTokenSecret = []byte("супер секретный ключ")
)

func profilePage(w http.ResponseWriter, r *http.Request) {
	parts := strings.Fields(r.Header.Get("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		jsonError(w, http.StatusUnauthorized, "bad token")
		return
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (any, error) {
		return ExampleTokenSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || !token.Valid {
		jsonError(w, http.StatusUnauthorized, "bad token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status": http.StatusOK,
		"data": map[string]any{
			"login": claims.Login,
			"name":  claims.Name,
			"role":  claims.Role,
			"id":    claims.ID,
		},
	})
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		jsonError(w, http.StatusBadRequest, "unknown payload")
		return
	}

	var form *LoginForm
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := decoder.Decode(&form); err != nil || form == nil {
		jsonError(w, http.StatusBadRequest, "cannot decode payload")
		return
	}
	// Reject trailing data and multiple JSON objects.
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		jsonError(w, http.StatusBadRequest, "cannot decode payload")
		return
	}

	user, exist := users[form.Login]
	if !exist || form.Password != ExamplePassword {
		jsonError(w, http.StatusUnauthorized, "bad login or password")
		return
	}

	// Create token with custom claims
	now := time.Now()
	claims := &Claims{
		Login: form.Login,
		Name:  user.FullName,
		Role:  user.Role,
		ID:    user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(ExampleTokenSecret)
	if err != nil {
		log.Printf("sign token: %v", err)
		jsonError(w, http.StatusInternalServerError, "cannot sign token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status": http.StatusOK,
		"data": map[string]any{
			"token": tokenString,
		},
	})
}

func main() {
	log.Println("starting server at http://localhost:8080")
	if err := http.ListenAndServe(":8080", newMux()); err != nil {
		log.Fatal(err)
	}
}

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /login", loginPage)
	mux.HandleFunc("GET /profile", profilePage)
	return mux
}

func jsonError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{
		"status": status,
		"error":  msg,
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode response: %v", err)
	}
}
