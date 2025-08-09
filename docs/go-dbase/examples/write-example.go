// Example: Writing DBF Files with go-dbase
// Based on https://github.com/Valentin-Kaiser/go-dbase/examples

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Valentin-Kaiser/go-dbase/dbase"
)

func main() {
	// Example 1: Create a new DBF file
	fmt.Println("=== Creating New DBF File ===")
	createNewTable()

	// Example 2: Modify existing DBF file
	fmt.Println("\n=== Modifying Existing DBF File ===")
	modifyExistingTable()

	// Example 3: Add records to existing table
	fmt.Println("\n=== Adding Records to Table ===")
	addRecordsToTable()

	// Example 4: Mark records as deleted
	fmt.Println("\n=== Marking Records as Deleted ===")
	deleteRecords()
}

// Example 1: Create a new DBF file with schema
func createNewTable() {
	// Define the table schema
	columns := []dbase.Column{
		// Character field (string)
		{
			Name:   "NAME",
			Type:   dbase.Character,
			Length: 50,
		},
		// Numeric field
		{
			Name:     "AGE",
			Type:     dbase.Numeric,
			Length:   3,
			Decimals: 0,
		},
		// Float field
		{
			Name:     "SALARY",
			Type:     dbase.Float,
			Length:   10,
			Decimals: 2,
		},
		// Date field
		{
			Name:   "BIRTHDATE",
			Type:   dbase.Date,
			Length: 8,
		},
		// Logical field (boolean)
		{
			Name:   "ACTIVE",
			Type:   dbase.Logical,
			Length: 1,
		},
		// Memo field (large text)
		{
			Name:   "NOTES",
			Type:   dbase.Memo,
			Length: 10,
		},
		// DateTime field
		{
			Name:   "CREATED",
			Type:   dbase.DateTime,
			Length: 8,
		},
	}

	// Create the new table
	table, err := dbase.CreateTable(&dbase.Config{
		Filename:  "employees.dbf",
		Columns:   columns,
		FileType:  dbase.FoxProAutoincrement, // Use FoxPro format
	})
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
	defer table.Close()

	// Add sample records
	records := []map[string]interface{}{
		{
			"NAME":      "John Doe",
			"AGE":       30,
			"SALARY":    75000.50,
			"BIRTHDATE": time.Date(1993, 5, 15, 0, 0, 0, 0, time.UTC),
			"ACTIVE":    true,
			"NOTES":     "Senior developer with 8 years experience",
			"CREATED":   time.Now(),
		},
		{
			"NAME":      "Jane Smith",
			"AGE":       28,
			"SALARY":    65000.00,
			"BIRTHDATE": time.Date(1995, 8, 22, 0, 0, 0, 0, time.UTC),
			"ACTIVE":    true,
			"NOTES":     "Project manager, excellent communication skills",
			"CREATED":   time.Now(),
		},
		{
			"NAME":      "Bob Johnson",
			"AGE":       45,
			"SALARY":    95000.75,
			"BIRTHDATE": time.Date(1978, 12, 3, 0, 0, 0, 0, time.UTC),
			"ACTIVE":    false,
			"NOTES":     "Former CTO, now consultant",
			"CREATED":   time.Now(),
		},
	}

	// Write records to the table
	for _, record := range records {
		row := table.NewRow()
		
		// Set field values
		for name, value := range record {
			field := row.FieldByName(name)
			if field == nil {
				fmt.Printf("Field %s not found\n", name)
				continue
			}
			
			// Set value based on type
			switch v := value.(type) {
			case string:
				field.SetString(v)
			case int:
				field.SetInt(int64(v))
			case int64:
				field.SetInt(v)
			case float64:
				field.SetFloat(v)
			case bool:
				field.SetBool(v)
			case time.Time:
				field.SetTime(v)
			}
		}
		
		// Write the row
		err := table.WriteRow(row)
		if err != nil {
			fmt.Printf("Error writing row: %v\n", err)
		}
	}

	// Save changes
	err = table.Save()
	if err != nil {
		log.Fatal("Failed to save table:", err)
	}

	fmt.Printf("Created table with %d records\n", len(records))
}

