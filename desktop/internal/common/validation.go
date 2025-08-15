package common

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// Email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	
	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)
	
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	
	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}
	
	return nil
}

// ValidateUsername validates a username
func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}
	if len(username) > 50 {
		return fmt.Errorf("username must be less than 50 characters")
	}
	
	// Check for valid characters (alphanumeric, underscore, dash)
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
	if !validUsername.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, underscores, and dashes")
	}
	
	return nil
}

// ValidateDateRange validates that start date is before end date
func ValidateDateRange(start, end time.Time) error {
	if start.After(end) {
		return fmt.Errorf("start date must be before end date")
	}
	return nil
}

// ValidateAmount validates a monetary amount
func ValidateAmount(amount float64) error {
	if amount < 0 {
		return fmt.Errorf("amount cannot be negative")
	}
	// Check for reasonable maximum (to catch data errors)
	if amount > 999999999.99 {
		return fmt.Errorf("amount exceeds maximum allowed value")
	}
	return nil
}

// ValidateAccountNumber validates an account number format
func ValidateAccountNumber(accountNumber string) error {
	accountNumber = strings.TrimSpace(accountNumber)
	if accountNumber == "" {
		return fmt.Errorf("account number cannot be empty")
	}
	if len(accountNumber) > 20 {
		return fmt.Errorf("account number is too long")
	}
	// Check for valid characters (alphanumeric and dash)
	validAccount := regexp.MustCompile(`^[a-zA-Z0-9\-]+$`)
	if !validAccount.MatchString(accountNumber) {
		return fmt.Errorf("account number contains invalid characters")
	}
	return nil
}

// ValidateCheckNumber validates a check number
func ValidateCheckNumber(checkNumber string) error {
	checkNumber = strings.TrimSpace(checkNumber)
	if checkNumber == "" {
		return fmt.Errorf("check number cannot be empty")
	}
	// Check numbers are typically numeric but can have prefixes
	validCheck := regexp.MustCompile(`^[a-zA-Z0-9\-]+$`)
	if !validCheck.MatchString(checkNumber) {
		return fmt.Errorf("check number contains invalid characters")
	}
	return nil
}

// ValidateCompanyName validates a company name
func ValidateCompanyName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("company name cannot be empty")
	}
	if len(name) > 100 {
		return fmt.Errorf("company name is too long")
	}
	// Check for SQL injection attempts
	dangerousPatterns := []string{"DROP", "DELETE", "INSERT", "UPDATE", "SELECT", "--", "/*", "*/"}
	upperName := strings.ToUpper(name)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(upperName, pattern) {
			return fmt.Errorf("company name contains invalid characters or patterns")
		}
	}
	return nil
}

// ValidateFilePath validates a file path (basic validation)
func ValidateFilePath(path string) error {
	if path == "" {
		return fmt.Errorf("file path cannot be empty")
	}
	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return fmt.Errorf("file path cannot contain '..'")
	}
	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("file path contains invalid characters")
	}
	return nil
}

// ValidatePagination validates pagination parameters
func ValidatePagination(page, pageSize int) error {
	if page < 1 {
		return fmt.Errorf("page must be greater than 0")
	}
	if pageSize < 1 {
		return fmt.Errorf("page size must be greater than 0")
	}
	if pageSize > 1000 {
		return fmt.Errorf("page size cannot exceed 1000")
	}
	return nil
}