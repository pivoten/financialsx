# Database Operations

This document describes the new database functionality that allows storing and editing PDF data before generation.

## Overview

The database system stores all the data that goes into the PDF forms in a SQLite database, allowing for manual review and modification before generating the final PDFs.

## What We Implemented

### 🎯 **Core Problem Solved**
- **Before**: PDF generation required reading from DBF files every time
- **After**: Data is processed once, stored in database, and PDFs use database data
- **Benefit**: Manual edits are preserved and reflected in PDFs

### 🏗️ **Database Architecture**
- **SQLite Database**: `sourcedata/sql/wv_operator.db`
- **Tables Created**:
  - `operator` - Producer/company information
  - `well` - Basic well information (location, status, etc.)
  - `well_financial_accounting` - Financial data filtered by accounting dates
  - `well_financial_production` - Financial data filtered by production dates
  - `lookup` - Reference data (county codes, well status, etc.)

### 🔄 **Smart Data Flow**
1. **Check Database First**: System checks if database exists and has data
2. **Use Database Data**: If available, uses database data (including your edits)
3. **Fallback to DBF**: If no database data, falls back to reading DBF files
4. **Generate PDFs**: Creates PDFs with the best available data

### 🛠️ **Key Features Implemented**
- **Automatic Directory Creation**: Creates `sourcedata/sql/` if it doesn't exist
- **Database Existence Check**: Prompts to create database if not found
- **Year-Specific Data Management**: Clears only data for the specified production year
- **Dual Date Type Storage**: Stores both accounting and production date data automatically
- **Smart PDF Generation**: Uses database data when available, falls back to DBF files

## Database Location

The SQLite database is stored at: `sourcedata/sql/wv_operator.db`

## Database Schema

The database contains the following tables:

### 1. `operator`
Stores company/producer information (Schedule 1 data)
- `operator_id` - Primary key
- `producer_name` - Company name
- `producer_code` - Producer code
- `address`, `city`, `state`, `zip_code` - Address information
- `phone`, `email` - Contact information

### 2. `well`
Stores well information (Schedule 1 & 2 data)
- `well_id` - Primary key
- `operator_id` - Foreign key to operator
- `county_name`, `county_number` - County information
- `nra_number`, `api_number` - Well identifiers
- `well_name` - Well name
- `land_acreage`, `lease_acreage` - Acreage information
- `status_code`, `formation_code` - Well status and formation
- `initial_production_date` - Production start date

### 3. `well_financial_accounting`
Stores financial data using accounting dates
- `well_id` - Foreign key to well
- `reporting_period_year` - Year for the data
- Production totals (BBL, MCF, NGL)
- Revenue data (Oil, Gas, NGL)
- Working interest revenue
- Royalty interest revenue
- Expense data

### 4. `well_financial_production`
Stores financial data using production dates
- Same structure as `well_financial_accounting` but filtered by production dates

### 5. `lookup`
Stores lookup values for:
- Gas types (ethane, propane, butane, etc.)
- Well statuses (Active, Plugged, Shut-in, etc.)
- Formations (Marcellus, Huron, etc.)

## Technical Implementation

### 🔧 **Database Functions Created**
- `ClearFinancialDataForYear()` - Clears data for specific production year only
- `InsertWellFinancialData()` - Inserts both accounting and production data
- `GetWellFromDatabase()` - Retrieves well data for PDF generation
- `GetProducerFromDatabase()` - Retrieves producer data
- `CheckDatabaseExists()` - Checks if database has data

### 📁 **File Structure Created**
```
database.go                    # Database management functions
cmd/db_viewer/main.go          # Database viewer application
docs/DATABASE_OPERATIONS.md    # This documentation
```

### 🔄 **Updated Functions**
- `handleDatabaseOperations()` - Enhanced with directory/database checks
- `displayWellDetails()` - Now uses database data when available
- Main PDF generation - Automatically checks database first

## Usage

### 1. Creating the Database

Run the main application and select "3. Database Operations":

```bash
./wv-operator-return
```

The system will:
1. **Check if SQL directory exists** - Creates `sourcedata/sql/` if needed
2. **Check if database exists** - Prompts to create if not found
3. **Clear existing data** for the specified production year
4. **Read data from DBF files** for both accounting and production dates
5. **Populate the database** with well and financial data for both date types
6. **Store all calculations** automatically (no need to choose between accounting/production)

### 2. Viewing and Editing Data

Use the database viewer to review and modify data:

```bash
./db_viewer
```

The database viewer provides:
- **View Operator Information** - See company details
- **View Wells Summary** - See a summary of all wells
- **View Well Details** - See detailed information for a specific well
- **Edit Well Data** - Modify production, revenue, or interest data
- **Export Data** - Export data to CSV format

### 3. Editing Data

The database viewer allows you to edit:
- **Production Data**: Oil, Gas, and NGL production volumes
- **Revenue Data**: Oil, Gas, and Other revenue amounts
- **Working Interest**: Working interest revenue for each resource
- **Royalty Interest**: Royalty interest revenue for each resource

