/**
 * Vendor Management Types
 * Explicit type definitions for vendor data model
 */

export interface Vendor {
  // Primary identification
  vendorId: string          // CVENDORID - Unique vendor identifier
  vendorName: string        // CVENDNAME - Company/vendor name
  
  // Contact information
  contactName: string       // CCONTACT - Primary contact person
  phone: string            // CPHONE - Primary phone number
  phoneExt: string         // CPHEXT - Phone extension
  fax: string              // CFAXPHONE - Fax number
  email: string            // CEMAIL - Email address
  
  // Address information
  address1: string         // CADDRESS1 - Street address line 1
  address2: string         // CADDRESS2 - Street address line 2
  city: string             // CCITY
  state: string            // CSTATE - 2-letter state code
  zipCode: string          // CZIP
  country: string          // CCOUNTRY
  
  // Billing address (if different from main)
  billAddress1: string     // CBADDR1
  billAddress2: string     // CBADDR2
  billCity: string         // BCITY
  billState: string        // BSTATE
  billZip: string          // BZIP
  billCountry: string      // BCOUNTRY
  
  // Financial information
  taxId: string            // CTAXID - Encrypted SSN/EIN
  accountNumber: string    // CACCTNO - GL account number
  terms: string            // CTERMS - Payment terms
  discountPercent: number  // NDISC - Discount percentage
  discountDays: number     // NDISCDAYS - Days to take discount
  netDays: number          // NNETDAYS - Net payment days
  creditLimit: number      // NCREDITLIM - Credit limit amount
  
  // Status flags
  isActive: boolean        // LINACTIVE - Active vendor flag (inverted logic)
  is1099: boolean          // LSEND1099 - Send 1099 flag
  isIntegrated: boolean    // LINTEGGL - GL integration flag
  
  // Internal tracking
  addedDate: Date | null   // DADDED - Date record added
  addedBy: string          // CADDEDBY - User who added record
  changedDate: Date | null // DCHANGED - Date last changed
  changedBy: string        // CCHANGEDBY - User who last changed
  
  // Additional fields
  notes: string            // CNOTES - General notes
  vendorType: string       // CVENDTYPE - Vendor category/type
  website: string          // CWEBSITE - Company website
  
  // DBF record tracking
  _rowIndex?: number       // Internal row index for updates
}

export interface VendorFormData {
  // Subset of fields that can be edited
  vendorId: string
  vendorName: string
  contactName: string
  phone: string
  phoneExt: string
  fax: string
  email: string
  address1: string
  address2: string
  city: string
  state: string
  zipCode: string
  country: string
  taxId?: string           // Optional - may be encrypted
  accountNumber: string
  terms: string
  discountPercent: number
  discountDays: number
  netDays: number
  creditLimit: number
  isActive: boolean
  is1099: boolean
  notes: string
  vendorType: string
  website: string
}

export interface VendorTableColumn {
  key: keyof Vendor
  label: string
  sortable?: boolean
  width?: string
  align?: 'left' | 'center' | 'right'
  format?: (value: any, row: Vendor) => string | JSX.Element
  className?: string
}

export interface VendorFilters {
  search: string
  showInactive: boolean
  vendorType?: string
  has1099?: boolean
}

export interface VendorSortConfig {
  key: keyof Vendor | null
  direction: 'asc' | 'desc'
}

// DBF field mapping
export const VENDOR_DBF_MAPPING = {
  CVENDORID: 'vendorId',
  CVENDNAME: 'vendorName',
  CCONTACT: 'contactName',
  CPHONE: 'phone',
  CPHEXT: 'phoneExt',
  CFAXPHONE: 'fax',
  CEMAIL: 'email',
  CADDRESS1: 'address1',
  CADDRESS2: 'address2',
  CCITY: 'city',
  CSTATE: 'state',
  CZIP: 'zipCode',
  CCOUNTRY: 'country',
  CBADDR1: 'billAddress1',
  CBADDR2: 'billAddress2',
  BCITY: 'billCity',
  BSTATE: 'billState',
  BZIP: 'billZip',
  BCOUNTRY: 'billCountry',
  CTAXID: 'taxId',
  CACCTNO: 'accountNumber',
  CTERMS: 'terms',
  NDISC: 'discountPercent',
  NDISCDAYS: 'discountDays',
  NNETDAYS: 'netDays',
  NCREDITLIM: 'creditLimit',
  LINACTIVE: 'isActive',
  LSEND1099: 'is1099',
  LINTEGGL: 'isIntegrated',
  DADDED: 'addedDate',
  CADDEDBY: 'addedBy',
  DCHANGED: 'changedDate',
  CCHANGEDBY: 'changedBy',
  CNOTES: 'notes',
  CVENDTYPE: 'vendorType',
  CWEBSITE: 'website'
} as const

// Reverse mapping for saving back to DBF
export const VENDOR_FIELD_TO_DBF = Object.entries(VENDOR_DBF_MAPPING).reduce(
  (acc, [dbf, field]) => ({ ...acc, [field]: dbf }),
  {} as Record<string, string>
)