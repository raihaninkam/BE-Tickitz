package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// validateEmail memvalidasi format email
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("format email tidak valid")
	}
	return nil
}

// validatePassword memvalidasi password sesuai requirement
func ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}

	// Check minimum length (8 characters)
	if len(password) < 8 {
		return fmt.Errorf("password minimal 8 karakter")
	}

	// Check for at least one uppercase letter
	hasUpper := false
	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		return fmt.Errorf("password harus mengandung minimal 1 huruf besar")
	}

	// Check for special characters
	specialChars := "!@#$%^&*()_+"
	hasSpecial := false
	for _, char := range password {
		if strings.ContainsRune(specialChars, char) {
			hasSpecial = true
			break
		}
	}
	if !hasSpecial {
		return fmt.Errorf("password harus mengandung minimal 1 karakter spesial (!@#$%%^&*()_+)")
	}

	return nil
}
