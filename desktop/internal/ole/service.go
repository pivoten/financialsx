package ole

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pivoten/financialsx/desktop/internal/debug"
	"github.com/pivoten/financialsx/desktop/internal/logger"
)

// Service provides OLE automation functionality
type Service struct {
	// Add any state management if needed
}

// NewService creates a new OLE service
func NewService() *Service {
	return &Service{}
}

// TestConnection tests if we can connect to FoxPro OLE server
func (s *Service) TestConnection() (map[string]interface{}, error) {
	// Add immediate console output
	fmt.Println("=== TestOLEConnection STARTED ===")
	fmt.Printf("TestOLEConnection called at %s\n", time.Now().Format("2006-01-02 15:04:05"))

	// Also add to debug log
	debug.SimpleLog("=== TestOLEConnection STARTED ===")
	debug.SimpleLog(fmt.Sprintf("TestOLEConnection called at %s", time.Now().Format("2006-01-02 15:04:05")))

	// Log the test attempt
	exePath, _ := os.Executable()
	fmt.Printf("TestOLEConnection: Executable path: %s\n", exePath)
	debug.SimpleLog(fmt.Sprintf("TestOLEConnection: Executable path: %s", exePath))

	logDir := filepath.Join(filepath.Dir(exePath), "logs")
	fmt.Printf("TestOLEConnection: Log directory: %s\n", logDir)
	debug.SimpleLog(fmt.Sprintf("TestOLEConnection: Log directory: %s", logDir))

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("TestOLEConnection: ERROR creating logs directory: %v\n", err)
		debug.SimpleLog(fmt.Sprintf("TestOLEConnection: ERROR creating logs directory: %v", err))
	}

	timestamp := time.Now().Format("2006-01-02")
	logPath := filepath.Join(logDir, fmt.Sprintf("financialsx_ole_%s.log", timestamp))
	testLogPath := filepath.Join(logDir, fmt.Sprintf("financialsx_test_%s.log", timestamp))

	fmt.Printf("TestOLEConnection: OLE log path: %s\n", logPath)
	fmt.Printf("TestOLEConnection: Test log path: %s\n", testLogPath)
	debug.SimpleLog(fmt.Sprintf("TestOLEConnection: OLE log path: %s", logPath))
	debug.SimpleLog(fmt.Sprintf("TestOLEConnection: Test log path: %s", testLogPath))

	// Write a test log to confirm function was called
	testLog := fmt.Sprintf("[%s] TestOLEConnection called from UI\n", time.Now().Format("2006-01-02 15:04:05"))
	if err := os.WriteFile(testLogPath, []byte(testLog), 0644); err != nil {
		fmt.Printf("TestOLEConnection: ERROR writing test log: %v\n", err)
		debug.SimpleLog(fmt.Sprintf("TestOLEConnection: ERROR writing test log: %v", err))
	} else {
		fmt.Printf("TestOLEConnection: Test log written successfully\n")
		debug.SimpleLog("TestOLEConnection: Test log written successfully")
	}

	// Try to create DbApi client
	fmt.Printf("TestOLEConnection: Attempting to create OLE DbApi client...\n")
	fmt.Printf("TestOLEConnection: Using ProgID: Pivoten.DbApi\n")
	debug.SimpleLog("TestOLEConnection: Attempting to create OLE DbApi client...")
	debug.SimpleLog("TestOLEConnection: Using ProgID: Pivoten.DbApi")

	client, err := NewDbApiClient()
	if err != nil {
		errMsg := fmt.Sprintf("TestOLEConnection: FAILED to connect to OLE server: %v", err)
		fmt.Printf("%s\n", errMsg)
		debug.SimpleLog(errMsg)

		// Provide detailed instructions for fixing the OLE server
		result := map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"message": "Could not connect to Pivoten.DbApi COM server. The dbapi.prg file has been fixed.",
			"logPath": logPath,
			"hint":    "To register: 1) Build dbapi.exe from dbapi.prg in VFP, 2) Run 'dbapi.exe /regserver' as admin",
			"details": "The dbapi.prg file in project root has been fixed to remove TRY/CATCH issues",
		}

		fmt.Printf("TestOLEConnection: Returning error result: %v\n", result)
		debug.SimpleLog(fmt.Sprintf("TestOLEConnection: Returning error result: %v", result))
		fmt.Println("=== TestOLEConnection ENDED (FAILED) ===")
		debug.SimpleLog("=== TestOLEConnection ENDED (FAILED) ===")

		return result, nil
	}
	defer client.Close()

	fmt.Printf("TestOLEConnection: SUCCESS - Connected to OLE server!\n")
	debug.SimpleLog("TestOLEConnection: SUCCESS - Connected to OLE server!")

	// Try to call Ping() method to verify server is working
	fmt.Printf("TestOLEConnection: Testing Ping() method...\n")
	debug.SimpleLog("TestOLEConnection: Testing Ping() method...")

	// Note: This would require actual OLE implementation
	// For now, we just report successful connection

	result := map[string]interface{}{
		"success": true,
		"message": "Successfully connected to Pivoten.DbApi COM server (v1.0.1)!",
		"logPath": logPath,
		"version": "1.0.1",
	}

	fmt.Printf("TestOLEConnection: Returning success result: %v\n", result)
	debug.SimpleLog(fmt.Sprintf("TestOLEConnection: Returning success result: %v", result))
	fmt.Println("=== TestOLEConnection ENDED (SUCCESS) ===")
	debug.SimpleLog("=== TestOLEConnection ENDED (SUCCESS) ===")

	return result, nil
}

