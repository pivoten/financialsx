# Owner Distribution Statements Feature Documentation

## Overview
The Owner Distribution Statements feature provides the ability to generate PDF statements for oil & gas well owner distributions. This feature reads DBF files from the legacy Visual FoxPro system and converts them to modern PDF documents for distribution to well owners.

## Implementation Date
**Completed**: August 2025 (based on commit 5207448)

## Feature Location
- **Access Path**: Financials → Reports → Owner Statements
- **Component**: `OwnerStatements.tsx`
- **Backend**: `main.go` lines 8027-8217

## Architecture

### Frontend Components
1. **OwnerStatements.tsx** (`desktop/frontend/src/components/`)
   - Main UI component for the feature
   - Handles file discovery and PDF generation
   - Provides user feedback for operations

### Backend Functions
1. **CheckOwnerStatementFiles** (`main.go:8027-8105`)
   - Checks if DBF files exist in `ownerstatements` subdirectory
   - Returns file availability status and error messages
   - Platform-aware path resolution (Windows vs macOS/Linux)

2. **GetOwnerStatementsList** (`main.go:8107-8167`)
   - Returns detailed list of available statement files
   - Includes file metadata (size, modification date, FPT companion files)
   - Formats dates for display

3. **GenerateOwnerStatementPDF** (`main.go:8169-8217`)
   - Reads DBF file data from ownerstatements directory
   - Currently analyzes DBF structure and returns summary
   - PDF generation implementation pending full structure analysis

## Data Flow

### File Discovery Process
1. User navigates to Owner Statements section
2. System checks for `{company_data}/ownerstatements/` directory
3. Scans for `.dbf` files in the directory
4. Returns list with metadata for each file

### Path Resolution
- **Windows**: Uses full path from company configuration
- **macOS/Linux**: Resolves relative to compmast.dbf location
- **Directory Structure**: `{company_folder}/ownerstatements/*.dbf`

## User Interface

### Main Features
- **Back Navigation**: Return to Reports section
- **File List Display**: Shows all available DBF files with:
  - Filename
  - Last modified date
  - File size
  - FPT indicator (memo field companion file)
- **Generate PDF Button**: Individual PDF generation per file
- **Refresh Button**: Re-scan directory for new files
- **Status Messages**: Success/error feedback

### Visual Design
- Follows standard dashboard layout pattern
- Clean card-based interface
- Color-coded status alerts:
  - Amber: No files found
  - Green: Success messages
  - Red: Error messages
- Loading spinner during operations

## Technical Details

### DBF File Structure
The system reads owner statement data from Visual FoxPro DBF files:
- Primary data in `.dbf` files
- Memo fields in companion `.fpt` files (if present)
- Uses `go-dbase` library for reading

### Supported Operations
1. **File Discovery**: Automatic detection of statement files
2. **Data Reading**: Parse DBF structure and records
3. **PDF Generation**: Convert DBF data to formatted PDFs (in progress)

### Error Handling
- Graceful handling of missing directories
- Clear error messages for access issues
- Fallback for missing or corrupted files

## Sample Files
Reference samples available in:
`docs/code references/accountingManager/ownerstatements/`
- `ownerstatement-sample1.pdf`
- `ownerstatement-sample2.pdf`
- `ownerstatement-sample3.pdf`
- `dmrostmt2a.fr2` (FoxPro report format)

## Current Status

### Completed
✅ UI Component implementation
✅ Directory scanning and file discovery
✅ DBF file reading capability with proper row conversion
✅ File metadata display
✅ Platform-specific path handling
✅ Error handling and user feedback
✅ Fixed DBF row format conversion ([][]interface{} to []map[string]interface{})
✅ **Interactive Owner Statement Viewer** (NEW)
  - Owner selection dropdown
  - Individual owner statement viewing
  - Summary cards with totals
  - Detailed line item display
  - Raw data viewer

### In Progress
🔄 PDF generation from DBF data
🔄 Formatting based on sample PDFs
🔄 Batch PDF generation

### Recent Updates (August 2025)

#### Bug Fix: DBF Reading
**Issue**: DBF files showing 0 records despite being 32MB+
**Root Cause**: Data format mismatch - `ReadDBFFile` returns rows as `[][]interface{}` (array of arrays) but `GenerateOwnerStatementPDF` expected `[]map[string]interface{}` (array of maps)
**Solution**: Added proper type conversion in `GenerateOwnerStatementPDF` to handle both formats and convert array rows to map format using column names as keys

#### New Feature: Interactive Statement Viewer
**Added**: Complete interactive viewer for owner statements
**Components**:
- `OwnerStatementViewer.tsx` - Full-featured viewer component
- `GetOwnersList()` - Backend function to get unique owners
- `GetOwnerStatementData()` - Backend function to get owner-specific data
**Features**:
- Select owner from dropdown list
- View statement summary with calculated totals
- Browse detailed line items
- See raw data in table format
- Automatic field formatting (currency, dates, etc.)

### Future Enhancements
- Email distribution capabilities
- Custom PDF templates
- Batch operations for multiple owners
- Historical statement archives
- Export to other formats (CSV, Excel)

## Integration Points

### With Existing Systems
- Uses same DBF reading infrastructure as other modules
- Follows established UI/UX patterns
- Integrates with company data path resolution
- Uses standard permission system (read access required)

### Database Tables
Currently reads from DBF files only. No SQLite persistence yet.

## Security Considerations
- Read-only access to DBF files
- No modification of source data
- PDF generation in memory (no temp files)
- Respects user permissions

## Performance Notes
- DBF files loaded on-demand
- No caching currently implemented
- PDF generation performance depends on record count
- Efficient for typical statement files (<1000 records)

## Developer Notes

### Adding PDF Generation
The framework is in place for PDF generation. To complete:
1. Analyze DBF column structure from actual data
2. Map fields to PDF layout based on samples
3. Implement formatting logic in `GenerateOwnerStatementPDF`
4. Use `gofpdf` library for PDF creation (already in project)

### Testing
- Test with various DBF file sizes
- Verify path resolution on different platforms
- Ensure proper error handling for edge cases
- Validate PDF output against samples

## Related Documentation
- See `CLAUDE.md` for general project structure
- Check sample PDFs for expected output format
- Review DBF field mappings when available

---

**Last Updated**: August 2025
**Status**: Core functionality complete, PDF generation in progress