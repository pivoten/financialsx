// Example: DBF Schema Inspection and Manipulation
// Demonstrates working with DBF table structure

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Valentin-Kaiser/go-dbase/dbase"
)

// SchemaAnalyzer provides schema inspection capabilities
type SchemaAnalyzer struct {
	filename string
	table    *dbase.File
}

// TableSchema represents complete table structure
type TableSchema struct {
	TableName    string
	Version      string
	Encoding     string
	Created      time.Time
	Modified     time.Time
	RecordCount  int
	RecordLength int
	Fields       []FieldSchema
	Statistics   SchemaStats
}

// FieldSchema represents field structure
type FieldSchema struct {
	Name         string
	Type         string
	TypeName     string
	Length       int
	Decimals     int
	Offset       int
	AllowNull    bool
	SystemField  bool
	Statistics   FieldStats
}

// SchemaStats provides table-level statistics
type SchemaStats struct {
	TotalFields      int
	CharacterFields  int
	NumericFields    int
	DateFields       int
	LogicalFields    int
	MemoFields       int
	OtherFields      int
	TotalDataSize    int64
	DeletedRecords   int
	ActiveRecords    int
}

// FieldStats provides field-level statistics
type FieldStats struct {
	NullCount     int
	UniqueValues  int
	MinValue      interface{}
	MaxValue      interface{}
	AvgLength     float64
	MostFrequent  interface{}
}

func main() {
	fmt.Println("=== DBF Schema Analysis ===\n")

	// Example 1: Basic schema inspection
	inspectSchema("sample.dbf")

	// Example 2: Field analysis
	analyzeFields("sample.dbf")

	// Example 3: Schema comparison
	compareSchemas("table1.dbf", "table2.dbf")

	// Example 4: Schema validation
	validateSchema("sample.dbf")

	// Example 5: Generate migration script
	generateMigration("old_schema.dbf", "new_schema.dbf")
}

// inspectSchema performs basic schema inspection
func inspectSchema(filename string) {
	fmt.Printf("=== Inspecting Schema: %s ===\n", filename)

	analyzer := NewSchemaAnalyzer(filename)
	schema, err := analyzer.Analyze()
	if err != nil {
		log.Printf("Error analyzing schema: %v", err)
		return
	}

	// Display table information
	fmt.Printf("\nTable: %s\n", schema.TableName)
	fmt.Printf("Version: %s\n", schema.Version)
	fmt.Printf("Encoding: %s\n", schema.Encoding)
	fmt.Printf("Modified: %s\n", schema.Modified.Format("2006-01-02 15:04:05"))
	fmt.Printf("Records: %d (Active: %d, Deleted: %d)\n",
		schema.RecordCount,
		schema.Statistics.ActiveRecords,
		schema.Statistics.DeletedRecords)
	fmt.Printf("Record Length: %d bytes\n", schema.RecordLength)

	// Display field summary
	fmt.Printf("\nField Summary:\n")
	fmt.Printf("  Total Fields: %d\n", schema.Statistics.TotalFields)
	fmt.Printf("  Character: %d\n", schema.Statistics.CharacterFields)
	fmt.Printf("  Numeric: %d\n", schema.Statistics.NumericFields)
	fmt.Printf("  Date: %d\n", schema.Statistics.DateFields)
	fmt.Printf("  Logical: %d\n", schema.Statistics.LogicalFields)
	fmt.Printf("  Memo: %d\n", schema.Statistics.MemoFields)

	// Display fields
	fmt.Printf("\nFields:\n")
	fmt.Printf("%-15s %-10s %-8s %-8s %-8s\n",
		"Name", "Type", "Length", "Decimals", "Offset")
	fmt.Println(strings.Repeat("-", 60))
	
	for _, field := range schema.Fields {
		fmt.Printf("%-15s %-10s %-8d %-8d %-8d\n",
			field.Name,
			field.TypeName,
			field.Length,
			field.Decimals,
			field.Offset)
	}
}

// NewSchemaAnalyzer creates a new schema analyzer
func NewSchemaAnalyzer(filename string) *SchemaAnalyzer {
	return &SchemaAnalyzer{
		filename: filename,
	}
}

