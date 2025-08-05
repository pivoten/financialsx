# WV Operator Return System - Complete Project Documentation

## Project Overview

The WV Operator Return System is a Go application designed to process FoxPro DBF files and generate West Virginia oil and gas annual return PDF forms. The system has evolved from a simple PDF generator to a comprehensive data management solution with database storage, editing capabilities, and smart data flow.

## Core Purpose

Convert FoxPro DBF data into properly formatted PDF annual return forms for West Virginia oil and gas producers, with the ability to review, edit, and validate data before PDF generation.

## Technical Architecture

### Programming Language
- **Go (Golang)**: Primary language for the application
- **SQLite**: Database for data storage and editing
- **pdfcpu**: CLI tool for PDF form filling

### Data Sources
- **FoxPro DBF Files**: Source data format
  - `VERSION.DBF` - Producer/company information
  - `WELLS.DBF` - Well information and metadata
  - `WELLHIST.DBF` - Production and financial history
  - `WELLINV.DBF` - Interest and ownership data

### Key Dependencies
- `github.com/Valentin-Kaiser/go-dbase` - DBF file reading
- `github.com/mattn/go-sqlite3` - SQLite database driver
- `pdfcpu` - PDF form filling (external CLI tool)

## System Components

### 1. Main Application (`main.go`)
**Purpose**: Core application logic, CLI interaction, data processing orchestration

**Key Functions**:
- `ReadProducerInfo()` - Reads producer data from VERSION.DBF
- `ReadWellsData()` - Reads and aggregates well data
- `aggregateWellhistData()` - Aggregates production/financial data
- `aggregateWellinvData()` - Aggregates interest/ownership data
- `handleDatabaseOperations()` - Database creation and population
- `handleDataAnalysisMode()` - Data analysis and reporting

**Configuration Structure**:
```go
type WVConfig struct {
    ReportType              ReportType
    DateType                DateType
    ReportingYear           int // 2-digit year for PDF header (e.g., 26)
    ProductionYear          int // 4-digit year for data filtering (e.g., 2024)
    ConsolidateWIToOperator bool
    SourceDataPath          string
}
```

### 2. PDF Form Filling (`pdf_form_fill.go`)
**Purpose**: Handles PDF form filling using pdfcpu

**Key Features**:
- JSON-based form data structure
- Support for both text fields and checkboxes
- Currency rounding to whole numbers
- Working interest adjustment for rounding errors
- Status checkbox mapping (Active, Plugged, Shut-in)

**Form Data Structure**:
```go
type FormData struct {
    Header struct { /* ... */ }
    Forms []struct {
        Textfield []FormField `json:"textfield"`
        Checkbox  []FormField `json:"checkbox"`
    } `json:"forms"`
}
```

### 3. Database Management (`database.go`)
**Purpose**: SQLite database operations and data persistence

**Database Schema**:
- `operator` - Producer/company information
- `well` - Basic well information
- `well_financial_accounting` - Financial data (accounting dates)
- `well_financial_production` - Financial data (production dates)
- `lookup` - Reference data (county codes, status, formations)

**Key Functions**:
- `ClearFinancialDataForYear()` - Clears data for specific year
- `InsertWellFinancialData()` - Inserts both accounting and production data
- `GetWellFromDatabase()` - Retrieves well data for PDF generation
- `CheckDatabaseExists()` - Checks if database has data

### 4. Database Viewer (`cmd/db_viewer/main.go`)
**Purpose**: Separate CLI application for viewing and editing database data

**Features**:
- View operator information
- View wells summary
- View detailed well data
- Edit production, revenue, and interest data
- Export data to CSV

## Data Processing Workflow

### 1. Data Reading and Aggregation
```go
// Read producer information
producer := ReadProducerInfo(sourceDataPath)

// Read and aggregate well data
wells := ReadWellsData(sourceDataPath, productionYear, dateType)

// Aggregate production/financial data
aggregateWellhistData(wells, sourceDataPath, productionYear, dateType)

// Aggregate interest/ownership data
aggregateWellinvData(wells, sourceDataPath, productionYear, dateType)
```

### 2. Date Filtering Logic
The system supports two date filtering methods:

**Accounting Date Filtering**:
- Uses `WELLHIST.HDATE` field
- Extracts year from MM/DD/YYYY or YYYY-MM-DD format
- Compares against 4-digit production year

**Production Date Filtering**:
- Uses `WELLHIST.HYEAR` field
- Direct integer comparison with production year

### 3. Interest Calculations
The system calculates working interest and royalty interest based on `WELLINV` data:

```go
// Interest groups from WELLINV
RoyaltyOilInterest      float64 // SUM(NREVOIL) for Royalty group
RoyaltyGasInterest      float64 // SUM(NREVGAS) for Royalty group
RoyaltyOtherInterest    float64 // SUM(NREVOTH) for Royalty group
WorkingOilInterest      float64 // SUM(NREVOIL) for Working Interest group
WorkingGasInterest      float64 // SUM(NREVGAS) for Working Interest group
WorkingOtherInterest    float64 // SUM(NREVOTH) for Working Interest group
```

