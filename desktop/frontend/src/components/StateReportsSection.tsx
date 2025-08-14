
import { useState, useEffect } from 'react'
import { Button } from './ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './ui/card'
import { Tabs, TabsList, TabsTrigger, TabsContent } from './ui/tabs'
import { Badge } from './ui/badge'
import { MapPin, FileText, Calendar, AlertCircle, Building, CheckCircle, Clock } from 'lucide-react'
import { User } from '../types'

export function StateReportsSection({ currentUser }: { currentUser: User | null }) {
  const [states, setStates] = useState([])
  const [loading, setLoading] = useState(true)
  const [selectedState, setSelectedState] = useState('')
  const [error, setError] = useState(null)

  const mockStatesData = [
    { code: 'WV', name: 'West Virginia', wellCount: 47, reports: [
      { name: 'Monthly Production Report', due: '2024-08-15', status: 'pending', type: 'monthly' },
      { name: 'Annual Tax Filing', due: '2025-03-31', status: 'upcoming', type: 'annual' },
      { name: 'Environmental Compliance', due: '2024-09-01', status: 'upcoming', type: 'quarterly' },
      { name: 'Royalty Statement', due: '2024-08-10', status: 'overdue', type: 'monthly' }
    ]},
    { code: 'PA', name: 'Pennsylvania', wellCount: 23, reports: [
      { name: 'Production Tax Report', due: '2024-08-20', status: 'pending', type: 'monthly' },
      { name: 'DEP Waste Report', due: '2024-08-25', status: 'upcoming', type: 'monthly' },
      { name: 'Annual Registration', due: '2025-01-31', status: 'upcoming', type: 'annual' }
    ]},
    { code: 'OH', name: 'Ohio', wellCount: 15, reports: [
      { name: 'ODNR Production Report', due: '2024-08-31', status: 'upcoming', type: 'monthly' },
      { name: 'Severance Tax Return', due: '2024-09-15', status: 'upcoming', type: 'monthly' }
    ]}
  ]

  useEffect(() => { loadStatesFromWells() }, [currentUser])

  const loadStatesFromWells = async () => {
    try {
      setLoading(true); setError(null)
      await new Promise(resolve => setTimeout(resolve, 1000))
      setStates(mockStatesData)
      if (mockStatesData.length > 0) setSelectedState(mockStatesData[0].code)
    } catch (err) {
      setError(err.message || 'Failed to load state reports')
    } finally { setLoading(false) }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'overdue': return 'destructive'
      case 'pending': return 'default'
      case 'upcoming': return 'secondary'
      case 'completed': return 'outline'
      default: return 'outline'
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'overdue': return <AlertCircle className="w-4 h-4" />
      case 'pending': return <Calendar className="w-4 h-4" />
      case 'upcoming': return <Clock className="w-4 h-4" />
      case 'completed': return <CheckCircle className="w-4 h-4" />
      default: return <FileText className="w-4 h-4" />
    }
  }

  if (loading) return (<Card><CardContent className="flex items-center justify-center pt-8"><div className="text-center"><div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div><p className="text-muted-foreground">Analyzing wells data to determine state reporting requirements...</p></div></CardContent></Card>)

  if (error) return (
    <Card className="border-red-200 bg-red-50">
      <CardHeader><CardTitle className="text-red-800">Error Loading Reports</CardTitle></CardHeader>
      <CardContent><p className="text-red-700">{error}</p><Button onClick={() => loadStatesFromWells()} className="mt-4" variant="outline">Try Again</Button></CardContent>
    </Card>
  )

  if (states.length === 0) return (
    <Card>
      <CardHeader><CardTitle>State Reporting</CardTitle><CardDescription>No wells data found or no states identified</CardDescription></CardHeader>
      <CardContent className="pt-6"><div className="text-center text-muted-foreground"><Building className="w-12 h-12 mx-auto mb-4 opacity-50" /><p>No wells found in WELLS.dbf file.</p><p className="text-sm mt-2">Import wells data to see state-specific reporting requirements.</p></div></CardContent>
    </Card>
  )

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2"><MapPin className="w-5 h-5" />State Reporting Dashboard</CardTitle>
          <CardDescription>Reports organized by state based on well locations in your WELLS.dbf file</CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs value={selectedState} onValueChange={setSelectedState} className="w-full">
            <TabsList className="grid w-full grid-cols-3">
              {states.map((state) => (
                <TabsTrigger key={state.code} value={state.code} className="flex items-center gap-2">
                  <MapPin className="w-4 h-4" />
                  {state.code}
                  <Badge variant="outline" className="ml-1 text-xs">{state.wellCount}</Badge>
                </TabsTrigger>
              ))}
            </TabsList>

            {states.map((state) => (
              <TabsContent key={state.code} value={state.code} className="space-y-4 mt-6">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-lg font-semibold">{state.name} Reporting</h3>
                    <p className="text-sm text-muted-foreground">{state.wellCount} wells • {state.reports.length} report types</p>
                  </div>
                  <Badge variant="outline">{state.reports.filter((r: any) => r.status === 'overdue').length} overdue</Badge>
                </div>

                <div className="grid gap-4 md:grid-cols-2">
                  {state.reports.map((report: any, index: number) => (
                    <Card key={index} className="relative">
                      <CardContent className="p-4">
                        <div className="flex items-start justify-between">
                          <div className="flex-1">
                            <div className="flex items-center gap-2 mb-2">
                              {getStatusIcon(report.status)}
                              <h4 className="font-semibold">{report.name}</h4>
                              <Badge variant={getStatusColor(report.status)} className="text-xs">{report.status}</Badge>
                            </div>
                            <p className="text-sm text-muted-foreground mb-2">Due: {new Date(report.due).toLocaleDateString()}</p>
                            <p className="text-xs text-muted-foreground">{report.type === 'monthly' ? 'Monthly filing' : report.type === 'quarterly' ? 'Quarterly filing' : 'Annual filing'}</p>
                          </div>
                          <div className="flex flex-col gap-2 ml-4">
                            <Button size="sm" variant="outline" disabled>Generate</Button>
                            <Button size="sm" variant="ghost" disabled>View</Button>
                          </div>
                        </div>
                      </CardContent>
                      {report.status === 'overdue' && (<div className="absolute top-2 right-2"><div className="w-2 h-2 bg-red-500 rounded-full animate-pulse"></div></div>)}
                    </Card>
                  ))}
                </div>

                <Card className="bg-muted/30">
                  <CardContent className="p-4">
                    <div className="flex items-center justify-between">
                      <div>
                        <h4 className="font-medium">Quick Actions for {state.name}</h4>
                        <p className="text-sm text-muted-foreground">Common reporting tasks</p>
                      </div>
                      <div className="flex gap-2">
                        <Button variant="outline" size="sm" disabled><FileText className="w-4 h-4 mr-2" />Export All Reports</Button>
                        <Button variant="outline" size="sm" disabled><Calendar className="w-4 h-4 mr-2" />Set Reminders</Button>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </TabsContent>
            ))}
          </Tabs>
        </CardContent>
      </Card>

      <Card className="border-blue-200 bg-blue-50">
        <CardContent className="pt-6">
          <div className="flex items-start gap-2 text-blue-800">
            <AlertCircle className="w-4 h-4 mt-0.5" />
            <div className="text-sm">
              <p className="font-medium">Development Note</p>
              <p className="mt-1">This interface shows a preview of state-based reporting. The system will automatically:</p>
              <ul className="list-disc list-inside mt-2 space-y-1 text-xs">
                <li>Scan WELLS.dbf to identify unique states from well locations</li>
                <li>Generate state-specific reporting requirements and due dates</li>
                <li>Provide report generation and submission workflows</li>
                <li>Track compliance status and send automated reminders</li>
              </ul>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
