// Example: Generating Documentation from DBF Structure
// Demonstrates creating documentation from database schema

package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Valentin-Kaiser/go-dbase/dbase"
)

// DocumentationGenerator creates documentation from DBF files
type DocumentationGenerator struct {
	tables     []TableDoc
	outputPath string
	format     string // html, markdown, json, xml
}

// TableDoc represents table documentation
type TableDoc struct {
	Name        string      `json:"name" xml:"name"`
	Filename    string      `json:"filename" xml:"filename"`
	Description string      `json:"description" xml:"description"`
	Records     int         `json:"records" xml:"records"`
	Modified    time.Time   `json:"modified" xml:"modified"`
	Fields      []FieldDoc  `json:"fields" xml:"fields"`
	Indexes     []IndexDoc  `json:"indexes" xml:"indexes"`
	SampleData  [][]string  `json:"sample_data,omitempty" xml:"sample_data,omitempty"`
}

// FieldDoc represents field documentation
type FieldDoc struct {
	Name        string `json:"name" xml:"name"`
	Type        string `json:"type" xml:"type"`
	Length      int    `json:"length" xml:"length"`
	Decimals    int    `json:"decimals" xml:"decimals"`
	Description string `json:"description" xml:"description"`
	Example     string `json:"example" xml:"example"`
}

// IndexDoc represents index documentation
type IndexDoc struct {
	Name       string `json:"name" xml:"name"`
	Expression string `json:"expression" xml:"expression"`
	Unique     bool   `json:"unique" xml:"unique"`
}

func main() {
	fmt.Println("=== DBF Documentation Generator ===\n")

	// Example 1: Generate HTML documentation
	generateHTMLDocs()

	// Example 2: Generate Markdown documentation
	generateMarkdownDocs()

	// Example 3: Generate JSON data dictionary
	generateJSONDictionary()

	// Example 4: Generate XML schema
	generateXMLSchema()

	// Example 5: Generate database diagram
	generateERDiagram()
}

// generateHTMLDocs creates HTML documentation
func generateHTMLDocs() {
	fmt.Println("=== Generating HTML Documentation ===")

	gen := NewDocumentationGenerator("html", "docs/html")
	
	// Scan for DBF files
	dbfFiles := []string{"customers.dbf", "orders.dbf", "products.dbf"}
	
	for _, file := range dbfFiles {
		tableDoc := gen.DocumentTable(file)
		gen.tables = append(gen.tables, tableDoc)
	}

	// Generate HTML
	htmlContent := gen.GenerateHTML()
	
	// Save to file
	outputFile := "database_documentation.html"
	fmt.Printf("Generated: %s\n", outputFile)
	
	// Preview
	fmt.Println("\nHTML Preview:")
	fmt.Println(htmlContent[:500] + "...")
}

// NewDocumentationGenerator creates a new generator
func NewDocumentationGenerator(format, outputPath string) *DocumentationGenerator {
	return &DocumentationGenerator{
		format:     format,
		outputPath: outputPath,
		tables:     []TableDoc{},
	}
}

// DocumentTable documents a single table
func (g *DocumentationGenerator) DocumentTable(filename string) TableDoc {
	doc := TableDoc{
		Name:     strings.TrimSuffix(filepath.Base(filename), ".dbf"),
		Filename: filename,
		Modified: time.Now(), // Would get from file
		Records:  1234,       // Would get from header
	}

	// Open and analyze table
	table, err := dbase.OpenTable(&dbase.Config{
		Filename:   filename,
		TrimSpaces: true,
		ReadOnly:   true,
	})
	if err != nil {
		fmt.Printf("Error opening %s: %v\n", filename, err)
		return doc
	}
	defer table.Close()

	// Document fields
	// This is simplified - actual implementation would read from header
	doc.Fields = []FieldDoc{
		{
			Name:        "CUST_ID",
			Type:        "Character",
			Length:      10,
			Description: "Unique customer identifier",
			Example:     "C000001",
		},
		{
			Name:        "COMPANY",
			Type:        "Character",
			Length:      40,
			Description: "Company name",
			Example:     "Acme Corp",
		},
		{
			Name:        "BALANCE",
			Type:        "Numeric",
			Length:      12,
			Decimals:    2,
			Description: "Current account balance",
			Example:     "1234.56",
		},
	}

	// Add sample data
	doc.SampleData = g.GetSampleData(table, 3)

	return doc
}

// GetSampleData retrieves sample records
func (g *DocumentationGenerator) GetSampleData(table *dbase.File, count int) [][]string {
	var samples [][]string
	
	for i := 0; i < count; i++ {
		row, err := table.Next()
		if err != nil {
			break
		}
		
		deleted, _ := table.Deleted()
		if deleted {
			i-- // Don't count deleted records
			continue
		}
		
		// Convert row to string array
		// Simplified - would actually extract field values
		samples = append(samples, []string{"Sample", "Data", "Row"})
	}
	
	return samples
}