### 4. Revenue Calculations
Revenue is calculated by applying interest percentages:

```go
// Working interest revenue
wi_oilRevenue = OilRevenue * (WorkingOilInterest / 100)
wi_gasRevenue = GasRevenue * (WorkingGasInterest / 100)
wi_nglsRevenue = OtherRevenue * (WorkingOtherInterest / 100)

// Royalty interest revenue
royalty_oilRevenue = OilRevenue * (RoyaltyOilInterest / 100)
royalty_gasRevenue = GasRevenue * (RoyaltyGasInterest / 100)
royalty_nglsRevenue = OtherRevenue * (RoyaltyOtherInterest / 100)
```

## Database Operations Workflow

### 1. Database Creation and Setup
```bash
./wv-operator-return
# Select "3. Database Operations"
```

**What happens**:
1. **Directory Check**: Creates `sourcedata/sql/` if needed
2. **Database Check**: Creates `wv_operator.db` if needed
3. **Schema Creation**: Initializes database tables
4. **Data Population**: Reads DBF files and stores data

### 2. Data Review and Editing
```bash
./db_viewer
# Use menu to view and edit data
```

### 3. PDF Generation with Database Data
```bash
./wv-operator-return
# Select "1. Process Data"
# System automatically uses database data if available
```

## PDF Field Mapping

### Schedule 1 - Well Information
- **County**: `WELLS.CCOUNTY` → county lookup
- **NRA Number**: `WELLS.CNRA1` (primary)
- **API Number**: `WELLS.CPERMIT1`
- **Well Name**: `WELLS.CWELLNAME`
- **Land Acreage**: `WELLS.NACRES`
- **Well Status**: `WELLS.CWELLSTAT` → status lookup
- **Formation**: `WELLS.CFORMATION` → formation lookup
- **Initial Production**: `WELLS.DPRODDATE`

### Schedule 2 - Production and Revenue
- **Production Totals**: Aggregated from `WELLHIST`
  - Oil: `SUM(NTOTBBL)`
  - Gas: `SUM(NTOTMCF)`
  - NGLs: `SUM(NTOTPROD)`
- **Revenue**: Aggregated from `WELLHIST`
  - Oil: `SUM(NGROSSOIL)`
  - Gas: `SUM(NGROSSGAS)`
  - Other: `SUM(NOTHINC)`
- **Expenses**: Aggregated from `WELLHIST`
  - Total: `SUM(NTOTALE + NEXPCL1-5)`

### Interest Calculations
- **Working Interest**: Calculated from `WELLINV` data
- **Royalty Interest**: Calculated from `WELLINV` data
- **Revenue Allocation**: Applied based on interest percentages

## Data Rounding and Validation

### Currency Rounding
All currency values are rounded to whole numbers:
```go
func roundToWholeNumber(value float64) float64 {
    return math.Round(value)
}
```

### Working Interest Adjustment
Ensures `Working Interest + Royalty Interest = Total Receipts`:
```go
func adjustWorkingInterestForRounding(totalReceipts, workingInterest, royaltyInterest float64) (float64, float64) {
    roundedTotal := roundToWholeNumber(totalReceipts)
    roundedWI := roundToWholeNumber(workingInterest)
    roundedRI := roundToWholeNumber(royaltyInterest)
    
    if roundedWI+roundedRI == roundedTotal {
        return roundedWI, roundedRI
    }
    
    adjustedWI := roundedTotal - roundedRI
    if adjustedWI < 0 {
        adjustedWI = 0
    }
    return adjustedWI, roundedRI
}
```

## Lookup Files

### County Codes (`seeds/wv_countycode.csv`)
Format: `CountyNumber,CountyName`
Example: `01,Barbour`

### Well Status (`seeds/wv_well_status.csv`)
Format: `StatusCode,StatusDescription`
Example: `A,Active`

### Formations (`seeds/wv_formations.csv`)
Format: `FormationCode,FormationName`
Example: `01,Oriskany`

### Producers (`seeds/wv_producers.csv`)
Format: `ProdCode,ProducerName`
Example: `90194,309 OIL`

## Error Handling and Troubleshooting

### Common Issues

#### 1. DBF File Access
- **Issue**: Cannot read DBF files
- **Solution**: Ensure files are in `sourcedata/` directory
- **Required Files**: `VERSION.DBF`, `WELLS.DBF`, `WELLHIST.DBF`, `WELLINV.DBF`

#### 2. Database Creation
- **Issue**: SQL directory or database not found
- **Solution**: System prompts to create automatically
- **Manual Fix**: `mkdir -p sourcedata/sql`

#### 3. PDF Generation
- **Issue**: Checkboxes not working
- **Solution**: Use boolean values in checkbox array
- **Format**: `"checkbox"` array with `true`/`false` values

