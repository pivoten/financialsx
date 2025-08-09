# go-dbase Examples

Complete examples from the [go-dbase library](https://github.com/Valentin-Kaiser/go-dbase) organized by category.

## Example Categories

Based on the go-dbase repository examples directory:

### 📁 Core Operations
- **[create](./create-example.go)** - Creating new DBF tables and databases
- **[read](./read-example.go)** - Reading DBF files and records
- **[write](./write-example.go)** - Writing and modifying DBF records
- **[search](./search-example.go)** - Searching and filtering DBF data

### 🔧 Advanced Features
- **[custom](./custom-example.go)** - Custom field types and conversions
- **[database](./database-example.go)** - Working with DBC database containers
- **[schema](./schema-example.go)** - Schema inspection and manipulation
- **[documentation](./documentation-example.go)** - Generating documentation from DBF structure

### 💼 Practical Implementation
- **[practical-example.go](./practical-example.go)** - Real-world usage in FoxPro Toolkit
- **[integration-guide.md](./integration-guide.md)** - How we integrate go-dbase

## Quick Start

Each example is a complete, runnable Go program demonstrating specific features:

```bash
# Run any example
go run create-example.go
go run read-example.go
go run search-example.go
# etc.
```

## Test Data

The examples reference test data from `test_data/` directory which would typically contain:
- Sample DBF files
- FPT memo files  
- DBC database containers
- Various encoding examples

## Building Examples

Use the Makefile from the go-dbase repository:

```bash
# Build all examples
make examples

# Run tests
make test

# Clean build artifacts
make clean
```

## Example Highlights

### Create Example
Shows how to:
- Define table schema
- Create DBF files with different versions
- Add various column types
- Handle memo fields

### Read Example  
Demonstrates:
- Opening existing DBF files
- Iterating through records
- Handling deleted records
- Reading into Go structs

### Write Example
Covers:
- Adding new records
- Modifying existing data
- Soft delete operations
- Saving changes

### Search Example
Illustrates:
- Implementing search filters
- Case-insensitive searching
- Field-specific searches
- Complex query conditions

### Custom Example
Shows:
- Custom type conversions
- Field validators
- Data transformations
- Extended field types

### Database Example
Demonstrates:
- Opening DBC containers
- Working with multiple tables
- Table relationships
- Database metadata

### Schema Example
Covers:
- Inspecting table structure
- Column metadata
- Index information
- Schema modification

### Documentation Example
Shows:
- Generating HTML docs
- Creating data dictionaries
- Exporting schema info
- Auto-documentation

## Notes

- All examples use the latest go-dbase API
- Error handling is included but simplified for clarity
- Examples can be adapted for production use
- Test data files are not included in this repository

For the complete go-dbase examples with test data, see:
https://github.com/Valentin-Kaiser/go-dbase/tree/main/examples