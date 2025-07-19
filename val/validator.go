package val

import (
	"fmt"
	"net/mail"
	"regexp"
)

const (
	USERNAME_MIN_LENGTH  = 3
	USERNAME_MAX_LENGTH  = 100
	PASSWORD_MIN_LENGTH  = 6
	PASSWORD_MAX_LENGTH  = 50
	EMAIL_MIN_LENGTH     = 3
	EMAIL_MAX_LENGTH     = 100
	FULL_NAME_MIN_LENGTH = 3
	FULL_NAME_MAX_LENGTH = 100
)

var (
	isValidUsername = regexp.MustCompile(`^[a-z0-9_]+$`).MatchString
	isValidFullName = regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString
)

func ValidateStringLength(value string, minLength int, maxLength int) error {
	n := len(value)
	if n < minLength || n > maxLength {
		return fmt.Errorf("invalid input length: must contain between %d-%d characters", minLength, maxLength)
	}

	return nil
}

func ValidateUsername(value string) error {
	if err := ValidateStringLength(value, USERNAME_MIN_LENGTH, USERNAME_MAX_LENGTH); err != nil {
		return err
	}

	if !isValidUsername(value) {
		return fmt.Errorf("username must contain only lowercase letters, digits, or underscores")
	}

	return nil
}

func ValidatePassword(value string) error {
	return ValidateStringLength(value, PASSWORD_MIN_LENGTH, PASSWORD_MAX_LENGTH)
}

func ValidateEmail(value string) error {
	if err := ValidateStringLength(value, EMAIL_MIN_LENGTH, EMAIL_MAX_LENGTH); err != nil {
		return err
	}

	if _, err := mail.ParseAddress(value); err != nil {
		return fmt.Errorf("invalid email")
	}

	return nil
}

func ValidateFullName(value string) error {
	if err := ValidateStringLength(value, FULL_NAME_MIN_LENGTH, FULL_NAME_MAX_LENGTH); err != nil {
		return err
	}

	if !isValidFullName(value) {
		return fmt.Errorf("full name must contain only letters and spaces")
	}

	return nil
}

func ValidateEmailID(value int64) error {
	if value <= 0 {
		return fmt.Errorf("must be a positive integer")
	}

	return nil
}

func ValidateSecretCode(value string) error {
	return ValidateStringLength(value, 32, 128)
}