// Analyze performs complete schema analysis
func (sa *SchemaAnalyzer) Analyze() (*TableSchema, error) {
	// Open table
	table, err := dbase.OpenTable(&dbase.Config{
		Filename:   sa.filename,
		TrimSpaces: true,
		ReadOnly:   true,
	})
	if err != nil {
		return nil, err
	}
	defer table.Close()

	sa.table = table
	header := table.Header()

	// Build schema
	schema := &TableSchema{
		TableName:    strings.TrimSuffix(sa.filename, ".dbf"),
		Version:      sa.getVersion(header),
		Encoding:     sa.getEncoding(header),
		Modified:     header.Modified(),
		RecordCount:  int(header.RecordsCount()),
		RecordLength: int(header.RecordLength()),
	}

	// Analyze fields
	schema.Fields = sa.analyzeFields(header)
	
	// Calculate statistics
	schema.Statistics = sa.calculateStats(schema.Fields)

	// Count deleted records
	deletedCount := 0
	for {
		_, err := table.Next()
		if err != nil {
			break
		}
		deleted, _ := table.Deleted()
		if deleted {
			deletedCount++
		}
	}
	schema.Statistics.DeletedRecords = deletedCount
	schema.Statistics.ActiveRecords = schema.RecordCount - deletedCount

	return schema, nil
}

func (sa *SchemaAnalyzer) getVersion(header *dbase.Header) string {
	// Determine DBF version based on header
	// This is simplified - actual version detection is more complex
	return "FoxPro 2.x"
}

func (sa *SchemaAnalyzer) getEncoding(header *dbase.Header) string {
	// Determine character encoding
	return "Windows-1252"
}

func (sa *SchemaAnalyzer) analyzeFields(header *dbase.Header) []FieldSchema {
	var fields []FieldSchema
	offset := 1 // First byte is deletion flag

	// Note: Actual column iteration would depend on the API
	// This is a conceptual example
	sampleFields := []struct {
		name     string
		typ      byte
		length   int
		decimals int
	}{
		{"ID", 'I', 4, 0},
		{"NAME", 'C', 50, 0},
		{"AMOUNT", 'N', 10, 2},
		{"DATE", 'D', 8, 0},
		{"ACTIVE", 'L', 1, 0},
	}

	for _, f := range sampleFields {
		field := FieldSchema{
			Name:     f.name,
			Type:     string(f.typ),
			TypeName: getFieldTypeName(f.typ),
			Length:   f.length,
			Decimals: f.decimals,
			Offset:   offset,
		}
		
		offset += f.length
		fields = append(fields, field)
	}

	return fields
}

func getFieldTypeName(fieldType byte) string {
	names := map[byte]string{
		'C': "Character",
		'N': "Numeric",
		'F': "Float",
		'D': "Date",
		'L': "Logical",
		'M': "Memo",
		'G': "General",
		'P': "Picture",
		'T': "DateTime",
		'I': "Integer",
		'Y': "Currency",
		'B': "Double",
		'V': "Varchar",
		'Q': "Varbinary",
	}
	
	if name, ok := names[fieldType]; ok {
		return name
	}
	return "Unknown"
}

func (sa *SchemaAnalyzer) calculateStats(fields []FieldSchema) SchemaStats {
	stats := SchemaStats{
		TotalFields: len(fields),
	}

	for _, field := range fields {
		switch field.Type {
		case "C", "V":
			stats.CharacterFields++
		case "N", "F", "I", "Y", "B":
			stats.NumericFields++
		case "D", "T":
			stats.DateFields++
		case "L":
			stats.LogicalFields++
		case "M", "G", "P":
			stats.MemoFields++
		default:
			stats.OtherFields++
		}
	}

	return stats
}

// analyzeFields performs detailed field analysis
func analyzeFields(filename string) {
	fmt.Printf("\n=== Field Analysis: %s ===\n", filename)

	table, err := dbase.OpenTable(&dbase.Config{
		Filename:   filename,
		TrimSpaces: true,
		ReadOnly:   true,
	})
	if err != nil {
		log.Printf("Error opening table: %v", err)
		return
	}
	defer table.Close()

	// Analyze each field's data
	fieldStats := make(map[string]*FieldStats)
	
	// Initialize field statistics
	// ... (would iterate through fields)

	// Scan all records to gather statistics
	recordCount := 0
	for {
		row, err := table.Next()
		if err != nil {
			break
		}

		deleted, _ := table.Deleted()
		if deleted {
			continue
		}

		recordCount++
		
		// Update field statistics
		// ... (would analyze each field value)
	}

	// Display field statistics
	fmt.Printf("\nField Statistics (%d records analyzed):\n", recordCount)
	// ... (would display the collected statistics)
}

