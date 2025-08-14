export interface SherWareForm {
  id: string
  title: string
  formName: string
  description: string
}

export interface SherWareCategory {
  category: string
  forms: SherWareForm[]
}

export const sherWareForms: SherWareCategory[] = [
  {
    category: "Company Management",
    forms: [
      { id: "compsetup", title: "Company Setup", formName: "compsetup", description: "Configure company settings" }
    ]
  },
  {
    category: "File Operations",
    forms: [
      { id: "filebackup", title: "File Backup", formName: "filebackup.scx", description: "Backup company data files" },
      { id: "filerestore", title: "File Restore", formName: "filerestore.scx", description: "Restore company data files" },
      { id: "utftpsend", title: "FTP Send", formName: "utftpsend.scx", description: "Send files via FTP" }
    ]
  },
  {
    category: "System Setup",
    forms: [
      { id: "preferences", title: "Preferences", formName: "preferences.scx", description: "User preferences and settings" },
      { id: "company", title: "Company Information", formName: "company.scx", description: "Company details and configuration" },
      { id: "frmRegistryEditor", title: "Registry Editor", formName: "frmRegistryEditor", description: "Edit system registry settings" },
      { id: "frmAccountEditor", title: "Account Editor", formName: "frmAccountEditor", description: "Edit account settings" },
      { id: "frmSetLogonPassword", title: "Set Logon Password", formName: "frmSetLogonPassword", description: "Change login password" },
      { id: "changereg", title: "Change Registration", formName: "changereg.scx", description: "Update registration information" }
    ]
  },
  {
    category: "General Ledger",
    forms: [
      { id: "glcoa", title: "Chart of Accounts", formName: "glcoa.scx", description: "Manage GL chart of accounts" },
      { id: "glacctyp", title: "Account Types", formName: "glacctyp.scx", description: "GL account type maintenance" },
      { id: "glassets", title: "GL Assets", formName: "glassets.scx", description: "Fixed assets management" },
      { id: "gldept", title: "GL Departments", formName: "gldept.scx", description: "Department code maintenance" },
      { id: "glterms", title: "Payment Terms", formName: "glterms.scx", description: "Payment terms setup" },
      { id: "gljourn", title: "Journal Entry", formName: "gljourn.scx", description: "General journal entries" },
      { id: "glrecur", title: "Recurring GL Entries", formName: "glrecur.scx", description: "Setup recurring journal entries" },
      { id: "glcloseprd", title: "Close GL Period", formName: "glcloseprd.scx", description: "Close general ledger period" }
    ]
  },
  {
    category: "Accounts Receivable",
    forms: [
      { id: "arcust", title: "AR Customers", formName: "arcust.scx", description: "Customer maintenance" },
      { id: "aritems", title: "AR Items", formName: "aritems.scx", description: "Inventory items setup" },
      { id: "arfinchg", title: "Finance Charges", formName: "arfinchg.scx", description: "Finance charge settings" },
      { id: "arsalestx", title: "Sales Tax", formName: "arsalestx.scx", description: "Sales tax configuration" },
      { id: "arinv", title: "AR Invoice Entry", formName: "arinv.scx", description: "Create and edit invoices" },
      { id: "arrecpmt", title: "Receive Payment", formName: "arrecpmt.scx", description: "Process customer payments" },
      { id: "arfingen", title: "Generate Finance Charges", formName: "arfingen.scx", description: "Calculate finance charges" },
      { id: "arrecur", title: "Recurring AR Entries", formName: "arrecur.scx", description: "Setup recurring invoices" }
    ]
  },
  {
    category: "Accounts Payable",
    forms: [
      { id: "apvendor", title: "AP Vendors", formName: "apvendor.scx", description: "Vendor maintenance" },
      { id: "apbill", title: "Bill Entry (Classic)", formName: "apbill.scx", description: "Enter vendor bills" },
      { id: "apbillnew", title: "Bill Entry (New)", formName: "apbillnew.scx", description: "Enhanced bill entry" },
      { id: "apbill2", title: "Bill Entry (Alternate)", formName: "apbill2.scx", description: "Alternative bill entry form" },
      { id: "apbillpay", title: "Pay Bills", formName: "apbillpay.scx", description: "Select and pay vendor bills" },
      { id: "aprecur", title: "Recurring AP Entries", formName: "aprecur.scx", description: "Setup recurring bills" }
    ]
  },
  {
    category: "Cash Management",
    forms: [
      { id: "csreceipt", title: "Cash Receipt", formName: "csreceipt.scx", description: "Record cash receipts" },
      { id: "csdisb", title: "Cash Disbursement", formName: "csdisb.scx", description: "Record cash disbursements" },
      { id: "csdeposit", title: "Bank Deposit", formName: "csdeposit.scx", description: "Prepare bank deposits" },
      { id: "csrecond", title: "Bank Reconciliation", formName: "csrecond.scx", description: "Reconcile bank statements" },
      { id: "cstransfer", title: "Cash Transfer", formName: "cstransfer.scx", description: "Transfer between accounts" },
      { id: "csexpense", title: "Cash Expense", formName: "csexpense.scx", description: "Record cash expenses" },
      { id: "cspospay2", title: "POS Payment", formName: "cspospay2.scx", description: "Point of sale payments" },
      { id: "csregister", title: "Check Register", formName: "csregister.scx", description: "View check register" },
      { id: "cschecks", title: "Check Maintenance", formName: "cschecks.scx", description: "Manage checks" },
      { id: "csvoidreissue", title: "Void/Reissue Checks", formName: "csvoidreissue.scx", description: "Void and reissue checks" },
      { id: "cschkpg", title: "Check Printing", formName: "cschkpg.scx", description: "Print checks" }
    ]
  },
  {
    category: "Oil & Gas - Wells",
    forms: [
      { id: "dmwellinfo", title: "Well Information", formName: "dmwellinfo.scx", description: "Well master data" },
      { id: "dmdoi", title: "Division of Interest", formName: "dmdoi.scx", description: "DOI maintenance" },
      { id: "dmoperators", title: "Operators", formName: "dmoperators.scx", description: "Operator information" },
      { id: "dmpumpers", title: "Pumpers", formName: "dmpumpers.scx", description: "Pumper information" },
      { id: "dmmeters", title: "Meters", formName: "dmmeters.scx", description: "Meter maintenance" },
      { id: "dmtanks", title: "Tanks", formName: "dmtanks.scx", description: "Tank information" },
      { id: "dmgroups", title: "Well Groups", formName: "dmgroups.scx", description: "Well grouping setup" },
      { id: "dmwellhistory", title: "Well History", formName: "dmwellhistory.scx", description: "Historical well data" },
      { id: "dmwelltot", title: "Well Totals", formName: "dmwelltot.scx", description: "Well production totals" },
      { id: "plugwell", title: "Plug Well", formName: "plugwell.scx", description: "Well plugging information" },
      { id: "plugwellbal", title: "Plug Well Balances", formName: "plugwellbal.scx", description: "Plugging cost balances" }
    ]
  },
  {
    category: "Oil & Gas - Revenue",
    forms: [
      { id: "dmrevsrc", title: "Revenue Sources", formName: "dmrevsrc.scx", description: "Revenue source setup" },
      { id: "dmprog", title: "Programs", formName: "dmprog.scx", description: "Revenue programs" },
      { id: "progunits", title: "Program Units", formName: "progunits.scx", description: "Program unit setup" },
      { id: "dmrevcat", title: "Revenue Categories", formName: "dmrevcat.scx", description: "Revenue category setup" },
      { id: "dmselrev", title: "Revenue Entry", formName: "dmselrev.scx", description: "Enter revenue data" },
      { id: "dmmeterall", title: "Meter Allocation", formName: "dmmeterall.scx", description: "Allocate meter readings" },
      { id: "dmcloserev", title: "Close Revenue Period", formName: "dmcloserev.scx", description: "Close revenue period" },
      { id: "dmstmtnotes", title: "Statement Notes", formName: "dmstmtnotes.scx", description: "Revenue statement notes" },
      { id: "dmsettlenotes", title: "Settlement Notes", formName: "dmsettlenotes.scx", description: "Settlement notes" }
    ]
  },
  {
    category: "Oil & Gas - Expenses",
    forms: [
      { id: "dmexpcat", title: "Expense Categories", formName: "dmexpcat.scx", description: "Expense category setup" },
      { id: "dmfixedexp", title: "Fixed Expenses", formName: "dmfixedexp.scx", description: "Fixed expense setup" },
      { id: "dmfixedpct", title: "Fixed Percentages", formName: "dmfixedpct.scx", description: "Fixed percentage setup" },
      { id: "dmexpall", title: "Expense Allocation", formName: "dmexpall.scx", description: "Allocate expenses" },
      { id: "dmselexp", title: "Expense Entry", formName: "dmselexp.scx", description: "Enter expense data" },
      { id: "dminvtrans", title: "Invoice Transfer", formName: "dminvtrans.scx", description: "Transfer invoices" },
      { id: "dmfixrel", title: "Fixed Release", formName: "dmfixrel.scx", description: "Release fixed expenses" },
      { id: "dmfixrelpct", title: "Fixed Release Percent", formName: "dmfixrelpct.scx", description: "Release fixed percentages" }
    ]
  }
]