// GenerateHTML creates HTML documentation
func (g *DocumentationGenerator) GenerateHTML() string {
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <title>Database Documentation</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        h2 { color: #666; border-bottom: 2px solid #ccc; }
        table { border-collapse: collapse; width: 100%; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .field-type { font-family: monospace; color: #008; }
        .example { font-style: italic; color: #666; }
    </style>
</head>
<body>
    <h1>Database Documentation</h1>
    <p>Generated: {{.Generated}}</p>
    
    {{range .Tables}}
    <h2>Table: {{.Name}}</h2>
    <p>File: {{.Filename}}</p>
    <p>Records: {{.Records}} | Modified: {{.Modified.Format "2006-01-02"}}</p>
    
    <h3>Fields</h3>
    <table>
        <tr>
            <th>Field Name</th>
            <th>Type</th>
            <th>Length</th>
            <th>Description</th>
            <th>Example</th>
        </tr>
        {{range .Fields}}
        <tr>
            <td><strong>{{.Name}}</strong></td>
            <td class="field-type">{{.Type}}</td>
            <td>{{.Length}}{{if .Decimals}}.{{.Decimals}}{{end}}</td>
            <td>{{.Description}}</td>
            <td class="example">{{.Example}}</td>
        </tr>
        {{end}}
    </table>
    {{end}}
</body>
</html>
`

	// Parse and execute template
	tmpl, _ := template.New("doc").Parse(htmlTemplate)
	
	data := struct {
		Generated time.Time
		Tables    []TableDoc
	}{
		Generated: time.Now(),
		Tables:    g.tables,
	}
	
	// Execute template to string
	var output strings.Builder
	tmpl.Execute(&output, data)
	
	return output.String()
}

// generateMarkdownDocs creates Markdown documentation
func generateMarkdownDocs() {
	fmt.Println("\n=== Generating Markdown Documentation ===")

	gen := NewDocumentationGenerator("markdown", "docs/md")
	
	// Document tables
	tables := []TableDoc{
		{
			Name:     "Customers",
			Filename: "customers.dbf",
			Records:  1500,
			Modified: time.Now(),
			Fields: []FieldDoc{
				{Name: "CUST_ID", Type: "C", Length: 10},
				{Name: "NAME", Type: "C", Length: 50},
				{Name: "BALANCE", Type: "N", Length: 12, Decimals: 2},
			},
		},
	}
	
	markdown := gen.GenerateMarkdown(tables)
	fmt.Println("\nMarkdown Output:")
	fmt.Println(markdown)
}

// GenerateMarkdown creates Markdown documentation
func (g *DocumentationGenerator) GenerateMarkdown(tables []TableDoc) string {
	var md strings.Builder
	
	md.WriteString("# Database Documentation\n\n")
	md.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	md.WriteString("## Tables\n\n")
	
	for _, table := range tables {
		md.WriteString(fmt.Sprintf("### %s\n\n", table.Name))
		md.WriteString(fmt.Sprintf("- **File**: %s\n", table.Filename))
		md.WriteString(fmt.Sprintf("- **Records**: %d\n", table.Records))
		md.WriteString(fmt.Sprintf("- **Modified**: %s\n\n", table.Modified.Format("2006-01-02")))
		
		md.WriteString("#### Fields\n\n")
		md.WriteString("| Field | Type | Length | Decimals | Description |\n")
		md.WriteString("|-------|------|--------|----------|-------------|\n")
		
		for _, field := range table.Fields {
			md.WriteString(fmt.Sprintf("| %s | %s | %d | %d | %s |\n",
				field.Name, field.Type, field.Length, field.Decimals, field.Description))
		}
		md.WriteString("\n")
	}
	
	return md.String()
}

// generateJSONDictionary creates JSON data dictionary
func generateJSONDictionary() {
	fmt.Println("\n=== Generating JSON Data Dictionary ===")

	dictionary := map[string]interface{}{
		"version":   "1.0",
		"generated": time.Now(),
		"database": map[string]interface{}{
			"name":   "Sample Database",
			"tables": []TableDoc{
				{
					Name:     "Customers",
					Filename: "customers.dbf",
					Records:  1500,
					Fields: []FieldDoc{
						{Name: "CUST_ID", Type: "C", Length: 10},
						{Name: "NAME", Type: "C", Length: 50},
					},
				},
			},
		},
	}
	
	jsonData, _ := json.MarshalIndent(dictionary, "", "  ")
	
	fmt.Println("\nJSON Dictionary:")
	fmt.Println(string(jsonData))
	
	// Save to file
	outputFile := "data_dictionary.json"
	os.WriteFile(outputFile, jsonData, 0644)
	fmt.Printf("\nSaved to: %s\n", outputFile)
}

// generateXMLSchema creates XML schema documentation
func generateXMLSchema() {
	fmt.Println("\n=== Generating XML Schema ===")

	schema := struct {
		XMLName xml.Name   `xml:"database"`
		Name    string     `xml:"name,attr"`
		Tables  []TableDoc `xml:"table"`
	}{
		Name: "SampleDB",
		Tables: []TableDoc{
			{
				Name:     "Customers",
				Filename: "customers.dbf",
				Fields: []FieldDoc{
					{Name: "CUST_ID", Type: "C", Length: 10},
				},
			},
		},
	}
	
	xmlData, _ := xml.MarshalIndent(schema, "", "  ")
	
	fmt.Println("\nXML Schema:")
	fmt.Println(string(xmlData))
}

// generateERDiagram creates entity relationship diagram
func generateERDiagram() {
	fmt.Println("\n=== Generating ER Diagram ===")
	
	// Generate PlantUML or Mermaid diagram
	diagram := `
@startuml
!define Table(name,desc) class name as "desc" << (T,#FFAAAA) >>
!define primary_key(x) <b>x</b>
!define foreign_key(x) <i>x</i>

Table(Customers, "Customers") {
  primary_key(CustomerID): VARCHAR(10)
  CompanyName: VARCHAR(40)
  ContactName: VARCHAR(30)
  Country: VARCHAR(15)
}

Table(Orders, "Orders") {
  primary_key(OrderID): INTEGER
  foreign_key(CustomerID): VARCHAR(10)
  OrderDate: DATE
  ShipDate: DATE
}

Table(OrderDetails, "Order Details") {
  foreign_key(OrderID): INTEGER
  foreign_key(ProductID): INTEGER
  Quantity: INTEGER
  UnitPrice: DECIMAL(10,2)
}

Table(Products, "Products") {
  primary_key(ProductID): INTEGER
  ProductName: VARCHAR(40)
  UnitPrice: DECIMAL(10,2)
  UnitsInStock: INTEGER
}

Customers ||--o{ Orders
Orders ||--o{ OrderDetails
Products ||--o{ OrderDetails

@enduml
`
	
	fmt.Println("PlantUML Diagram:")
	fmt.Println(diagram)
	
	// Alternative: Mermaid diagram
	mermaid := `
erDiagram
    CUSTOMERS ||--o{ ORDERS : places
    ORDERS ||--o{ ORDER_DETAILS : contains
    PRODUCTS ||--o{ ORDER_DETAILS : includes
    
    CUSTOMERS {
        string CustomerID PK
        string CompanyName
        string ContactName
    }
    
    ORDERS {
        int OrderID PK
        string CustomerID FK
        date OrderDate
    }
    
    ORDER_DETAILS {
        int OrderID FK
        int ProductID FK
        int Quantity
    }
    
    PRODUCTS {
        int ProductID PK
        string ProductName
        decimal UnitPrice
    }
`
	
	fmt.Println("\nMermaid Diagram:")
	fmt.Println(mermaid)
}

// Additional documentation features

// GenerateDataDictionary creates comprehensive data dictionary
func GenerateDataDictionary(w io.Writer, tables []TableDoc) {
	fmt.Fprintln(w, "DATA DICTIONARY")
	fmt.Fprintln(w, "="*50)
	fmt.Fprintln(w)
	
	for _, table := range tables {
		fmt.Fprintf(w, "TABLE: %s\n", strings.ToUpper(table.Name))
		fmt.Fprintf(w, "File: %s\n", table.Filename)
		fmt.Fprintf(w, "Records: %d\n", table.Records)
		fmt.Fprintln(w)
		
		fmt.Fprintln(w, "FIELDS:")
		for _, field := range table.Fields {
			fmt.Fprintf(w, "  %-15s %-10s %3d", field.Name, field.Type, field.Length)
			if field.Decimals > 0 {
				fmt.Fprintf(w, ".%d", field.Decimals)
			}
			if field.Description != "" {
				fmt.Fprintf(w, "  -- %s", field.Description)
			}
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w)
	}
}

// GenerateChangeLog documents schema changes
func GenerateChangeLog(oldSchema, newSchema []TableDoc) string {
	var log strings.Builder
	
	log.WriteString("SCHEMA CHANGE LOG\n")
	log.WriteString(fmt.Sprintf("Date: %s\n\n", time.Now().Format("2006-01-02")))
	
	// Compare schemas and document changes
	// ... (implementation would compare and log differences)
	
	return log.String()
}

// Tips for documentation:
// 1. Keep documentation up-to-date
// 2. Include examples and samples
// 3. Document relationships
// 4. Explain business rules
// 5. Version your documentation
// 6. Include data types and constraints
// 7. Generate diagrams for visualization