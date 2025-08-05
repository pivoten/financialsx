# DBF File Structure Documentation

This document details the structure and field definitions for all DBF files used in the FinancialsX system.

## GLMASTER.dbf - General Ledger Master

**Purpose**: Contains all general ledger transactions and entries

**Total Columns**: 28

| Column Index | Field Name | Data Type | Description |
|--------------|------------|-----------|-------------|
| 0 | CIDGLMA | Character | GL Master ID |
| 1 | CBATCH | Character | Batch Number |
| 2 | CYEAR | Character | Year |
| 3 | CPERIOD | Character | Period |
| 4 | CSOURCE | Character | Source |
| 5 | CREF | Character | Reference |
| 6 | DDATE | Date | Transaction Date |
| 7 | CDESC | Character | Description |
| 8 | CACCTNO | Character | **Account Number** (Key field for balance calculations) |
| 9 | CUNITNO | Character | Unit Number |
| 10 | CDEPTNO | Character | Department Number |
| 11 | NDEBITS | Numeric | **Debit Amount** (Used for balance calculations) |
| 12 | NCREDITS | Numeric | **Credit Amount** (Used for balance calculations) |
| 13 | CDMBATCH | Character | DM Batch |
| 14 | CID | Character | ID |
| 15 | CBUNCH | Character | Bunch |
| 16 | DLASTMODIF | Date | Last Modified Date |
| 17 | CUSER | Character | User |
| 18 | CCATCODE | Character | Category Code |
| 19 | MNOTES | Memo | Notes |
| 20 | CIDCHEC | Character | Check ID |
| 21 | CPRODYR | Character | Production Year |
| 22 | CPRODPRD | Character | Production Period |
| 23 | CAFENO | Character | AFE Number |
| 24 | DADDED | Date | Date Added |
| 25 | CADDEDBY | Character | Added By |
| 26 | DCHANGED | Date | Date Changed |
| 27 | CCHANGEDBY | Character | Changed By |

### Key Fields for Banking
- **CACCTNO** (Column 8): Account number to match with bank accounts
- **NDEBITS** (Column 11): Debit amounts (increase bank balance)
- **NCREDITS** (Column 12): Credit amounts (decrease bank balance)
- **Balance Calculation**: `NDEBITS - NCREDITS` for each account

---

## COA.dbf - Chart of Accounts

**Purpose**: Defines the chart of accounts structure including bank account identification

**Key Columns**:

| Column Index | Field Name | Data Type | Description |
|--------------|------------|-----------|-------------|
| 0 | CACCTNO | Character | **Account Number** (Primary key) |
| 1 | NACCTTYPE | Numeric | Account Type |
| 2 | CACCTDESC | Character | Account Description |
| 3 | CPARENT | Character | Parent Account |
| 4 | LACCTUNIT | Logical | Account Unit Flag |
| 5 | LACCTDEPT | Logical | Department Flag |
| 6 | LBANKACCT | Logical | **Bank Account Flag** (TRUE = Bank Account) |

### Key Fields for Banking
- **CACCTNO** (Column 0): Account number (must match GLMASTER.CACCTNO)
- **LBANKACCT** (Column 6): When TRUE, identifies the account as a bank account
- **CACCTDESC** (Column 2): Display name for the bank account

---

## CHECKS.dbf - Check Transactions

**Purpose**: Contains check transaction records for audit and reconciliation

**Total Columns**: 25

| Column Index | Field Name | Data Type | Description |
|--------------|------------|-----------|-------------|
| 0 | CCHECKNO | Character | **Check Number** (Key field for identification) |
| 1 | CID | Character | Check ID |
| 2 | CPAYEE | Character | **Payee Name** (Who the check is paid to) |
| 3 | CIDCHEC | Character | Check ID Reference |
| 4 | CIDTYPE | Character | Check Type |
| 5 | DCHECKDATE | Date | **Check Date** (When check was written) |
| 6 | NAMOUNT | Numeric | **Check Amount** (Dollar amount) |
| 7 | CPERIOD | Character | Period |
| 8 | CYEAR | Character | Year |
| 9 | CMEMO | Character | Memo/Description |
| 10 | LPRINTED | Logical | Printed Flag |
| 11 | LVOID | Logical | Void Flag |
| 12 | NVOIDAMT | Numeric | Void Amount |
| 13 | CGROUP | Character | Group |
| 14 | LCLEARED | Logical | **Cleared Flag** (TRUE = cleared by bank) |
| 15 | CACCTNO | Character | **Account Number** (Bank account) |
| 16 | LMANUAL | Logical | Manual Entry Flag |
| 17 | CENTRYTYPE | Character | Entry Type |
| 18 | CBATCH | Character | **Batch Number** (Used for audit matching) |
| 19 | DPOSTDATE | Date | Post Date |
| 20 | LHIST | Logical | History Flag |
| 21 | CSOURCE | Character | Source |
| 22 | LDELETED | Logical | Deleted Flag |
| 23 | DRECDATE | Date | Record Date |
| 24 | LDEPOSITED | Logical | Deposited Flag |

### Key Fields for Outstanding Checks
- **CCHECKNO** (Column 0): Check number for identification
- **LCLEARED** (Column 14): When FALSE, check is outstanding
- **DCHECKDATE** (Column 5): Used to calculate days outstanding
- **NAMOUNT** (Column 6): Check amount
- **CPAYEE** (Column 2): Payee name
- **CACCTNO** (Column 15): Bank account number

### Outstanding Check Logic
- **Outstanding**: `LCLEARED = FALSE` and `LVOID = FALSE`
- **Days Outstanding**: `TODAY - DCHECKDATE`
- **Total Outstanding**: SUM(NAMOUNT) WHERE LCLEARED = FALSE

---

## EXPENSE.dbf - Expense Records

**Purpose**: Contains expense transaction records

**Status**: Structure to be documented
**Record Count**: ~48,829 active records (varies by company)

---

## INCOME.dbf - Revenue Records

**Purpose**: Contains revenue/income transaction records

**Status**: Structure to be documented

---

## WELLS.dbf - Well Information

**Purpose**: Contains well master data and information

**Status**: Structure to be documented
**Record Count**: ~2,640 active records (varies by company)

---

## Field Naming Conventions

### Prefixes
- **C**: Character/String fields
- **N**: Numeric fields
- **D**: Date fields
- **L**: Logical/Boolean fields
- **M**: Memo fields

### Common Field Patterns
- **CACCTNO**: Account Number (found in multiple files)
- **CBATCH**: Batch Number (used for grouping transactions)
- **DADDED/CADDEDBY**: Audit trail for record creation
- **DCHANGED/CCHANGEDBY**: Audit trail for record modification
- **DLASTMODIF**: Last modification timestamp

---

## Data Integration Notes

### Banking Module Integration
1. **Account Discovery**: COA.dbf filtered by LBANKACCT = TRUE
2. **Balance Calculation**: GLMASTER.dbf summed by CACCTNO
3. **Balance Formula**: SUM(NDEBITS - NCREDITS) WHERE CACCTNO = account_number

### Audit System Integration
1. **Check Matching**: CHECKS.dbf matched with GLMASTER.dbf
2. **Primary Key**: Check number or CBATCH field
3. **Amount Comparison**: Check amounts vs GL amounts for discrepancy detection

---

*Last Updated: August 2, 2025*
*This documentation is based on analysis of FinancialsX desktop application DBF file structures*