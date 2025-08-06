# DBF Date Parsing Documentation

## Issue Summary
Visual FoxPro DBF files contain date fields that need proper parsing. Currently there's a discrepancy between what we see in raw CHECKS.DBF (multiple unchecked checks from 2025) vs. what the audit shows (only 1 outstanding check).

## Current Date Parsing Methods

### Method 1: Direct time.Time (Used in AuditSingleBankAccount)
```go
// Check if it's already a time.Time object from DBF
if t, ok := row[dateIdx].(time.Time); ok {
    recDate = t
}
```
**Location**: `main.go:1698` in AuditSingleBankAccount function  
**Status**: ✅ Working correctly

### Method 2: String Conversion + Multiple Format Parsing  
```go
if dateStr := fmt.Sprintf("%v", row[dateIdx]); dateStr != "" {
    // Handle various string formats
    for _, format := range []string{
        "2006-01-02 15:04:05 -0700 MST", // Go time.Time string format
        "2006-01-02T15:04:05Z", 
        "2006-01-02T15:04:05", 
        "2006-01-02", 
        "01/02/2006", 
        "1/2/2006"} {
        if parsedDate, err := time.Parse(format, dateStr); err == nil {
            recDate = parsedDate
            break
        }
    }
}
```
**Location**: `main.go:1409-1422` in AuditBankReconciliation function  
**Status**: ✅ Fallback method working

### Method 3: Simple String Conversion (Used in GetOutstandingChecks)
```go
if dateIdx != -1 && len(row) > dateIdx && row[dateIdx] != nil {
    check["date"] = fmt.Sprintf("%v", row[dateIdx])
}
```
**Location**: `main.go:~130` in GetOutstandingChecks function  
**Status**: ⚠️ No actual date parsing, just string conversion

## Critical Issue: Outstanding Checks Calculation

### Problem
The `RefreshOutstandingChecks` function in `internal/database/balance_cache.go` **DOES NOT** consider dates when calculating outstanding checks:

```go
// Only filters by these criteria:
if !isCleared && !isVoided {
    totalOutstanding += parseFloat(row[amountIdx])
    checkCount++
}
```

**Missing**: Date filtering logic  
**Result**: Old checks from years ago are included in outstanding total

### Evidence
- **CHECKS.DBF shows**: Multiple 2025 checks marked `LCLEARED = False`
- **Audit shows**: Only 1 outstanding check for $3,250.00
- **Discrepancy**: The cached outstanding total doesn't match visible uncashed checks

## Recommended Fix

### Option 1: Add Date Filtering to RefreshOutstandingChecks
Filter checks to only include those from recent periods (e.g., last 12 months):

```go
// Add date column detection
var dateIdx int = -1
for i, col := range checksColumns {
    colUpper := strings.ToUpper(col)
    if colUpper == "DCHECKDATE" {
        dateIdx = i
        break
    }
}

// Add date filtering in the check processing loop
if !isCleared && !isVoided {
    // Add date filter
    if dateIdx != -1 && len(row) > dateIdx && row[dateIdx] != nil {
        if checkDate, ok := row[dateIdx].(time.Time); ok {
            // Only include checks from last 12 months
            cutoffDate := time.Now().AddDate(-1, 0, 0)
            if checkDate.Before(cutoffDate) {
                continue // Skip old checks
            }
        }
    }
    
    totalOutstanding += parseFloat(row[amountIdx])
    checkCount++
}
```

### Option 2: Make Date Cutoff Configurable
Allow users to configure how far back to look for outstanding checks:
- 30 days (current period)
- 90 days (quarterly)  
- 365 days (annual)
- All time (no date filter)

## DBF Date Field Standards

### Common DBF Date Column Names
- `DCHECKDATE` - Check date (CHECKS.DBF)
- `DRECDATE` - Reconciliation date (CHECKREC.DBF)
- `DDATE` - Generic date field
- `DATE` - Simple date field

### Data Types Returned by go-dbase Library
1. **time.Time** - Preferred, direct parsing
2. **string** - Fallback, requires format detection
3. **nil** - Empty/null date

### Best Practice Date Parsing Function
```go
func parseDBFDate(value interface{}) (time.Time, error) {
    // Direct time.Time (preferred)
    if t, ok := value.(time.Time); ok {
        return t, nil
    }
    
    // String parsing (fallback)
    if dateStr := fmt.Sprintf("%v", value); dateStr != "" && dateStr != "<nil>" {
        // Try multiple common formats
        formats := []string{
            "2006-01-02 15:04:05 -0700 MST", // Go time.Time string
            "2006-01-02T15:04:05Z",          // RFC3339 UTC
            "2006-01-02T15:04:05",           // RFC3339 local
            "2006-01-02",                    // Date only
            "01/02/2006",                    // US format
            "1/2/2006",                      // US format short
            "2006/01/02",                    // ISO-ish
        }
        
        for _, format := range formats {
            if parsedDate, err := time.Parse(format, dateStr); err == nil {
                return parsedDate, nil
            }
        }
    }
    
    return time.Time{}, fmt.Errorf("unable to parse date: %v", value)
}
```

## Action Items

1. ✅ **Document the issue** (this file)
2. 🔄 **Fix RefreshOutstandingChecks** to include date filtering  
3. 🔄 **Add configuration** for outstanding check date cutoff
4. 🔄 **Test with real data** to verify accuracy
5. 🔄 **Update CLAUDE.md** with date parsing standards

## Testing Strategy

1. **Before Fix**: Note current outstanding check totals
2. **Apply Fix**: Add date filtering to RefreshOutstandingChecks
3. **Refresh Balances**: Run account balance refresh
4. **Compare Results**: Verify outstanding checks match visible data
5. **Edge Cases**: Test with various date formats and null values

---
**Created**: 2025-08-05  
**Issue**: Outstanding checks calculation ignores dates  
**Impact**: Inaccurate bank reconciliation audits  
**Priority**: High - affects core banking functionality