// Example: Working with DBC Database Containers
// Demonstrates FoxPro database container operations

package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/Valentin-Kaiser/go-dbase/dbase"
)

// DatabaseInfo represents a DBC database container
type DatabaseInfo struct {
	Name        string
	Path        string
	Tables      []TableInfo
	Views       []ViewInfo
	Connections []ConnectionInfo
	Relations   []RelationInfo
}

// TableInfo represents a table in the database
type TableInfo struct {
	Name        string
	Path        string
	Alias       string
	Fields      []FieldInfo
	Indexes     []IndexInfo
	Triggers    []TriggerInfo
	Rules       []RuleInfo
}

// FieldInfo represents a field with extended properties
type FieldInfo struct {
	Name         string
	Type         string
	Length       int
	Decimals     int
	AllowNull    bool
	DefaultValue interface{}
	Caption      string
	Comment      string
	Format       string
	InputMask    string
	Validation   string
}

// IndexInfo represents an index
type IndexInfo struct {
	Name       string
	Expression string
	Type       string // PRIMARY, CANDIDATE, REGULAR
	Unique     bool
	Descending bool
}

// ViewInfo represents a database view
type ViewInfo struct {
	Name       string
	SQL        string
	Updateable bool
	Fields     []string
}

// ConnectionInfo represents a remote connection
type ConnectionInfo struct {
	Name         string
	DataSource   string
	ConnectString string
	UserID       string
}

// RelationInfo represents a table relationship
type RelationInfo struct {
	Name        string
	ParentTable string
	ChildTable  string
	ParentKey   string
	ChildKey    string
	Type        string // ONE_TO_MANY, ONE_TO_ONE
}

// TriggerInfo represents a table trigger
type TriggerInfo struct {
	Name    string
	Type    string // INSERT, UPDATE, DELETE
	Timing  string // BEFORE, AFTER
	Code    string
}

// RuleInfo represents a validation rule
type RuleInfo struct {
	Name       string
	Expression string
	Message    string
	Type       string // FIELD, RECORD
}

func main() {
	fmt.Println("=== FoxPro Database Container (DBC) Examples ===\n")

	// Example 1: Open and explore a database
	exploreDatabase()

	// Example 2: List all tables in database
	listDatabaseTables()

	// Example 3: Show table relationships
	showRelationships()

	// Example 4: Extract database schema
	extractSchema()

	// Example 5: Work with views
	workWithViews()
}

// exploreDatabase opens and explores a DBC file
func exploreDatabase() {
	fmt.Println("=== Exploring Database Container ===")

	// Note: DBC support in go-dbase may be limited
	// This shows the conceptual approach
	
	dbPath := "northwind.dbc"
	fmt.Printf("Opening database: %s\n", dbPath)

	// In FoxPro, a DBC contains metadata about:
	// - Tables and their locations
	// - Field properties
	// - Indexes
	// - Relations
	// - Stored procedures
	// - Connections
	// - Views

	db := openDatabase(dbPath)
	if db == nil {
		fmt.Println("Failed to open database")
		return
	}

	fmt.Printf("\nDatabase: %s\n", db.Name)
	fmt.Printf("Tables: %d\n", len(db.Tables))
	fmt.Printf("Views: %d\n", len(db.Views))
	fmt.Printf("Relations: %d\n", len(db.Relations))
	fmt.Printf("Connections: %d\n", len(db.Connections))

	// List tables
	fmt.Println("\nTables in database:")
	for _, table := range db.Tables {
		fmt.Printf("  - %s (%s)\n", table.Name, table.Path)
	}
}

