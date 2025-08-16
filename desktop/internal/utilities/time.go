package utilities

import "time"

// GetFreshnessStatus determines the freshness status based on time
func GetFreshnessStatus(lastUpdated time.Time) string {
	age := time.Since(lastUpdated).Hours()
	if age < 1 {
		return "fresh"
	} else if age < 24 {
		return "aging"
	}
	return "stale"
}