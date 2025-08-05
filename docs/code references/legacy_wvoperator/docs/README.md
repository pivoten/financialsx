We are converting a FoxPro process to Go, which will read tables and create data based on the processing defined in wvannualretcreate.txt file.  Any references in FoxPro that uses m.goApp.{variable} is created earlier and read from a database table.

Table references are in foxpro_table.csv
Table - Field references are in foxpro_table_fields.csv

The output of this program eventually needs to be:


1. PDF form
2. Excel or CSV output
3. DBF file for updating the main external legacy app.


# go-dbase-main

A Go library for reading and processing DBF (dBASE/FoxPro) files. Useful for legacy data integration, reporting, and migration tasks.

---

## Features

- Open and inspect DBF files
- Read all records as Go maps
- Access field metadata (names, types, lengths)
- Safe conversion helpers for DBF field values
- Example code for common use cases
- Advanced options for encoding, error handling, and field mapping

---

## Installation

```sh
go get github.com/Valentin-Kaiser/go-dbase
```

---

## Basic Usage

### Opening a DBF File

```go
import "github.com/Valentin-Kaiser/go-dbase/dbase"

db, err := dbase.OpenDatabase(dbase.NewConfig().WithFilepath("path/to/file.DBF"))
if err != nil {
    // handle error
}
defer db.Close()
```

### Inspecting File Metadata

```go
header := db.Header()
fmt.Printf("Records: %d, Fields: %d\n", header.RecordsCount(), header.FieldsCount())
for _, field := range header.Fields() {
    fmt.Printf("Name: %s, Type: %c, Length: %d\n", field.Name(), field.Type(), field.Length())
}
```

### Reading All Records

```go
records, err := db.AllRecordsAsMap()
if err != nil {
    // handle error
}
for _, record := range records {
    fmt.Println(record) // map[string]interface{}
}
```

### Safe Field Access Helpers

```go
func getStringField(record map[string]interface{}, fieldName string) string {
    if val, ok := record[fieldName]; ok && val != nil {
        return fmt.Sprintf("%v", val)
    }
    return ""
}
```

---

## Example: Filtering Records

```go
for _, record := range records {
    state := getStringField(record, "CSTATE")
    if state == "WV" {
        // process West Virginia records
    }
}
```

---

## Field Types

- String: `string`
- Numeric: `float64`, `int`
- Date: `time.Time` (may require parsing)
- Boolean: `bool`

---

## Advanced Usage

### Custom Encoding

Some DBF files use non-UTF8 encodings (e.g., CP850, CP437). You can specify encoding in the config:

```go
db, err := dbase.OpenDatabase(
    dbase.NewConfig().
        WithFilepath("file.DBF").
        WithEncoding("cp850"),
)
```

### Selective Field Reading

To optimize performance, read only specific fields:

```go
fields := []string{"CSTATE", "CWELLID", "CWELLNAME"}
records, err := db.RecordsAsMap(fields)
```

### Handling Large Files

For very large DBF files, process records in batches:

```go
for i := 0; i < header.RecordsCount(); i += batchSize {
    batch, err := db.RecordsAsMapRange(i, i+batchSize)
    // process batch
}
```

### Date Parsing

DBF date fields may be strings. Use robust parsing:

```go
func getDateField(record map[string]interface{}, fieldName string) time.Time {
    // ...see main.go for full implementation...
}
```

### Boolean Handling

DBF boolean fields may be "T", "F", "Y", "N", "1", "0", etc. Use a helper:

```go
func getBoolField(record map[string]interface{}, fieldName string) bool {
    // ...see main.go for full implementation...
}
```

---

## Examples Folder

The `examples/` directory contains sample Go programs demonstrating how to use the library for various DBF operations. These are ideal for learning, testing, or adapting to your own projects.

**How to use:**
- Browse the `examples/` folder for ready-to-run code.
- Each example is self-contained and shows practical usage patterns.
- Modify the examples to fit your data and workflow.

---

## Troubleshooting

### Common Issues

- **File Not Found:**  
  Ensure the DBF file path is correct and the file exists.

