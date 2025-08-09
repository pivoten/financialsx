// Example: Searching DBF Files with go-dbase
// Demonstrates various search and filter techniques

package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/Valentin-Kaiser/go-dbase/dbase"
)

// SearchCriteria defines search parameters
type SearchCriteria struct {
	Field    string      // Specific field to search (empty = all fields)
	Value    interface{} // Value to search for
	Operator string      // Comparison operator: =, !=, >, <, >=, <=, LIKE, IN
	CaseSensitive bool   // Case-sensitive search
}

// SearchResult holds search results
type SearchResult struct {
	RowIndex int
	Data     map[string]interface{}
	Matched  []string // Fields that matched
}

func main() {
	// Example 1: Simple text search
	simpleSearch()

	// Example 2: Field-specific search
	fieldSearch()

	// Example 3: Numeric range search
	rangeSearch()

	// Example 4: Date range search
	dateSearch()

	// Example 5: Complex multi-criteria search
	complexSearch()

	// Example 6: Regex pattern search
	regexSearch()
}

// simpleSearch demonstrates basic text search across all fields
func simpleSearch() {
	fmt.Println("=== Simple Text Search ===")

	table, err := dbase.OpenTable(&dbase.Config{
		Filename:   "customers.dbf",
		TrimSpaces: true,
		ReadOnly:   true,
	})
	if err != nil {
		log.Fatal("Failed to open table:", err)
	}
	defer table.Close()

	searchTerm := "smith"
	results := searchAllFields(table, searchTerm)

	fmt.Printf("Found %d records containing '%s':\n", len(results), searchTerm)
	for _, result := range results {
		fmt.Printf("  Row %d: %v\n", result.RowIndex, result.Data)
		fmt.Printf("    Matched in: %v\n", result.Matched)
	}
}

// searchAllFields searches for a term across all fields
func searchAllFields(table *dbase.File, searchTerm string) []SearchResult {
	var results []SearchResult
	searchLower := strings.ToLower(searchTerm)
	rowIndex := 0

	for {
		row, err := table.Next()
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				break
			}
			log.Printf("Error reading row: %v", err)
			continue
		}

		// Skip deleted records
		deleted, _ := table.Deleted()
		if deleted {
			continue
		}

		rowData := make(map[string]interface{})
		var matchedFields []string

		// Check each field
		for i := 0; i < row.FieldCount(); i++ {
			field := row.Field(i)
			value, _ := field.GetValue()
			fieldName := field.Name()
			rowData[fieldName] = value

			// Convert to string and search
			strValue := fmt.Sprintf("%v", value)
			if strings.Contains(strings.ToLower(strValue), searchLower) {
				matchedFields = append(matchedFields, fieldName)
			}
		}

		// Add to results if matched
		if len(matchedFields) > 0 {
			results = append(results, SearchResult{
				RowIndex: rowIndex,
				Data:     rowData,
				Matched:  matchedFields,
			})
		}

		rowIndex++
	}

	return results
}

// fieldSearch demonstrates searching specific fields
func fieldSearch() {
	fmt.Println("\n=== Field-Specific Search ===")

	table, err := dbase.OpenTable(&dbase.Config{
		Filename:   "customers.dbf",
		TrimSpaces: true,
		ReadOnly:   true,
	})
	if err != nil {
		log.Fatal("Failed to open table:", err)
	}
	defer table.Close()

	// Search for specific customer ID
	criteria := SearchCriteria{
		Field:    "CUST_ID",
		Value:    "C0123",
		Operator: "=",
	}

	results := searchByField(table, criteria)
	fmt.Printf("Found %d records where %s %s %v\n", 
		len(results), criteria.Field, criteria.Operator, criteria.Value)

	for _, result := range results {
		fmt.Printf("  Row %d: %v\n", result.RowIndex, result.Data[criteria.Field])
	}
}

// searchByField searches a specific field with criteria
func searchByField(table *dbase.File, criteria SearchCriteria) []SearchResult {
	var results []SearchResult
	rowIndex := 0

	for {
		row, err := table.Next()
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				break
			}
			continue
		}

		deleted, _ := table.Deleted()
		if deleted {
			continue
		}

		// Find the specific field
		var fieldValue interface{}
		found := false

		for i := 0; i < row.FieldCount(); i++ {
			field := row.Field(i)
			if field.Name() == criteria.Field {
				fieldValue, _ = field.GetValue()
				found = true
				break
			}
		}

		if !found {
			continue
		}

		// Apply comparison
		if matchesCriteria(fieldValue, criteria) {
			rowData := extractRowData(row)
			results = append(results, SearchResult{
				RowIndex: rowIndex,
				Data:     rowData,
				Matched:  []string{criteria.Field},
			})
		}

		rowIndex++
	}

	return results
}

// matchesCriteria checks if a value matches search criteria
func matchesCriteria(value interface{}, criteria SearchCriteria) bool {
	switch criteria.Operator {
	case "=", "==":
		return compareEqual(value, criteria.Value, criteria.CaseSensitive)
	case "!=", "<>":
		return !compareEqual(value, criteria.Value, criteria.CaseSensitive)
	case ">":
		return compareGreater(value, criteria.Value)
	case "<":
		return compareLess(value, criteria.Value)
	case ">=":
		return compareGreater(value, criteria.Value) || compareEqual(value, criteria.Value, true)
	case "<=":
		return compareLess(value, criteria.Value) || compareEqual(value, criteria.Value, true)
	case "LIKE":
		return compareLike(value, criteria.Value, criteria.CaseSensitive)
	case "IN":
		return compareIn(value, criteria.Value)
	default:
		return false
	}
}

