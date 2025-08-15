package common

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"regexp"
)

// ParseFloat safely parses float values from various types (DBF compatibility)
func ParseFloat(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		// Remove commas and parse
		cleanStr := strings.ReplaceAll(v, ",", "")
		if f, err := strconv.ParseFloat(cleanStr, 64); err == nil {
			return f
		}
	}
	return 0.0
}

// ParseCSVLine properly parses a CSV line handling quoted fields
func ParseCSVLine(line string) []string {
	var fields []string
	var current strings.Builder
	inQuotes := false
	
	for i := 0; i < len(line); i++ {
		ch := line[i]
		
		if ch == '"' {
			inQuotes = !inQuotes
		} else if ch == ',' && !inQuotes {
			fields = append(fields, current.String())
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}
	
	// Don't forget the last field
	fields = append(fields, current.String())
	
	return fields
}

// ExtractCheckNumber extracts check number from a description string
func ExtractCheckNumber(description string) string {
	// Look for patterns like "Check #1234" or "CHK 1234" or just "1234"
	re := regexp.MustCompile(`(?i)(?:check|chk|ck)?\s*#?\s*(\d+)`)
	matches := re.FindStringSubmatch(description)
	if len(matches) > 1 {
		return matches[1]
	}
	
	// If no pattern found, look for any number at the beginning
	re = regexp.MustCompile(`^\d+`)
	if match := re.FindString(description); match != "" {
		return match
	}
	
	return ""
}

// ParseDate parses various date formats
func ParseDate(dateStr string) (time.Time, error) {
	// Try common date formats
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"1/2/2006",
		"01-02-2006",
		"1-2-2006",
		"2006/01/02",
		"Jan 2, 2006",
		"January 2, 2006",
		"02-Jan-2006",
		"Mon, 02 Jan 2006",
	}
	
	dateStr = strings.TrimSpace(dateStr)
	
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// GetMapKeys returns all keys from a map (useful for debugging)
func GetMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ParseBool parses various boolean representations
func ParseBool(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		lower := strings.ToLower(strings.TrimSpace(v))
		return lower == "true" || lower == "t" || lower == ".t." || 
		       lower == "yes" || lower == "y" || lower == "1"
	case int:
		return v != 0
	case float64:
		return v != 0
	}
	return false
}

// SanitizeString removes potentially dangerous characters
func SanitizeString(s string) string {
	// Remove null bytes and control characters
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.TrimSpace(s)
	
	// Remove other control characters
	result := strings.Builder{}
	for _, r := range s {
		if r >= 32 && r != 127 { // Printable characters only
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

// FormatCurrency formats a float as currency string
func FormatCurrency(amount float64) string {
	if amount < 0 {
		return fmt.Sprintf("-$%.2f", -amount)
	}
	return fmt.Sprintf("$%.2f", amount)
}

// ParseCurrency parses a currency string to float64
func ParseCurrency(s string) (float64, error) {
	// Remove currency symbols and commas
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "(", "-") // Handle (100.00) format
	s = strings.ReplaceAll(s, ")", "")
	
	return strconv.ParseFloat(s, 64)
}