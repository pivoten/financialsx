// Package audit provides financial auditing functionality
package audit

import (
	"fmt"
	"strings"
	"time"

	"github.com/pivoten/financialsx/desktop/internal/currency"
)

// Service handles all audit operations
type Service struct {
	// Add dependencies as needed
}

// NewService creates a new audit service
func NewService() *Service {
	return &Service{}
}

// AuditResult represents the result of an audit operation
type AuditResult struct {
	Success         bool                   `json:"success"`
	Message         string                 `json:"message,omitempty"`
	Error           string                 `json:"error,omitempty"`
	TotalChecks     int                    `json:"totalChecks,omitempty"`
	Issues          []AuditIssue          `json:"issues,omitempty"`
	Summary         map[string]interface{} `json:"summary,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// AuditIssue represents a single issue found during an audit
type AuditIssue struct {
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"` // "error", "warning", "info"
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details"`
	RowData     map[string]interface{} `json:"rowData,omitempty"`
}

// CheckBatchAudit represents the result of a check batch audit
type CheckBatchAudit struct {
	TotalChecks       int                      `json:"totalChecks"`
	ChecksWithBatch   int                      `json:"checksWithBatch"`
	ChecksWithoutBatch int                     `json:"checksWithoutBatch"`
	MatchedEntries    int                      `json:"matchedEntries"`
	MissingEntries    []map[string]interface{} `json:"missingEntries"`
	MismatchedAmounts []map[string]interface{} `json:"mismatchedAmounts"`
	CheckColumns      []string                 `json:"checkColumns"`
}

// DuplicateCheck represents a duplicate check entry
type DuplicateCheck struct {
	CIDCHEC     string    `json:"cidchec"`
	CheckNumber string    `json:"checkNumber"`
	CheckDate   time.Time `json:"checkDate"`
	Payee       string    `json:"payee"`
	Amount      float64   `json:"amount"`
	AccountNo   string    `json:"accountNo"`
	Batch       string    `json:"batch"`
	Void        bool      `json:"void"`
	Cleared     bool      `json:"cleared"`
	Count       int       `json:"count"`
}

// VoidCheckAudit represents void check audit results
type VoidCheckAudit struct {
	TotalVoidChecks   int                      `json:"totalVoidChecks"`
	ProperlyVoided    int                      `json:"properlyVoided"`
	Issues            []map[string]interface{} `json:"issues"`
	VoidWithNonZero   []map[string]interface{} `json:"voidWithNonZeroAmount"`
	NonVoidWithZero   []map[string]interface{} `json:"nonVoidWithZeroAmount"`
	VoidButCleared    []map[string]interface{} `json:"voidButCleared"`
}

// GLMatchingAudit represents GL matching audit results  
type GLMatchingAudit struct {
	Period            string                   `json:"period"`
	TotalChecks       int                      `json:"totalChecks"`
	MatchedChecks     int                      `json:"matchedChecks"`
	UnmatchedChecks   []map[string]interface{} `json:"unmatchedChecks"`
	TotalGLEntries    int                      `json:"totalGLEntries"`
	MatchedGLEntries  int                      `json:"matchedGLEntries"`
	UnmatchedGLEntries []map[string]interface{} `json:"unmatchedGLEntries"`
}

// PayeeCIDVerification represents payee verification results
type PayeeCIDVerification struct {
	TotalChecks        int                      `json:"totalChecks"`
	ChecksWithCID      int                      `json:"checksWithCID"`
	ChecksWithoutCID   int                      `json:"checksWithoutCID"`
	UniquePayees       int                      `json:"uniquePayees"`
	PayeesWithMultipleCIDs []map[string]interface{} `json:"payeesWithMultipleCIDs"`
	MissingCIDChecks   []map[string]interface{} `json:"missingCIDChecks"`
}

// ReconciliationAudit represents bank reconciliation audit results
type ReconciliationAudit struct {
	TotalAccounts      int                      `json:"totalAccounts"`
	ReconciledAccounts int                      `json:"reconciledAccounts"`
	UnreconciledAccounts []map[string]interface{} `json:"unreconciledAccounts"`
	OutOfBalanceAccounts []map[string]interface{} `json:"outOfBalanceAccounts"`
	StaleReconciliations []map[string]interface{} `json:"staleReconciliations"`
}

// Helper function to parse amount from various formats
func parseAmount(value interface{}) float64 {
	c := currency.ParseFromDBF(value)
	return c.ToFloat64()
}

// Helper function to parse date from various formats
func parseDate(value interface{}) (time.Time, error) {
	if value == nil {
		return time.Time{}, fmt.Errorf("date is nil")
	}
	
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		// Try various date formats
		formats := []string{
			"2006-01-02",
			"01/02/2006",
			"01/02/06",
			"2006-01-02 15:04:05",
		}
		
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("unable to parse date: %s", v)
	default:
		return time.Time{}, fmt.Errorf("unsupported date type: %T", value)
	}
}

// Helper function to safely get string value
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok && val != nil {
		return strings.TrimSpace(fmt.Sprintf("%v", val))
	}
	return ""
}

// Helper function to safely get bool value
func getBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case bool:
			return v
		case string:
			v = strings.ToLower(strings.TrimSpace(v))
			return v == "true" || v == "t" || v == ".t." || v == "1"
		case int:
			return v != 0
		case float64:
			return v != 0
		}
	}
	return false
}