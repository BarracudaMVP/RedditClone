package id

import "fmt"

type ID string

const MinLengthID = 1

func ParseID(value string) (ID, error) {
	if len(value) < MinLengthID {
		return "", fmt.Errorf("ParseID error")
	}
	return ID(value), nil
}
