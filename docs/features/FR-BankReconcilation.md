# FR-BankReconcilation

## 1. Purpose
Explain how bank reconciliation works (in the style of Xero) and define requirements for building a similar reconciliation assistant. The goal is to align internal transaction records with external bank statement lines, surface suggested matches, provide automation via rules, support manual reconciliation, and produce summary reports with auditability and learning.

## 2. Key Concepts & Definitions
- **Bank Statement Line**: A line item imported from the bank (via automatic feed or file upload) representing money in or out on a specific date/amount. citeturn0search0
- **Account Transaction**: A recorded internal transaction (invoices paid, payments, transfers, etc.), which may or may not have cleared the bank. citeturn0search3
- **Match**: Linking a bank statement line to one or more account transactions; suggested using heuristics and historical behavior. citeturn0search0
- **Bank Rule**: User-defined, ordered rule to automate categorization or matching for recurring/predictable transactions. citeturn0search8
- **Reconciliation Session**: A snapshot for a given period reflecting statement balance, internal balance, outstanding items, and discrepancies. citeturn0search6
- **Manual Reconciliation**: Marking transactions as reconciled when the bank line is missing or delayed, creating reconciliation metadata to prevent repeated mismatches. citeturn0search12turn0search9

## 3. System Architecture Overview
### Core Entities
- `BankStatementLine`: id, bank_account_id, date, amount, description, reference, imported_source, matched_flag, reconciliation_timestamp.  
- `AccountTransaction`: id, type (invoice, bill, payment, transfer), amount, date, status (cleared/uncleared), counterparty, metadata/tags.  
- `Match`: links statement lines to account transactions with confidence score, user_override flag, timestamp, and approver.  
- `BankRule`: id, name, conditions (description patterns, amount ranges, counterparties), action (categorize, suggest match, create transaction), and priority.  
- `ReconciliationSession`: period snapshot capturing balances and outstanding items.

### Services / Components
- **Ingestion Service**: Imports bank statement lines via secure feeds or uploads. citeturn0search7  
- **Transaction Collector**: Aggregates candidate internal transactions (receipts, payments, transfers, etc.). citeturn0search10  
- **Matching Engine**: Suggests matches based on heuristics, historical patterns, and bank rules. citeturn0search0turn0search16  
- **Rule Engine**: Applies ordered bank rules to automate matching/categorization. citeturn0search8  
- **UI Layer**: Dual-pane reconciliation interface, quick actions, overrides, and summaries. citeturn0search0  
- **Reporting/Audit Module**: Produces reconciliation summaries and maintains audit trails. citeturn0search6  
- **Learning/Feedback Loop**: Detects repeated user behavior to suggest rules. citeturn0search11  

## 4. Detailed Workflow
1. **Import Bank Data**: Securely ingest statement lines (feed or file) to be reconciled. citeturn0search7  
2. **Collect Internal Transactions**: Gather potential matching internal transactions. citeturn0search10  
3. **Generate Suggested Matches**: Use matching logic combining amount/date tolerance, name similarity, and past behavior; rank candidates. citeturn0search0turn0search16  
4. **Apply Bank Rules**: Auto-categorize or auto-match based on ordered rules. citeturn0search8  
5. **User Review & Action**: Present suggestions; allow accept, create new, split, override. citeturn0search0turn0search16  
6. **Exception Handling / Manual Reconciliation**: Support marking items reconciled when external data missing. citeturn0search12turn0search9  
7. **Completion Summary**: Show final balances, outstanding items, and discrepancies. citeturn0search6  

## 5. Bank Rule Specification
- Conditions: description contains/regex, amount range, counterparties, frequency.  
- Actions: assign category, suggest or auto-apply match, create standard transaction template, apply tags.  
- Priority: Rules ordered; first match may take precedence unless overridden.  
- Suggest rule creation when users repeatedly accept similar manual matches. citeturn0search11
- Preview before bulk application. citeturn0search8

## 6. Matching Algorithm Details
```pseudo
for each bank_statement_line:
    candidates = find account_transactions where 
        abs(transaction.amount - line.amount) <= tolerance AND
        date_within_window(transaction.date, line.date, days=3)

    rank candidates by:
        - exact amount/date
        - historical match patterns
        - fuzzy payee/name similarity
        - existing bank rules
    if top candidate.score > threshold:
        suggest match
    else:
        prompt user to create new transaction or split
```
Include confidence scoring; incorporate feedback loop to surface rule suggestions for repeated patterns. citeturn0search16turn0search11

## 7. UI/UX Guidelines
- Dual-pane view (statement lines vs. suggested matches). citeturn0search0  
- Inline actions: accept, create, split, override.  
- Highlight automation and rule-applied items.  
- Provide clear audit trail and ability to undo/adjust.  
- Summary dashboard with reconciliation progress and discrepancies. citeturn0search6  

## 8. Edge Cases & Exception Handling
- Partial matches (one statement line to multiple transactions and vice versa).  
- Currency conversions for multi-currency accounts.  
- Duplicate imports or transactions.  
- Timing differences (late clearing).  
- Missing bank feed lines requiring manual reconciliation. citeturn0search12turn0search9  

## 9. Reporting & Audit Trails
- Statement vs. internal balance reconciliation summary. citeturn0search6  
- Drill-down capability on each match.  
- Exportable reports for compliance or review.  
- Historical audit logs with who accepted/overrode and timestamps.  

## 10. Example Scenarios / Use Cases
- Recurring vendor payment auto-matched by rule.  
- Customer payment split across two invoices.  
- Missing bank feed line manually marked reconciled. citeturn0search12turn0search9  

## 11. Metrics & Health Checks
- Percentage of automated matches accepted.  
- Time to reconcile per period.  
- Discrepancy rate (difference between statement and internal).  
- Rule effectiveness (how often suggested rules match and are accepted).  

## 12. API Contracts (suggested)
- `POST /bank-statements/import` – ingest statement lines.  
- `GET /transactions/candidates` – fetch candidate internal transactions for a statement line.  
- `POST /matches/suggest` – run matching logic and return suggestions.  
- `POST /matches/apply` – apply/confirm a match.  
- `POST /rules` – create/update bank rules.  
- `GET /reconciliation-summary` – retrieve session summary.  
- `GET /audit-log` – fetch reconciliation audit history.  

## 13. Security & Permissions
- Role-based access: who can reconcile, create/override matches, manage rules.  
- Immutable audit log (with versioning of overrides).  
- Data integrity checks to prevent tampering.  

## 14. Starter Prompt for AI
> Build me a bank reconciliation assistant like Xero. It should ingest bank statement lines, pull internal recorded transactions, suggest matches based on amount/date similarity and learned patterns, allow the user to accept or create transactions, support user-defined ordered bank rules for automation, handle manual reconciliation when statement lines are missing, and produce a reconciliation summary report showing the statement balance, internal balance, and outstanding items. Include confidence scoring for suggestions, an audit trail of all actions, and a feedback loop that proposes new rules when repetitive matches are accepted. citeturn0search0turn0search8turn0search6turn0search11turn0search16

## 15. References
Include the original detailed explanation sources (Xero documentation, blog posts, product UX descriptions, etc.) as captured in the internal citation markers above.
