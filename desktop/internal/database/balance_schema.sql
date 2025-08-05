-- Bank Account Balance Caching System
-- This system caches GL balances and calculates available balances

-- Main balance cache table
CREATE TABLE IF NOT EXISTS account_balances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    company_name TEXT NOT NULL,
    account_number TEXT NOT NULL,
    account_name TEXT NOT NULL,
    account_type INTEGER NOT NULL,
    
    -- GL Balance (from GLMASTER.dbf scan)
    gl_balance DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    gl_last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    gl_record_count INTEGER NOT NULL DEFAULT 0, -- Number of GL entries processed
    
    -- Outstanding Checks (from CHECKS.dbf scan)
    outstanding_checks_total DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    outstanding_checks_count INTEGER NOT NULL DEFAULT 0,
    outstanding_checks_last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Calculated Bank Balance (GL + Uncleared Checks)
    bank_balance DECIMAL(15,2) GENERATED ALWAYS AS (gl_balance + outstanding_checks_total) STORED,
    
    -- Metadata
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_bank_account BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Additional info as JSON
    metadata JSON DEFAULT '{}', -- Store things like last check number, reconciliation status, etc.
    
    UNIQUE(company_name, account_number)
);

-- Balance update history for audit trail
CREATE TABLE IF NOT EXISTS balance_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_balance_id INTEGER NOT NULL,
    company_name TEXT NOT NULL,
    account_number TEXT NOT NULL,
    
    -- What changed
    change_type TEXT NOT NULL CHECK (change_type IN ('gl_refresh', 'checks_refresh', 'manual_adjustment', 'reconciliation')),
    
    -- Before/After values
    old_gl_balance DECIMAL(15,2),
    new_gl_balance DECIMAL(15,2),
    old_outstanding_total DECIMAL(15,2),
    new_outstanding_total DECIMAL(15,2),
    old_available_balance DECIMAL(15,2),
    new_available_balance DECIMAL(15,2),
    
    -- Change details
    change_reason TEXT,
    changed_by TEXT, -- username
    change_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Additional context
    metadata JSON DEFAULT '{}',
    
    FOREIGN KEY (account_balance_id) REFERENCES account_balances(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_account_balances_company_account 
    ON account_balances(company_name, account_number);

CREATE INDEX IF NOT EXISTS idx_account_balances_company_active 
    ON account_balances(company_name, is_active, is_bank_account);

CREATE INDEX IF NOT EXISTS idx_balance_history_account_timestamp 
    ON balance_history(account_balance_id, change_timestamp);

CREATE INDEX IF NOT EXISTS idx_balance_history_company_account 
    ON balance_history(company_name, account_number, change_timestamp);

-- Trigger to update the updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_account_balances_timestamp 
    AFTER UPDATE ON account_balances
    FOR EACH ROW
BEGIN
    UPDATE account_balances 
    SET updated_at = CURRENT_TIMESTAMP 
    WHERE id = NEW.id;
END;

-- View for easy balance retrieval with age information
CREATE VIEW IF NOT EXISTS account_balance_summary AS
SELECT 
    ab.*,
    -- Age calculations
    ROUND((julianday('now') - julianday(ab.gl_last_updated)) * 24, 2) as gl_age_hours,
    ROUND((julianday('now') - julianday(ab.outstanding_checks_last_updated)) * 24, 2) as checks_age_hours,
    
    -- Freshness indicators
    CASE 
        WHEN (julianday('now') - julianday(ab.gl_last_updated)) * 24 > 24 THEN 'stale'
        WHEN (julianday('now') - julianday(ab.gl_last_updated)) * 24 > 4 THEN 'aging'
        ELSE 'fresh'
    END as gl_freshness,
    
    CASE 
        WHEN (julianday('now') - julianday(ab.outstanding_checks_last_updated)) * 24 > 4 THEN 'stale'
        WHEN (julianday('now') - julianday(ab.outstanding_checks_last_updated)) * 24 > 1 THEN 'aging'
        ELSE 'fresh'
    END as checks_freshness,
    
    -- Last update info
    (SELECT change_timestamp FROM balance_history bh 
     WHERE bh.account_balance_id = ab.id 
     ORDER BY change_timestamp DESC LIMIT 1) as last_change_timestamp,
     
    (SELECT changed_by FROM balance_history bh 
     WHERE bh.account_balance_id = ab.id 
     ORDER BY change_timestamp DESC LIMIT 1) as last_changed_by

FROM account_balances ab
WHERE ab.is_active = TRUE;