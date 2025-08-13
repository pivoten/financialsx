package vfp

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

// Settings represents VFP connection configuration
type Settings struct {
	ID       int    `json:"id"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Enabled  bool   `json:"enabled"`
	Timeout  int    `json:"timeout"` // in seconds
	UpdatedAt string `json:"updated_at"`
}

// Command represents a VFP form launch command
type Command struct {
	Action   string `json:"action,omitempty"`      // action to perform (launchForm, getCompany, setCompany)
	Form     string `json:"form,omitempty"`        // legacy field for backward compatibility
	FormName string `json:"formName,omitempty"`    // form to launch
	Arg      string `json:"arg,omitempty"`         // legacy field
	Argument string `json:"argument,omitempty"`    // argument for form
	Company  string `json:"company,omitempty"`     // company context
	Token    string `json:"token,omitempty"`
}

// VFPClient handles communication with Visual FoxPro
type VFPClient struct {
	db       *sql.DB
	settings *Settings
}

// NewVFPClient creates a new VFP integration client
func NewVFPClient(db *sql.DB) *VFPClient {
	return &VFPClient{
		db: db,
	}
}

// InitializeSchema creates the VFP settings table if it doesn't exist
func (v *VFPClient) InitializeSchema() error {
	query := `
		CREATE TABLE IF NOT EXISTS vfp_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			host TEXT NOT NULL DEFAULT 'localhost',
			port INTEGER NOT NULL DEFAULT 23456,
			enabled BOOLEAN NOT NULL DEFAULT 0,
			timeout INTEGER NOT NULL DEFAULT 5,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := v.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create vfp_settings table: %w", err)
	}

	// Insert default settings if none exist
	var count int
	err = v.db.QueryRow("SELECT COUNT(*) FROM vfp_settings").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err = v.db.Exec(`
			INSERT INTO vfp_settings (host, port, enabled, timeout) 
			VALUES ('localhost', 23456, 0, 5)
		`)
		if err != nil {
			return fmt.Errorf("failed to insert default settings: %w", err)
		}
	}

	return nil
}

// GetSettings retrieves the current VFP connection settings
func (v *VFPClient) GetSettings() (*Settings, error) {
	settings := &Settings{}
	err := v.db.QueryRow(`
		SELECT id, host, port, enabled, timeout, 
		       datetime(updated_at, 'localtime') as updated_at
		FROM vfp_settings 
		ORDER BY id DESC 
		LIMIT 1
	`).Scan(&settings.ID, &settings.Host, &settings.Port, 
	        &settings.Enabled, &settings.Timeout, &settings.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			// Return default settings
			return &Settings{
				Host:    "localhost",
				Port:    23456,
				Enabled: false,
				Timeout: 5,
			}, nil
		}
		return nil, err
	}

	v.settings = settings
	return settings, nil
}

// SaveSettings updates the VFP connection settings
func (v *VFPClient) SaveSettings(settings *Settings) error {
	// Check if settings exist
	var count int
	err := v.db.QueryRow("SELECT COUNT(*) FROM vfp_settings").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		// Update existing settings
		_, err = v.db.Exec(`
			UPDATE vfp_settings 
			SET host = ?, port = ?, enabled = ?, timeout = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = (SELECT MAX(id) FROM vfp_settings)
		`, settings.Host, settings.Port, settings.Enabled, settings.Timeout)
	} else {
		// Insert new settings
		_, err = v.db.Exec(`
			INSERT INTO vfp_settings (host, port, enabled, timeout) 
			VALUES (?, ?, ?, ?)
		`, settings.Host, settings.Port, settings.Enabled, settings.Timeout)
	}

	if err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	v.settings = settings
	return nil
}

// TestConnection verifies connectivity to the VFP listener
func (v *VFPClient) TestConnection() error {
	// Load current settings if not cached
	if v.settings == nil {
		settings, err := v.GetSettings()
		if err != nil {
			return fmt.Errorf("failed to load settings: %w", err)
		}
		v.settings = settings
	}

	if !v.settings.Enabled {
		return fmt.Errorf("VFP integration is disabled")
	}

	// Try to connect
	address := fmt.Sprintf("%s:%d", v.settings.Host, v.settings.Port)
	timeout := time.Duration(v.settings.Timeout) * time.Second
	
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	// Send a test command (empty form to just test connectivity)
	testCmd := Command{Form: "TEST"}
	payload, _ := json.Marshal(testCmd)
	payload = append(payload, '\n')

	conn.SetWriteDeadline(time.Now().Add(timeout))
	if _, err := conn.Write(payload); err != nil {
		return fmt.Errorf("failed to send test command: %w", err)
	}

	// Read response
	conn.SetReadDeadline(time.Now().Add(timeout))
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	response = strings.TrimSpace(response)
	if !strings.HasPrefix(response, "OK") && !strings.HasPrefix(response, "ERR") {
		return fmt.Errorf("unexpected response: %s", response)
	}

	return nil
}