// Example 2: Modify existing records
func modifyExistingTable() {
	// Open existing table for writing
	table, err := dbase.OpenTable(&dbase.Config{
		Filename: "employees.dbf",
		ReadOnly: false,
	})
	if err != nil {
		log.Fatal("Failed to open table:", err)
	}
	defer table.Close()

	// Find and modify specific records
	modifiedCount := 0
	
	for table.Next() {
		if table.Deleted() {
			continue
		}
		
		row := table.Row()
		
		// Get the NAME field
		nameField := row.FieldByName("NAME")
		if nameField == nil {
			continue
		}
		
		name := nameField.String()
		
		// Update salary for John Doe
		if name == "John Doe" {
			salaryField := row.FieldByName("SALARY")
			if salaryField != nil {
				// Give a raise
				currentSalary := salaryField.Float()
				newSalary := currentSalary * 1.10 // 10% raise
				salaryField.SetFloat(newSalary)
				
				// Update the notes
				notesField := row.FieldByName("NOTES")
				if notesField != nil {
					notesField.SetString("Senior developer - Received 10% raise")
				}
				
				// Save the modified row
				err := table.WriteRow(row)
				if err != nil {
					fmt.Printf("Error updating row: %v\n", err)
				} else {
					modifiedCount++
					fmt.Printf("Updated salary for %s: %.2f -> %.2f\n", 
						name, currentSalary, newSalary)
				}
			}
		}
	}

	// Commit changes
	err = table.Save()
	if err != nil {
		log.Fatal("Failed to save changes:", err)
	}

	fmt.Printf("Modified %d records\n", modifiedCount)
}

// Example 3: Add new records to existing table
func addRecordsToTable() {
	// Open table for appending
	table, err := dbase.OpenTable(&dbase.Config{
		Filename: "employees.dbf",
		ReadOnly: false,
	})
	if err != nil {
		log.Fatal("Failed to open table:", err)
	}
	defer table.Close()

	// Create new employee record
	newEmployee := map[string]interface{}{
		"NAME":      "Alice Wilson",
		"AGE":       26,
		"SALARY":    55000.00,
		"BIRTHDATE": time.Date(1997, 3, 10, 0, 0, 0, 0, time.UTC),
		"ACTIVE":    true,
		"NOTES":     "Junior developer, fast learner",
		"CREATED":   time.Now(),
	}

	// Create new row
	row := table.NewRow()
	
	// Set field values
	for name, value := range newEmployee {
		field := row.FieldByName(name)
		if field == nil {
			continue
		}
		
		switch v := value.(type) {
		case string:
			field.SetString(v)
		case int:
			field.SetInt(int64(v))
		case float64:
			field.SetFloat(v)
		case bool:
			field.SetBool(v)
		case time.Time:
			field.SetTime(v)
		}
	}
	
	// Append the new row
	err = table.WriteRow(row)
	if err != nil {
		log.Fatal("Failed to write new row:", err)
	}
	
	// Save changes
	err = table.Save()
	if err != nil {
		log.Fatal("Failed to save new record:", err)
	}

	fmt.Println("Added new employee record")
}

// Example 4: Mark records as deleted (soft delete)
func deleteRecords() {
	// Open table for modification
	table, err := dbase.OpenTable(&dbase.Config{
		Filename: "employees.dbf",
		ReadOnly: false,
	})
	if err != nil {
		log.Fatal("Failed to open table:", err)
	}
	defer table.Close()

	deletedCount := 0
	
	// Find and mark inactive employees as deleted
	for table.Next() {
		if table.Deleted() {
			continue // Already deleted
		}
		
		row := table.Row()
		
		// Check ACTIVE field
		activeField := row.FieldByName("ACTIVE")
		if activeField != nil && !activeField.Bool() {
			// Mark as deleted
			err := table.Delete()
			if err != nil {
				fmt.Printf("Error deleting row: %v\n", err)
			} else {
				nameField := row.FieldByName("NAME")
				if nameField != nil {
					fmt.Printf("Marked %s as deleted\n", nameField.String())
				}
				deletedCount++
			}
		}
	}

	// Save changes
	err = table.Save()
	if err != nil {
		log.Fatal("Failed to save deletions:", err)
	}

	fmt.Printf("Marked %d records as deleted\n", deletedCount)
	
	// Note: Records marked as deleted are not physically removed from the file
	// They are just marked with a deletion flag and will be skipped during reads
	// To permanently remove them, you would need to "pack" the database
}