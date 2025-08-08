package ole

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// writeLog writes debug information to a log file
func writeLog(message string) {
	// Get executable directory
	exePath, _ := os.Executable()
	logDir := filepath.Join(filepath.Dir(exePath), "logs")
	os.MkdirAll(logDir, 0755)
	
	dateStamp := time.Now().Format("2006-01-02")
	logPath := filepath.Join(logDir, fmt.Sprintf("financialsx_ole_%s.log", dateStamp))
	
	// Open or create log file
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	
	// Write timestamp and message
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	file.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, message))
}

// getCurrentDirectory gets the current working directory
func getCurrentDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}

// XMLResult represents the FoxPro XML cursor result
type XMLResult struct {
	XMLName xml.Name `xml:"VFPData"`
	Rows    []map[string]interface{} `xml:">"` // Will parse dynamically
}

// DbApiClient handles OLE automation with Pivoten.DbApi COM server
type DbApiClient struct {
	oleObject *ole.IDispatch
	dbPath    string
}

// NewDbApiClient creates a new connection to the Pivoten.DbApi COM server
func NewDbApiClient() (*DbApiClient, error) {
	// Recover from any panics
	defer func() {
		if r := recover(); r != nil {
			writeLog(fmt.Sprintf("PANIC RECOVERED in NewDbApiClient: %v", r))
		}
	}()
	
	writeLog("=== Starting Pivoten.DbApi OLE Connection ===")
	writeLog(fmt.Sprintf("Running on: %s", os.Getenv("COMPUTERNAME")))
	writeLog(fmt.Sprintf("User: %s", os.Getenv("USERNAME")))
	writeLog(fmt.Sprintf("Working directory: %s", getCurrentDirectory()))
	
	// Initialize COM
	writeLog("Initializing COM...")
	err := ole.CoInitialize(0)
	if err != nil {
		// Check if already initialized
		if err.Error() == "Incorrect function." || err.Error() == "CoInitialize has already been called." {
			writeLog("COM was already initialized, continuing...")
		} else {
			errMsg := fmt.Sprintf("Failed to initialize COM: %v", err)
			writeLog(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
	} else {
		writeLog("COM initialized successfully")
	}

	// Try to connect to Pivoten.DbApi
	serverName := "Pivoten.DbApi"
	writeLog(fmt.Sprintf("Attempting to create OLE object '%s'...", serverName))
	
	unknown, err := oleutil.CreateObject(serverName)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "Invalid class string") || strings.Contains(errStr, "Class not registered") {
			writeLog(fmt.Sprintf("'%s' is not registered. Please run: pivoten.exe /regserver", serverName))
		} else {
			writeLog(fmt.Sprintf("Failed to create object '%s': %v", serverName, err))
		}
		return nil, fmt.Errorf("failed to create Pivoten.DbApi: %w", err)
	}
	
	writeLog("Successfully created Pivoten.DbApi object")

	// Get IDispatch interface
	writeLog("Getting IDispatch interface...")
	oleObject, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to get IDispatch interface: %v", err)
		writeLog(errMsg)
		return nil, fmt.Errorf(errMsg)
	}
	
	writeLog("IDispatch interface obtained successfully")
	
	// Test the connection with Ping
	writeLog("Testing connection with Ping()...")
	result, err := oleutil.CallMethod(oleObject, "Ping")
	if err != nil {
		writeLog(fmt.Sprintf("Ping failed: %v", err))
	} else {
		pingResult := result.ToString()
		writeLog(fmt.Sprintf("Ping successful: %s", pingResult))
	}
	
	writeLog("=== Pivoten.DbApi OLE Connection Successful ===")

	return &DbApiClient{
		oleObject: oleObject,
	}, nil
}

// Initialize sets the base folder for the DbApi (optional)
func (c *DbApiClient) Initialize(rootPath string) error {
	writeLog(fmt.Sprintf("Initializing DbApi with root: %s", rootPath))
	_, err := oleutil.CallMethod(c.oleObject, "Initialize", rootPath)
	if err != nil {
		writeLog(fmt.Sprintf("Initialize failed: %v", err))
		return fmt.Errorf("failed to initialize: %w", err)
	}
	writeLog("Initialize successful")
	return nil
}

