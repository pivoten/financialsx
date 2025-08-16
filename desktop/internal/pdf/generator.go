package pdf

import (
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf/v2"
)

// Generator is the main PDF generator with common functionality
type Generator struct {
	pdf         *gofpdf.Fpdf
	config      *Config
	companyInfo *CompanyInfo
	currentPage int
	totalPages  int
	reportTitle string // Store the report title
}

// Config holds PDF configuration
type Config struct {
	Orientation string  // "P" (portrait) or "L" (landscape)
	Unit        string  // "mm", "pt", "in"
	Size        string  // "Letter", "A4", "Legal"
	FontFamily  string  // Default font family
	FontSize    float64 // Default font size
	Margins     Margins
	HeaderStyle HeaderStyle
	FooterStyle FooterStyle
}

// Margins defines page margins
type Margins struct {
	Top    float64
	Right  float64
	Bottom float64
	Left   float64
}

// HeaderStyle defines header appearance
type HeaderStyle struct {
	Height          float64
	ShowLogo        bool
	ShowCompanyName bool
	ShowAddress     bool
	ShowDate        bool
	ShowReportTitle bool
	Alignment       string // "L", "C", "R"
	BackgroundColor RGB
	TextColor       RGB
}

// FooterStyle defines footer appearance
type FooterStyle struct {
	Height           float64
	ShowPageNumbers  bool
	ShowDate         bool
	ShowCompanyName  bool
	ShowCustomText   bool
	CustomText       string
	Alignment        string // "L", "C", "R"
	BackgroundColor  RGB
	TextColor        RGB
}

// RGB represents a color
type RGB struct {
	R, G, B int
}

// CompanyInfo holds company details for reports
type CompanyInfo struct {
	Name      string
	Address1  string
	Address2  string
	City      string
	State     string
	Zip       string
	Phone     string
	Email     string
	Website   string
	LogoPath  string
}

// DefaultConfig returns standard PDF configuration
func DefaultConfig() *Config {
	return &Config{
		Orientation: "P",
		Unit:        "mm",
		Size:        "Letter",
		FontFamily:  "Arial",
		FontSize:    10,
		Margins: Margins{
			Top:    25.4,  // 1 inch
			Right:  19.05, // 0.75 inch
			Bottom: 25.4,  // 1 inch
			Left:   19.05, // 0.75 inch
		},
		HeaderStyle: HeaderStyle{
			Height:          20,
			ShowCompanyName: true,
			ShowDate:        true,
			ShowReportTitle: true,
			Alignment:       "C",
			TextColor:       RGB{0, 0, 0},
		},
		FooterStyle: FooterStyle{
			Height:          15,
			ShowPageNumbers: true,
			ShowDate:        true,
			Alignment:       "C",
			TextColor:       RGB{128, 128, 128},
		},
	}
}

// LandscapeConfig returns landscape-oriented configuration
func LandscapeConfig() *Config {
	config := DefaultConfig()
	config.Orientation = "L"
	return config
}

// NewGenerator creates a new PDF generator
func NewGenerator(config *Config) *Generator {
	if config == nil {
		config = DefaultConfig()
	}

	pdf := gofpdf.New(config.Orientation, config.Unit, config.Size, "")
	pdf.SetFont(config.FontFamily, "", config.FontSize)
	pdf.SetMargins(config.Margins.Left, config.Margins.Top, config.Margins.Right)
	pdf.SetAutoPageBreak(true, config.Margins.Bottom)

	gen := &Generator{
		pdf:    pdf,
		config: config,
	}

	// Set up header and footer callbacks
	pdf.SetHeaderFunc(gen.headerCallback)
	pdf.SetFooterFunc(gen.footerCallback)

	return gen
}

// SetCompanyInfo sets the company information for the report
func (g *Generator) SetCompanyInfo(info *CompanyInfo) {
	g.companyInfo = info
}

// SetReportTitle sets the title that appears in the header
func (g *Generator) SetReportTitle(title string) {
	g.reportTitle = title
	g.pdf.SetTitle(title, true)
}

// AddPage adds a new page to the PDF
func (g *Generator) AddPage() {
	g.pdf.AddPage()
	g.currentPage++
}

// headerCallback is called automatically for each page
func (g *Generator) headerCallback() {
	if g.config.HeaderStyle.Height == 0 {
		return
	}

	pdf := g.pdf
	style := g.config.HeaderStyle

	// Save current position
	x, y := pdf.GetXY()
	
	// Set header colors
	if style.BackgroundColor.R > 0 || style.BackgroundColor.G > 0 || style.BackgroundColor.B > 0 {
		pdf.SetFillColor(style.BackgroundColor.R, style.BackgroundColor.G, style.BackgroundColor.B)
		width, _ := pdf.GetPageSize()
		pdf.Rect(0, 0, width, style.Height, "F")
	}

	pdf.SetTextColor(style.TextColor.R, style.TextColor.G, style.TextColor.B)
	
	// Company name
	if style.ShowCompanyName && g.companyInfo != nil && g.companyInfo.Name != "" {
		pdf.SetFont(g.config.FontFamily, "B", 14)
		pdf.SetY(5)
		g.alignText(g.companyInfo.Name, style.Alignment)
		pdf.Ln(6)
	}

	// Address
	if style.ShowAddress && g.companyInfo != nil {
		pdf.SetFont(g.config.FontFamily, "", 9)
		if g.companyInfo.Address1 != "" {
			g.alignText(g.companyInfo.Address1, style.Alignment)
			pdf.Ln(4)
		}
		if g.companyInfo.City != "" && g.companyInfo.State != "" {
			address2 := fmt.Sprintf("%s, %s %s", g.companyInfo.City, g.companyInfo.State, g.companyInfo.Zip)
			g.alignText(address2, style.Alignment)
			pdf.Ln(4)
		}
	}

	// Report title
	if style.ShowReportTitle && g.reportTitle != "" {
		pdf.SetFont(g.config.FontFamily, "B", 12)
		pdf.SetY(style.Height - 8)
		g.alignText(g.reportTitle, style.Alignment)
	}

	// Date
	if style.ShowDate {
		pdf.SetFont(g.config.FontFamily, "", 9)
		pdf.SetY(5)
		pdf.SetX(-40)
		pdf.Cell(35, 5, time.Now().Format("January 2, 2006"))
	}

	// Restore position
	pdf.SetXY(x, y)
	pdf.SetTextColor(0, 0, 0)
}

