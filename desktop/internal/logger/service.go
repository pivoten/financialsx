package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pivoten/financialsx/desktop/internal/config"
)

// Service provides logging functionality as a service
type Service struct {
	debugMode bool
	logDir    string
}

// NewService creates a new logging service
func NewService() *Service {
	return &Service{}
}

// InitializeLogging sets up the logging system
func (s *Service) InitializeLogging(debugMode bool) map[string]interface{} {
	// Get user's home directory for log storage
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Failed to get home directory: " + err.Error(),
		}
	}

	// Create logs directory in user's app data
	s.logDir = filepath.Join(homeDir, ".financialsx", "logs")
	s.debugMode = debugMode

	// Initialize the logger with the directory
	if err := os.MkdirAll(s.logDir, 0755); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Failed to create log directory: " + err.Error(),
		}
	}

	// Initialize the main logger system
	if err := Initialize(); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Failed to initialize logger: " + err.Error(),
		}
	}

	// Clean logs older than 30 days in background
	go s.cleanOldLogs(30)

	return map[string]interface{}{
		"success":   true,
		"logDir":    s.logDir,
		"debugMode": debugMode,
	}
}

// LogMessage logs a message with context
func (s *Service) LogMessage(level, message, component string, data map[string]interface{}) map[string]interface{} {
	// Convert level to appropriate logging function
	switch level {
	case "error":
		WriteError(component, message)
	case "warning":
		WriteWarning(component, message)
	case "info":
		WriteInfo(component, message)
	case "debug":
		if s.debugMode {
			WriteDebug(component, message)
		}
	default:
		WriteInfo(component, message)
	}

	return map[string]interface{}{
		"success": true,
	}
}

// SetDebugMode enables or disables debug logging
func (s *Service) SetDebugMode(enabled bool) map[string]interface{} {
	s.debugMode = enabled

	// Save preference to config
	if err := config.SetDebugMode(enabled); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to save debug mode preference: %v", err),
		}
	}

	return map[string]interface{}{
		"success":   true,
		"debugMode": enabled,
	}
}

// GetDebugMode returns the current debug mode status
func (s *Service) GetDebugMode() bool {
	return s.debugMode
}

// GetLogFilePath returns the path to the current log file
func (s *Service) GetLogFilePath() string {
	return GetLogPath()
}

// LogError logs an error from the frontend
func (s *Service) LogError(errorMessage string, stackTrace string) {
	if GetLogPath() != "" {
		WriteError("Frontend", fmt.Sprintf("Error: %s\nStack: %s", errorMessage, stackTrace))
	}
	fmt.Printf("Frontend Error: %s\n", errorMessage)
}

// TestLogging tests the logging system
func (s *Service) TestLogging() string {
	testMsg := "Test log entry"
	WriteInfo("LogTest", testMsg)
	WriteWarning("LogTest", "Test warning")
	WriteError("LogTest", "Test error")
	WriteDebug("LogTest", "Test debug message")
	WriteCrash("LogTest", fmt.Errorf("Test crash (not a real crash)"), nil)
	
	result := fmt.Sprintf("Logging test completed. Check log file at: %s", GetLogPath())
	return result
}

// cleanOldLogs removes log files older than the specified number of days
func (s *Service) cleanOldLogs(daysToKeep int) {
	if s.logDir == "" {
		return
	}

	cutoffTime := time.Now().AddDate(0, 0, -daysToKeep)

	filepath.Walk(s.logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Only remove .log files
		if filepath.Ext(path) == ".log" && info.ModTime().Before(cutoffTime) {
			os.Remove(path)
		}

		return nil
	})
}