// OpenDbc opens a database container by path
func (c *DbApiClient) OpenDbc(dbcPath string) error {
	// Recover from any panics
	defer func() {
		if r := recover(); r != nil {
			writeLog(fmt.Sprintf("PANIC RECOVERED in OpenDbc: %v", r))
		}
	}()
	
	// Check if oleObject is valid
	if c == nil || c.oleObject == nil {
		writeLog("ERROR: DbApiClient or oleObject is nil")
		return fmt.Errorf("DbApi client not initialized")
	}
	
	writeLog(fmt.Sprintf("Opening DBC: %s", dbcPath))
	result, err := oleutil.CallMethod(c.oleObject, "OpenDbc", dbcPath)
	if err != nil {
		writeLog(fmt.Sprintf("OpenDbc failed: %v", err))
		return fmt.Errorf("failed to open DBC: %w", err)
	}
	
	success := result.Value().(bool)
	if !success {
		lastError := c.GetLastError()
		writeLog(fmt.Sprintf("OpenDbc returned false: %s", lastError))
		return fmt.Errorf("failed to open DBC: %s", lastError)
	}
	
	c.dbPath = dbcPath
	writeLog("OpenDbc successful")
	return nil
}

// SelectToXML executes a SELECT query and returns XML results
func (c *DbApiClient) SelectToXML(sql string) (string, error) {
	// Recover from any panics
	defer func() {
		if r := recover(); r != nil {
			writeLog(fmt.Sprintf("PANIC RECOVERED in SelectToXML: %v", r))
		}
	}()
	
	writeLog(fmt.Sprintf("Executing SELECT: %s", sql))
	
	// Check if oleObject is valid
	if c == nil || c.oleObject == nil {
		writeLog("ERROR: DbApiClient or oleObject is nil")
		return "", fmt.Errorf("DbApi client not initialized")
	}
	
	result, err := oleutil.CallMethod(c.oleObject, "SelectToXml", sql)
	if err != nil {
		writeLog(fmt.Sprintf("SelectToXml failed: %v", err))
		return "", fmt.Errorf("failed to execute SELECT: %w", err)
	}
	
	xmlData := result.ToString()
	if xmlData == "" {
		lastError := c.GetLastError()
		if lastError != "" {
			writeLog(fmt.Sprintf("SelectToXml error: %s", lastError))
			return "", fmt.Errorf("SELECT failed: %s", lastError)
		}
	}
	
	writeLog(fmt.Sprintf("SelectToXml returned %d bytes", len(xmlData)))
	return xmlData, nil
}

// QueryToJson executes a SELECT query and returns results as JSON
func (c *DbApiClient) QueryToJson(sql string) (string, error) {
	if c == nil || c.oleObject == nil {
		return "", fmt.Errorf("DbApi client not initialized")
	}
	
	writeLog(fmt.Sprintf("Calling QueryToJson with SQL: %s", sql))
	
	result, err := oleutil.CallMethod(c.oleObject, "QueryToJson", sql)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to execute QueryToJson: %v", err)
		writeLog(errMsg)
		return "", fmt.Errorf(errMsg)
	}
	
	if result.Value() == nil {
		writeLog("QueryToJson returned nil")
		return "{}", nil
	}
	
	jsonResult := result.ToString()
	writeLog(fmt.Sprintf("QueryToJson returned %d bytes", len(jsonResult)))
	
	return jsonResult, nil
}

// GetTableListSimple returns a JSON array of table names
func (c *DbApiClient) GetTableListSimple() (string, error) {
	if c == nil || c.oleObject == nil {
		return "[]", fmt.Errorf("DbApi client not initialized")
	}
	
	writeLog("Calling GetTableListSimple")
	
	result, err := oleutil.CallMethod(c.oleObject, "GetTableListSimple")
	if err != nil {
		errMsg := fmt.Sprintf("Failed to execute GetTableListSimple: %v", err)
		writeLog(errMsg)
		return "[]", fmt.Errorf(errMsg)
	}
	
	if result.Value() == nil {
		writeLog("GetTableListSimple returned nil")
		return "[]", nil
	}
	
	tableList := result.ToString()
	writeLog(fmt.Sprintf("GetTableListSimple returned: %s", tableList))
	
	return tableList, nil
}