// footerCallback is called automatically for each page
func (g *Generator) footerCallback() {
	if g.config.FooterStyle.Height == 0 {
		return
	}

	pdf := g.pdf
	style := g.config.FooterStyle

	// Position at bottom of page
	pdf.SetY(-style.Height)
	
	// Set footer colors
	pdf.SetTextColor(style.TextColor.R, style.TextColor.G, style.TextColor.B)
	pdf.SetFont(g.config.FontFamily, "", 8)

	// Build footer text
	var footerText string

	// Page numbers
	if style.ShowPageNumbers {
		footerText = fmt.Sprintf("Page %d", pdf.PageNo())
	}

	// Date
	if style.ShowDate {
		if footerText != "" {
			footerText += " | "
		}
		footerText += time.Now().Format("01/02/2006")
	}

	// Company name
	if style.ShowCompanyName && g.companyInfo != nil && g.companyInfo.Name != "" {
		if footerText != "" {
			footerText += " | "
		}
		footerText += g.companyInfo.Name
	}

	// Custom text
	if style.ShowCustomText && style.CustomText != "" {
		if footerText != "" {
			footerText += " | "
		}
		footerText += style.CustomText
	}

	// Align and print footer text
	g.alignText(footerText, style.Alignment)
	
	// Reset text color
	pdf.SetTextColor(0, 0, 0)
}

// alignText aligns text based on alignment setting
func (g *Generator) alignText(text string, alignment string) {
	pdf := g.pdf
	width, _ := pdf.GetPageSize()
	
	switch alignment {
	case "C":
		pdf.SetX((width - pdf.GetStringWidth(text)) / 2)
	case "R":
		pdf.SetX(width - pdf.GetStringWidth(text) - g.config.Margins.Right)
	default: // "L"
		pdf.SetX(g.config.Margins.Left)
	}
	
	pdf.Cell(pdf.GetStringWidth(text), 5, text)
}

// Utility methods for reports

// AddTitle adds a centered title to the document
func (g *Generator) AddTitle(title string, fontSize float64) {
	g.pdf.SetFont(g.config.FontFamily, "B", fontSize)
	width, _ := g.pdf.GetPageSize()
	g.pdf.SetX((width - g.pdf.GetStringWidth(title)) / 2)
	g.pdf.Cell(g.pdf.GetStringWidth(title), fontSize*0.5, title)
	g.pdf.Ln(fontSize * 0.7)
	g.pdf.SetFont(g.config.FontFamily, "", g.config.FontSize)
}

// AddSubtitle adds a subtitle to the document
func (g *Generator) AddSubtitle(subtitle string) {
	g.pdf.SetFont(g.config.FontFamily, "", 10)
	g.pdf.Cell(0, 6, subtitle)
	g.pdf.Ln(8)
}

// AddTable adds a table to the PDF
func (g *Generator) AddTable(headers []string, data [][]string, widths []float64) {
	pdf := g.pdf
	
	// Table header
	pdf.SetFont(g.config.FontFamily, "B", 9)
	pdf.SetFillColor(240, 240, 240)
	pdf.SetTextColor(0, 0, 0)
	
	for i, header := range headers {
		width := widths[i]
		pdf.CellFormat(width, 7, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)
	
	// Table rows
	pdf.SetFont(g.config.FontFamily, "", 8)
	pdf.SetFillColor(255, 255, 255)
	
	for _, row := range data {
		for i, cell := range row {
			if i < len(widths) {
				width := widths[i]
				alignment := "L"
				if i > 0 && isNumeric(cell) {
					alignment = "R"
				}
				pdf.CellFormat(width, 6, cell, "1", 0, alignment, false, 0, "")
			}
		}
		pdf.Ln(-1)
	}
}

// AddSeparator adds a horizontal line separator
func (g *Generator) AddSeparator() {
	pdf := g.pdf
	width, _ := pdf.GetPageSize()
	y := pdf.GetY()
	pdf.Line(g.config.Margins.Left, y, width-g.config.Margins.Right, y)
	pdf.Ln(3)
}

// GetPDF returns the underlying gofpdf instance for custom operations
func (g *Generator) GetPDF() *gofpdf.Fpdf {
	return g.pdf
}

// Output generates the PDF and returns it as bytes
func (g *Generator) Output() ([]byte, error) {
	var buf []byte
	buffer := &pdfBuffer{data: &buf}
	err := g.pdf.Output(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}
	return buf, nil
}

// OutputToFile saves the PDF to a file
func (g *Generator) OutputToFile(filename string) error {
	return g.pdf.OutputFileAndClose(filename)
}

// Helper types

type pdfBuffer struct {
	data *[]byte
}

func (p *pdfBuffer) Write(b []byte) (int, error) {
	*p.data = append(*p.data, b...)
	return len(b), nil
}

// isNumeric checks if a string represents a number
func isNumeric(s string) bool {
	// Simple check - could be enhanced
	if len(s) == 0 {
		return false
	}
	// Check for currency symbols, numbers, decimals, commas
	for _, c := range s {
		if c != '$' && c != '.' && c != ',' && c != '-' && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}