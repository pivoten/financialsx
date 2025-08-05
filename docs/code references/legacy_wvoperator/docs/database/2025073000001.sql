-- main.sql: Added owner master table + associated accounting/production interest tables
PRAGMA foreign_keys = ON;

-- 1) Operator (top-level company info)
CREATE TABLE operator (
  operator_id               TEXT    PRIMARY KEY,
  sw_version_cid_comp       TEXT,
  producer_name             TEXT,
  producer_code             TEXT,
  address                   TEXT,
  city                      TEXT,
  state                     TEXT,
  zip_code                  TEXT,
  dba_attn_agent            TEXT,
  phone                     TEXT,
  email                     TEXT,
  created_at                DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at                DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 2) Generic lookup table for enums
CREATE TABLE lookup (
  domain      TEXT    NOT NULL,   -- 'gas', 'status', 'formation', etc.
  code        TEXT    NOT NULL,
  description TEXT    NOT NULL,
  PRIMARY KEY(domain, code)
);

-- 3) Well master table (Schedule 1 & 2)
CREATE TABLE well (
  well_id                    TEXT    PRIMARY KEY,
  operator_id                TEXT    NOT NULL REFERENCES operator(operator_id) ON DELETE CASCADE,
  sw_cwell_id                TEXT    NOT NULL UNIQUE,

  -- Schedule 1 fields
  county_name                TEXT,
  county_number              TEXT,
  nra_number                 TEXT,
  api_number                 TEXT,
  well_name                  TEXT,
  land_acreage               REAL,
  lease_acreage              REAL,

  -- Schedule 2 fields
  status_domain              TEXT    NOT NULL DEFAULT 'status',
  status_code                TEXT    NOT NULL,
  formation_domain           TEXT    NOT NULL DEFAULT 'formation',
  formation_code             TEXT    NOT NULL,
  initial_production_date    DATE,

  created_at                 DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at                 DATETIME DEFAULT CURRENT_TIMESTAMP,

  FOREIGN KEY(status_domain, status_code)     REFERENCES lookup(domain, code),
  FOREIGN KEY(formation_domain, formation_code) REFERENCES lookup(domain, code)
);
CREATE INDEX idx_well_sw_cwell_id    ON well(sw_cwell_id);
CREATE INDEX idx_well_status_code     ON well(status_code);
CREATE INDEX idx_well_formation_code  ON well(formation_code);

-- 4) Well-gas bridge (many-to-many for gas types)
CREATE TABLE well_gas (
  well_id      TEXT NOT NULL REFERENCES well(well_id) ON DELETE CASCADE,
  gas_domain   TEXT NOT NULL DEFAULT 'gas',
  gas_code     TEXT NOT NULL,
  created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY(well_id, gas_code),
  FOREIGN KEY(gas_domain, gas_code) REFERENCES lookup(domain, code)
);

-- 5) Owner master table
CREATE TABLE owner (
  owner_id                   TEXT    PRIMARY KEY,
  sw_cowner_id               TEXT    NOT NULL UNIQUE,
  last_name                  TEXT    NOT NULL,
  first_name                 TEXT,
  address                    TEXT,
  created_at                 DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at                 DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_owner_sw_cowner_id ON owner(sw_cowner_id);

-- 6) Accounting financial snapshot (with reporting_period_year)
CREATE TABLE well_financial_accounting (
  id                                INTEGER PRIMARY KEY AUTOINCREMENT,
  well_id                           TEXT    NOT NULL REFERENCES well(well_id) ON DELETE CASCADE,
  reporting_period_year             INTEGER NOT NULL,

  production_total_bbl              REAL,
  production_total_mcf              REAL,
  production_total_ngl              REAL,

  revenue_gross_oil                 REAL,
  revenue_gross_gas                 REAL,
  revenue_gross_ngl                 REAL,

  revenue_working_interest_net_oil  REAL,
  revenue_working_interest_net_gas  REAL,
  revenue_working_interest_net_ngl  REAL,

  expenses_working_interest_gross_oil REAL,
  expenses_working_interest_gross_gas REAL,
  expenses_working_interest_gross_ngl REAL,

  revenue_royalty_interest_net_oil  REAL,
  revenue_royalty_interest_net_gas  REAL,
  revenue_royalty_interest_net_ngl  REAL,

  total_revenue_working_interest    REAL,
  total_revenue_royalty_interest    REAL,

  created_at                        DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at                        DATETIME DEFAULT CURRENT_TIMESTAMP,

  UNIQUE(well_id, reporting_period_year)
);
CREATE INDEX idx_wfa_well_reporting_period_year ON well_financial_accounting(well_id, reporting_period_year);

