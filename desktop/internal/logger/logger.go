package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var (
	logFile   *os.File
	logMutex  sync.Mutex
	logPath   string
	crashPath string
)

// Initialize sets up the logging system
func Initialize() error {
	// Get executable directory
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDir := filepath.Dir(exePath)
	
	// Create logs directory
	logsDir := filepath.Join(exeDir, "logs")
	err = os.MkdirAll(logsDir, 0755)
	if err != nil {
		// Try to create in temp directory as fallback
		tempDir := os.TempDir()
		logsDir = filepath.Join(tempDir, "financialsx_logs")
		os.MkdirAll(logsDir, 0755)
	}
	
	// Clean up old logs (keep last 10 days)
	cleanupOldLogs(logsDir, 10)
	
	// Set up log files
	timestamp := time.Now().Format("2006-01-02")
	logPath = filepath.Join(logsDir, fmt.Sprintf("financialsx_%s.log", timestamp))
	crashPath = filepath.Join(logsDir, fmt.Sprintf("financialsx_crash_%s.log", timestamp))
	
	// Open main log file
	logFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Try to write to console and a simple error file
		fmt.Printf("Failed to open log file %s: %v\n", logPath, err)
		
		// Create a simple error file in exe directory
		errorFile := filepath.Join(exeDir, "log_error.txt")
		os.WriteFile(errorFile, []byte(fmt.Sprintf("Failed to create log file: %v\nAttempted path: %s\n", err, logPath)), 0644)
		return err
	}
	
	// Write startup message
	WriteInfo("Application", "Starting FinancialsX Desktop")
	WriteInfo("Application", fmt.Sprintf("Log file: %s", logPath))
	WriteInfo("Application", fmt.Sprintf("OS: %s, Arch: %s", runtime.GOOS, runtime.GOARCH))
	
	return nil
}

// Close closes the log file
func Close() {
	if logFile != nil {
		WriteInfo("Application", "Shutting down FinancialsX Desktop")
		logFile.Close()
	}
}

// WriteInfo writes an info message to the log
func WriteInfo(module, message string) {
	write("INFO", module, message)
}

// WriteError writes an error message to the log
func WriteError(module, message string) {
	write("ERROR", module, message)
}

// WriteWarning writes a warning message to the log
func WriteWarning(module, message string) {
	write("WARN", module, message)
}

// WriteDebug writes a debug message to the log
func WriteDebug(module, message string) {
	write("DEBUG", module, message)
}

// WriteCrash writes crash information to both logs
func WriteCrash(module string, err interface{}, stackTrace []byte) {
	crashMsg := fmt.Sprintf("CRASH in %s: %v", module, err)
	
	// Write to main log
	write("CRASH", module, crashMsg)
	
	// Write detailed crash info to crash log
	logMutex.Lock()
	defer logMutex.Unlock()
	
	crashFile, err2 := os.OpenFile(crashPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err2 == nil {
		defer crashFile.Close()
		
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		crashFile.WriteString(fmt.Sprintf("\n========================================\n"))
		crashFile.WriteString(fmt.Sprintf("[%s] CRASH REPORT\n", timestamp))
		crashFile.WriteString(fmt.Sprintf("Module: %s\n", module))
		crashFile.WriteString(fmt.Sprintf("Error: %v\n", err))
		crashFile.WriteString(fmt.Sprintf("Stack Trace:\n%s\n", stackTrace))
		crashFile.WriteString(fmt.Sprintf("========================================\n"))
	}
}

// write is the internal function that actually writes to the log
func write(level, module, message string) {
	logMutex.Lock()
	defer logMutex.Unlock()
	
	if logFile == nil {
		// If log file isn't open, try to write to console
		fmt.Printf("[%s] %s: %s - %s\n", time.Now().Format("15:04:05"), level, module, message)
		return
	}
	
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	logEntry := fmt.Sprintf("[%s] [%s] [%s] %s\n", timestamp, level, module, message)
	
	// Write to file
	_, writeErr := logFile.WriteString(logEntry)
	if writeErr != nil {
		fmt.Printf("Failed to write to log: %v\n", writeErr)
	}
	logFile.Sync() // Force write to disk immediately
	
	// Also print to console for debugging
	fmt.Print(logEntry)
}

// RecoverPanic recovers from a panic and logs it
func RecoverPanic(module string) {
	if r := recover(); r != nil {
		// Get stack trace
		stackBuf := make([]byte, 4096)
		stackSize := runtime.Stack(stackBuf, false)
		stackTrace := stackBuf[:stackSize]
		
		// Log the crash
		WriteCrash(module, r, stackTrace)
		
		// Try to write a user-friendly error file
		writeUserErrorFile(module, r)
	}
}

// writeUserErrorFile writes a user-friendly error message
func writeUserErrorFile(module string, err interface{}) {
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	errorPath := filepath.Join(exeDir, "LAST_ERROR.txt")
	
	errorMsg := fmt.Sprintf(`FinancialsX Desktop Error Report
================================
Date: %s
Module: %s
Error: %v

The application encountered an error and had to close.
Please check the logs folder for more details.

Log files location: %s\logs\

If this problem persists, please contact support with:
1. This error file
2. The log files from the logs folder
3. What you were doing when the error occurred

================================
`, time.Now().Format("2006-01-02 15:04:05"), module, err, exeDir)
	
	os.WriteFile(errorPath, []byte(errorMsg), 0644)
}

// GetLogPath returns the current log file path
func GetLogPath() string {
	return logPath
}

// GetCrashPath returns the current crash log file path  
func GetCrashPath() string {
	return crashPath
}

// cleanupOldLogs removes log files older than the specified number of days
func cleanupOldLogs(logsDir string, keepDays int) {
	// Calculate cutoff time
	cutoff := time.Now().AddDate(0, 0, -keepDays)
	
	// Read directory
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		// Can't read directory, skip cleanup
		return
	}
	
	// Check each file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		// Check if it's a log file
		name := entry.Name()
		if !isLogFile(name) {
			continue
		}
		
		// Get file info
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		// Check if file is older than cutoff
		if info.ModTime().Before(cutoff) {
			// Delete old log file
			filePath := filepath.Join(logsDir, name)
			os.Remove(filePath)
			fmt.Printf("Removed old log file: %s\n", name)
		}
	}
}

// isLogFile checks if a filename looks like one of our log files
func isLogFile(name string) bool {
	// Check for our log file patterns
	patterns := []string{
		"financialsx_",
		"debug_",
		"startup",
	}
	
	for _, pattern := range patterns {
		if len(name) >= len(pattern) && name[:len(pattern)] == pattern {
			return true
		}
	}
	
	return false
}