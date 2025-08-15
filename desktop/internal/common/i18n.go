// Package common provides i18n (internationalization) support
package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// I18n handles internationalization
type I18n struct {
	mu           sync.RWMutex
	translations map[string]map[string]interface{} // locale -> translations
	fallback     string                             // fallback locale
	current      string                             // current locale
}

// NewI18n creates a new i18n instance
func NewI18n(fallbackLocale string) *I18n {
	return &I18n{
		translations: make(map[string]map[string]interface{}),
		fallback:     fallbackLocale,
		current:      fallbackLocale,
	}
}

// LoadLocale loads translations from a JSON file
func (i *I18n) LoadLocale(locale string, filePath string) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read locale file %s: %w", filePath, err)
	}

	var translations map[string]interface{}
	if err := json.Unmarshal(data, &translations); err != nil {
		return fmt.Errorf("failed to parse locale file %s: %w", filePath, err)
	}

	i.translations[locale] = translations
	return nil
}

// LoadLocalesFromDir loads all locale files from a directory
func (i *I18n) LoadLocalesFromDir(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to list locale files: %w", err)
	}

	for _, file := range files {
		locale := strings.TrimSuffix(filepath.Base(file), ".json")
		if err := i.LoadLocale(locale, file); err != nil {
			return fmt.Errorf("failed to load locale %s: %w", locale, err)
		}
	}

	return nil
}

// SetLocale sets the current locale
func (i *I18n) SetLocale(locale string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	
	if _, exists := i.translations[locale]; exists {
		i.current = locale
	}
}

// GetLocale returns the current locale
func (i *I18n) GetLocale() string {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.current
}

// GetAvailableLocales returns all loaded locales
func (i *I18n) GetAvailableLocales() []string {
	i.mu.RLock()
	defer i.mu.RUnlock()
	
	locales := make([]string, 0, len(i.translations))
	for locale := range i.translations {
		locales = append(locales, locale)
	}
	return locales
}

// T translates a key to the current locale
func (i *I18n) T(key string, args ...interface{}) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// Try current locale first
	if text := i.getTranslation(i.current, key); text != "" {
		if len(args) > 0 {
			return fmt.Sprintf(text, args...)
		}
		return text
	}

	// Fall back to default locale
	if i.current != i.fallback {
		if text := i.getTranslation(i.fallback, key); text != "" {
			if len(args) > 0 {
				return fmt.Sprintf(text, args...)
			}
			return text
		}
	}

	// Return the key itself if no translation found
	return key
}

// getTranslation gets a translation from nested keys (e.g., "login.errors.invalidCredentials")
func (i *I18n) getTranslation(locale, key string) string {
	translations, exists := i.translations[locale]
	if !exists {
		return ""
	}

	keys := strings.Split(key, ".")
	var current interface{} = translations

	for _, k := range keys {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[k]
		default:
			return ""
		}
	}

	if str, ok := current.(string); ok {
		return str
	}

	return ""
}

// Pluralize returns the correct plural form based on count
func (i *I18n) Pluralize(key string, count int) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// Simple English pluralization rules
	// You can extend this for other languages
	if count == 1 {
		return i.T(key + ".one")
	}
	return i.T(key+".other", count)
}

// Currency formats a number as currency for the current locale
func (i *I18n) Currency(amount float64) string {
	// Simple currency formatting
	// You can extend this based on locale
	switch i.current {
	case "es":
		return fmt.Sprintf("€%.2f", amount)
	default:
		return fmt.Sprintf("$%.2f", amount)
	}
}

// Date formats a date string for the current locale
func (i *I18n) Date(date string) string {
	// Simple date formatting
	// You can extend this based on locale
	switch i.current {
	case "es":
		// Spanish format: DD/MM/YYYY
		return date // Implement proper date formatting
	default:
		// US format: MM/DD/YYYY
		return date // Implement proper date formatting
	}
}

// Global i18n instance (optional - you can also inject this)
var globalI18n *I18n

// InitI18n initializes the global i18n instance
func InitI18n(localesDir string, defaultLocale string) error {
	globalI18n = NewI18n(defaultLocale)
	return globalI18n.LoadLocalesFromDir(localesDir)
}

// T is a convenience function for global translations
func T(key string, args ...interface{}) string {
	if globalI18n == nil {
		return key
	}
	return globalI18n.T(key, args...)
}

// SetGlobalLocale sets the locale for the global instance
func SetGlobalLocale(locale string) {
	if globalI18n != nil {
		globalI18n.SetLocale(locale)
	}
}

// GetGlobalLocale gets the current locale from the global instance
func GetGlobalLocale() string {
	if globalI18n != nil {
		return globalI18n.GetLocale()
	}
	return "en"
}