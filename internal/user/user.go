package user

import (
	"RedditClone/internal/id"
	"context"
	"errors"
	"fmt"
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
type userKey struct{}
type Username string

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
func ParseToken(tokenString string, secret []byte) (User, error) {
	claims := &JwtClaims{}
	getSecret := func(*jwt.Token) (any, error) {
		return secret, nil
	}
	_, err := jwt.ParseWithClaims(tokenString, claims, getSecret, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
	if err != nil {
		return User{}, fmt.Errorf("ParseToken error: %w", err)
	}
	if _, err := id.ParseID(string(claims.User.ID)); err != nil {
		return User{}, fmt.Errorf("validation ID error: %w", err)
	}
	if err := ValidateUsername(claims.User.Username); err != nil {
		return User{}, fmt.Errorf("validation username error: %w", err)
	}

	if claims.IssuedAt == nil {
		return User{}, fmt.Errorf("token issue error")
	}
	if claims.ExpiresAt == nil || !claims.ExpiresAt.Time.After(claims.IssuedAt.Time) {
		return User{}, fmt.Errorf("token expire error")
	}
	return claims.User, nil
}

func UserFromContext(ctx context.Context) (*User, error) {
	sess, ok := ctx.Value(userKey{}).(*User)
	if !ok || sess == nil {
		return nil, fmt.Errorf("unauthorized")
	}
	return sess, nil
}

func ContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}
