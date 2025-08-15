package legacy

import (
	"fmt"
	"github.com/pivoten/financialsx/desktop/internal/vfp"
)

// VFPWrapper provides a wrapper around VFP integration functionality
type VFPWrapper struct {
	vfpClient *vfp.VFPClient
}

// NewVFPWrapper creates a new VFP wrapper
func NewVFPWrapper(vfpClient *vfp.VFPClient) *VFPWrapper {
	return &VFPWrapper{
		vfpClient: vfpClient,
	}
}

// GetVFPSettings retrieves the current VFP connection settings
func (w *VFPWrapper) GetVFPSettings() (map[string]interface{}, error) {
	if w.vfpClient == nil {
		return nil, fmt.Errorf("VFP client not initialized")
	}
	
	settings, err := w.vfpClient.GetSettings()
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"host":    settings.Host,
		"port":    settings.Port,
		"enabled": settings.Enabled,
		"timeout": settings.Timeout,
		"updated_at": settings.UpdatedAt,
	}, nil
}

// SaveVFPSettings updates the VFP connection settings
func (w *VFPWrapper) SaveVFPSettings(host string, port int, enabled bool, timeout int) error {
	if w.vfpClient == nil {
		return fmt.Errorf("VFP client not initialized")
	}
	
	settings := &vfp.Settings{
		Host:    host,
		Port:    port,
		Enabled: enabled,
		Timeout: timeout,
	}
	
	return w.vfpClient.SaveSettings(settings)
}

// TestVFPConnection tests the connection to the VFP listener
func (w *VFPWrapper) TestVFPConnection() (map[string]interface{}, error) {
	if w.vfpClient == nil {
		return nil, fmt.Errorf("VFP client not initialized")
	}
	
	err := w.vfpClient.TestConnection()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}, nil
	}
	
	return map[string]interface{}{
		"success": true,
		"message": "Connection successful",
	}, nil
}

// LaunchVFPForm launches a VFP form with optional argument and company synchronization
func (w *VFPWrapper) LaunchVFPForm(formName string, argument string) (map[string]interface{}, error) {
	if w.vfpClient == nil {
		return nil, fmt.Errorf("VFP client not initialized")
	}
	
	// Don't send company for now - user will ensure correct company is open
	response, err := w.vfpClient.LaunchForm(formName, argument, "")
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}, nil
	}
	
	return map[string]interface{}{
		"success": true,
		"message": response,
	}, nil
}

// SyncVFPCompany synchronizes the company between FinancialsX and VFP
func (w *VFPWrapper) SyncVFPCompany() (map[string]interface{}, error) {
	// Get current company from FinancialsX
	// For now, don't sync company - user will ensure correct company is open
	currentCompany := ""
	if w.vfpClient == nil {
		return map[string]interface{}{
			"success": false,
			"message": "VFP integration not initialized",
		}, nil
	}

	// Set it in VFP
	err := w.vfpClient.SetVFPCompany(currentCompany)
	if err != nil {
		// Try to get VFP's current company for info
		vfpCompany, _ := w.vfpClient.GetVFPCompany()
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
			"financialsxCompany": currentCompany,
			"vfpCompany": vfpCompany,
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"message": "Company synchronized",
		"company": currentCompany,
	}, nil
}

// GetVFPCompany gets the current company from VFP
func (w *VFPWrapper) GetVFPCompany() (map[string]interface{}, error) {
	if w.vfpClient == nil {
		return map[string]interface{}{
			"success": false,
			"message": "VFP integration not initialized",
		}, nil
	}

	company, err := w.vfpClient.GetVFPCompany()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"company": company,
	}, nil
}

// GetVFPFormList returns a list of available VFP forms
func (w *VFPWrapper) GetVFPFormList() []map[string]string {
	if w.vfpClient == nil {
		return []map[string]string{}
	}
	
	return w.vfpClient.GetFormList()
}