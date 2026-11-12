package user

import (
	"RedditClone/internal/id"
	"errors"
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
var ErrInvalidUsernameLength = errors.New("incorrect username length")
var ErrInvalidUsernameFormat = errors.New("incorrect symbols in username")
var ErrInvalidPasswordLength = errors.New("incorrect password length")

const (
	MinLengthUsername = 1
	MaxLengthUsername = 32
	MinLengthPassword = 8
	MaxLengthPassword = 72
)

type User struct {
	ID       id.ID    `json:"id"`
	Username Username `json:"username"`
	Admin    bool     `json:"admin,omitempty"`
}

type StoredUser struct {
	User
	PasswordHash string `json:"-"`
}
type Username string

func ValidateUsername(username Username) error {
	if len(username) < MinLengthUsername || len(username) > MaxLengthUsername {
		return ErrInvalidUsernameLength
	}
	if !usernamePattern.MatchString(string(username)) {
		return ErrInvalidUsernameFormat
	}
	return nil
}

func ValidatePassword(password string) error {
	if utf8.RuneCountInString(password) < MinLengthPassword || len(password) > MaxLengthPassword {
		return ErrInvalidPasswordLength
	}
	return nil
}

func HashPassword(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

func CheckPasswordHash(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CreateToken(user User, secret []byte, tokenDuration time.Duration) (string, error) {
	now := time.Now()
	claims := &JwtClaims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenDuration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

type Credentials struct {
	Username Username `json:"username"`
	Password string   `json:"password"`
}
type JwtClaims struct {
	User User `json:"user"`
	jwt.RegisteredClaims
}
type AuthenticationResponse struct {
	Token string `json:"token"`
}
