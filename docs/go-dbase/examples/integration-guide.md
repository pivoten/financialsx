# go-dbase Integration Guide

This guide shows how the FoxPro Toolkit integrates and uses the go-dbase library for DBF file operations.

## Library Overview

The [go-dbase](https://github.com/Valentin-Kaiser/go-dbase) library provides:
- Read and write support for dBase/FoxPro DBF files
- Memo field support (.FPT/.DBT files)
- Multiple character encodings
- Thread-safe operations
- Efficient memory usage with streaming

## How We Use go-dbase in FoxPro Toolkit

### 1. Opening DBF Files

In `internal/dbf/reader.go`:

```go
import "github.com/Valentin-Kaiser/go-dbase/dbase"

func (r *Reader) openTable(filePath string) (*dbase.File, error) {
    // Open with read-only and auto-trim spaces
    table, err := dbase.OpenTable(&dbase.Config{
        Filename:   filePath,
        TrimSpaces: true,
        ReadOnly:   true,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to open DBF: %w", err)
    }
    return table, nil
}
```

### 2. Reading Column Information

```go
func (r *Reader) getColumns(table *dbase.File) []ColumnInfo {
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
```

### 3. Reading Records with Pagination

```go
func (r *Reader) readWithPagination(table *dbase.File, offset, limit int) ([][]interface{}, error) {
    rows := make([][]interface{}, 0, limit)
    currentPos := 0
    recordCount := 0
    
    // Iterate through records
    for {
        row, err := table.Next()
        if err != nil {
            if err.Error() == "EOF" {
                break // End of file
            }
            return nil, err
        }
        
        // Check if deleted
        deleted, err := table.Deleted()
        if err != nil {
            return nil, err
        }
        
        if deleted {
            continue // Skip deleted records
        }
        
        // Handle pagination
        if currentPos < offset {
            currentPos++
            continue
        }
        
        if recordCount >= limit {
            break
        }
        
        // Extract row data
        rowData := r.extractRowData(row)
        rows = append(rows, rowData)
        recordCount++
    }
    
    return rows, nil
}
```

### 4. Extracting Row Data

```go
func (r *Reader) extractRowData(row *dbase.Row) []interface{} {
    data := make([]interface{}, row.FieldCount())
    
    for i := 0; i < row.FieldCount(); i++ {
        field := row.Field(i)
        
        // Get value based on field type
        value, err := field.GetValue()
        if err != nil {
            data[i] = nil
            continue
        }
        
        // Handle special formatting
        switch field.Type() {
        case 'D': // Date
            if t, ok := value.(time.Time); ok {
                data[i] = t.Format("2006-01-02")
            } else {
                data[i] = value
            }
        case 'T': // DateTime
            if t, ok := value.(time.Time); ok {
                data[i] = t.Format("2006-01-02 15:04:05")
            } else {
                data[i] = value
            }
        case 'L': // Logical
            if b, ok := value.(bool); ok {
                data[i] = b
            } else {
                data[i] = value
            }
        default:
            data[i] = value
        }
    }
    
    return data
}
```

### 5. Search Implementation

```go
func (r *Reader) searchRecords(table *dbase.File, searchTerm string) ([][]interface{}, error) {
    results := make([][]interface{}, 0)
    searchLower := strings.ToLower(searchTerm)
    
    for {
        row, err := table.Next()
        if err != nil {
            if err.Error() == "EOF" {
                break
            }
            return nil, err
        }
        
        deleted, _ := table.Deleted()
        if deleted {
            continue
        }
        
        // Search across all fields
        found := false
        for i := 0; i < row.FieldCount(); i++ {
            field := row.Field(i)
            value, err := field.GetValue()
            if err != nil {
                continue
            }
            
            // Convert to string and search
            strValue := fmt.Sprintf("%v", value)
            if strings.Contains(strings.ToLower(strValue), searchLower) {
                found = true
                break
            }
        }
        
        if found {
            rowData := r.extractRowData(row)
            results = append(results, rowData)
        }
    }
    
    return results, nil
}
```

### 6. Export Functionality

#### Export to JSON

```go
func (r *Reader) exportToJSON(filePath string) error {
    table, err := r.openTable(filePath)
    if err != nil {
        return err
    }
    defer table.Close()
    
    // Prepare JSON structure
    output := map[string]interface{}{
        "columns": r.getColumns(table),
        "records": []map[string]interface{}{},
    }
    
    records := []map[string]interface{}{}
    
    for {
        row, err := table.Next()
        if err != nil {
            if err.Error() == "EOF" {
                break
            }
            return err
        }
        
        deleted, _ := table.Deleted()
        if deleted {
            continue
        }
        
        // Convert row to map
        record := make(map[string]interface{})
        for i := 0; i < row.FieldCount(); i++ {
            field := row.Field(i)
            value, _ := field.GetValue()
            record[field.Name()] = value
        }
        
        records = append(records, record)
    }
    
    output["records"] = records
    
    // Write to file
    jsonData, err := json.MarshalIndent(output, "", "  ")
    if err != nil {
        return err
    }
    
    outputPath := strings.TrimSuffix(filePath, ".dbf") + ".json"
    return os.WriteFile(outputPath, jsonData, 0644)
}
```

#### Export to CSV

```go
func (r *Reader) exportToCSV(filePath string) error {
    table, err := r.openTable(filePath)
    if err != nil {
        return err
    }
    defer table.Close()
    
    outputPath := strings.TrimSuffix(filePath, ".dbf") + ".csv"
    file, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer file.Close()
    
    writer := csv.NewWriter(file)
    defer writer.Flush()
    
    // Write headers
    header := table.Header()
    headers := make([]string, len(header.Columns()))
    for i, col := range header.Columns() {
        headers[i] = col.Name()
    }
    writer.Write(headers)
    
    // Write data
    for {
        row, err := table.Next()
        if err != nil {
            if err.Error() == "EOF" {
                break
            }
            return err
        }
        
        deleted, _ := table.Deleted()
        if deleted {
            continue
        }
        
        // Convert row to string slice
        record := make([]string, row.FieldCount())
        for i := 0; i < row.FieldCount(); i++ {
            field := row.Field(i)
            value, _ := field.GetValue()
            record[i] = fmt.Sprintf("%v", value)
        }
        
        writer.Write(record)
    }
    
    return nil
}
```

## Complete Working Example

Here's a complete example showing how to use go-dbase in your own project:

```go
package main

import (
    "fmt"
    "log"
    "github.com/Valentin-Kaiser/go-dbase/dbase"
)

func main() {
    // Open DBF file
    table, err := dbase.OpenTable(&dbase.Config{
        Filename:   "customers.dbf",
        TrimSpaces: true,
        ReadOnly:   true,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer table.Close()
    
    // Print header info
    header := table.Header()
    fmt.Printf("Records: %d\n", header.RecordsCount())
    fmt.Printf("Modified: %s\n", header.Modified())
    
    // Print column info
    fmt.Println("\nColumns:")
    for _, col := range header.Columns() {
        fmt.Printf("  %s (%c, %d)\n", 
            col.Name(), col.Type(), col.Length)
    }
    
    // Read first 10 records
    fmt.Println("\nFirst 10 Records:")
    count := 0
    
    for count < 10 {
        row, err := table.Next()
        if err != nil {
            if err.Error() == "EOF" {
                break
            }
            log.Fatal(err)
        }
        
        // Check if deleted
        deleted, err := table.Deleted()
        if err != nil {
            log.Fatal(err)
        }
        
        if deleted {
            continue
        }
        
        fmt.Printf("Record %d:\n", count+1)
        
        // Print all fields
        for i := 0; i < row.FieldCount(); i++ {
            field := row.Field(i)
            value, _ := field.GetValue()
            fmt.Printf("  %s: %v\n", field.Name(), value)
        }
        
        count++
    }
}
```

## Error Handling Best Practices

```go
// Always check for EOF
row, err := table.Next()
if err != nil {
    if err.Error() == "EOF" {
        // Normal end of file
        return nil
    }
    // Actual error
    return fmt.Errorf("read error: %w", err)
}

// Handle deleted records
deleted, err := table.Deleted()
if err != nil {
    // Log but continue
    log.Printf("Error checking deleted status: %v", err)
}

// Safe field access
field := row.FieldByName("CUSTOMER_ID")
if field != nil {
    value, err := field.GetValue()
    if err != nil {
        // Handle missing or corrupt field
        value = nil
    }
}
```

## Performance Optimization

1. **Use Read-Only Mode**: When just reading, always use `ReadOnly: true`
2. **Enable TrimSpaces**: Automatically trim character fields with `TrimSpaces: true`
3. **Close Tables**: Always defer `table.Close()` to free resources
4. **Stream Large Files**: Process row-by-row instead of loading all into memory
5. **Skip Deleted Records**: Check `Deleted()` early to avoid processing

## Character Encoding

Handle different encodings:

```go
// Specify encoding
table, err := dbase.OpenTable(&dbase.Config{
    Filename: "data.dbf",
    Encoding: "Windows-1252", // Western European
})

// Common encodings for FoxPro files:
// - "UTF-8"
// - "Windows-1252" (Western European)
// - "Windows-1251" (Cyrillic)
// - "CP437" (DOS)
// - "CP850" (DOS Latin-1)
```

## Limitations to Consider

1. **No Write Support in Toolkit**: We only use read operations
2. **Memo Fields**: Limited search capability in memo fields
3. **Large Files**: Implement pagination for files > 10,000 records
4. **Concurrent Access**: One reader at a time per file
5. **Index Files**: .CDX/.MDX files are not used

## Testing with Sample Data

Create test DBF files:

```go
func createTestDBF() {
    // This would require write support
    // For testing, use existing DBF files or tools like:
    // - LibreOffice Calc
    // - DBF Viewer
    // - Original FoxPro application
}
```

## Debugging

Enable debug output:

```go
import "io"
import "os"

// Enable debug logging
dbase.Debug(true, os.Stdout)

// Or log to file
f, _ := os.Create("debug.log")
defer f.Close()
dbase.Debug(true, f)
```

## Resources

- [go-dbase Documentation](https://pkg.go.dev/github.com/Valentin-Kaiser/go-dbase)
- [DBF Format Specification](http://www.dbf2002.com/)
- [FoxPro Toolkit Source](https://github.com/pivoten/foxprotoolkit)

## Summary

The go-dbase library provides robust DBF file reading capabilities that power the FoxPro Toolkit. By following the patterns shown above, you can:

1. Read DBF files efficiently
2. Handle pagination for large datasets
3. Search across records
4. Export to modern formats
5. Properly handle deleted records and errors

The key is to always:
- Check for EOF when iterating
- Handle deleted records
- Close tables when done
- Use read-only mode for safety
- Implement proper error handling