- **Encoding Errors:**  
  If you see garbled text, set the correct encoding in the config.

- **Field Name Mismatch:**  
  DBF field names are often uppercase and may be truncated. Use exact names.

- **Date/Type Conversion Errors:**  
  Use helper functions to safely convert types.

- **Permission Denied:**  
  Make sure you have read permissions for the DBF file.

### Debugging Tips

- Print the DBF header and field names before processing records.
- Print sample records to verify field mapping.
- Use Go’s error handling idioms (`if err != nil { ... }`).
- For large files, process in batches to avoid memory issues.

### Example Debug Output

```go
header := db.Header()
fmt.Printf("Fields: ")
for _, field := range header.Fields() {
    fmt.Printf("%s ", field.Name())
}
fmt.Println()

records, err := db.AllRecordsAsMap()
for i, record := range records {
    if i < 3 {
        fmt.Printf("Record #%d: %v\n", i+1, record)
    }
}
```

---

## Best Practices

- Always close DBF files with `defer db.Close()`.
- Use helper functions to safely extract and convert field values.
- Inspect field names and types before processing records.
- Refer to the `examples/` folder for real-world scenarios.
- Handle errors gracefully and log useful debug information.

---

## References

- [go-dbase GitHub](https://github.com/Valentin-Kaiser/go-dbase)
- [dBASE File Format](https://en.wikipedia.org/wiki/.dbf)

---

If you need more advanced examples or troubleshooting tips, see the `examples/` folder or reach out for help!

---
go doc github.com/Valentin-Kaiser/go-dbase/dbase
---

package dbase // import "github.com/Valentin-Kaiser/go-dbase/dbase"
This go-dbase package offers tools for managing dBase-format database files.
It supports tailored I/O operations for Unix and Windows platforms, provides
flexible data representations like maps, JSON, and Go structs, and ensures safe
concurrent operations with built-in mutex locks.
The package facilitates defining, manipulating, and querying columns and rows
in dBase tables, converting between dBase-specific data types and Go data types,
and performing systematic error handling.
Typical use cases include data retrieval from legacy dBase systems, conversion
of dBase files to modern formats, and building applications that interface with
dBase databases.
const MaxColumnNameLength = 10 ...
var ErrEOF = errors.New("EOF") ...
func Debug(enabled bool, out io.Writer)
func RegisterCustomEncoding(codePageMark byte, encoding encoding.Encoding)
func ValidateFileVersion(version byte, untested bool) error
type Column struct{ ... }
    func NewColumn(name string, dataType DataType, length uint8, decimals uint8, nullable bool) (*Column, error)
type ColumnFlag byte
    const HiddenFlag ColumnFlag = 0x01 ...
type Config struct{ ... }
type DataType byte
    const Character DataType = 0x43 ...
type Database struct{ ... }
    func OpenDatabase(config *Config) (*Database, error)
type DefaultConverter struct{ ... }
    func ConverterFromCodePage(codePageMark byte) DefaultConverter
    func NewDefaultConverter(encoding encoding.Encoding) DefaultConverter
type EncodingConverter interface{ ... }
type Error struct{ ... }
    func NewError(err string) Error
    func NewErrorf(format string, a ...interface{}) Error
    func WrapError(err error) Error
type Field struct{ ... }
type File struct{ ... }
    func NewTable(version FileVersion, config *Config, columns []*Column, memoBlockSize uint16, ...) (*File, error)
    func OpenTable(config *Config) (*File, error)
type FileExtension string
    const DBC FileExtension = ".DBC" ...
type FileVersion byte
    const FoxPro FileVersion = 0x30 ...
    const FoxBase FileVersion = 0x02 ...
type GenericIO struct{ ... }
type Header struct{ ... }
type IO interface{ ... }
type Marker byte
    const Null Marker = 0x00 ...
type MemoHeader struct{ ... }
type Modification struct{ ... }
type Row struct{ ... }
type Table struct{ ... }
type TableFlag byte
    const StructuralFlag TableFlag = 0x01 ...
type UnixIO struct{}
    var DefaultIO UnixIO
chriscantrell@ChrisCallsM4MBP legacy_wvoperator %