// Comparison helper functions
func compareEqual(a, b interface{}, caseSensitive bool) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	
	if !caseSensitive {
		aStr = strings.ToLower(aStr)
		bStr = strings.ToLower(bStr)
	}
	
	return aStr == bStr
}

func compareGreater(a, b interface{}) bool {
	// Handle numeric comparison
	switch va := a.(type) {
	case int, int32, int64:
		if vb, ok := b.(int); ok {
			return va.(int) > vb
		}
	case float32, float64:
		if vb, ok := b.(float64); ok {
			return va.(float64) > vb
		}
	case time.Time:
		if vb, ok := b.(time.Time); ok {
			return va.After(vb)
		}
	}
	return false
}

func compareLess(a, b interface{}) bool {
	// Similar to compareGreater but opposite
	switch va := a.(type) {
	case int, int32, int64:
		if vb, ok := b.(int); ok {
			return va.(int) < vb
		}
	case float32, float64:
		if vb, ok := b.(float64); ok {
			return va.(float64) < vb
		}
	case time.Time:
		if vb, ok := b.(time.Time); ok {
			return va.Before(vb)
		}
	}
	return false
}

func compareLike(value, pattern interface{}, caseSensitive bool) bool {
	valStr := fmt.Sprintf("%v", value)
	patStr := fmt.Sprintf("%v", pattern)
	
	if !caseSensitive {
		valStr = strings.ToLower(valStr)
		patStr = strings.ToLower(patStr)
	}
	
	// Convert SQL LIKE pattern to regex
	// % -> .*
	// _ -> .
	patStr = strings.ReplaceAll(patStr, "%", ".*")
	patStr = strings.ReplaceAll(patStr, "_", ".")
	patStr = "^" + patStr + "$"
	
	matched, _ := regexp.MatchString(patStr, valStr)
	return matched
}

func compareIn(value, list interface{}) bool {
	// Check if value is in a list
	if values, ok := list.([]interface{}); ok {
		for _, v := range values {
			if compareEqual(value, v, false) {
				return true
			}
		}
	}
	return false
}

// rangeSearch demonstrates numeric range searching
func rangeSearch() {
	fmt.Println("\n=== Numeric Range Search ===")

	// Find all records with balance between 1000 and 5000
	fmt.Println("Searching for balance between 1000 and 5000...")
	
	// Implementation would search numeric fields
	// Example: WHERE balance >= 1000 AND balance <= 5000
}

// dateSearch demonstrates date range searching
func dateSearch() {
	fmt.Println("\n=== Date Range Search ===")

	// Find records from last 30 days
	startDate := time.Now().AddDate(0, 0, -30)
	endDate := time.Now()

	fmt.Printf("Searching for dates between %s and %s\n",
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"))

	// Implementation would search date fields
	// Example: WHERE created >= startDate AND created <= endDate
}

// complexSearch demonstrates multi-criteria searching
func complexSearch() {
	fmt.Println("\n=== Complex Multi-Criteria Search ===")

	// Complex query example:
	// Find active customers with balance > 1000 
	// AND (name contains "Corp" OR city = "New York")
	
	criteria := []SearchCriteria{
		{Field: "ACTIVE", Value: true, Operator: "="},
		{Field: "BALANCE", Value: 1000.0, Operator: ">"},
		// Additional OR conditions would be handled separately
	}

	fmt.Println("Complex search criteria:")
	for _, c := range criteria {
		fmt.Printf("  %s %s %v\n", c.Field, c.Operator, c.Value)
	}

	// Implementation would combine multiple criteria
}

// regexSearch demonstrates pattern matching with regex
func regexSearch() {
	fmt.Println("\n=== Regex Pattern Search ===")

	patterns := []string{
		`^\d{3}-\d{3}-\d{4}$`, // Phone number
		`^[A-Z]{2}\d{4}$`,      // Product code
		`@[a-z]+\.com$`,        // Email domain
	}

	for _, pattern := range patterns {
		fmt.Printf("Searching with pattern: %s\n", pattern)
		// Implementation would use regex matching
	}
}

// extractRowData converts a row to a map
func extractRowData(row *dbase.Row) map[string]interface{} {
	data := make(map[string]interface{})
	
	for i := 0; i < row.FieldCount(); i++ {
		field := row.Field(i)
		value, _ := field.GetValue()
		data[field.Name()] = value
	}
	
	return data
}

// Utility function for building SQL-like queries
func buildSQLQuery(criteria []SearchCriteria) string {
	var conditions []string
	
	for _, c := range criteria {
		condition := fmt.Sprintf("%s %s %v", c.Field, c.Operator, c.Value)
		conditions = append(conditions, condition)
	}
	
	return "WHERE " + strings.Join(conditions, " AND ")
}

// Performance tips for searching:
// 1. Use specific field searches when possible
// 2. Index frequently searched fields (if supported)
// 3. Limit result set size with pagination
// 4. Cache search results for repeated queries
// 5. Use compiled regex for pattern matching
// 6. Consider parallel search for large files
// 7. Skip deleted records early in the loop