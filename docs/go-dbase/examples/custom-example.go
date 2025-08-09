// Example: Custom Field Types and Conversions with go-dbase
// Demonstrates custom data handling and transformations

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Valentin-Kaiser/go-dbase/dbase"
)

// Custom types for specialized data handling
type (
	// PhoneNumber with validation and formatting
	PhoneNumber string

	// Currency with precision handling
	Currency float64

	// EmailAddress with validation
	EmailAddress string

	// JSONField for storing JSON in character fields
	JSONField map[string]interface{}
)

// CustomRecord demonstrates custom field mapping
type CustomRecord struct {
	ID       int          `dbase:"ID"`
	Name     string       `dbase:"NAME"`
	Phone    PhoneNumber  `dbase:"PHONE"`
	Email    EmailAddress `dbase:"EMAIL"`
	Balance  Currency     `dbase:"BALANCE"`
	Metadata JSONField    `dbase:"METADATA"` // JSON stored as string
	Tags     []string     `dbase:"TAGS"`     // Array stored as delimited string
}

func main() {
	fmt.Println("=== Custom Field Types and Conversions ===\n")

	// Example 1: Custom validation
	demonstrateValidation()

	// Example 2: Custom formatting
	demonstrateFormatting()

	// Example 3: Custom transformations
	demonstrateTransformations()

	// Example 4: JSON storage in DBF
	demonstrateJSONStorage()

	// Example 5: Array handling
	demonstrateArrayHandling()
}

// demonstrateValidation shows field validation
func demonstrateValidation() {
	fmt.Println("=== Field Validation ===")

	// Phone number validation
	phones := []string{
		"123-456-7890",
		"(123) 456-7890",
		"123.456.7890",
		"invalid-phone",
	}

	for _, phone := range phones {
		pn := PhoneNumber(phone)
		if pn.IsValid() {
			fmt.Printf("✓ Valid phone: %s -> %s\n", phone, pn.Format())
		} else {
			fmt.Printf("✗ Invalid phone: %s\n", phone)
		}
	}

	// Email validation
	emails := []string{
		"user@example.com",
		"invalid.email",
		"user+tag@domain.co.uk",
	}

	fmt.Println("\nEmail Validation:")
	for _, email := range emails {
		em := EmailAddress(email)
		if em.IsValid() {
			fmt.Printf("✓ Valid email: %s\n", email)
		} else {
			fmt.Printf("✗ Invalid email: %s\n", email)
		}
	}
}

// PhoneNumber methods
func (p PhoneNumber) IsValid() bool {
	// Remove all non-digits
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, string(p))

	// US phone number should have 10 digits
	return len(digits) == 10
}

func (p PhoneNumber) Format() string {
	// Extract digits
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, string(p))

	if len(digits) == 10 {
		return fmt.Sprintf("(%s) %s-%s", digits[:3], digits[3:6], digits[6:])
	}
	return string(p)
}

// EmailAddress methods
func (e EmailAddress) IsValid() bool {
	parts := strings.Split(string(e), "@")
	if len(parts) != 2 {
		return false
	}
	return len(parts[0]) > 0 && strings.Contains(parts[1], ".")
}