// openDatabase simulates opening a DBC file
func openDatabase(dbcPath string) *DatabaseInfo {
	// In reality, you would:
	// 1. Open the DBC file (it's actually a DBF file)
	// 2. Read the metadata tables
	// 3. Parse the structure

	// For demonstration, return sample data
	return &DatabaseInfo{
		Name: filepath.Base(dbcPath),
		Path: dbcPath,
		Tables: []TableInfo{
			{
				Name:  "Customers",
				Path:  "customers.dbf",
				Alias: "CUST",
				Fields: []FieldInfo{
					{
						Name:      "CustomerID",
						Type:      "C",
						Length:    10,
						AllowNull: false,
						Caption:   "Customer ID",
						Comment:   "Unique customer identifier",
					},
					{
						Name:         "CompanyName",
						Type:         "C",
						Length:       40,
						AllowNull:    false,
						Caption:      "Company",
						DefaultValue: "",
					},
				},
				Indexes: []IndexInfo{
					{
						Name:       "PK_CustomerID",
						Expression: "CustomerID",
						Type:       "PRIMARY",
						Unique:     true,
					},
					{
						Name:       "CompanyName",
						Expression: "UPPER(CompanyName)",
						Type:       "REGULAR",
					},
				},
			},
			{
				Name:  "Orders",
				Path:  "orders.dbf",
				Alias: "ORDR",
			},
			{
				Name:  "Products",
				Path:  "products.dbf",
				Alias: "PROD",
			},
		},
		Relations: []RelationInfo{
			{
				Name:        "CustomerOrders",
				ParentTable: "Customers",
				ChildTable:  "Orders",
				ParentKey:   "CustomerID",
				ChildKey:    "CustomerID",
				Type:        "ONE_TO_MANY",
			},
		},
	}
}

// listDatabaseTables shows all tables with details
func listDatabaseTables() {
	fmt.Println("\n=== Database Tables Detail ===")

	db := openDatabase("sample.dbc")
	if db == nil {
		return
	}

	for _, table := range db.Tables {
		fmt.Printf("\nTable: %s\n", table.Name)
		fmt.Printf("  File: %s\n", table.Path)
		fmt.Printf("  Alias: %s\n", table.Alias)
		
		// Open the actual DBF file
		dbfTable, err := dbase.OpenTable(&dbase.Config{
			Filename:   table.Path,
			TrimSpaces: true,
			ReadOnly:   true,
		})
		if err != nil {
			fmt.Printf("  Error opening table: %v\n", err)
			continue
		}
		defer dbfTable.Close()

		header := dbfTable.Header()
		fmt.Printf("  Records: %d\n", header.RecordsCount())
		fmt.Printf("  Last Modified: %s\n", header.Modified())
		
		// Show fields
		fmt.Println("  Fields:")
		for _, field := range table.Fields {
			fmt.Printf("    - %s (%s%d): %s\n",
				field.Name,
				field.Type,
				field.Length,
				field.Caption)
		}

		// Show indexes
		if len(table.Indexes) > 0 {
			fmt.Println("  Indexes:")
			for _, idx := range table.Indexes {
				fmt.Printf("    - %s: %s (%s)\n",
					idx.Name,
					idx.Expression,
					idx.Type)
			}
		}
	}
}

// showRelationships displays database relationships
func showRelationships() {
	fmt.Println("\n=== Database Relationships ===")

	db := openDatabase("sample.dbc")
	if db == nil {
		return
	}

	if len(db.Relations) == 0 {
		fmt.Println("No relationships defined")
		return
	}

	for _, rel := range db.Relations {
		fmt.Printf("\nRelation: %s\n", rel.Name)
		fmt.Printf("  Type: %s\n", rel.Type)
		fmt.Printf("  Parent: %s.%s\n", rel.ParentTable, rel.ParentKey)
		fmt.Printf("  Child: %s.%s\n", rel.ChildTable, rel.ChildKey)
		
		// Show referential integrity rules
		fmt.Println("  Rules:")
		fmt.Println("    - Cascade Delete: No")
		fmt.Println("    - Cascade Update: No")
		fmt.Println("    - Restrict Delete: Yes")
	}

	// Visual representation
	fmt.Println("\n=== Relationship Diagram ===")
	fmt.Println("Customers (1)")
	fmt.Println("    |")
	fmt.Println("    | CustomerID")
	fmt.Println("    |")
	fmt.Println("    v")
	fmt.Println("Orders (∞)")
}

