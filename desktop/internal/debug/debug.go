package debug

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var (
	debugMutex sync.Mutex
	debugFile  *os.File
	debugPath  string
)

// SimpleLog writes a debug message directly to a file in the executable directory
// This is a fallback logging mechanism for Windows debugging
func SimpleLog(message string) {
	debugMutex.Lock()
	defer debugMutex.Unlock()
	
	// Get executable directory
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("DEBUG: Failed to get exe path: %v\n", err)
		return
	}
	exeDir := filepath.Dir(exePath)
	
	// Use a simple debug file in the exe directory
	if debugPath == "" {
		timestamp := time.Now().Format("2006-01-02")
		debugPath = filepath.Join(exeDir, fmt.Sprintf("debug_%s.txt", timestamp))
	}
	
	// Open or create the debug file
	if debugFile == nil {
		debugFile, err = os.OpenFile(debugPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			// If we can't open in exe dir, try temp dir
			tempPath := filepath.Join(os.TempDir(), fmt.Sprintf("financialsx_debug_%s.txt", time.Now().Format("2006-01-02")))
			debugFile, err = os.OpenFile(tempPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("DEBUG: Failed to open any debug file: %v\n", err)
				return
			}
			debugPath = tempPath
		}
		
		// Write header
		debugFile.WriteString(fmt.Sprintf("\n=== FinancialsX Debug Log Started %s ===\n", time.Now().Format("2006-01-02 15:04:05")))
		debugFile.WriteString(fmt.Sprintf("OS: %s, Arch: %s\n", runtime.GOOS, runtime.GOARCH))
		debugFile.WriteString(fmt.Sprintf("Exe Dir: %s\n", exeDir))
		debugFile.WriteString(fmt.Sprintf("Debug File: %s\n\n", debugPath))
	}
	
	// Write the message
	timestamp := time.Now().Format("15:04:05.000")
	logEntry := fmt.Sprintf("[%s] %s\n", timestamp, message)
	debugFile.WriteString(logEntry)
	debugFile.Sync() // Force immediate write
	
	// Also print to console
	fmt.Print("DEBUG: " + logEntry)
}

// LogError logs an error with context
func LogError(context string, err error) {
	if err != nil {
		SimpleLog(fmt.Sprintf("ERROR in %s: %v", context, err))
	}
}

// LogInfo logs an informational message
func LogInfo(context string, message string) {
	SimpleLog(fmt.Sprintf("INFO [%s]: %s", context, message))
}

// Close closes the debug file
func Close() {
	debugMutex.Lock()
	defer debugMutex.Unlock()
	
	if debugFile != nil {
		debugFile.WriteString(fmt.Sprintf("\n=== Debug Log Closed %s ===\n", time.Now().Format("2006-01-02 15:04:05")))
		debugFile.Close()
		debugFile = nil
	}
}

// GetDebugPath returns the current debug file path
func GetDebugPath() string {
	return debugPath
}