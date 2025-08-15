// Package utilities provides helper functions and utilities used throughout the application
package utilities

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// FormatCurrency formats a float64 as a currency string
func FormatCurrency(amount float64) string {
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}
	return fmt.Sprintf("%s$%,.2f", sign, amount)
}

// ParseDate parses various date formats
func ParseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"01-02-2006",
		"2006/01/02",
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// SanitizeFilename removes invalid characters from a filename
func SanitizeFilename(filename string) string {
	// Remove invalid characters for Windows/Mac/Linux
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	sanitized := invalidChars.ReplaceAllString(filename, "_")
	
	// Trim spaces and dots from the ends
	sanitized = strings.Trim(sanitized, " .")
	
	// Ensure it's not empty
	if sanitized == "" {
		sanitized = "file"
	}
	
	return sanitized
}

// NormalizePath normalizes a file path for the current OS
func NormalizePath(path string) string {
	// Convert backslashes to forward slashes
	path = strings.ReplaceAll(path, "\\", "/")
	
	// Clean the path
	path = filepath.Clean(path)
	
	// Convert to OS-specific separators
	return filepath.FromSlash(path)
}

// IsValidEmail validates an email address
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidTaxID validates a US tax ID (EIN or SSN format)
func IsValidTaxID(taxID string) bool {
	// Remove all non-digits
	cleaned := regexp.MustCompile(`\D`).ReplaceAllString(taxID, "")
	
	// Check for SSN (9 digits) or EIN (9 digits)
	if len(cleaned) != 9 {
		return false
	}
	
	// Basic validation - not all zeros
	if cleaned == "000000000" {
		return false
	}
	
	return true
}

// Encrypt encrypts a string using AES
func Encrypt(text, key string) (string, error) {
	// Ensure key is 32 bytes for AES-256
	keyBytes := make([]byte, 32)
	copy(keyBytes, []byte(key))
	
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}
	
	plaintext := []byte(text)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	
	// Generate IV
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts an AES-encrypted string
func Decrypt(encrypted, key string) (string, error) {
	// Ensure key is 32 bytes for AES-256
	keyBytes := make([]byte, 32)
	copy(keyBytes, []byte(key))
	
	ciphertext, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}
	
	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	
	return string(ciphertext), nil
}

// CalculatePercentage calculates percentage with proper handling of zero division
func CalculatePercentage(part, whole float64) float64 {
	if whole == 0 {
		return 0
	}
	return (part / whole) * 100
}

// RoundToDecimalPlaces rounds a float to specified decimal places
func RoundToDecimalPlaces(value float64, places int) float64 {
	shift := 1.0
	for i := 0; i < places; i++ {
		shift *= 10
	}
	return float64(int(value*shift+0.5)) / shift
}

// GenerateID generates a unique ID for records
func GenerateID(prefix string) string {
	timestamp := time.Now().UnixNano()
	random := make([]byte, 4)
	rand.Read(random)
	randomStr := base64.URLEncoding.EncodeToString(random)[:6]
	
	if prefix != "" {
		return fmt.Sprintf("%s_%d_%s", prefix, timestamp, randomStr)
	}
	return fmt.Sprintf("%d_%s", timestamp, randomStr)
}