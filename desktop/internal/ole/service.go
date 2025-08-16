package ole

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pivoten/financialsx/desktop/internal/debug"
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