// compareSchemas compares two table schemas
func compareSchemas(file1, file2 string) {
	fmt.Printf("\n=== Schema Comparison ===\n")
	fmt.Printf("Comparing: %s vs %s\n", file1, file2)

	analyzer1 := NewSchemaAnalyzer(file1)
	schema1, err1 := analyzer1.Analyze()
	
	analyzer2 := NewSchemaAnalyzer(file2)
	schema2, err2 := analyzer2.Analyze()

	if err1 != nil || err2 != nil {
		fmt.Println("Error loading schemas")
		return
	}

	// Compare basic properties
	fmt.Printf("\nTable Properties:\n")
	fmt.Printf("  Records: %d vs %d\n", schema1.RecordCount, schema2.RecordCount)
	fmt.Printf("  Fields: %d vs %d\n", len(schema1.Fields), len(schema2.Fields))

	// Compare fields
	fmt.Printf("\nField Differences:\n")
	
	// Find fields only in schema1
	for _, f1 := range schema1.Fields {
		found := false
		for _, f2 := range schema2.Fields {
			if f1.Name == f2.Name {
				found = true
				// Compare field properties
				if f1.Type != f2.Type || f1.Length != f2.Length {
					fmt.Printf("  Modified: %s (%s%d -> %s%d)\n",
						f1.Name, f1.Type, f1.Length, f2.Type, f2.Length)
				}
				break
			}
		}
		if !found {
			fmt.Printf("  Removed: %s\n", f1.Name)
		}
	}

	// Find fields only in schema2
	for _, f2 := range schema2.Fields {
		found := false
		for _, f1 := range schema1.Fields {
			if f1.Name == f2.Name {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("  Added: %s (%s%d)\n", f2.Name, f2.Type, f2.Length)
		}
	}
}

// validateSchema checks schema for issues
func validateSchema(filename string) {
	fmt.Printf("\n=== Schema Validation: %s ===\n", filename)

	analyzer := NewSchemaAnalyzer(filename)
	schema, err := analyzer.Analyze()
	if err != nil {
		log.Printf("Error analyzing schema: %v", err)
		return
	}

	issues := []string{}

	// Check for common issues
	for _, field := range schema.Fields {
		// Check field name length
		if len(field.Name) > 10 {
			issues = append(issues, fmt.Sprintf("Field name too long: %s", field.Name))
		}

		// Check for reserved names
		reserved := []string{"DATE", "TIME", "USER", "DELETE"}
		for _, r := range reserved {
			if strings.ToUpper(field.Name) == r {
				issues = append(issues, fmt.Sprintf("Reserved field name: %s", field.Name))
			}
		}

		// Check character field length
		if field.Type == "C" && field.Length > 254 {
			issues = append(issues, fmt.Sprintf("Character field too long: %s (%d)", field.Name, field.Length))
		}

		// Check numeric precision
		if field.Type == "N" && field.Decimals >= field.Length {
			issues = append(issues, fmt.Sprintf("Invalid numeric precision: %s", field.Name))
		}
	}

	// Check table-level issues
	if schema.RecordLength > 4000 {
		issues = append(issues, fmt.Sprintf("Record length very large: %d bytes", schema.RecordLength))
	}

	if len(schema.Fields) > 255 {
		issues = append(issues, "Too many fields (>255)")
	}

	// Display results
	if len(issues) == 0 {
		fmt.Println("✓ No schema issues found")
	} else {
		fmt.Printf("Found %d issues:\n", len(issues))
		for i, issue := range issues {
			fmt.Printf("  %d. %s\n", i+1, issue)
		}
	}
}

// generateMigration creates migration script between schemas
func generateMigration(oldFile, newFile string) {
	fmt.Printf("\n=== Migration Script Generation ===\n")
	fmt.Printf("From: %s\n", oldFile)
	fmt.Printf("To: %s\n", newFile)

	// This would analyze differences and generate:
	// 1. ALTER TABLE statements for field changes
	// 2. Data type conversions
	// 3. Default value assignments
	// 4. Index recreations

	fmt.Println("\n-- Migration Script --")
	fmt.Println("-- Add new fields")
	fmt.Println("ALTER TABLE customers ADD COLUMN email C(100);")
	fmt.Println("\n-- Modify existing fields")
	fmt.Println("-- Note: DBF doesn't support ALTER COLUMN directly")
	fmt.Println("-- Would need to create new table and copy data")
	fmt.Println("\n-- Create indexes")
	fmt.Println("INDEX ON email TAG email_idx")
}

// Tips for schema management:
// 1. Always backup before schema changes
// 2. Test migrations on copies first
// 3. Document all schema changes
// 4. Consider field order for performance
// 5. Use appropriate field types and sizes
// 6. Avoid reserved words as field names
// 7. Plan for future growth in field sizes