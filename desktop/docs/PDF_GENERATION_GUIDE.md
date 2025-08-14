# PDF Generation Guide for FinancialsX

## Overview
This guide documents how PDF generation is implemented in FinancialsX, specifically for the Chart of Accounts report. The system uses the `gofpdf` library to create professional PDF reports with custom headers, footers, and formatted content.

## Library Used
- **Package**: `github.com/jung-kurt/gofpdf`
- **Import**: `import "github.com/jung-kurt/gofpdf"`

## PDF Structure Components

### 1. Document Setup
```go
// Create a new PDF document with landscape orientation
pdf := gofpdf.New("L", "mm", "Letter", "")
pdf.SetAutoPageBreak(true, 20) // 20mm bottom margin for footer
```

**Parameters explained:**
- `"L"` - Landscape orientation (use `"P"` for Portrait)
- `"mm"` - Units in millimeters (alternatives: `"pt"`, `"cm"`, `"in"`)
- `"Letter"` - Page size (alternatives: `"A4"`, `"Legal"`, etc.)

### 2. Page Dimensions
- **Letter Landscape**: 279.4mm × 215.9mm
- **Usable Width**: 259mm (with 10mm margins on each side)
- **Content Area**: X: 10 to 269, Y: varies based on content

## Header Implementation

### Company Header (First Page)
The header contains company information pulled from VERSION.DBF:

```go
// Company name (large, bold, centered)
pdf.SetFont("Helvetica", "B", 16)
pdf.SetTextColor(52, 73, 94) // Dark blue-gray
pdf.Cell(0, 10, displayCompanyName)

// Company address (smaller, centered)
pdf.SetFont("Helvetica", "", 10)
pdf.SetTextColor(100, 100, 100) // Gray
pdf.Cell(0, 6, companyAddress)

// City, State, Zip
pdf.Cell(0, 6, companyCityStateZip)

// Report title
pdf.SetFont("Helvetica", "B", 14)
pdf.SetTextColor(52, 73, 94)
pdf.Cell(0, 10, "Chart of Accounts")

// Separator line
pdf.SetDrawColor(200, 200, 200) // Light gray
pdf.SetLineWidth(0.5)
pdf.Line(10, pdf.GetY(), 269, pdf.GetY())
```

### Data Source for Header
Company information is retrieved from VERSION.DBF:
- **CPRODUCER**: Company name
- **CADDRESS1/CADDRESS2**: Street address
- **CCITY**: City
- **CSTATE**: State
- **CZIPCODE**: Zip code

## Footer Implementation

### Setting Up Footer Function
The footer must be defined BEFORE adding pages to ensure it applies to all pages:

```go
pdf.SetFooterFunc(func() {
    pdf.SetY(-15) // Position 15mm from bottom
    pdf.SetFont("Helvetica", "", 7) // Small font
    pdf.SetTextColor(128, 128, 128) // Gray text
    
    // Draw separator line
    pdf.SetDrawColor(200, 200, 200)
    pdf.Line(10, pdf.GetY(), 269, pdf.GetY())
    pdf.Ln(2) // Small gap after line
    
    // Footer content goes here...
})
```

### Footer Content Layout

The footer is divided into three sections:

#### Left Section - Branding
```go
// Position at left margin
pdf.SetX(10)

// Write "Pivoten" with superscript TM
pdf.SetFont("Helvetica", "", 7)
pivotText := "Pivoten"
pivotWidth := pdf.GetStringWidth(pivotText)
pdf.Cell(pivotWidth, 5, pivotText)

// Add superscript TM
currentX := pdf.GetX()
currentY := pdf.GetY()
pdf.SetXY(currentX-0.5, currentY-1.2) // Move up and slightly left
pdf.SetFont("Helvetica", "", 4) // Smaller font for superscript
pdf.Cell(2, 2, "TM")

// Continue with version info
pdf.SetXY(currentX+2, currentY) // Reset position
pdf.SetFont("Helvetica", "", 7)
pdf.Cell(50, 5, " - Financials 2026 - BETA 2025-08-13")
```

