// Practical Example: How FoxPro Toolkit Uses go-dbase
// This shows the actual implementation patterns used in our codebase

package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Valentin-Kaiser/go-dbase/dbase"
)

// ColumnInfo represents DBF column metadata
type ColumnInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Length   int    `json:"length"`
	Decimals int    `json:"decimals"`
}

// DBFStats provides statistics about the DBF file
type DBFStats struct {
	TotalRecords   int    `json:"totalRecords"`
	DeletedRecords int    `json:"deletedRecords"`
	LoadedRecords  int    `json:"loadedRecords"`
	HasMoreRecords bool   `json:"hasMoreRecords"`
	SearchTerm     string `json:"searchTerm,omitempty"`
}

// DBFResult contains the complete result of a DBF read operation
type DBFResult struct {
	Columns []ColumnInfo    `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Stats   DBFStats        `json:"stats"`
}

// DBFReader handles all DBF file operations
type DBFReader struct {
	table *dbase.File
}

// NewDBFReader creates a new DBF reader instance
func NewDBFReader() *DBFReader {
	return &DBFReader{}
}

// ReadFile reads a DBF file with pagination and optional search
func (r *DBFReader) ReadFile(filePath string, offset, limit int, searchTerm string) (*DBFResult, error) {
	// Open the DBF file
	table, err := dbase.OpenTable(&dbase.Config{
		Filename:   filePath,
		TrimSpaces: true,
		ReadOnly:   true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open DBF file: %w", err)
	}
	defer table.Close()

	// Get column information
	columns := r.extractColumns(table)

	// Get statistics
	stats := DBFStats{
		TotalRecords: int(table.Header().RecordsCount()),
		SearchTerm:   searchTerm,
	}

	// Read records with pagination and search
	rows, deletedCount, hasMore, err := r.readRecords(table, offset, limit, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("failed to read records: %w", err)
	}

	stats.DeletedRecords = deletedCount
	stats.LoadedRecords = len(rows)
	stats.HasMoreRecords = hasMore

	return &DBFResult{
		Columns: columns,
		Rows:    rows,
		Stats:   stats,
	}, nil
}

// extractColumns gets column metadata from the table
func (r *DBFReader) extractColumns(table *dbase.File) []ColumnInfo {
	header := table.Header()
	columns := make([]ColumnInfo, 0, len(header.Columns()))

	for _, col := range header.Columns() {
		columns = append(columns, ColumnInfo{
			Name:     col.Name(),
			Type:     string(col.Type()),
			Length:   int(col.Length),
			Decimals: int(col.Decimals()),
		})
	}

	return columns
}

// readRecords reads records with pagination and search
func (r *DBFReader) readRecords(table *dbase.File, offset, limit int, searchTerm string) ([][]interface{}, int, bool, error) {
	rows := make([][]interface{}, 0, limit)
	currentPos := 0
	deletedCount := 0
	hasMore := false
	searchLower := strings.ToLower(searchTerm)

	for {
		// Get next row
		row, err := table.Next()
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				break
			}
			return nil, 0, false, err
		}

		// Check if deleted
		deleted, err := table.Deleted()
		if err != nil {
			// Log error but continue
			fmt.Printf("Warning: error checking deleted status: %v\n", err)
		}

		if deleted {
			deletedCount++
			continue
		}

		// Apply search filter if provided
		if searchTerm != "" {
			if !r.matchesSearch(row, searchLower) {
				continue
			}
		}

		// Handle pagination
		if currentPos < offset {
			currentPos++
			continue
		}

		if len(rows) >= limit {
			hasMore = true
			break
		}

		// Extract row data
		rowData := r.extractRowData(row)
		rows = append(rows, rowData)
		currentPos++
	}

	return rows, deletedCount, hasMore, nil
}

// matchesSearch checks if a row matches the search term
func (r *DBFReader) matchesSearch(row *dbase.Row, searchTerm string) bool {
	for i := 0; i < row.FieldCount(); i++ {
		field := row.Field(i)
		value, err := field.GetValue()
		if err != nil {
			continue
		}

		// Convert to string and check for match
		strValue := fmt.Sprintf("%v", value)
		if strings.Contains(strings.ToLower(strValue), searchTerm) {
			return true
		}
	}
	return false
}

// extractRowData converts a dbase.Row to a slice of values
func (r *DBFReader) extractRowData(row *dbase.Row) []interface{} {
	data := make([]interface{}, row.FieldCount())

	for i := 0; i < row.FieldCount(); i++ {
		field := row.Field(i)
		value, err := field.GetValue()
		if err != nil {
			data[i] = nil
			continue
		}

		// Format special types
		data[i] = r.formatValue(value, field.Type())
	}

	return data
}

// formatValue formats a value based on its type
func (r *DBFReader) formatValue(value interface{}, fieldType byte) interface{} {
	switch fieldType {
	case 'D': // Date
		if t, ok := value.(time.Time); ok {
			return t.Format("2006-01-02")
		}
	case 'T': // DateTime
		if t, ok := value.(time.Time); ok {
			return t.Format("2006-01-02 15:04:05")
		}
	case 'L': // Logical
		if b, ok := value.(bool); ok {
			if b {
				return "True"
			}
			return "False"
		}
	case 'M': // Memo
		if value == nil {
			return "[MEMO]"
		}
	}
	return value
}

// ExportToJSON exports a DBF file to JSON format
func (r *DBFReader) ExportToJSON(filePath string) (string, error) {
	// Read all data
	result, err := r.ReadFile(filePath, 0, 999999, "")
	if err != nil {
		return "", err
	}

	// Convert rows to maps for better JSON structure
	records := make([]map[string]interface{}, len(result.Rows))
	for i, row := range result.Rows {
		record := make(map[string]interface{})
		for j, col := range result.Columns {
			if j < len(row) {
				record[col.Name] = row[j]
			}
		}
		records[i] = record
	}

	// Create output structure
	output := map[string]interface{}{
		"metadata": map[string]interface{}{
			"source":   filePath,
			"exported": time.Now().Format(time.RFC3339),
			"records":  result.Stats.TotalRecords,
			"deleted":  result.Stats.DeletedRecords,
		},
		"columns": result.Columns,
		"records": records,
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}

	// Write to file
	outputPath := strings.TrimSuffix(filePath, ".dbf") + "_export.json"
	err = os.WriteFile(outputPath, jsonData, 0644)
	if err != nil {
		return "", err
	}

	return outputPath, nil
}

// ExportToCSV exports a DBF file to CSV format
func (r *DBFReader) ExportToCSV(filePath string) (string, error) {
	// Read all data
	result, err := r.ReadFile(filePath, 0, 999999, "")
	if err != nil {
		return "", err
	}

	// Create output file
	outputPath := strings.TrimSuffix(filePath, ".dbf") + "_export.csv"
	file, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	headers := make([]string, len(result.Columns))
	for i, col := range result.Columns {
		headers[i] = col.Name
	}
	if err := writer.Write(headers); err != nil {
		return "", err
	}

	// Write data rows
	for _, row := range result.Rows {
		record := make([]string, len(row))
		for i, value := range row {
			if value == nil {
				record[i] = ""
			} else {
				record[i] = fmt.Sprintf("%v", value)
			}
		}
		if err := writer.Write(record); err != nil {
			return "", err
		}
	}

	return outputPath, nil
}

// GetFileInfo returns metadata about a DBF file without reading all records
func (r *DBFReader) GetFileInfo(filePath string) (*DBFResult, error) {
	table, err := dbase.OpenTable(&dbase.Config{
		Filename:   filePath,
		TrimSpaces: true,
		ReadOnly:   true,
	})
	if err != nil {
		return nil, err
	}
	defer table.Close()

	// Get column information
	columns := r.extractColumns(table)

	// Count deleted records
	deletedCount := 0
	for {
		_, err := table.Next()
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				break
			}
			return nil, err
		}

		deleted, _ := table.Deleted()
		if deleted {
			deletedCount++
		}
	}

	return &DBFResult{
		Columns: columns,
		Rows:    [][]interface{}{}, // Empty rows for info only
		Stats: DBFStats{
			TotalRecords:   int(table.Header().RecordsCount()),
			DeletedRecords: deletedCount,
		},
	}, nil
}

// Example usage
func main() {
	reader := NewDBFReader()

	// Example 1: Read with pagination
	fmt.Println("=== Reading with Pagination ===")
	result, err := reader.ReadFile("customers.dbf", 0, 10, "")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Loaded %d of %d records\n", result.Stats.LoadedRecords, result.Stats.TotalRecords)
	fmt.Printf("Deleted records: %d\n", result.Stats.DeletedRecords)
	fmt.Printf("Has more records: %v\n", result.Stats.HasMoreRecords)

	// Display columns
	fmt.Println("\nColumns:")
	for _, col := range result.Columns {
		fmt.Printf("  %s (%s, %d)\n", col.Name, col.Type, col.Length)
	}

	// Display first few rows
	fmt.Println("\nFirst few records:")
	for i, row := range result.Rows {
		if i >= 3 {
			break
		}
		fmt.Printf("Record %d: %v\n", i+1, row)
	}

	// Example 2: Search
	fmt.Println("\n=== Searching for 'Smith' ===")
	searchResult, err := reader.ReadFile("customers.dbf", 0, 100, "Smith")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Found %d records matching 'Smith'\n", searchResult.Stats.LoadedRecords)

	// Example 3: Export to JSON
	fmt.Println("\n=== Exporting to JSON ===")
	jsonPath, err := reader.ExportToJSON("customers.dbf")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Exported to: %s\n", jsonPath)

	// Example 4: Export to CSV
	fmt.Println("\n=== Exporting to CSV ===")
	csvPath, err := reader.ExportToCSV("customers.dbf")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Exported to: %s\n", csvPath)

	// Example 5: Get file info
	fmt.Println("\n=== Getting File Info ===")
	info, err := reader.GetFileInfo("customers.dbf")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Total columns: %d\n", len(info.Columns))
	fmt.Printf("Total records: %d (including %d deleted)\n",
		info.Stats.TotalRecords, info.Stats.DeletedRecords)
}