### 4. Generating PDFs from Database

After reviewing and editing the data, you can generate PDFs using the modified data. The system will automatically:

1. **Check for database data** first
2. **Use database data** if available (including your edits)
3. **Fall back to DBF files** if no database data exists
4. **Generate PDFs** with the best available data

This means your manual edits in the database will be reflected in the generated PDFs!

## Database Viewer Menu

```
=== WV Operator Database Viewer ===
1. View Operator Information
2. View Wells Summary
3. View Well Details
4. View Financial Data
5. Edit Well Data
6. Export Data
0. Exit
```

## SQL Commands

You can also use SQLite command line tools to work with the database:

```bash
# Connect to the database
sqlite3 sourcedata/sql/wv_operator.db

# View all wells
SELECT * FROM well;

# View financial data for a specific year
SELECT * FROM well_financial_accounting WHERE reporting_period_year = 2024;

# Export data to CSV
.mode csv
.headers on
SELECT * FROM well_financial_accounting > export.csv
```

## Complete Workflow

### 🔄 **Step 1: Create Database Data**
```bash
./wv-operator-return
# Select "3. Database Operations"
# Enter your production year (e.g., 2024)
```

**What happens:**
1. **Directory Check**: System checks if `sourcedata/sql/` exists
   - If not: Prompts to create directory
2. **Database Check**: System checks if `wv_operator.db` exists
   - If not: Prompts to create database
3. **Data Processing**: 
   - Clears existing data for the specified year
   - Reads DBF files for both accounting and production dates
   - Stores all calculations in database tables

### 🔍 **Step 2: Review and Edit (Optional)**
```bash
./db_viewer
# Use the menu to view and edit data
```

**What you can do:**
- View operator information
- Review well summaries
- Examine detailed well data
- Edit production, revenue, or interest data
- Export data to CSV

### 📄 **Step 3: Generate PDFs**
```bash
./wv-operator-return
# Select "1. Process Data"
# The system will automatically use database data if available
```

**Smart Data Selection:**
1. **Database Check**: System checks if database exists and has data
2. **Use Database**: If available, uses database data (including your edits)
3. **Fallback**: If no database data, falls back to DBF files
4. **Generate**: Creates PDFs with the best available data

### 🎯 **Key Benefits of This Workflow**
- **Data Persistence**: Your edits are saved and reused
- **No Re-processing**: Don't need to re-read DBF files every time
- **Flexibility**: Can edit data without changing source files
- **Reliability**: Automatic fallback ensures PDFs always generate

## Benefits

1. **Data Review**: Review all data before generating PDFs
2. **Manual Corrections**: Fix any data issues manually
3. **Audit Trail**: Track changes with timestamps
4. **Data Validation**: Ensure data accuracy before submission
5. **Flexibility**: Modify data without changing source files
6. **Automatic Fallback**: Uses database data when available, falls back to DBF files
7. **Both Date Types**: Stores both accounting and production date data automatically

## File Structure

```
sourcedata/
└── sql/
    └── wv_operator.db          # SQLite database
cmd/
└── db_viewer/
    └── main.go                 # Database viewer source
db_viewer                       # Database viewer executable
wv-operator-return              # Main application executable
```

## Troubleshooting

### Directory/Database Creation Issues

#### SQL Directory Not Found
**Error**: `❌ SQL directory not found: sourcedata/sql`

**Solution**: 
- The system will prompt you to create the directory
- Answer `y` to create it automatically
- Or manually create: `mkdir -p sourcedata/sql`

#### Database Not Found
**Error**: `❌ Database not found: sourcedata/sql/wv_operator.db`

**Solution**:
- The system will prompt you to create the database
- Answer `y` to create it automatically
- The database will be created and populated with data

### Permission Errors
If you get permission errors:
```bash
# Create directory with proper permissions
mkdir -p sourcedata/sql
chmod 755 sourcedata/sql

# Check file permissions
ls -la sourcedata/sql/
```

### Data Not Showing
If data is not showing in the viewer:
1. **Check Database Creation**: Ensure "Database Operations" completed successfully
2. **Verify Year**: Check that the reporting year matches your data
3. **Check Date Type**: Verify accounting vs production date filtering
4. **Database Integrity**: Use SQLite to check if data exists:
   ```bash
   sqlite3 sourcedata/sql/wv_operator.db
   SELECT COUNT(*) FROM well;
   SELECT COUNT(*) FROM well_financial_accounting;
   ```

### PDF Generation Issues
If PDFs are not using database data:
1. **Check Database Exists**: Ensure `sourcedata/sql/wv_operator.db` exists
2. **Verify Data**: Check that database has data for your production year
3. **Fallback Working**: System should automatically fall back to DBF files
4. **Check Logs**: Look for "Using database data" or "Falling back to DBF files" messages 