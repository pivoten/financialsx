// Package common provides shared types, constants, and utilities used across the application
package common

import (
	"time"
)

// Common types that are used across multiple packages

// Company represents a company in the system
type Company struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Alias        string    `json:"alias"`
	DataPath     string    `json:"dataPath"`
	IsActive     bool      `json:"isActive"`
	LastAccessed time.Time `json:"lastAccessed"`
}

// DBFRecord represents a generic DBF record
type DBFRecord map[string]interface{}

// Result represents a standard API response
type Result struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Count   int         `json:"count,omitempty"`
}

// Pagination parameters
type PaginationParams struct {
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
	Sort   string `json:"sort"`
	Order  string `json:"order"` // "asc" or "desc"
}

// Common constants
const (
	DefaultPageSize = 50
	MaxPageSize     = 1000
	DateFormat      = "2006-01-02"
	DateTimeFormat  = "2006-01-02 15:04:05"
)

// Common error types
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}