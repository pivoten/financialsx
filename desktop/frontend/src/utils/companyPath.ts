// Utility function to get the correct company identifier for DBF operations
// Returns the company path if available, otherwise returns the company name (legacy)
export function getCompanyDataPath(): string | null {
  const companyPath = localStorage.getItem('company_path')
  const companyName = localStorage.getItem('company_name')
  
  // Use the actual data path if available, otherwise use company name (legacy)
  return companyPath || companyName
}

// Get just the company name for display purposes
export function getCompanyName(): string | null {
  return localStorage.getItem('company_name')
}

// Get company path for informational purposes
export function getCompanyPath(): string | null {
  return localStorage.getItem('company_path')
}