// GetTableCount returns a simple count of tables in the DBC
func (c *DbApiClient) GetTableCount() (string, error) {
	if c == nil || c.oleObject == nil {
		return "", fmt.Errorf("DbApi client not initialized")
	}
	
	writeLog("Calling GetTableCount")
	
	result, err := oleutil.CallMethod(c.oleObject, "GetTableCount")
	if err != nil {
		errMsg := fmt.Sprintf("Failed to execute GetTableCount: %v", err)
		writeLog(errMsg)
		return "", fmt.Errorf(errMsg)
	}
	
	if result.Value() == nil {
		writeLog("GetTableCount returned nil")
		return "0", nil
	}
	
	count := result.ToString()
	writeLog(fmt.Sprintf("GetTableCount returned: %s", count))
	
	return count, nil
}

// ExecNonQuery executes INSERT/UPDATE/DELETE statements
func (c *DbApiClient) ExecNonQuery(sql string) error {
	writeLog(fmt.Sprintf("Executing non-query: %s", sql))
	result, err := oleutil.CallMethod(c.oleObject, "ExecNonQuery", sql)
	if err != nil {
		writeLog(fmt.Sprintf("ExecNonQuery failed: %v", err))
		return fmt.Errorf("failed to execute non-query: %w", err)
	}
	
	success := result.Value().(bool)
	if !success {
		lastError := c.GetLastError()
		writeLog(fmt.Sprintf("ExecNonQuery returned false: %s", lastError))
		return fmt.Errorf("non-query failed: %s", lastError)
	}
	
	writeLog("ExecNonQuery successful")
	return nil
}

// GetLastError returns the last error from DbApi
func (c *DbApiClient) GetLastError() string {
	result, err := oleutil.CallMethod(c.oleObject, "GetLastError")
	if err != nil {
		return fmt.Sprintf("failed to get last error: %v", err)
	}
	return result.ToString()
}

// IsDbcOpen checks if a database is currently open
func (c *DbApiClient) IsDbcOpen() bool {
	result, err := oleutil.CallMethod(c.oleObject, "IsDbcOpen")
	if err != nil {
		return false
	}
	return result.Value().(bool)
}

// GetDbcPath returns the current DBC path
func (c *DbApiClient) GetDbcPath() string {
	result, err := oleutil.CallMethod(c.oleObject, "GetDbcPath")
	if err != nil {
		return ""
	}
	return result.ToString()
}

// CloseDbc closes the current database
func (c *DbApiClient) CloseDbc() error {
	_, err := oleutil.CallMethod(c.oleObject, "CloseDbc")
	if err != nil {
		return fmt.Errorf("failed to close DBC: %w", err)
	}
	c.dbPath = ""
	return nil
}

// Close releases the OLE object and terminates the server process
func (c *DbApiClient) Close() {
	if c.oleObject != nil {
		// Try to close DBC first
		c.CloseDbc()
		
		// Don't call Quit when using connection pool - we want to reuse connections
		// Only call Quit if this is a standalone connection (not from pool)
		// The pool will manage process lifecycle
		
		// Release the OLE object
		c.oleObject.Release()
		c.oleObject = nil
		
		// Don't call CoUninitialize here - it should be called once per thread
		// ole.CoUninitialize()
	}
}

// ParseXMLToRows parses FoxPro XML result to rows
func ParseXMLToRows(xmlData string) ([]map[string]interface{}, error) {
	if xmlData == "" {
		return nil, nil
	}
	
	// Parse the XML data
	var result struct {
		XMLName xml.Name `xml:"VFPData"`
		Rows    []struct {
			Data map[string]interface{} `xml:",any"`
		} `xml:">"` 
	}
	
	err := xml.Unmarshal([]byte(xmlData), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}
	
	// Convert to simple map slice
	rows := make([]map[string]interface{}, 0)
	for _, row := range result.Rows {
		rows = append(rows, row.Data)
	}
	
	return rows, nil
}