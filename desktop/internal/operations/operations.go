// Package operations handles business operations including vendors, purchases, inventory, and workflow
package operations

import (
	"fmt"
	"time"
)

// Service handles all business operations
type Service struct {
	// Add dependencies here as needed
}

// NewService creates a new operations service
func NewService() *Service {
	return &Service{}
}

// Vendor represents a vendor/supplier
type Vendor struct {
	VendorID      string    `json:"vendorId"`
	VendorName    string    `json:"vendorName"`
	Address1      string    `json:"address1"`
	Address2      string    `json:"address2"`
	City          string    `json:"city"`
	State         string    `json:"state"`
	Zip           string    `json:"zip"`
	Country       string    `json:"country"`
	TaxID         string    `json:"taxId"`
	Phone         string    `json:"phone"`
	Contact       string    `json:"contact"`
	Email         string    `json:"email"`
	Terms         string    `json:"terms"`
	IsActive      bool      `json:"isActive"`
	Is1099        bool      `json:"is1099"`
	CreditLimit   float64   `json:"creditLimit"`
	CurrentBalance float64  `json:"currentBalance"`
	LastPurchase  *time.Time `json:"lastPurchase,omitempty"`
	LastPayment   *time.Time `json:"lastPayment,omitempty"`
}

// Purchase represents a purchase/bill
type Purchase struct {
	PurchaseID    string    `json:"purchaseId"`
	VendorID      string    `json:"vendorId"`
	InvoiceNumber string    `json:"invoiceNumber"`
	InvoiceDate   time.Time `json:"invoiceDate"`
	DueDate       time.Time `json:"dueDate"`
	Amount        float64   `json:"amount"`
	Description   string    `json:"description"`
	Status        string    `json:"status"` // open, paid, partial, void
	PaidAmount    float64   `json:"paidAmount"`
	BatchNumber   string    `json:"batchNumber"`
}

// Well represents an oil/gas well
type Well struct {
	WellID       string  `json:"wellId"`
	WellName     string  `json:"wellName"`
	APINumber    string  `json:"apiNumber"`
	County       string  `json:"county"`
	State        string  `json:"state"`
	Field        string  `json:"field"`
	Operator     string  `json:"operator"`
	WorkingInterest float64 `json:"workingInterest"`
	NetRevenue   float64 `json:"netRevenue"`
	Status       string  `json:"status"` // active, inactive, plugged
	SpudDate     *time.Time `json:"spudDate,omitempty"`
	FirstProdDate *time.Time `json:"firstProdDate,omitempty"`
}

// Owner represents a working interest owner
type Owner struct {
	OwnerID        string  `json:"ownerId"`
	OwnerName      string  `json:"ownerName"`
	Address        string  `json:"address"`
	City           string  `json:"city"`
	State          string  `json:"state"`
	Zip            string  `json:"zip"`
	TaxID          string  `json:"taxId"`
	OwnershipType  string  `json:"ownershipType"` // WI, RI, ORRI
	InterestPercent float64 `json:"interestPercent"`
	IsActive       bool    `json:"isActive"`
}

// Example function structures - to be implemented by moving from main.go
func (s *Service) GetVendors(companyName string) ([]Vendor, error) {
	// TODO: Move implementation from main.go
	return nil, fmt.Errorf("not implemented yet - move from main.go")
}

func (s *Service) UpdateVendor(companyName string, vendorID string, vendor Vendor) error {
	// TODO: Move implementation from main.go
	return fmt.Errorf("not implemented yet - move from main.go")
}

func (s *Service) GetWells(companyName string) ([]Well, error) {
	// TODO: Move implementation from main.go
	return nil, fmt.Errorf("not implemented yet - move from main.go")
}

func (s *Service) GetOwners(companyName string) ([]Owner, error) {
	// TODO: Move implementation from main.go
	return nil, fmt.Errorf("not implemented yet - move from main.go")
}