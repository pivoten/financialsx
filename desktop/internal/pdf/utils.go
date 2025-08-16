package pdf

import (
	"fmt"
	"strings"
)

// FormatCurrency formats a number as currency
func FormatCurrency(amount float64) string {
	if amount < 0 {
		return fmt.Sprintf("($%,.2f)", -amount)
	}
	return fmt.Sprintf("$%,.2f", amount)
}

// FormatNumber formats a number with thousands separator
func FormatNumber(num float64, decimals int) string {
	format := fmt.Sprintf("%%.%df", decimals)
	str := fmt.Sprintf(format, num)
	
	// Add thousands separator
	parts := strings.Split(str, ".")
	intPart := parts[0]
	
	// Add commas
	result := ""
	negative := false
	if intPart[0] == '-' {
		negative = true
		intPart = intPart[1:]
	}
	
	for i, digit := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result += ","
		}
		result += string(digit)
	}
	
	if negative {
		result = "-" + result
	}
	
	if len(parts) > 1 {
		result += "." + parts[1]
	}
	
	return result
}

// FormatDate formats a date string
func FormatDate(dateStr string) string {
	// Convert various date formats to MM/DD/YYYY
	// This is simplified - could be enhanced with proper date parsing
	if len(dateStr) >= 10 {
		// Assume YYYY-MM-DD format
		if dateStr[4] == '-' && dateStr[7] == '-' {
			return dateStr[5:7] + "/" + dateStr[8:10] + "/" + dateStr[0:4]
		}
	}
	return dateStr
}

// TruncateText truncates text to fit within a specified width
func TruncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	if maxLen <= 3 {
		return text[:maxLen]
	}
	return text[:maxLen-3] + "..."
}

// SanitizeFileName removes invalid characters from filename
func SanitizeFileName(name string) string {
	// Remove invalid filename characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}

// CalculateColumnWidths calculates optimal column widths for a table
func CalculateColumnWidths(headers []string, totalWidth float64) []float64 {
	numCols := len(headers)
	if numCols == 0 {
		return []float64{}
	}
	
	// Simple equal distribution - could be enhanced with content analysis
	baseWidth := totalWidth / float64(numCols)
	widths := make([]float64, numCols)
	
	for i := range widths {
		widths[i] = baseWidth
	}
	
	return widths
}

// WrapText wraps text to fit within a specified width
func WrapText(text string, maxWidth int) []string {
	if len(text) <= maxWidth {
		return []string{text}
	}
	
	var lines []string
	words := strings.Fields(text)
	currentLine := ""
	
	for _, word := range words {
		if len(currentLine)+len(word)+1 <= maxWidth {
			if currentLine != "" {
				currentLine += " "
			}
			currentLine += word
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}
	
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	
	return lines
}