// sendCommand sends a command to VFP and returns the response
func (v *VFPClient) sendCommand(cmd Command) (map[string]interface{}, error) {
	// Load current settings if not cached
	if v.settings == nil {
		settings, err := v.GetSettings()
		if err != nil {
			return nil, fmt.Errorf("failed to load settings: %w", err)
		}
		v.settings = settings
	}

	if !v.settings.Enabled {
		return nil, fmt.Errorf("VFP integration is disabled")
	}

	// Connect to VFP
	address := fmt.Sprintf("%s:%d", v.settings.Host, v.settings.Port)
	timeout := time.Duration(v.settings.Timeout) * time.Second
	
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	// Send command
	payload, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}
	payload = append(payload, '\n')

	conn.SetWriteDeadline(time.Now().Add(timeout))
	if _, err := conn.Write(payload); err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	// Read response
	conn.SetReadDeadline(time.Now().Add(timeout))
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	response = strings.TrimSpace(response)
	
	// Try to parse as JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// If not JSON, treat as simple string response
		if strings.HasPrefix(response, "ERR") {
			return nil, fmt.Errorf("VFP error: %s", strings.TrimPrefix(response, "ERR "))
		}
		return map[string]interface{}{"response": response}, nil
	}

	return result, nil
}

// LaunchForm sends a command to open a VFP form with company synchronization
func (v *VFPClient) LaunchForm(formName string, argument string, companyName string) (string, error) {
	cmd := Command{
		Action:   "launchForm",
		FormName: formName,
		Argument: argument,
		// Don't send company for now - user will ensure correct company is open
	}

	// Debug log the command
	fmt.Printf("LaunchForm sending command: %+v\n", cmd)
	
	response, err := v.sendCommand(cmd)
	if err != nil {
		return "", err
	}

	// Check if company change is needed
	if needsChange, ok := response["needsCompanyChange"].(bool); ok && needsChange {
		currentCompany := ""
		requestedCompany := ""
		
		if curr, ok := response["currentCompany"].(string); ok {
			currentCompany = curr
		}
		if req, ok := response["requestedCompany"].(string); ok {
			requestedCompany = req
		}
		
		return "", fmt.Errorf("company mismatch: FoxPro has '%s' open, FinancialsX has '%s'", 
			currentCompany, requestedCompany)
	}

	// Check response for success
	if success, ok := response["success"].(bool); ok && success {
		if msg, ok := response["message"].(string); ok {
			return msg, nil
		}
		return "Form launched successfully", nil
	}

	// Get error message if available
	if msg, ok := response["message"].(string); ok {
		return "", fmt.Errorf(msg)
	}

	return "", fmt.Errorf("unknown error launching form")
}

// GetVFPCompany gets the current company from VFP
func (v *VFPClient) GetVFPCompany() (string, error) {
	cmd := Command{
		Action: "getCompany",
	}

	response, err := v.sendCommand(cmd)
	if err != nil {
		return "", err
	}

	if company, ok := response["company"].(string); ok {
		return company, nil
	}

	return "", fmt.Errorf("could not get company from VFP")
}

// SetVFPCompany sets the current company in VFP
func (v *VFPClient) SetVFPCompany(companyName string) error {
	cmd := Command{
		Action:  "setCompany",
		Company: companyName,
	}

	response, err := v.sendCommand(cmd)
	if err != nil {
		return err
	}

	if success, ok := response["success"].(bool); ok && success {
		return nil
	}

	if msg, ok := response["message"].(string); ok {
		return fmt.Errorf(msg)
	}

	return fmt.Errorf("failed to set company in VFP")
}

// GetFormList returns a list of commonly used VFP forms
func (v *VFPClient) GetFormList() []map[string]string {
	// This could be extended to read from a configuration file or database
	return []map[string]string{
		{"name": "Customer", "description": "Customer Management"},
		{"name": "Invoice", "description": "Invoice Entry"},
		{"name": "Payment", "description": "Payment Processing"},
		{"name": "Reports", "description": "Report Generation"},
		{"name": "GLEntry", "description": "General Ledger Entry"},
		{"name": "APBill", "description": "Accounts Payable Bills"},
		{"name": "ARInvoice", "description": "Accounts Receivable"},
		{"name": "CheckPrint", "description": "Check Printing"},
		{"name": "BankRec", "description": "Bank Reconciliation"},
		{"name": "Vendor", "description": "Vendor Management"},
	}
}