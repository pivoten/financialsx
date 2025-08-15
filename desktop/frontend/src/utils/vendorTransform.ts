/**
 * Vendor Data Transformation Utilities
 * Handles conversion between DBF format and application types
 */

import { Vendor, VENDOR_DBF_MAPPING, VENDOR_FIELD_TO_DBF } from '../types/vendor'
import { decryptTaxId, isEncryptedTaxId } from './sherwareEncryption'

/**
 * Transform raw DBF record to Vendor type
 */
export function dbfToVendor(dbfRecord: any, rowIndex: number): Vendor {
  const vendor: Vendor = {
    // Primary identification
    vendorId: dbfRecord.CVENDORID || '',
    vendorName: dbfRecord.CVENDNAME || '',
    
    // Contact information
    contactName: dbfRecord.CCONTACT || '',
    phone: dbfRecord.CPHONE || '',
    phoneExt: dbfRecord.CPHEXT || '',
    fax: dbfRecord.CFAXPHONE || '',
    email: dbfRecord.CEMAIL || '',
    
    // Address information
    address1: dbfRecord.CADDRESS1 || '',
    address2: dbfRecord.CADDRESS2 || '',
    city: dbfRecord.CCITY || '',
    state: dbfRecord.CSTATE || '',
    zipCode: dbfRecord.CZIP || '',
    country: dbfRecord.CCOUNTRY || '',
    
    // Billing address
    billAddress1: dbfRecord.CBADDR1 || '',
    billAddress2: dbfRecord.CBADDR2 || '',
    billCity: dbfRecord.BCITY || '',
    billState: dbfRecord.BSTATE || '',
    billZip: dbfRecord.BZIP || '',
    billCountry: dbfRecord.BCOUNTRY || '',
    
    // Financial information
    taxId: dbfRecord.CTAXID || '',  // Keep encrypted as-is
    accountNumber: dbfRecord.CACCTNO || '',
    terms: dbfRecord.CTERMS || '',
    discountPercent: parseFloat(dbfRecord.NDISC) || 0,
    discountDays: parseInt(dbfRecord.NDISCDAYS) || 0,
    netDays: parseInt(dbfRecord.NNETDAYS) || 0,
    creditLimit: parseFloat(dbfRecord.NCREDITLIM) || 0,
    
    // Status flags - LINACTIVE is inverted (false = active)
    isActive: dbfRecord.LINACTIVE === false,
    is1099: dbfRecord.LSEND1099 === true,
    isIntegrated: dbfRecord.LINTEGGL === true,
    
    // Internal tracking
    addedDate: parseDbfDate(dbfRecord.DADDED),
    addedBy: dbfRecord.CADDEDBY || '',
    changedDate: parseDbfDate(dbfRecord.DCHANGED),
    changedBy: dbfRecord.CCHANGEDBY || '',
    
    // Additional fields
    notes: dbfRecord.CNOTES || '',
    vendorType: dbfRecord.CVENDTYPE || '',
    website: dbfRecord.CWEBSITE || '',
    
    // DBF record tracking
    _rowIndex: rowIndex
  }
  
  // Trim all string fields
  Object.keys(vendor).forEach(key => {
    const value = (vendor as any)[key]
    if (typeof value === 'string') {
      (vendor as any)[key] = value.trim()
    }
  })
  
  return vendor
}

/**
 * Transform array of DBF records to Vendor array
 */
export function dbfArrayToVendors(dbfData: any): Vendor[] {
  if (!dbfData || !dbfData.rows || !Array.isArray(dbfData.rows)) {
    return []
  }
  
  const { rows, columns } = dbfData
  
  // Convert array rows to objects if needed
  const records = rows.map((row: any, index: number) => {
    if (Array.isArray(row)) {
      const record: any = {}
      columns.forEach((col: string, colIndex: number) => {
        record[col] = row[colIndex]
      })
      return dbfToVendor(record, index)
    }
    return dbfToVendor(row, index)
  })
  
  return records
}

/**
 * Transform Vendor to DBF format for saving
 */
export function vendorToDbf(vendor: Partial<Vendor>): Record<string, any> {
  const dbfRecord: Record<string, any> = {}
  
  Object.entries(vendor).forEach(([key, value]) => {
    const dbfField = VENDOR_FIELD_TO_DBF[key]
    if (dbfField) {
      // Special handling for certain fields
      if (key === 'isActive') {
        // Invert the logic for LINACTIVE
        dbfRecord[dbfField] = !value
      } else if (value instanceof Date) {
        // Format dates for DBF
        dbfRecord[dbfField] = formatDateForDbf(value)
      } else {
        dbfRecord[dbfField] = value
      }
    }
  })
  
  return dbfRecord
}

/**
 * Parse DBF date field
 */
function parseDbfDate(value: any): Date | null {
  if (!value) return null
  
  // DBF dates might be strings in YYYYMMDD format or Date objects
  if (typeof value === 'string') {
    if (value.length === 8 && /^\d{8}$/.test(value)) {
      // YYYYMMDD format
      const year = parseInt(value.substring(0, 4))
      const month = parseInt(value.substring(4, 6)) - 1
      const day = parseInt(value.substring(6, 8))
      return new Date(year, month, day)
    }
  }
  
  try {
    const date = new Date(value)
    return isNaN(date.getTime()) ? null : date
  } catch {
    return null
  }
}

/**
 * Format date for DBF storage
 */
function formatDateForDbf(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}${month}${day}`
}

/**
 * Get display value for tax ID (handles encryption)
 */
export function getDisplayTaxId(taxId: string, showDecrypted: boolean): string {
  const trimmed = taxId?.trim() || ''
  
  if (!trimmed) {
    return ''
  }
  
  if (isEncryptedTaxId(trimmed)) {
    if (showDecrypted) {
      return decryptTaxId(trimmed)
    }
    return 'Encrypted'
  }
  
  // Plain text tax ID
  if (showDecrypted) {
    return trimmed
  }
  return 'Hidden'
}

/**
 * Format phone number for display
 */
export function formatPhone(phone: string): string {
  const cleaned = phone?.replace(/\D/g, '') || ''
  
  if (cleaned.length === 10) {
    return `${cleaned.slice(0, 3)}-${cleaned.slice(3, 6)}-${cleaned.slice(6)}`
  }
  
  return phone || ''
}

/**
 * Validate vendor data before saving
 */
export function validateVendor(vendor: Partial<Vendor>): string[] {
  const errors: string[] = []
  
  if (!vendor.vendorId?.trim()) {
    errors.push('Vendor ID is required')
  }
  
  if (!vendor.vendorName?.trim()) {
    errors.push('Vendor name is required')
  }
  
  if (vendor.email && !isValidEmail(vendor.email)) {
    errors.push('Invalid email address')
  }
  
  if (vendor.state && vendor.state.length !== 2) {
    errors.push('State must be 2 characters')
  }
  
  if (vendor.discountPercent && (vendor.discountPercent < 0 || vendor.discountPercent > 100)) {
    errors.push('Discount percent must be between 0 and 100')
  }
  
  return errors
}

/**
 * Check if email is valid
 */
function isValidEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}