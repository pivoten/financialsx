
import { useState, useEffect } from 'react'
import { GetCompanyInfo, UpdateCompanyInfo, TestOLEConnection } from '../../wailsjs/go/main/App'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './ui/card'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Building2, FileText, Settings, Save, X, Edit, Wifi } from 'lucide-react'
import { User } from '../types'
import logger from '../services/logger'

const toast = ({ title, description, variant }: { title: string; description: string; variant?: string }): void => {
  if (variant === 'destructive') logger.error(title, { description })
  else logger.debug(title, { description })
}

export default function CompanyInformation({ currentUser }: { currentUser: User | null }) {
  const [companyData, setCompanyData] = useState<any>(null)
  const [loading, setLoading] = useState(false)
  const [editing, setEditing] = useState(false)
  const [editedData, setEditedData] = useState<any>(null)
  const [testingConnection, setTestingConnection] = useState(false)
  const canEdit = currentUser && (currentUser.is_root || currentUser.role_name === 'Admin')

  useEffect(() => {
    if (currentUser?.company_name) {
      const timer = setTimeout(() => { loadCompanyData() }, 100)
      return () => clearTimeout(timer)
    }
    return undefined
  }, [currentUser])

  const loadCompanyData = async () => {
    setLoading(true)
    try {
      if (!window.go) throw new Error('Wails runtime not initialized')
      if (typeof GetCompanyInfo !== 'function') throw new Error('GetCompanyInfo function not available')
      const companyPath = localStorage.getItem('company_path')
      const companyName = localStorage.getItem('company_name')
      const companyIdentifier = companyPath || currentUser.company_name || companyName
      const result = await GetCompanyInfo(companyIdentifier)
      if (result.success) {
        setCompanyData(result.data)
        setEditedData(result.data)
        if (result.mock) {
          toast({ title: 'Using Mock Data', description: result.error || 'Could not connect to FoxPro OLE server', variant: 'warning' })
        }
      }
    } catch (error) {
      toast({ title: 'Error', description: error.message || 'Failed to load company information', variant: 'destructive' })
    } finally { setLoading(false) }
  }

  const handleEdit = () => { setEditing(true); setEditedData({ ...companyData }) }
  const handleCancel = () => { setEditing(false); setEditedData(companyData) }

  const handleSave = async () => {
    setLoading(true)
    try {
      const result = await UpdateCompanyInfo(JSON.stringify(editedData))
      if (result.success) { setCompanyData(editedData); setEditing(false); toast({ title: 'Success', description: 'Company information updated successfully' }) }
    } catch (error) {
      toast({ title: 'Error', description: error.message || 'Failed to save company information', variant: 'destructive' })
    } finally { setLoading(false) }
  }

  const handleInputChange = (field: string, value: string) => { setEditedData((prev: any) => ({ ...prev, [field]: value })) }

  if (loading && !companyData) return (<div className="flex items-center justify-center p-8"><p className="text-muted-foreground">Loading company information...</p></div>)
  if (!companyData) return (<div className="flex items-center justify-center p-8"><p className="text-muted-foreground">No company data available</p></div>)

  return (
    <>
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 mb-8">
        <Card className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]">
          <CardContent className="p-8">
            <div className="flex items-start justify-between mb-4">
              <div className="space-y-1">
                <p className="text-sm font-medium text-muted-foreground">Organization</p>
                <h3 className="text-2xl font-bold">Company Information</h3>
                <p className="text-sm text-muted-foreground mt-2">View company details</p>
              </div>
              <div className="p-3 bg-primary/10 rounded-lg">
                <Building2 className="w-5 h-5 text-primary" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Company Information</CardTitle>
              <CardDescription>View and manage company details and settings</CardDescription>
            </div>
            {canEdit && !editing && (<Button onClick={handleEdit} variant="outline"><Edit className="w-4 h-4 mr-2" />Edit</Button>)}
            {editing && (
              <div className="flex gap-2">
                <Button onClick={handleCancel} variant="outline"><X className="w-4 h-4 mr-2" />Cancel</Button>
                <Button onClick={handleSave} disabled={loading}><Save className="w-4 h-4 mr-2" />Save</Button>
              </div>
            )}
          </div>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="grid gap-6 md:grid-cols-2">
            <div className="space-y-4">
              <div>
                <Label className="text-sm font-medium text-muted-foreground">Company Name</Label>
                {editing ? (<Input value={editedData.company_name || ''} onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('company_name', e.target.value)} className="mt-1" />) : (<p className="text-lg font-semibold">{companyData.company_name || 'Not Available'}</p>)}
              </div>
              <div>
                <Label className="text-sm font-medium text-muted-foreground">Tax ID / EIN</Label>
                {editing ? (<Input value={editedData.tax_id || ''} onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('tax_id', e.target.value)} className="mt-1" />) : (<p className="text-lg">{companyData.tax_id || 'Not Available'}</p>)}
              </div>
              <div>
                <Label className="text-sm font-medium text-muted-foreground">Contact Person</Label>
                {editing ? (<Input value={editedData.contact || ''} onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('contact', e.target.value)} className="mt-1" />) : (<p className="text-lg">{companyData.contact || 'Not Available'}</p>)}
              </div>
              <div>
                <Label className="text-sm font-medium text-muted-foreground">Fiscal Year End</Label>
                {editing ? (<Input value={editedData.fiscal_year_end || ''} onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('fiscal_year_end', e.target.value)} className="mt-1" />) : (<p className="text-lg">{companyData.fiscal_year_end || 'December 31'}</p>)}
              </div>
            </div>
            <div className="space-y-4">
              <div>
                <Label className="text-sm font-medium text-muted-foreground">Primary Address</Label>
                {editing ? (
                  <>
                    <Input value={editedData.address1 || ''} onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('address1', e.target.value)} className="mt-1 mb-2" placeholder="Address Line 1" />
                    <Input value={editedData.address2 || ''} onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('address2', e.target.value)} className="mb-2" placeholder="Address Line 2" />
                    <div className="grid grid-cols-2 gap-2">
                      <Input value={editedData.city || ''} onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('city', e.target.value)} placeholder="City" />
                      <div className="grid grid-cols-2 gap-2">
                        <Input value={editedData.state || ''} onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('state', e.target.value)} placeholder="State" maxLength={2} />
                        <Input value={editedData.zip_code || ''} onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('zip_code', e.target.value)} placeholder="ZIP" />
                      </div>
                    </div>
                  </>
                ) : (
                  <>
                    <p className="text-lg">{companyData.address1 || ''}</p>
                    {companyData.address2 && <p className="text-lg">{companyData.address2}</p>}
                    <p className="text-lg">{companyData.city && `${companyData.city}, `}{companyData.state} {companyData.zip_code}</p>
                  </>
                )}
              </div>
              <div>
                <Label className="text-sm font-medium text-muted-foreground">Phone</Label>
                {editing ? (<Input value={editedData.phone || ''} onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('phone', e.target.value)} className="mt-1" />) : (<p className="text-lg">{companyData.phone || 'Not Available'}</p>)}
              </div>
              <div>
                <Label className="text-sm font-medium text-muted-foreground">Fax</Label>
                {editing ? (<Input value={editedData.fax || ''} onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('fax', e.target.value)} className="mt-1" />) : (<p className="text-lg">{companyData.fax || 'Not Available'}</p>)}
              </div>
              <div>
                <Label className="text-sm font-medium text-muted-foreground">Email</Label>
                {editing ? (<Input type="email" value={editedData.email || ''} onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('email', e.target.value)} className="mt-1" />) : (<p className="text-lg">{companyData.email || 'Not Available'}</p>)}
              </div>
            </div>
          </div>

          <div className="border-t my-6"></div>

          <div className="space-y-4">
            <h3 className="text-lg font-semibold">Additional Information</h3>
            <div className="grid gap-4 md:grid-cols-2">
              <div>
                <Label className="text-sm font-medium text-muted-foreground">Processor</Label>
                {editing ? (<Input value={editedData.processor || ''} onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('processor', e.target.value)} className="mt-1" />) : (<p className="text-lg">{companyData.processor || 'Not Available'}</p>)}
              </div>
              <div>
                <Label className="text-sm font-medium text-muted-foreground">Domain</Label>
                {editing ? (<Input value={editedData.domain || ''} onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('domain', e.target.value)} className="mt-1" />) : (<p className="text-lg">{companyData.domain || 'Not Available'}</p>)}
              </div>
            </div>
          </div>

          <div className="border-t my-6"></div>

          <div className="space-y-4">
            <h3 className="text-lg font-semibold">System Settings</h3>
            <div className="grid gap-4 md:grid-cols-2">
              <div>
                <Label className="text-sm font-medium text-muted-foreground">Data Directory</Label>
                <p className="text-sm font-mono bg-muted p-2 rounded mt-1">{companyData.data_path || `../datafiles/${currentUser?.company_name || 'company'}/`}</p>
              </div>
              <div>
                <Label className="text-sm font-medium text-muted-foreground">System Version</Label>
                <p className="text-lg">{companyData.version || 'Not Available'}</p>
              </div>
              <div>
                <Label className="text-sm font-medium text-muted-foreground">Company ID</Label>
                <p className="text-lg font-mono">{companyData.company_id || 'Not Available'}</p>
              </div>
              <div>
                <Label className="text-sm font-medium text-muted-foreground">Alias</Label>
                <p className="text-lg">{companyData.alias || 'Not Available'}</p>
              </div>
            </div>
          </div>

          {!editing && (
            <div className="flex gap-3 pt-4">
              <Button variant="outline"><FileText className="w-4 h-4 mr-2" />Export Company Info</Button>
              {canEdit && (<Button variant="outline" onClick={handleEdit}><Settings className="w-4 h-4 mr-2" />Edit Settings</Button>)}
              <Button variant="outline" onClick={async () => {
                setTestingConnection(true)
                try { const result = await TestOLEConnection(); toast({ title: result.success ? 'Connection Successful' : 'Connection Failed', description: result.message + (result.logPath ? ` Log: ${result.logPath}` : ''), variant: result.success ? 'default' : 'destructive' }) }
                catch (error) { toast({ title: 'Test Failed', description: error.message || 'Unknown error occurred', variant: 'destructive' }) }
                finally { setTestingConnection(false) }
              }} disabled={testingConnection}>
                <Wifi className="w-4 h-4 mr-2" />
                {testingConnection ? 'Testing...' : 'Test OLE Connection'}
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </>
  )
}
