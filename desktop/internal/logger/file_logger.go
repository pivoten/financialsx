package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileLogger handles logging to local files
type FileLogger struct {
	mu          sync.Mutex
	file        *os.File
	debugMode   bool
	logDir      string
	maxFileSize int64 // in bytes
	currentSize int64
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp   string                 `json:"timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Component   string                 `json:"component,omitempty"`
	UserID      string                 `json:"userId,omitempty"`
	CompanyName string                 `json:"companyName,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Stack       string                 `json:"stack,omitempty"`
}

var (
	instance *FileLogger
	once     sync.Once
)

// GetLogger returns the singleton logger instance
func GetLogger() *FileLogger {
	once.Do(func() {
		instance = &FileLogger{
			debugMode:   false,
			maxFileSize: 10 * 1024 * 1024, // 10MB default
		}
	})
	return instance
}

// Initialize sets up the file logger
func (l *FileLogger) Initialize(debugMode bool, logDir string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.debugMode = debugMode
	l.logDir = logDir

	if !debugMode {
		return nil // Don't create log files if not in debug mode
	}

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open or create today's log file
	return l.rotateLogFile()
}

// rotateLogFile creates a new log file or rotates if needed
func (l *FileLogger) rotateLogFile() error {
	if l.file != nil {
		l.file.Close()
	}

	// Generate filename with date
	filename := fmt.Sprintf("financialsx_%s.log", time.Now().Format("2006-01-02"))
	fullPath := filepath.Join(l.logDir, filename)

	// Check if file exists and get its size
	if info, err := os.Stat(fullPath); err == nil {
		l.currentSize = info.Size()
		
		// If file is too large, create a new one with timestamp
		if l.currentSize >= l.maxFileSize {
			timestamp := time.Now().Format("150405") // HHMMSS
			filename = fmt.Sprintf("financialsx_%s_%s.log", 
				time.Now().Format("2006-01-02"), timestamp)
			fullPath = filepath.Join(l.logDir, filename)
			l.currentSize = 0
		}
	} else {
		l.currentSize = 0
	}

	// Open file in append mode
	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	l.file = file
	return nil
}

// Log writes a log entry to file
func (l *FileLogger) Log(level, message, component string, data map[string]interface{}) error {
	if !l.debugMode {
		return nil // Skip logging if not in debug mode
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Component: component,
		Data:      data,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Add newline
	jsonData = append(jsonData, '\n')

	// Check if rotation is needed
	if l.currentSize+int64(len(jsonData)) > l.maxFileSize {
		if err := l.rotateLogFile(); err != nil {
			return err
		}
	}

	// Write to file
	n, err := l.file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	l.currentSize += int64(n)

	// Sync to disk for important logs
	if level == "ERROR" || level == "FATAL" {
		l.file.Sync()
	}

	return nil
}

// SetDebugMode enables or disables debug logging
func (l *FileLogger) SetDebugMode(enabled bool) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.debugMode = enabled

	if !enabled && l.file != nil {
		l.file.Close()
		l.file = nil
	} else if enabled && l.file == nil {
		return l.rotateLogFile()
	}

	return nil
}

// GetDebugMode returns the current debug mode status
func (l *FileLogger) GetDebugMode() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.debugMode
}

// Close closes the log file
func (l *FileLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}
	return nil
}

// CleanOldLogs removes log files older than specified days
func (l *FileLogger) CleanOldLogs(daysToKeep int) error {
	if l.logDir == "" {
		return nil
	}

	cutoffTime := time.Now().AddDate(0, 0, -daysToKeep)

	entries, err := os.ReadDir(l.logDir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if it's a log file
		if filepath.Ext(entry.Name()) != ".log" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Remove if older than cutoff
		if info.ModTime().Before(cutoffTime) {
			os.Remove(filepath.Join(l.logDir, entry.Name()))
		}
	}

	return nil
}