// Package financials handles all financial operations including GL, banking, reconciliation, and accounting
package financials

import (
	"fmt"
	"time"

	"github.com/pivoten/financialsx/desktop/internal/common"
)

// Service handles all financial operations
type Service struct {
	// Add dependencies here as needed
}

// NewService creates a new financials service
func NewService() *Service {
	return &Service{}
}

// BankAccount represents a bank account
type BankAccount struct {
	AccountNumber      string    `json:"accountNumber"`
	AccountName        string    `json:"accountName"`
	AccountType        string    `json:"accountType"`
	BankName           string    `json:"bankName"`
	RoutingNumber      string    `json:"routingNumber"`
	GLBalance          float64   `json:"glBalance"`
	OutstandingChecks  float64   `json:"outstandingChecks"`
	BankBalance        float64   `json:"bankBalance"` // GL + Outstanding
	LastReconciled     time.Time `json:"lastReconciled"`
	IsActive           bool      `json:"isActive"`
}

// Check represents a check transaction
type Check struct {
	CheckNumber   string    `json:"checkNumber"`
	CIDCHEC       string    `json:"cidchec"`      // Unique check ID
	Date          time.Time `json:"date"`
	Payee         string    `json:"payee"`
	Amount        float64   `json:"amount"`
	Memo          string    `json:"memo"`
	AccountNumber string    `json:"accountNumber"`
	IsCleared     bool      `json:"isCleared"`
	IsVoid        bool      `json:"isVoid"`
	ClearedDate   *time.Time `json:"clearedDate,omitempty"`
	BatchNumber   string    `json:"batchNumber"`
}

// GLEntry represents a general ledger entry
type GLEntry struct {
	AccountNumber string    `json:"accountNumber"`
	Date          time.Time `json:"date"`
	Description   string    `json:"description"`
	Debit         float64   `json:"debit"`
	Credit        float64   `json:"credit"`
	Balance       float64   `json:"balance"`
	Source        string    `json:"source"` // AP, AR, CHECK, etc.
	Reference     string    `json:"reference"`
	BatchNumber   string    `json:"batchNumber"`
}

// Reconciliation represents a bank reconciliation
type Reconciliation struct {
	ID               int       `json:"id"`
	AccountNumber    string    `json:"accountNumber"`
	ReconcileDate    time.Time `json:"reconcileDate"`
	StatementDate    time.Time `json:"statementDate"`
	BeginningBalance float64   `json:"beginningBalance"`
	EndingBalance    float64   `json:"endingBalance"`
	StatementBalance float64   `json:"statementBalance"`
	Status           string    `json:"status"` // draft, committed, archived
	SelectedChecks   []string  `json:"selectedChecks"` // CIDCHECs
}

// Example function structures - to be implemented by moving from main.go
func (s *Service) GetBankAccounts(companyName string) ([]BankAccount, error) {
	// TODO: Move implementation from main.go
	return nil, fmt.Errorf("not implemented yet - move from main.go")
}

func (s *Service) GetOutstandingChecks(companyName, accountNumber string) ([]Check, error) {
	// TODO: Move implementation from main.go
	return nil, fmt.Errorf("not implemented yet - move from main.go")
}

func (s *Service) GetGLBalance(companyName, accountNumber string) (float64, error) {
	// TODO: Move implementation from main.go
	return 0, fmt.Errorf("not implemented yet - move from main.go")
}

func (s *Service) ReconcileBank(companyName string, reconciliation Reconciliation) error {
	// TODO: Move implementation from main.go
	return fmt.Errorf("not implemented yet - move from main.go")
}