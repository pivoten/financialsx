package audit

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pivoten/financialsx/desktop/internal/company"
)

// PayeeCIDVerification audits payee CID consistency
func (s *Service) PayeeCIDVerification(companyName string) (*PayeeCIDVerification, error) {
	// Read checks.dbf
	checksData, err := company.ReadDBFFile(companyName, "CHECKS.DBF", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read CHECKS.DBF: %w", err)
	}

	checks, ok := checksData["rows"].([]map[string]interface{})
	if !ok || len(checks) == 0 {
		return &PayeeCIDVerification{
			TotalChecks:            0,
			ChecksWithCID:          0,
			ChecksWithoutCID:       0,
			UniquePayees:           0,
			PayeesWithMultipleCIDs: []map[string]interface{}{},
			MissingCIDChecks:       []map[string]interface{}{},
		}, nil
	}

	// Check if CID field exists (could be CIDPAYEE or similar)
	cidField := ""
	if len(checks) > 0 {
		for field := range checks[0] {
			upperField := strings.ToUpper(field)
			if strings.Contains(upperField, "CID") && strings.Contains(upperField, "PAYEE") {
				cidField = field
				break
			}
		}
		// Fallback to CIDCHEC if no payee-specific CID field
		if cidField == "" {
			if _, hasCIDCHEC := checks[0]["CIDCHEC"]; hasCIDCHEC {
				cidField = "CIDCHEC"
			}
		}
	}

	result := &PayeeCIDVerification{
		PayeesWithMultipleCIDs: []map[string]interface{}{},
		MissingCIDChecks:       []map[string]interface{}{},
	}

	// Group by payee
	payeeGroups := make(map[string]map[string]int) // payee -> CID -> count
	payeeCIDList := make(map[string][]string)      // payee -> list of unique CIDs
	totalChecks := 0
	checksWithCID := 0
	checksWithoutCID := 0

	for _, check := range checks {
		// Skip void checks
		if getBool(check, "LVOID") {
			continue
		}

		totalChecks++
		payee := strings.TrimSpace(getString(check, "CPAYEE"))
		
		if payee == "" {
			continue
		}

		var cid string
		if cidField != "" {
			cid = strings.TrimSpace(getString(check, cidField))
		}

		if cid != "" {
			checksWithCID++
			
			// Initialize payee group if needed
			if payeeGroups[payee] == nil {
				payeeGroups[payee] = make(map[string]int)
			}
			
			// Count this CID for this payee
			payeeGroups[payee][cid]++
			
			// Track unique CIDs per payee
			cidFound := false
			for _, existingCID := range payeeCIDList[payee] {
				if existingCID == cid {
					cidFound = true
					break
				}
			}
			if !cidFound {
				payeeCIDList[payee] = append(payeeCIDList[payee], cid)
			}
		} else {
			checksWithoutCID++
			
			// Add to missing CID list
			result.MissingCIDChecks = append(result.MissingCIDChecks, map[string]interface{}{
				"checkNumber": getString(check, "CCHECKNO"),
				"checkDate":   check["DCHECKDATE"],
				"payee":       payee,
				"amount":      parseAmount(check["NAMOUNT"]),
				"accountNo":   getString(check, "CACCTNO"),
			})
		}
	}

	// Find payees with multiple CIDs
	for payee, cidList := range payeeCIDList {
		if len(cidList) > 1 {
			cidCounts := payeeGroups[payee]
			
			// Build CID details
			cidDetails := []map[string]interface{}{}
			for cid, count := range cidCounts {
				cidDetails = append(cidDetails, map[string]interface{}{
					"cid":   cid,
					"count": count,
				})
			}
			
			// Sort by count descending
			sort.Slice(cidDetails, func(i, j int) bool {
				return cidDetails[i]["count"].(int) > cidDetails[j]["count"].(int)
			})
			
			result.PayeesWithMultipleCIDs = append(result.PayeesWithMultipleCIDs, map[string]interface{}{
				"payee":        payee,
				"cidCount":     len(cidList),
				"cids":         cidList,
				"cidDetails":   cidDetails,
				"totalChecks":  sumCounts(cidCounts),
			})
		}
	}

	// Sort payees with multiple CIDs by number of different CIDs
	sort.Slice(result.PayeesWithMultipleCIDs, func(i, j int) bool {
		return result.PayeesWithMultipleCIDs[i]["cidCount"].(int) > 
			   result.PayeesWithMultipleCIDs[j]["cidCount"].(int)
	})

	result.TotalChecks = totalChecks
	result.ChecksWithCID = checksWithCID
	result.ChecksWithoutCID = checksWithoutCID
	result.UniquePayees = len(payeeGroups)

	return result, nil
}

// Helper function to sum counts in a map
func sumCounts(counts map[string]int) int {
	total := 0
	for _, count := range counts {
		total += count
	}
	return total
}