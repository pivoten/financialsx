// Utility function to get the correct company identifier for DBF operations
// Always use company name for consistency (company_path may contain Windows paths or numeric values)
export function getCompanyDataPath(): string | null {
  const companyName = localStorage.getItem('company_name')
  
  // Always use company name, ignore company_path which may have bad values
  return companyName
}

// Get just the company name for display purposes
export function getCompanyName(): string | null {
  return localStorage.getItem('company_name')
}

// Get company path for informational purposes
export function getCompanyPath(): string | null {
  return localStorage.getItem('company_path')
}