#### 4. Data Accuracy
- **Issue**: Revenue calculations incorrect
- **Solution**: Check interest percentages in `WELLINV` data
- **Debug**: Use single well analysis to verify calculations

### Debug Features
- **Single Well Analysis**: Detailed breakdown of calculations
- **Raw Record Display**: Shows source DBF records used
- **Database Viewer**: Interactive data inspection
- **CSV Export**: Data export for external analysis

## File Structure

```
legacy_wvoperator/
├── main.go                           # Main application
├── pdf_form_fill.go                  # PDF generation
├── database.go                       # Database operations
├── cmd/
│   └── db_viewer/
│       └── main.go                   # Database viewer
├── seeds/
│   ├── wv_countycode.csv            # County lookup
│   ├── wv_well_status.csv           # Status lookup
│   ├── wv_formations.csv            # Formation lookup
│   └── wv_producers.csv             # Producer lookup
├── docs/
│   ├── DATABASE_OPERATIONS.md       # Database documentation
│   ├── Current_PDF_Field_Mapping.csv # Field mapping
│   ├── PDF_FORM_FILLING.md          # PDF generation guide
│   └── stc1235.instructions.pdf     # Form instructions
├── sourcedata/
│   ├── VERSION.DBF                  # Producer data
│   ├── WELLS.DBF                    # Well data
│   ├── WELLHIST.DBF                 # Production history
│   ├── WELLINV.DBF                  # Interest data
│   └── sql/
│       └── wv_operator.db           # SQLite database
├── wv-operator-return               # Main executable
└── db_viewer                        # Database viewer executable
```

## Build and Installation

### Prerequisites
```bash
# Install Go
# Install pdfcpu CLI tool
# Ensure DBF files are in sourcedata/ directory
```

### Build Commands
```bash
# Build main application
go build -o wv-operator-return

# Build database viewer
go build -o db_viewer cmd/db_viewer/main.go
```

### Dependencies
```bash
go mod init wv-operator-return
go get github.com/Valentin-Kaiser/go-dbase
go get github.com/mattn/go-sqlite3
```

## Usage Examples

### 1. Basic PDF Generation
```bash
./wv-operator-return
# Select "1. Process Data"
# Enter production year: 2024
# Select date type: Accounting Date
# PDFs generated in output/ directory
```

### 2. Database Operations
```bash
./wv-operator-return
# Select "3. Database Operations"
# Enter production year: 2024
# Database created and populated
```

### 3. Data Analysis
```bash
./wv-operator-return
# Select "2. Analyze Data"
# Choose analysis type
# View detailed reports
```

### 4. Database Viewing/Editing
```bash
./db_viewer
# Use menu to navigate
# View and edit data
# Export to CSV
```

## Key Features Summary

### ✅ **Data Processing**
- FoxPro DBF file reading and parsing
- Data aggregation and calculations
- Interest and revenue calculations
- Date filtering (accounting vs production)

### ✅ **Database Management**
- SQLite database storage
- Automatic directory and database creation
- Year-specific data management
- Dual date type storage (accounting + production)

### ✅ **PDF Generation**
- Form field mapping and population
- Checkbox and text field support
- Currency rounding and validation
- Working interest adjustment

### ✅ **Data Analysis**
- Single well analysis
- Bulk data analysis
- Statistical reporting
- Data validation

### ✅ **User Interface**
- Interactive CLI menus
- Database viewer application
- Data editing capabilities
- Export functionality

### ✅ **Error Handling**
- Graceful fallbacks
- Comprehensive error messages
- Debug information
- Data validation

## Future Enhancements

### Potential Improvements
1. **Web Interface**: Browser-based data viewing and editing
2. **Batch Processing**: Multiple year processing
3. **Data Validation**: Enhanced validation rules
4. **Reporting**: Additional report types
5. **Integration**: API endpoints for external systems
6. **Backup**: Automated database backup
7. **Audit Trail**: Enhanced change tracking
8. **Performance**: Optimized data processing

### Scalability Considerations
- **Large Datasets**: Handle thousands of wells efficiently
- **Concurrent Access**: Multiple users editing data
- **Data Migration**: Version upgrade procedures
- **Backup/Recovery**: Data protection strategies

## Conclusion

The WV Operator Return System provides a comprehensive solution for processing oil and gas data and generating regulatory reports. The system's evolution from a simple PDF generator to a full-featured data management platform demonstrates its flexibility and extensibility.

Key strengths include:
- **Data Integrity**: Comprehensive validation and error handling
- **User Flexibility**: Multiple ways to view and edit data
- **Reliability**: Automatic fallbacks and robust error handling
- **Maintainability**: Well-documented code and clear architecture
- **Extensibility**: Modular design for future enhancements

The system successfully addresses the core requirement of converting FoxPro DBF data into properly formatted PDF forms while providing additional value through data management, validation, and analysis capabilities. 