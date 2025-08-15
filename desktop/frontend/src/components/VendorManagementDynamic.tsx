import React from 'react'
import DynamicDBFTable from './DynamicDBFTable'

interface Props {
  companyName: string
  currentUser?: any
}

export default function VendorManagementDynamic({ companyName, currentUser }: Props) {
  // For now, allow editing for testing (until auth is implemented)
  const canEdit = true
  
  console.log('%c🎯 VendorManagementDynamic loaded', 'background: blue; color: white; font-weight: bold', {
    companyName,
    currentUser,
    canEdit
  })
  
  // Define which fields should be shown in the table by default
  const primaryFields = [
    'CVENDORID',
    'CVENDNAME', 
    'CCONTACT',
    'CPHONE',
    'CEMAIL',
    'LINACTIVE'
  ]
  
  return (
    <DynamicDBFTable
      tableName="VENDOR.dbf"
      companyName={companyName}
      title="Vendor Management"
      description="Manage vendor information and records"
      canEdit={canEdit}
      primaryFields={primaryFields}
      maxTableColumns={6}
    />
  )
}