package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	HTTPAddr     string
	StaticDir    string
	JWT          JWTConfig
	BcryptConfig BcryptConfig
}
type JWTConfig struct {
	Secret []byte
	TTL    time.Duration
}
type BcryptConfig struct {
	DummyHash string
	Cost      int
}
type envReader struct {
	errs []error
}

func (r *envReader) str(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func (r *envReader) int(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := strconv.ParseInt(v, 10, 0)
	if err != nil {
		r.errs = append(r.errs, fmt.Errorf("%s: %w", key, err))
		return def
	}
	return int(d)
}

func (r *envReader) duration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		r.errs = append(r.errs, fmt.Errorf("%s: %w", key, err))
		return def
	}
	return d
}
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, errors.New("cannot load .env")
	}
	var r envReader
	cfg := Config{
		HTTPAddr:  r.str("HTTP_ADDR", ":8080"),
		StaticDir: r.str("STATIC_DIR", "./static"),
		JWT: JWTConfig{
			Secret: []byte(r.str("JWT_SECRET", "")),
			TTL:    r.duration("JWT_TTL", 7*24*time.Hour),
		},
		BcryptConfig: BcryptConfig{
			Cost: r.int("BCRYPT_COST", bcrypt.DefaultCost),
		},
	}
	if err := errors.Join(r.errs...); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	dummyPassword := r.str("DUMMY_PASSWORD", "dummy_password")
	hash, err := bcrypt.GenerateFromPassword([]byte(dummyPassword), cfg.BcryptConfig.Cost)
	if err != nil {
		return nil, fmt.Errorf("hash DUMMY_PASSWORD: %w", err)
	}
	cfg.BcryptConfig.DummyHash = string(hash)
	return &cfg, nil
}

func (c *Config) Validate() error {
	var errs []error

	if c.HTTPAddr == "" {
		errs = append(errs, errors.New("HTTP_ADDR is required"))
	}

	if c.StaticDir == "" {
		errs = append(errs, errors.New("STATIC_DIR is required"))
	}
	if len(c.JWT.Secret) < 32 {
		errs = append(errs, errors.New("JWT_SECRET must be at least 32 bytes"))
	}

	if c.JWT.TTL <= 0 {
		errs = append(errs, errors.New("JWT_TTL must be positive"))
	}
	if c.BcryptConfig.Cost < bcrypt.MinCost || c.BcryptConfig.Cost > bcrypt.MaxCost {
		errs = append(errs, fmt.Errorf("BCRYPT_COST must be between %d and %d", bcrypt.MinCost, bcrypt.MaxCost))
	}
	return errors.Join(errs...)
}