// extractSchema exports database schema
func extractSchema() {
	fmt.Println("\n=== Extracting Database Schema ===")

	db := openDatabase("sample.dbc")
	if db == nil {
		return
	}

	// Generate SQL DDL
	fmt.Println("\n-- SQL Schema Export --")
	
	for _, table := range db.Tables {
		fmt.Printf("\nCREATE TABLE %s (\n", table.Name)
		
		for i, field := range table.Fields {
			sqlType := dbfToSQLType(field.Type, field.Length, field.Decimals)
			nullConstraint := ""
			if !field.AllowNull {
				nullConstraint = " NOT NULL"
			}
			
			comma := ","
			if i == len(table.Fields)-1 && len(table.Indexes) == 0 {
				comma = ""
			}
			
			fmt.Printf("    %s %s%s%s\n",
				field.Name, sqlType, nullConstraint, comma)
		}
		
		// Add primary key
		for _, idx := range table.Indexes {
			if idx.Type == "PRIMARY" {
				fmt.Printf("    PRIMARY KEY (%s)\n", idx.Expression)
			}
		}
		
		fmt.Println(");")
	}

	// Add foreign keys
	for _, rel := range db.Relations {
		fmt.Printf("\nALTER TABLE %s\n", rel.ChildTable)
		fmt.Printf("    ADD FOREIGN KEY (%s)\n", rel.ChildKey)
		fmt.Printf("    REFERENCES %s(%s);\n", rel.ParentTable, rel.ParentKey)
	}
}

// dbfToSQLType converts DBF field types to SQL
func dbfToSQLType(dbfType string, length, decimals int) string {
	switch dbfType {
	case "C":
		return fmt.Sprintf("VARCHAR(%d)", length)
	case "N":
		if decimals > 0 {
			return fmt.Sprintf("DECIMAL(%d,%d)", length, decimals)
		}
		return fmt.Sprintf("INTEGER")
	case "D":
		return "DATE"
	case "T":
		return "TIMESTAMP"
	case "L":
		return "BOOLEAN"
	case "M":
		return "TEXT"
	case "F":
		return fmt.Sprintf("FLOAT(%d)", length)
	case "I":
		return "INTEGER"
	case "Y":
		return "DECIMAL(19,4)"
	default:
		return fmt.Sprintf("VARCHAR(%d)", length)
	}
}

// workWithViews demonstrates working with database views
func workWithViews() {
	fmt.Println("\n=== Working with Views ===")

	// In FoxPro, views can be:
	// 1. Local views (based on local tables)
	// 2. Remote views (based on remote data)

	views := []ViewInfo{
		{
			Name: "CustomerOrders",
			SQL: `SELECT c.CompanyName, o.OrderID, o.OrderDate, o.Total
				  FROM Customers c
				  JOIN Orders o ON c.CustomerID = o.CustomerID
				  WHERE o.Status = 'Active'`,
			Updateable: false,
			Fields:     []string{"CompanyName", "OrderID", "OrderDate", "Total"},
		},
		{
			Name: "ProductInventory",
			SQL: `SELECT ProductName, UnitsInStock, ReorderLevel,
				         UnitsInStock - ReorderLevel as Available
				  FROM Products
				  WHERE Discontinued = .F.`,
			Updateable: true,
			Fields:     []string{"ProductName", "UnitsInStock", "ReorderLevel", "Available"},
		},
	}

	for _, view := range views {
		fmt.Printf("\nView: %s\n", view.Name)
		fmt.Printf("  Updateable: %v\n", view.Updateable)
		fmt.Printf("  Fields: %v\n", view.Fields)
		fmt.Printf("  SQL:\n%s\n", view.SQL)
	}
}

// Additional database operations

// validateDatabase checks database integrity
func validateDatabase(db *DatabaseInfo) []string {
	var issues []string

	// Check if all table files exist
	for _, table := range db.Tables {
		// Check if DBF file exists
		// In real implementation, use os.Stat
		fmt.Printf("Checking table file: %s\n", table.Path)
	}

	// Check relationships validity
	for _, rel := range db.Relations {
		// Verify parent and child tables exist
		// Verify key fields exist
		fmt.Printf("Validating relationship: %s\n", rel.Name)
	}

	// Check for orphaned indexes
	// Check for missing memo files
	// Validate stored procedures

	return issues
}

// Tips for working with DBC files:
// 1. DBC is actually a special DBF file
// 2. Metadata is stored in hidden tables
// 3. Use VALIDATE DATABASE in FoxPro
// 4. Back up DBC before modifications
// 5. Keep DBC and DBF files together
// 6. Handle long field names (DBC feature)
// 7. Consider stored procedures and triggers