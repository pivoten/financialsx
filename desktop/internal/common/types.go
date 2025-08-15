package common

import "time"

// StandardResponse is the common response format for API calls
type StandardResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}


// PaginatedResponse contains paginated data
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	TotalCount int         `json:"totalCount"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}

// FilterParams contains common filter parameters
type FilterParams struct {
	SearchTerm string                 `json:"searchTerm,omitempty"`
	DateFrom   *time.Time             `json:"dateFrom,omitempty"`
	DateTo     *time.Time             `json:"dateTo,omitempty"`
	Status     string                 `json:"status,omitempty"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
}

// BankTransaction represents a bank transaction from CSV import
type BankTransaction struct {
	ID              int     `json:"id"`
	TransactionDate string  `json:"transaction_date"`
	Description     string  `json:"description"`
	Amount          float64 `json:"amount"`
	Balance         float64 `json:"balance,omitempty"`
	CheckNumber     string  `json:"check_number,omitempty"`
	TransactionType string  `json:"transaction_type"` // "debit" or "credit"
	Matched         bool    `json:"matched"`
	MatchedCheckID  string  `json:"matched_check_id,omitempty"`
}

// MatchResult represents the result of matching a bank transaction to a check
type MatchResult struct {
	BankTransactionID int     `json:"bank_transaction_id"`
	CheckID           string  `json:"check_id"`
	CheckNumber       string  `json:"check_number"`
	Confidence        float64 `json:"confidence"`
	MatchType         string  `json:"match_type"` // "auto", "manual", "suggested"
}

// AuditResult represents the result of an audit operation
type AuditResult struct {
	Type        string                 `json:"type"`
	Company     string                 `json:"company"`
	Timestamp   time.Time              `json:"timestamp"`
	TotalCount  int                    `json:"totalCount"`
	IssueCount  int                    `json:"issueCount"`
	Details     []interface{}          `json:"details"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ErrorResponse creates a standard error response
func ErrorResponse(err error) StandardResponse {
	return StandardResponse{
		Success: false,
		Error:   err.Error(),
	}
}

// SuccessResponse creates a standard success response
func SuccessResponse(data interface{}) StandardResponse {
	return StandardResponse{
		Success: true,
		Data:    data,
	}
}

// MessageResponse creates a response with a message
func MessageResponse(success bool, message string) StandardResponse {
	return StandardResponse{
		Success: success,
		Message: message,
	}
}