#### Center Section - Report Details
```go
// Generated timestamp
pdf.SetX(65)
pdf.Cell(50, 5, fmt.Sprintf("Generated: %s", time.Now().Format("January 2, 2006 3:04 PM")))

// Total accounts
pdf.SetX(115)
pdf.Cell(30, 5, fmt.Sprintf("Total: %d", len(accounts)))

// Sort order
pdf.SetX(145)
pdf.Cell(40, 5, fmt.Sprintf("Sort: %s", sortText))

// Filter status
pdf.SetX(185)
pdf.Cell(40, 5, fmt.Sprintf("Filter: %s", filterText))
```

#### Right Section - Page Number
```go
// Right-aligned page number
pageText := fmt.Sprintf("Page %d", pdf.PageNo())
pageWidth := pdf.GetStringWidth(pageText)
pdf.SetX(269 - pageWidth) // Position at right margin minus text width
pdf.Cell(pageWidth, 5, pageText)
```

## Table Implementation

### Table Setup
```go
// Set table styling
pdf.SetFillColor(245, 245, 245) // Light gray for headers
pdf.SetTextColor(40, 40, 40) // Dark gray text
pdf.SetDrawColor(200, 200, 200) // Light gray borders
pdf.SetLineWidth(0.2)

// Define column widths (must total ≤ 259 for landscape)
colWidths := []float64{38, 122, 33, 22, 22, 22} // Total: 259
headers := []string{"Account #", "Description", "Type", "Bank", "Unit", "Dept"}
```

### Table Headers
```go
// Draw headers with background fill
for i, header := range headers {
    align := "L" // Left align by default
    if i >= 3 { // Center align for boolean columns
        align = "C"
    } else if i == 2 { // Center align Type column
        align = "C"
    }
    pdf.CellFormat(colWidths[i], 7, header, "1", 0, align, true, 0, "")
}
pdf.Ln(-1) // Move to next line
```

### Table Rows
```go
// Alternate row colors for readability
for rowNum, account := range accounts {
    if rowNum%2 == 1 {
        pdf.SetFillColor(250, 250, 250) // Very light gray
    } else {
        pdf.SetFillColor(255, 255, 255) // White
    }
    
    // Add cells for each column
    pdf.CellFormat(colWidths[0], 6, accountNumber, "1", 0, "L", true, 0, "")
    pdf.CellFormat(colWidths[1], 6, accountName, "1", 0, "L", true, 0, "")
    // ... more cells
    
    pdf.Ln(-1) // Next row
}
```

## File Naming Convention

### Filename Format
```
YYYY-MM-DD - {Company Name} - Chart of Accounts.pdf
```

Example: `2025-08-13 - Knox Oil & Gas - Chart of Accounts.pdf`

### Filename Sanitization
```go
// Clean company name for filename
cleanCompanyName := displayCompanyName

// Remove problematic characters
cleanCompanyName = strings.ReplaceAll(cleanCompanyName, "\\", "_")
cleanCompanyName = strings.ReplaceAll(cleanCompanyName, "/", "_")
cleanCompanyName = strings.ReplaceAll(cleanCompanyName, ":", "_")
cleanCompanyName = strings.ReplaceAll(cleanCompanyName, "*", "_")
cleanCompanyName = strings.ReplaceAll(cleanCompanyName, "?", "_")
cleanCompanyName = strings.ReplaceAll(cleanCompanyName, "\"", "_")
cleanCompanyName = strings.ReplaceAll(cleanCompanyName, "<", "_")
cleanCompanyName = strings.ReplaceAll(cleanCompanyName, ">", "_")
cleanCompanyName = strings.ReplaceAll(cleanCompanyName, "|", "_")

// Generate filename
datePrefix := time.Now().Format("2006-01-02")
defaultFilename := fmt.Sprintf("%s - %s - Chart of Accounts.pdf", datePrefix, cleanCompanyName)
```

## Save Dialog Integration

Using Wails runtime for native save dialog:

```go
selectedFile, err := wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
    Title:           "Save Chart of Accounts Report",
    DefaultFilename: defaultFilename,
    Filters: []wailsruntime.FileFilter{
        {
            DisplayName: "PDF Files (*.pdf)",
            Pattern:     "*.pdf",
        },
        {
            DisplayName: "All Files (*.*)",
            Pattern:     "*.*",
        },
    },
})

if selectedFile == "" {
    return "", fmt.Errorf("save cancelled by user")
}

// Write PDF to selected file
err = pdf.OutputFileAndClose(selectedFile)
```

