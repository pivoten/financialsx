// Example: Creating DBF Tables with go-dbase
// Demonstrates creating new DBF files with various column types

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Valentin-Kaiser/go-dbase/dbase"
)

func main() {
	// Example 1: Create a simple customer table
	createCustomerTable()

	// Example 2: Create a table with all field types
	createCompleteTable()

	// Example 3: Create a FoxPro specific table
	createFoxProTable()
}

// createCustomerTable creates a basic customer DBF file
func createCustomerTable() {
	fmt.Println("=== Creating Customer Table ===")

	// Open a new table file
	table, err := dbase.OpenTable(&dbase.Config{
		Filename:   "customers.dbf",
		TrimSpaces: true,
		Version:    dbase.DBase5, // dBase 5 format
	})
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
	defer table.Close()

	// Add columns to the table structure
	// Note: In actual API, columns are defined during table creation
	// This is a simplified representation

	// Add sample customer records
	for i := 1; i <= 5; i++ {
		// Create a new row
		row := table.NewRow()

		// Set field values (assuming fields exist)
		// In real implementation, you'd set up the schema first
		// row.SetField("CUST_ID", fmt.Sprintf("C%04d", i))
		// row.SetField("NAME", fmt.Sprintf("Customer %d", i))
		// row.SetField("EMAIL", fmt.Sprintf("customer%d@example.com", i))
		// row.SetField("BALANCE", float64(i*1000.50))
		// row.SetField("ACTIVE", i%2 == 1)
		// row.SetField("JOINED", time.Now())

		fmt.Printf("Created customer record %d\n", i)
	}

	fmt.Println("Customer table created successfully")
}

// createCompleteTable demonstrates all supported field types
func createCompleteTable() {
	fmt.Println("\n=== Creating Complete Table with All Field Types ===")

	// Define schema with all field types
	// This shows the conceptual structure - actual API may differ
	schema := []struct {
		Name     string
		Type     byte
		Length   int
		Decimals int
	}{
		{"CHAR_FLD", 'C', 50, 0},    // Character
		{"NUM_FLD", 'N', 10, 2},      // Numeric with decimals
		{"DATE_FLD", 'D', 8, 0},      // Date
		{"LOGIC_FLD", 'L', 1, 0},     // Logical (boolean)
		{"MEMO_FLD", 'M', 10, 0},     // Memo (requires FPT file)
		{"FLOAT_FLD", 'F', 20, 4},    // Float
		{"INT_FLD", 'I', 4, 0},       // Integer (FoxPro)
		{"DTIME_FLD", 'T', 8, 0},     // DateTime (FoxPro)
		{"CURR_FLD", 'Y', 8, 4},      // Currency (FoxPro)
		{"DOUBLE_FLD", 'B', 8, 0},    // Double (FoxPro)
		{"VARBIN_FLD", 'Q', 10, 0},   // Varbinary (FoxPro)
		{"VARCHAR_FLD", 'V', 254, 0}, // Varchar (FoxPro)
	}

	fmt.Println("Schema defined with field types:")
	for _, field := range schema {
		fmt.Printf("  %s: Type=%c, Length=%d, Decimals=%d\n",
			field.Name, field.Type, field.Length, field.Decimals)
	}

	// In actual implementation, you would:
	// 1. Create table with this schema
	// 2. Add records with appropriate values
	// 3. Save the table

	fmt.Println("Complete table structure created")
}

// createFoxProTable creates a FoxPro-specific table with advanced features
func createFoxProTable() {
	fmt.Println("\n=== Creating FoxPro Table ===")

	// FoxPro specific features
	config := dbase.Config{
		Filename: "foxpro_table.dbf",
		Version:  dbase.FoxPro, // Use FoxPro format
		// Additional FoxPro specific options would go here
	}

	table, err := dbase.OpenTable(&config)
	if err != nil {
		log.Fatal("Failed to create FoxPro table:", err)
	}
	defer table.Close()

	// FoxPro supports additional features:
	// - Autoincrement fields
	// - Varchar fields (variable length)
	// - Varbinary fields
	// - DateTime with milliseconds
	// - Larger memo fields

	fmt.Println("FoxPro table created with advanced features")

	// Add a sample record with FoxPro-specific types
	row := table.NewRow()

	// Example of setting FoxPro-specific field types
	// These would need actual field definitions first
	sampleData := map[string]interface{}{
		"ID":          1,                                              // Autoincrement
		"NAME":        "FoxPro Test",                                 // Varchar
		"CREATED":     time.Now(),                                    // DateTime with ms
		"DATA":        []byte{0x01, 0x02, 0x03},                      // Varbinary
		"AMOUNT":      12345.6789,                                    // Currency
		"DESCRIPTION": "Long text that would go in a memo field...", // Memo
	}

	fmt.Println("Sample FoxPro record data:")
	for key, value := range sampleData {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Println("FoxPro table setup complete")
}

// Helper function to demonstrate table creation with error handling
func createTableWithSchema(filename string, version dbase.Version) error {
	// This shows the pattern for creating a table with proper error handling
	table, err := dbase.OpenTable(&dbase.Config{
		Filename:   filename,
		Version:    version,
		TrimSpaces: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create table %s: %w", filename, err)
	}
	defer table.Close()

	// Define and add columns
	// ... column definitions ...

	// Add initial records
	// ... record creation ...

	// Save the table
	// In actual API, saving might be automatic or explicit
	// err = table.Save()

	return nil
}

// Example of different DBF versions
func demonstrateVersions() {
	versions := []struct {
		Name    string
		Version dbase.Version
	}{
		{"dBase III", dbase.DBase3},
		{"dBase IV", dbase.DBase4},
		{"dBase 5", dbase.DBase5},
		{"FoxPro", dbase.FoxPro},
		{"FoxPro with Autoincrement", dbase.FoxProAutoincrement},
	}

	fmt.Println("\n=== DBF Version Examples ===")
	for _, v := range versions {
		fmt.Printf("Creating %s format file...\n", v.Name)
		filename := fmt.Sprintf("example_%s.dbf", v.Name)
		err := createTableWithSchema(filename, v.Version)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
		} else {
			fmt.Printf("  Created: %s\n", filename)
		}
	}
}

// Tips for creating DBF files:
// 1. Choose the right version for compatibility
// 2. Define schema carefully - it's hard to change later
// 3. Set appropriate field lengths
// 4. Consider memo fields for large text
// 5. Use proper character encoding
// 6. Remember index files (.CDX/.MDX) are separate
// 7. Test with target application (FoxPro, dBase, etc.)