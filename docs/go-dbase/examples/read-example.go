// Example: Reading DBF Files with go-dbase
// Based on https://github.com/Valentin-Kaiser/go-dbase/examples

package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Valentin-Kaiser/go-dbase/dbase"
)

// Product struct with dbase field mappings
type Product struct {
	// The dbase tag contains the table name and column name separated by a dot
	// Column names are case insensitive
	ID          int32     `dbase:"TEST.PRODUCTID"`
	Name        string    `dbase:"TEST.PRODNAME"`
	Price       float64   `dbase:"TEST.PRICE"`
	Double      float64   `dbase:"TEST.DOUBLE"`
	Date        time.Time `dbase:"TEST.DATE"`
	DateTime    time.Time `dbase:"TEST.DATETIME"`
	Integer     int32     `dbase:"TEST.INTEGER"`
	Float       float64   `dbase:"TEST.FLOAT"`
	Active      bool      `dbase:"TEST.ACTIVE"`
	Description string    `dbase:"TEST.DESC"`
	Tax         float64   `dbase:"TEST.TAX"`
	Stock       int64     `dbase:"TEST.INSTOCK"`
	Blob        []byte    `dbase:"TEST.BLOB"`
	Varbinary   []byte    `dbase:"TEST.VARBIN_NIL"`
	Varchar     string    `dbase:"TEST.VAR_NIL"`
	Var         string    `dbase:"TEST.VAR"`
}

func main() {
	// Enable debug logging (optional)
	f, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	// Enable debug mode to see what's happening
	dbase.Debug(true, io.MultiWriter(os.Stdout, f))

	// Open the DBF table
	table, err := dbase.OpenTable(&dbase.Config{
		Filename:   "path/to/your/file.dbf",
		TrimSpaces: true, // Automatically trim spaces from character fields
	})
	if err != nil {
		panic(err)
	}
	defer table.Close()

	// Example 1: Read all records at once
	fmt.Println("=== Reading All Records ===")
	readAllExample(table)

	// Reset to beginning
	table.GoTo(0)

	// Example 2: Iterate through records
	fmt.Println("\n=== Iterating Through Records ===")
	iterateExample(table)

	// Reset to beginning
	table.GoTo(0)

	// Example 3: Read into struct
	fmt.Println("\n=== Reading Into Struct ===")
	readStructExample(table)

	// Example 4: Access specific fields
	table.GoTo(0)
	fmt.Println("\n=== Accessing Specific Fields ===")
	accessFieldsExample(table)
}

// Example 1: Read all records at once
func readAllExample(table *dbase.Table) {
	rows, err := table.ReadAll()
	if err != nil {
		fmt.Printf("Error reading all: %v\n", err)
		return
	}

	fmt.Printf("Total records: %d\n", len(rows))
	
	// Print first 3 records
	for i, row := range rows {
		if i >= 3 {
			break
		}
		fmt.Printf("Record %d: %v\n", i+1, row)
	}
}

// Example 2: Iterate through records (memory efficient)
func iterateExample(table *dbase.Table) {
	recordCount := 0
	deletedCount := 0

	for table.Next() {
		// Check if record is deleted
		if table.Deleted() {
			deletedCount++
			continue
		}

		row := table.Row()
		recordCount++

		// Print first 3 non-deleted records
		if recordCount <= 3 {
			fmt.Printf("Record %d:\n", recordCount)
			
			// Access fields by position
			for i := 0; i < row.FieldCount(); i++ {
				field := row.Field(i)
				value, _ := field.GetValue()
				fmt.Printf("  %s: %v\n", field.Name(), value)
			}
		}
	}

	if err := table.Err(); err != nil {
		fmt.Printf("Error during iteration: %v\n", err)
	}

	fmt.Printf("Active records: %d, Deleted records: %d\n", recordCount, deletedCount)
}

// Example 3: Read into struct
func readStructExample(table *dbase.Table) {
	// Read first 3 records into Product structs
	products := make([]Product, 0, 3)
	count := 0

	for table.Next() && count < 3 {
		if table.Deleted() {
			continue
		}

		var product Product
		err := table.Scan(&product)
		if err != nil {
			fmt.Printf("Error scanning into struct: %v\n", err)
			continue
		}

		products = append(products, product)
		count++
	}

	// Display products
	for i, p := range products {
		fmt.Printf("Product %d:\n", i+1)
		fmt.Printf("  ID: %d\n", p.ID)
		fmt.Printf("  Name: %s\n", p.Name)
		fmt.Printf("  Price: %.2f\n", p.Price)
		fmt.Printf("  Active: %v\n", p.Active)
		fmt.Printf("  Date: %s\n", p.Date.Format("2006-01-02"))
	}
}

// Example 4: Access specific fields by name
func accessFieldsExample(table *dbase.Table) {
	fmt.Println("Header Information:")
	header := table.Header()
	fmt.Printf("  Records: %d\n", header.RecordsCount())
	fmt.Printf("  Fields: %d\n", header.FieldsCount())
	fmt.Printf("  Modified: %s\n", header.Modified().Format("2006-01-02"))

	fmt.Println("\nColumn Information:")
	for _, col := range header.Columns() {
		fmt.Printf("  %s: Type=%c, Length=%d\n", 
			col.Name(), col.Type(), col.Length)
	}

	// Read specific fields from first record
	if table.Next() && !table.Deleted() {
		row := table.Row()
		
		fmt.Println("\nFirst Record Fields:")
		
		// Access by field name
		if field := row.FieldByName("NAME"); field != nil {
			name, _ := field.GetValue()
			fmt.Printf("  Name: %v\n", name)
		}

		// Access numeric field
		if field := row.FieldByName("PRICE"); field != nil {
			price := field.Float()
			fmt.Printf("  Price: %.2f\n", price)
		}

		// Access date field
		if field := row.FieldByName("DATE"); field != nil {
			date := field.Time()
			fmt.Printf("  Date: %s\n", date.Format("2006-01-02"))
		}

		// Access logical field
		if field := row.FieldByName("ACTIVE"); field != nil {
			active := field.Bool()
			fmt.Printf("  Active: %v\n", active)
		}
	}
}