func (e EmailAddress) Domain() string {
	parts := strings.Split(string(e), "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// Currency methods
func (c Currency) Format() string {
	return fmt.Sprintf("$%.2f", float64(c))
}

func (c Currency) Cents() int {
	return int(float64(c) * 100)
}

// demonstrateFormatting shows custom formatting
func demonstrateFormatting() {
	fmt.Println("\n=== Custom Formatting ===")

	// Currency formatting
	amounts := []Currency{1234.56, 0.99, -500.00, 1000000.01}
	
	fmt.Println("Currency Formatting:")
	for _, amount := range amounts {
		fmt.Printf("  %f -> %s (%d cents)\n", 
			amount, amount.Format(), amount.Cents())
	}

	// Date formatting for DBF
	dates := []time.Time{
		time.Now(),
		time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1980, 6, 15, 14, 30, 0, 0, time.UTC),
	}

	fmt.Println("\nDate Formatting for DBF:")
	for _, date := range dates {
		dbfDate := formatDateForDBF(date)
		dbfDateTime := formatDateTimeForDBF(date)
		fmt.Printf("  %s -> Date: %s, DateTime: %s\n",
			date.Format(time.RFC3339),
			dbfDate, dbfDateTime)
	}
}

func formatDateForDBF(t time.Time) string {
	return t.Format("20060102") // YYYYMMDD
}

func formatDateTimeForDBF(t time.Time) string {
	// Some DBF formats store datetime as seconds since epoch
	return strconv.FormatInt(t.Unix(), 10)
}

// demonstrateTransformations shows data transformations
func demonstrateTransformations() {
	fmt.Println("\n=== Data Transformations ===")

	// Transform DBF row to custom struct
	table, err := openSampleTable()
	if err != nil {
		log.Printf("Error opening table: %v", err)
		return
	}
	defer table.Close()

	fmt.Println("Transforming DBF rows to custom structs...")
	
	// Example transformation pipeline
	transformer := NewDataTransformer()
	
	// Add transformation rules
	transformer.AddRule("PHONE", normalizePhone)
	transformer.AddRule("EMAIL", normalizeEmail)
	transformer.AddRule("BALANCE", parseCurrency)
	
	fmt.Println("Transformation rules applied:")
	fmt.Println("  - Phone: Normalize to (XXX) XXX-XXXX")
	fmt.Println("  - Email: Convert to lowercase")
	fmt.Println("  - Balance: Parse currency string to float")
}

// DataTransformer handles field transformations
type DataTransformer struct {
	rules map[string]TransformFunc
}

type TransformFunc func(interface{}) (interface{}, error)

func NewDataTransformer() *DataTransformer {
	return &DataTransformer{
		rules: make(map[string]TransformFunc),
	}
}

func (dt *DataTransformer) AddRule(field string, fn TransformFunc) {
	dt.rules[field] = fn
}

func (dt *DataTransformer) Transform(field string, value interface{}) (interface{}, error) {
	if fn, ok := dt.rules[field]; ok {
		return fn(value)
	}
	return value, nil
}

// Transformation functions
func normalizePhone(val interface{}) (interface{}, error) {
	str := fmt.Sprintf("%v", val)
	phone := PhoneNumber(str)
	return phone.Format(), nil
}

func normalizeEmail(val interface{}) (interface{}, error) {
	str := fmt.Sprintf("%v", val)
	return strings.ToLower(str), nil
}

func parseCurrency(val interface{}) (interface{}, error) {
	str := fmt.Sprintf("%v", val)
	// Remove currency symbols and commas
	str = strings.ReplaceAll(str, "$", "")
	str = strings.ReplaceAll(str, ",", "")
	return strconv.ParseFloat(str, 64)
}

// demonstrateJSONStorage shows storing JSON in DBF character fields
func demonstrateJSONStorage() {
	fmt.Println("\n=== JSON Storage in DBF ===")

	// Example: Store complex data as JSON in a character field
	metadata := JSONField{
		"created_by": "admin",
		"version":    "1.0",
		"features":   []string{"feature1", "feature2"},
		"settings": map[string]interface{}{
			"enabled": true,
			"limit":   100,
		},
	}

	// Convert to JSON string for storage
	jsonStr, err := json.Marshal(metadata)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}

	fmt.Printf("Original: %+v\n", metadata)
	fmt.Printf("Stored as: %s\n", jsonStr)

	// Retrieve and parse
	var retrieved JSONField
	err = json.Unmarshal(jsonStr, &retrieved)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return
	}

	fmt.Printf("Retrieved: %+v\n", retrieved)
}

// demonstrateArrayHandling shows handling arrays in DBF
func demonstrateArrayHandling() {
	fmt.Println("\n=== Array Handling in DBF ===")

	// Arrays can be stored as delimited strings
	tags := []string{"customer", "vip", "priority", "wholesale"}
	
	// Store as pipe-delimited string
	stored := strings.Join(tags, "|")
	fmt.Printf("Array: %v\n", tags)
	fmt.Printf("Stored as: %s\n", stored)

	// Retrieve and split
	retrieved := strings.Split(stored, "|")
	fmt.Printf("Retrieved: %v\n", retrieved)

	// Alternative: Store as JSON
	jsonTags, _ := json.Marshal(tags)
	fmt.Printf("\nAlternative (JSON): %s\n", jsonTags)
}

// Helper function to open a sample table
func openSampleTable() (*dbase.File, error) {
	// This would open an actual DBF file
	// For demo purposes, we're returning a mock
	return dbase.OpenTable(&dbase.Config{
		Filename:   "sample.dbf",
		TrimSpaces: true,
		ReadOnly:   true,
	})
}

// CustomFieldReader demonstrates reading with custom type conversion
type CustomFieldReader struct {
	table       *dbase.File
	transformer *DataTransformer
}

func NewCustomFieldReader(table *dbase.File) *CustomFieldReader {
	return &CustomFieldReader{
		table:       table,
		transformer: NewDataTransformer(),
	}
}

func (r *CustomFieldReader) ReadRecord() (*CustomRecord, error) {
	row, err := r.table.Next()
	if err != nil {
		return nil, err
	}

	record := &CustomRecord{}
	
	// Map fields with custom conversion
	for i := 0; i < row.FieldCount(); i++ {
		field := row.Field(i)
		value, _ := field.GetValue()
		
		// Apply transformations
		transformed, _ := r.transformer.Transform(field.Name(), value)
		
		// Map to struct fields (simplified)
		switch field.Name() {
		case "ID":
			record.ID, _ = transformed.(int)
		case "NAME":
			record.Name, _ = transformed.(string)
		case "PHONE":
			record.Phone = PhoneNumber(fmt.Sprintf("%v", transformed))
		case "EMAIL":
			record.Email = EmailAddress(fmt.Sprintf("%v", transformed))
		case "BALANCE":
			if f, ok := transformed.(float64); ok {
				record.Balance = Currency(f)
			}
		case "METADATA":
			// Parse JSON from string field
			var meta JSONField
			json.Unmarshal([]byte(fmt.Sprintf("%v", transformed)), &meta)
			record.Metadata = meta
		case "TAGS":
			// Parse delimited string to array
			str := fmt.Sprintf("%v", transformed)
			record.Tags = strings.Split(str, "|")
		}
	}
	
	return record, nil
}

// Tips for custom field handling:
// 1. Validate data on read and write
// 2. Use consistent formatting
// 3. Document custom formats
// 4. Handle null/empty values
// 5. Consider field length limits
// 6. Test with target applications
// 7. Implement error recovery