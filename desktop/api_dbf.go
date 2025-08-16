package main

import (
	"fmt"
)

// ============================================================================
// DBF FILE API
// This file contains all DBF file-related API methods
// ============================================================================

// GetDBFFiles returns a list of DBF files in the company directory
func (a *App) GetDBFFiles(companyName string) ([]string, error) {
	return a.Services.DBF.GetFiles(companyName)
}

// GetDBFTableData reads a DBF file and returns its data
func (a *App) GetDBFTableData(companyName, fileName string) (map[string]interface{}, error) {
	return a.Services.DBF.GetTableData(companyName, fileName)
}

// GetDBFTableDataPaged reads a DBF file with pagination
func (a *App) GetDBFTableDataPaged(companyName, fileName string, offset, limit int, sortColumn, sortDirection string) (map[string]interface{}, error) {
	return a.Services.DBF.GetTableDataPaged(companyName, fileName, offset, limit, sortColumn, sortDirection)
}

// SearchDBFTable searches within a DBF file
func (a *App) SearchDBFTable(companyName, fileName, searchTerm string) (map[string]interface{}, error) {
	return a.Services.DBF.SearchTable(companyName, fileName, searchTerm)
}

// UpdateDBFRecord updates a record in a DBF file
func (a *App) UpdateDBFRecord(companyName, fileName string, rowIndex, colIndex int, value string) error {
	if err := a.requirePermission("dbf.write"); err != nil {
		return err
	}

	return a.Services.DBF.UpdateRecord(companyName, fileName, rowIndex, colIndex, value)
}

// GetTableList returns a list of available DBF tables with metadata
func (a *App) GetTableList(companyName string) (map[string]interface{}, error) {
	return a.Services.DBF.GetTableList(companyName)
}

// GetDBFTableInfo returns metadata about a DBF table
func (a *App) GetDBFTableInfo(companyName, fileName string) (map[string]interface{}, error) {
	return a.Services.DBF.GetTableInfo(companyName, fileName)
}

// ExportDBFToCSV exports a DBF file to CSV format
func (a *App) ExportDBFToCSV(companyName, fileName, outputPath string) error {
	if err := a.requirePermission("export.data"); err != nil {
		return err
	}

	// TODO: Implement in DBF service
	return fmt.Errorf("export to CSV not yet implemented")
}

// ValidateDBFStructure validates the structure of a DBF file
func (a *App) ValidateDBFStructure(companyName, fileName string) (map[string]interface{}, error) {
	// TODO: Implement in DBF service
	return nil, fmt.Errorf("DBF structure validation not yet implemented")
}