-- 7) Production financial snapshot (with reporting_period_year)
CREATE TABLE well_financial_production (
  id                                INTEGER PRIMARY KEY AUTOINCREMENT,
  well_id                           TEXT    NOT NULL REFERENCES well(well_id) ON DELETE CASCADE,
  reporting_period_year             INTEGER NOT NULL,

  production_total_bbl              REAL,
  production_total_mcf              REAL,
  production_total_ngl              REAL,

  revenue_gross_oil                 REAL,
  revenue_gross_gas                 REAL,
  revenue_gross_ngl                 REAL,

  revenue_working_interest_net_oil  REAL,
  revenue_working_interest_net_gas  REAL,
  revenue_working_interest_net_ngl  REAL,

  expenses_working_interest_gross_oil REAL,
  expenses_working_interest_gross_gas REAL,
  expenses_working_interest_gross_ngl REAL,

  revenue_royalty_interest_net_oil  REAL,
  revenue_royalty_interest_net_gas  REAL,
  revenue_royalty_interest_net_ngl  REAL,

  total_revenue_working_interest    REAL,
  total_revenue_royalty_interest    REAL,

  created_at                        DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at                        DATETIME DEFAULT CURRENT_TIMESTAMP,

  UNIQUE(well_id, reporting_period_year)
);
CREATE INDEX idx_wfp_well_reporting_period_year ON well_financial_production(well_id, reporting_period_year);

-- 8) Accounting owner interest snapshot (linked to owner)
CREATE TABLE well_owner_interest_accounting (
  id                                INTEGER PRIMARY KEY AUTOINCREMENT,
  well_id                           TEXT    NOT NULL REFERENCES well(well_id) ON DELETE CASCADE,
  owner_id                          TEXT    NOT NULL REFERENCES owner(owner_id) ON DELETE CASCADE,
  reporting_period_year             INTEGER NOT NULL,

  decimal_interest                  REAL    NOT NULL,
  income                            REAL    NOT NULL,

  created_at                        DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at                        DATETIME DEFAULT CURRENT_TIMESTAMP,

  UNIQUE(well_id, owner_id, reporting_period_year)
);
CREATE INDEX idx_woi_a_well_reporting_period_year ON well_owner_interest_accounting(well_id, reporting_period_year);

-- 9) Production owner interest snapshot (linked to owner)
CREATE TABLE well_owner_interest_production (
  id                                INTEGER PRIMARY KEY AUTOINCREMENT,
  well_id                           TEXT    NOT NULL REFERENCES well(well_id) ON DELETE CASCADE,
  owner_id                          TEXT    NOT NULL REFERENCES owner(owner_id) ON DELETE CASCADE,
  reporting_period_year             INTEGER NOT NULL,

  decimal_interest                  REAL    NOT NULL,
  income                            REAL    NOT NULL,

  created_at                        DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at                        DATETIME DEFAULT CURRENT_TIMESTAMP,

  UNIQUE(well_id, owner_id, reporting_period_year)
);
CREATE INDEX idx_woi_p_well_reporting_period_year ON well_owner_interest_production(well_id, reporting_period_year);