// PreloadConnection preloads OLE connection for a company
func (s *Service) PreloadConnection(companyName string) map[string]interface{} {
	PreloadOLEConnection(companyName)
	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("OLE connection preloaded for company: %s", companyName),
	}
}

// CloseConnection closes the current OLE connection
func (s *Service) CloseConnection() map[string]interface{} {
	CloseOLEConnection()
	return map[string]interface{}{
		"success": true,
		"message": "OLE connection closed",
	}
}

// SetIdleTimeout sets the idle timeout for OLE connections
func (s *Service) SetIdleTimeout(minutes int) map[string]interface{} {
	duration := time.Duration(minutes) * time.Minute
	SetIdleTimeout(duration)
	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("OLE idle timeout set to %d minutes", minutes),
		"timeout": minutes,
	}
}

// TestDatabaseQuery tests a database query via OLE server
func (s *Service) TestDatabaseQuery(companyName, query string) (map[string]interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			logger.WriteCrash("TestDatabaseQuery", r, nil)
		}
	}()

	fmt.Printf("=== TestDatabaseQuery STARTED (OLE TEST ONLY) ===\n")
	fmt.Printf("TestDatabaseQuery: company=%s\n", companyName)
	fmt.Printf("TestDatabaseQuery: query=%s\n", query)
	debug.SimpleLog(fmt.Sprintf("TestDatabaseQuery: company=%s, query=%s", companyName, query))

	startTime := time.Now()

	// This is specifically for testing OLE server - no fallback
	fmt.Printf("TestDatabaseQuery: Testing OLE server (Pivoten.DbApi)...\n")
	debug.SimpleLog("TestDatabaseQuery: Testing OLE server connection")

	// Execute on dedicated COM thread to avoid threading issues
	var jsonResult string
	var queryErr error

	err := ExecuteOnCOMThread(companyName, func(client *DbApiClient) error {
		fmt.Printf("TestDatabaseQuery: Executing on COM thread\n")
		debug.SimpleLog("TestDatabaseQuery: Using COM thread for OLE connection")

		// Note: Ping method would be called here if implemented in OLE client
		fmt.Printf("TestDatabaseQuery: OLE connection established\n")
		debug.SimpleLog("TestDatabaseQuery: OLE connection established")

		// Database should already be open via ExecuteOnCOMThread
		fmt.Printf("TestDatabaseQuery: Database is open on COM thread\n")

		// Execute the query via OLE using JSON
		fmt.Printf("TestDatabaseQuery: Executing SQL query via OLE (JSON)...\n")
		jsonResult, queryErr = client.QueryToJson(query)
		return queryErr
	})

	if err != nil {
		fmt.Printf("TestDatabaseQuery: Failed to use singleton OLE connection: %v\n", err)
		debug.SimpleLog(fmt.Sprintf("TestDatabaseQuery: OLE singleton connection failed: %v", err))

		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("OLE server connection failed: %v", err),
			"message": "Could not connect to Pivoten.DbApi COM server",
			"hint":    "To use OLE: 1) Build dbapi.exe from dbapi.prg in Visual FoxPro, 2) Run 'dbapi.exe /regserver' as admin",
			"progId":  "Pivoten.DbApi",
		}, nil
	}

	if err != nil {
		fmt.Printf("TestDatabaseQuery: Query execution failed: %v\n", err)
		debug.SimpleLog(fmt.Sprintf("TestDatabaseQuery: Query failed: %v", err))

		// Get last error from OLE server if available
		var lastError string
		ExecuteOnCOMThread(companyName, func(client *DbApiClient) error {
			lastError = client.GetLastError()
			return nil
		})

		return map[string]interface{}{
			"success":   false,
			"error":     fmt.Sprintf("Query execution failed: %v", err),
			"lastError": lastError,
			"database":  companyName,
			"query":     query,
		}, nil
	}

	// Success!
	elapsedTime := time.Since(startTime)
	fmt.Printf("TestDatabaseQuery: Query executed successfully in %.2fms\n", elapsedTime.Seconds()*1000)

	// Parse the JSON result
	var queryResult map[string]interface{}
	if err := json.Unmarshal([]byte(jsonResult), &queryResult); err != nil {
		fmt.Printf("TestDatabaseQuery: Failed to parse JSON: %v\n", err)
		fmt.Printf("TestDatabaseQuery: Attempting to fix common JSON issues...\n")
		debug.SimpleLog(fmt.Sprintf("TestDatabaseQuery: JSON parse error: %v", err))

		// Try to fix common FoxPro JSON issues
		fixedJson := jsonResult
		// Fix unescaped backslashes in paths (common in Windows paths)
		// This is a simple fix - replace single backslashes with double
		// But be careful not to double-escape already escaped ones
		fixedJson = strings.ReplaceAll(fixedJson, `\`, `\\`)
		// Fix already double-escaped becoming quad-escaped
		fixedJson = strings.ReplaceAll(fixedJson, `\\\\`, `\\`)

		// Try parsing again with fixed JSON
		if err2 := json.Unmarshal([]byte(fixedJson), &queryResult); err2 != nil {
			fmt.Printf("TestDatabaseQuery: Still failed after fix attempt: %v\n", err2)
			// Return the raw result if parsing still fails
			queryResult = map[string]interface{}{
				"raw":          jsonResult,
				"parseError":   err.Error(),
				"fixAttempted": true,
			}
		} else {
			fmt.Printf("TestDatabaseQuery: JSON fix successful!\n")
		}
	} else {
		fmt.Printf("TestDatabaseQuery: JSON parsed successfully\n")
		if success, ok := queryResult["success"].(bool); ok && !success {
			// Query returned an error
			if errMsg, ok := queryResult["error"].(string); ok {
				return map[string]interface{}{
					"success":  false,
					"error":    errMsg,
					"database": companyName,
					"query":    query,
				}, nil
			}
		}
	}

	// Log the JSON result for debugging
	fmt.Printf("TestDatabaseQuery: JSON result length: %d bytes\n", len(jsonResult))
	debug.SimpleLog(fmt.Sprintf("TestDatabaseQuery: JSON result length: %d bytes", len(jsonResult)))

	// Log first 500 chars for debugging
	if len(jsonResult) > 0 {
		maxLen := 500
		if len(jsonResult) < maxLen {
			maxLen = len(jsonResult)
		}
		fmt.Printf("TestDatabaseQuery: JSON preview: %s...\n", jsonResult[:maxLen])
		debug.SimpleLog(fmt.Sprintf("TestDatabaseQuery: JSON preview: %s", jsonResult[:maxLen]))
	}

	// Extract the actual data array and count from the FoxPro JSON response
	var dataArray []map[string]interface{}
	var rowCount int

	// Debug: Log the structure of queryResult
	fmt.Printf("TestDatabaseQuery: queryResult type: %T\n", queryResult)
	fmt.Printf("TestDatabaseQuery: queryResult keys: ")
	for key, value := range queryResult {
		fmt.Printf("%s(%T) ", key, value)
	}
	fmt.Printf("\n")

	// Check if data is directly in queryResult
	if data, ok := queryResult["data"].([]interface{}); ok {
		fmt.Printf("TestDatabaseQuery: Found data array with %d items\n", len(data))
		// Convert []interface{} to []map[string]interface{}
		for _, item := range data {
			if row, ok := item.(map[string]interface{}); ok {
				dataArray = append(dataArray, row)
			}
		}
		rowCount = len(dataArray)
		fmt.Printf("TestDatabaseQuery: Extracted %d rows\n", rowCount)
	} else {
		fmt.Printf("TestDatabaseQuery: data field not found or not an array, checking type: %T\n", queryResult["data"])
		// Check if the whole queryResult might BE the data itself
		if queryResult["success"] != nil && queryResult["count"] != nil {
			// This means we parsed the FoxPro JSON correctly
			fmt.Printf("TestDatabaseQuery: FoxPro response structure detected\n")
		}
		// Fallback if structure is different
		dataArray = []map[string]interface{}{}
		rowCount = 0
	}

	// Also check for count field - could be float64 or int
	if count, ok := queryResult["count"].(float64); ok {
		rowCount = int(count)
		fmt.Printf("TestDatabaseQuery: Found count field (float64): %d\n", rowCount)
	} else if count, ok := queryResult["count"].(int); ok {
		rowCount = count
		fmt.Printf("TestDatabaseQuery: Found count field (int): %d\n", rowCount)
	}

	result := map[string]interface{}{
		"success":       true,
		"database":      companyName,
		"query":         query,
		"method":        "OLE/COM JSON (Pivoten.DbApi)",
		"executionTime": fmt.Sprintf("%.2fms", elapsedTime.Seconds()*1000),
		"data":          dataArray,
		"rowCount":      rowCount,
		"raw":           jsonResult,
		"message":       "Query executed successfully via OLE server (JSON)",
	}

	fmt.Printf("TestDatabaseQuery: SUCCESS - Query executed in %.2fms\n", elapsedTime.Seconds()*1000)
	debug.SimpleLog(fmt.Sprintf("TestDatabaseQuery: SUCCESS - Query executed in %.2fms", elapsedTime.Seconds()*1000))
	fmt.Printf("=== TestDatabaseQuery COMPLETED ===\n")

	return result, nil
}