// Export just the key categories for the quick access grid
export const quickAccessForms = [
  { id: "dmwellinfo", title: "Well Information", formName: "dmwellinfo.scx", description: "Well master data", category: "Wells" },
  { id: "arcust", title: "AR Customers", formName: "arcust.scx", description: "Customer maintenance", category: "AR" },
  { id: "apvendor", title: "AP Vendors", formName: "apvendor.scx", description: "Vendor maintenance", category: "AP" },
  { id: "glcoa", title: "Chart of Accounts", formName: "glcoa.scx", description: "Manage GL chart of accounts", category: "GL" },
  { id: "arinv", title: "AR Invoice Entry", formName: "arinv.scx", description: "Create and edit invoices", category: "AR" },
  { id: "apbill", title: "Bill Entry", formName: "apbill.scx", description: "Enter vendor bills", category: "AP" },
  { id: "gljourn", title: "Journal Entry", formName: "gljourn.scx", description: "General journal entries", category: "GL" },
  { id: "csrecond", title: "Bank Reconciliation", formName: "csrecond.scx", description: "Reconcile bank statements", category: "Cash" },
  { id: "dmselrev", title: "Revenue Entry", formName: "dmselrev.scx", description: "Enter revenue data", category: "Revenue" },
  { id: "dmselexp", title: "Expense Entry", formName: "dmselexp.scx", description: "Enter expense data", category: "Expense" },
  { id: "cschecks", title: "Check Maintenance", formName: "cschecks.scx", description: "Manage checks", category: "Cash" },
  { id: "dmrvstmt", title: "Revenue Statements", formName: "dmrvstmt.scx", description: "Revenue statement generation", category: "Reports" }
]