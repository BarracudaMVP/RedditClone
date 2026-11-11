package id

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type ID string

const MinLengthID = 1

func ParseID(value string) (ID, error) {
	if len(value) < MinLengthID {
		return "", fmt.Errorf("ParseID error")
	}
	return ID(value), nil
}
func NewID() (ID, error) {
	bytes := make([]byte, 12)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("NewID error: %w", err)
	}
	return ID(hex.EncodeToString(bytes)), nil
}