## Special Formatting Techniques

### Superscript Text (for TM symbol)
```go
// Save current position
currentX := pdf.GetX()
currentY := pdf.GetY()

// Move up and slightly left for superscript
pdf.SetXY(currentX-0.5, currentY-1.2)
pdf.SetFont("Helvetica", "", 4) // Smaller font
pdf.Cell(2, 2, "TM")

// Reset position for normal text
pdf.SetXY(currentX+2, currentY)
pdf.SetFont("Helvetica", "", 7) // Normal font
```

### Indentation for Sub-Accounts
```go
// Check if account has parent
if hasParent {
    accountNumber = "  " + accountNumber // Add spaces for indentation
}
```

### Dynamic Column Alignment
```go
// Right-align text dynamically
text := "Page 1"
textWidth := pdf.GetStringWidth(text)
rightMargin := 269
pdf.SetX(rightMargin - textWidth)
pdf.Cell(textWidth, 5, text)
```

## Color Scheme

The PDF uses a professional color scheme:

- **Headers**: Dark blue-gray (RGB: 52, 73, 94)
- **Body Text**: Dark gray (RGB: 40, 40, 40)
- **Footer Text**: Medium gray (RGB: 128, 128, 128)
- **Borders**: Light gray (RGB: 200, 200, 200)
- **Table Header Background**: Very light gray (RGB: 245, 245, 245)
- **Alternate Row Background**: Off-white (RGB: 250, 250, 250)

## Best Practices

1. **Always Define Footer Before Adding Pages**: The footer function must be set before `pdf.AddPage()` to ensure it appears on all pages.

2. **Check Page Margins**: Ensure content doesn't exceed page boundaries:
   - Landscape Letter: Keep X coordinates between 10 and 269
   - Account for footer height when setting auto page break

3. **Font Management**: 
   - Define fonts at the beginning of each section
   - Reset font after special formatting (like superscripts)

4. **Error Handling**: Always check for errors in:
   - File operations
   - Data retrieval
   - PDF generation

5. **Memory Management**: For large reports, consider:
   - Streaming data instead of loading all at once
   - Using `pdf.OutputFileAndClose()` to properly close resources

## Complete Example Function Structure

```go
func GeneratePDF(data []Account) (string, error) {
    // 1. Create PDF document
    pdf := gofpdf.New("L", "mm", "Letter", "")
    pdf.SetAutoPageBreak(true, 20)
    
    // 2. Define footer (before adding pages!)
    pdf.SetFooterFunc(func() {
        // Footer implementation
    })
    
    // 3. Add first page
    pdf.AddPage()
    
    // 4. Add header
    // Company info, title, etc.
    
    // 5. Add table headers
    // Column definitions
    
    // 6. Add table data
    // Loop through records
    
    // 7. Save with dialog
    selectedFile, err := wailsruntime.SaveFileDialog(...)
    if err != nil {
        return "", err
    }
    
    // 8. Output PDF
    err = pdf.OutputFileAndClose(selectedFile)
    if err != nil {
        return "", err
    }
    
    return selectedFile, nil
}
```

## Troubleshooting

### Common Issues and Solutions

1. **Footer not appearing on all pages**
   - Solution: Ensure `SetFooterFunc()` is called before `AddPage()`

2. **Content cut off at page edges**
   - Solution: Check that line endpoints don't exceed 269 for landscape
   - Verify margins are consistent

3. **Page numbers incorrect**
   - Solution: Use `pdf.PageNo()` inside the footer function for dynamic page numbers

4. **Special characters not displaying**
   - Solution: Use ASCII alternatives or manual drawing (like superscript TM)

5. **Table rows splitting across pages**
   - Solution: Set appropriate `SetAutoPageBreak()` margin
   - Consider checking `pdf.GetY()` before adding rows

## Future Enhancements

Potential improvements for PDF generation:

1. **Add charts/graphs** using the drawing primitives
2. **Include images** like company logos
3. **Generate multiple report formats** from same data
4. **Add bookmarks** for multi-section reports
5. **Implement digital signatures** for security
6. **Create fillable PDF forms** for data entry
7. **Add watermarks** for draft versions
8. **Support for multiple languages** and RTL text

---

*Last Updated: August 2025*
*Version: 1.0*