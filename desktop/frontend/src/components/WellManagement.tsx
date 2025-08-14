import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from './ui/card';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from './ui/tabs';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui/table';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { Badge } from './ui/badge';
import { Checkbox } from './ui/checkbox';
import { 
  Search, 
  Plus, 
  Edit, 
  Trash2, 
  Save, 
  X, 
  MapPin, 
  Droplet,
  Activity,
  FileText,
  Users,
  DollarSign,
  Calendar,
  AlertCircle,
  CheckCircle,
  BarChart3,
  TrendingUp,
  Settings,
  Download,
  Upload,
  ChevronLeft
} from 'lucide-react';
// Check if Wails API is available
const isWailsAvailable = typeof window !== 'undefined' && (window as any).go?.main?.App;

// Import Wails functions or use mock fallbacks
import * as WailsApp from '../../wailsjs/go/main/App';

const GetDBFTableData = isWailsAvailable 
  ? WailsApp.GetDBFTableData 
  : async () => { 
      console.warn('GetDBFTableData not available - using mock data'); 
      return null; 
    };

const SearchDBFTable = isWailsAvailable
  ? WailsApp.SearchDBFTable
  : async () => { 
      console.warn('SearchDBFTable not available - using mock data'); 
      return null; 
    };

interface Well {
  id: string;
  wellId: string;
  wellName: string;
  apiNumber: string;
  leaseId: string;
  fieldName: string;
  county: string;
  state: string;
  status: 'Active' | 'Inactive' | 'Suspended' | 'Abandoned' | 'Drilling';
  type: 'Oil' | 'Gas' | 'Oil & Gas' | 'Injection' | 'SWD';
  operator: string;
  workingInterest: number;
  netRevenueInterest: number;
  spudDate?: Date;
  completionDate?: Date;
  latitude?: number;
  longitude?: number;
  depth?: number;
  formation?: string;
  currentProduction?: {
    oil: number;
    gas: number;
    water: number;
    date: Date;
  };
}

interface WellManagementProps {
  companyName?: string;
  currentUser?: any;
}

export default function WellManagement({ companyName, currentUser }: WellManagementProps) {
  const [wells, setWells] = useState<Well[]>([]);
  const [selectedWell, setSelectedWell] = useState<Well | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [isAddingNew, setIsAddingNew] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterStatus, setFilterStatus] = useState<string>('all');
  const [filterType, setFilterType] = useState<string>('all');
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('general');
  const [showDetails, setShowDetails] = useState(false);

  // Form state for editing/adding
  const [formData, setFormData] = useState<Partial<Well>>({
    status: 'Active',
    type: 'Oil & Gas'
  });

  // Load wells on component mount or company change
  useEffect(() => {
    if (companyName || !isWailsAvailable) {
      loadWells();
    }
  }, [companyName]);

  const loadWells = async () => {
    // Check if Wails is available (running through Wails dev, not browser)
    if (!isWailsAvailable) {
      console.log('Wails not available (browser mode), loading mock wells');
      setWells(getMockWells());
      setLoading(false);
      return;
    }

    if (!companyName) {
      console.log('No company name, loading mock wells');
      setWells(getMockWells());
      return;
    }
    
    setLoading(true);
    try {
      console.log('Attempting to load wells from DBF for company:', companyName);
      const result = await GetDBFTableData(companyName, 'WELLS.dbf');
      if (result?.rows) {
        const wellData = result.rows.map((row: any, index: number) => ({
          id: row[0] || `well-${index}`,
          wellId: row[0] || '',
          wellName: row[1] || '',
          apiNumber: row[2] || '',
          leaseId: row[3] || '',
          fieldName: row[4] || '',
          county: row[5] || '',
          state: row[6] || '',
          status: row[7] || 'Active',
          type: row[8] || 'Oil & Gas',
          operator: row[9] || '',
          workingInterest: parseFloat(row[10]) || 0,
          netRevenueInterest: parseFloat(row[11]) || 0,
          spudDate: row[12] ? new Date(row[12]) : undefined,
          completionDate: row[13] ? new Date(row[13]) : undefined,
          latitude: parseFloat(row[14]) || undefined,
          longitude: parseFloat(row[15]) || undefined,
          depth: parseFloat(row[16]) || undefined,
          formation: row[17] || ''
        }));
        console.log('Loaded', wellData.length, 'wells from DBF');
        setWells(wellData);
      } else {
        console.log('No wells found in DBF, loading mock data');
        setWells(getMockWells());
      }
    } catch (error) {
      console.error('Error loading wells from DBF:', error);
      console.log('Loading mock wells due to error');
      // Use mock data for development
      setWells(getMockWells());
    } finally {
      setLoading(false);
    }
  };

  const getMockWells = (): Well[] => [
    {
      id: '1',
      wellId: 'LIME-001',
      wellName: 'Limestone Creek #1',
      apiNumber: '42-123-45678',
      leaseId: 'LC-2024-001',
      fieldName: 'Limestone Creek Field',
      county: 'Midland',
      state: 'TX',
      status: 'Active',
      type: 'Oil & Gas',
      operator: 'LimeCreek Energy LLC',
      workingInterest: 87.5,
      netRevenueInterest: 75.0,
      spudDate: new Date('2024-01-15'),
      completionDate: new Date('2024-03-20'),
      latitude: 31.9973,
      longitude: -102.0779,
      depth: 8500,
      formation: 'Wolfcamp',
      currentProduction: {
        oil: 450,
        gas: 1200,
        water: 150,
        date: new Date()
      }
    },
    {
      id: '2',
      wellId: 'LIME-002',
      wellName: 'Limestone Creek #2',
      apiNumber: '42-123-45679',
      leaseId: 'LC-2024-002',
      fieldName: 'Limestone Creek Field',
      county: 'Midland',
      state: 'TX',
      status: 'Drilling',
      type: 'Oil & Gas',
      operator: 'LimeCreek Energy LLC',
      workingInterest: 87.5,
      netRevenueInterest: 75.0,
      spudDate: new Date('2024-06-01'),
      depth: 6200,
      formation: 'Wolfcamp'
    },
    {
      id: '3',
      wellId: 'SALT-001',
      wellName: 'Salt Water Disposal #1',
      apiNumber: '42-123-45680',
      leaseId: 'SWD-2023-001',
      fieldName: 'Limestone Creek Field',
      county: 'Midland',
      state: 'TX',
      status: 'Active',
      type: 'SWD',
      operator: 'LimeCreek Energy LLC',
      workingInterest: 100,
      netRevenueInterest: 100,
      completionDate: new Date('2023-09-15'),
      depth: 5000,
      formation: 'San Andres'
    }
  ];

  const handleSaveWell = async () => {
    // TODO: Implement save to DBF
    console.log('Saving well:', formData);
    console.log('isAddingNew:', isAddingNew, 'selectedWell:', selectedWell);
    
    if (isAddingNew) {
      const newWell: Well = {
        ...formData as Well,
        id: Date.now().toString()
      };
      setWells([...wells, newWell]);
    } else if (selectedWell) {
      setWells(wells.map(w => 
        w.id === selectedWell.id ? {...w, ...formData} : w
      ));
    }
    
    setIsEditing(false);
    setIsAddingNew(false);
    setShowDetails(false);
  };

  const handleCancel = () => {
    setIsEditing(false);
    setIsAddingNew(false);
    if (selectedWell) {
      setFormData(selectedWell);
    } else {
      setShowDetails(false);
    }
  };

  const handleDeleteWell = async (wellId: string) => {
    if (confirm('Are you sure you want to delete this well?')) {
      setWells(wells.filter(w => w.id !== wellId));
    }
  };

  const handleAddNew = () => {
    console.log('New Well button clicked');
    setIsAddingNew(true);
    setIsEditing(true);
    setSelectedWell(null);
    setShowDetails(true);
    setFormData({ 
      status: 'Active', 
      type: 'Oil & Gas',
      wellId: '',
      wellName: '',
      apiNumber: '',
      workingInterest: 100,
      netRevenueInterest: 87.5
    });
    setActiveTab('general');
    console.log('showDetails set to:', true);
  };

  // Filter wells based on search and filters
  const filteredWells = wells.filter(well => {
    const matchesSearch = searchTerm === '' || 
      well.wellId.toLowerCase().includes(searchTerm.toLowerCase()) ||
      well.wellName.toLowerCase().includes(searchTerm.toLowerCase()) ||
      well.apiNumber.toLowerCase().includes(searchTerm.toLowerCase());
    
    const matchesStatus = filterStatus === 'all' || well.status === filterStatus;
    const matchesType = filterType === 'all' || well.type === filterType;
    
    return matchesSearch && matchesStatus && matchesType;
  });

  const getStatusBadge = (status: Well['status']) => {
    const colors = {
      'Active': 'bg-green-100 text-green-800',
      'Inactive': 'bg-gray-100 text-gray-800',
      'Suspended': 'bg-yellow-100 text-yellow-800',
      'Abandoned': 'bg-red-100 text-red-800',
      'Drilling': 'bg-blue-100 text-blue-800'
    };
    
    return (
      <Badge className={colors[status]}>
        {status}
      </Badge>
    );
  };

  const getTypeBadge = (type: Well['type']) => {
    const colors = {
      'Oil': 'bg-amber-100 text-amber-800',
      'Gas': 'bg-blue-100 text-blue-800',
      'Oil & Gas': 'bg-purple-100 text-purple-800',
      'Injection': 'bg-cyan-100 text-cyan-800',
      'SWD': 'bg-gray-100 text-gray-800'
    };
    
    return (
      <Badge className={colors[type]}>
        {type}
      </Badge>
    );
  };

  // Well Details View - Separate screen
  if (showDetails) {
    return (
      <div className="space-y-6">
        {/* Header with Back Button */}
        <div className="bg-white rounded-lg shadow-sm p-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  setShowDetails(false);
                  setSelectedWell(null);
                  setIsEditing(false);
                  setIsAddingNew(false);
                }}
                className="mr-4"
              >
                <ChevronLeft className="h-4 w-4 mr-1" />
                Back to Wells
              </Button>
              <Droplet className="h-8 w-8 text-blue-600" />
              <div>
                <h1 className="text-2xl font-bold text-gray-900">
                  {isAddingNew ? 'Add New Well' : `Well Details: ${selectedWell?.wellId}`}
                </h1>
                <p className="text-sm text-gray-500">
                  {isAddingNew ? 'Enter information for the new well' : 'View and edit well information'}
                </p>
              </div>
            </div>
            <div className="flex items-center space-x-2">
              {isEditing ? (
                <>
                  <Button onClick={handleSaveWell} className="bg-green-600 hover:bg-green-700">
                    <Save className="h-4 w-4 mr-2" />
                    Save
                  </Button>
                  <Button 
                    variant="outline" 
                    onClick={handleCancel}
                  >
                    <X className="h-4 w-4 mr-2" />
                    Cancel
                  </Button>
                </>
              ) : (
                <>
                  <Button onClick={() => setIsEditing(true)}>
                    <Edit className="h-4 w-4 mr-2" />
                    Edit
                  </Button>
                  <Button 
                    variant="outline"
                    onClick={() => {
                      setShowDetails(false);
                      setSelectedWell(null);
                      setIsEditing(false);
                      setIsAddingNew(false);
                    }}
                  >
                    <X className="h-4 w-4 mr-2" />
                    Close
                  </Button>
                </>
              )}
            </div>
          </div>
        </div>

        {/* Detail Form in White Background */}
        <div className="bg-white rounded-lg shadow-sm p-6">
          {/* Tabs for Well Details */}
          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList className="grid w-full grid-cols-6">
              <TabsTrigger value="general">General</TabsTrigger>
              <TabsTrigger value="location">Location</TabsTrigger>
              <TabsTrigger value="production">Production</TabsTrigger>
              <TabsTrigger value="ownership">Ownership</TabsTrigger>
              <TabsTrigger value="financial">Financial</TabsTrigger>
              <TabsTrigger value="documents">Documents</TabsTrigger>
            </TabsList>

            {/* General Tab */}
            <TabsContent value="general" className="space-y-4 mt-6">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="wellId">Well ID *</Label>
                  <Input
                    id="wellId"
                    value={formData.wellId || ''}
                    onChange={(e) => setFormData({...formData, wellId: e.target.value})}
                    disabled={!isEditing}
                    className={!isEditing ? 'bg-gray-50' : ''}
                  />
                </div>
                <div>
                  <Label htmlFor="wellName">Well Name *</Label>
                  <Input
                    id="wellName"
                    value={formData.wellName || ''}
                    onChange={(e) => setFormData({...formData, wellName: e.target.value})}
                    disabled={!isEditing}
                    className={!isEditing ? 'bg-gray-50' : ''}
                  />
                </div>
                <div>
                  <Label htmlFor="apiNumber">API Number</Label>
                  <Input
                    id="apiNumber"
                    value={formData.apiNumber || ''}
                    onChange={(e) => setFormData({...formData, apiNumber: e.target.value})}
                    disabled={!isEditing}
                    className={!isEditing ? 'bg-gray-50' : ''}
                    placeholder="42-XXX-XXXXX"
                  />
                </div>
                <div>
                  <Label htmlFor="leaseId">Lease ID</Label>
                  <Input
                    id="leaseId"
                    value={formData.leaseId || ''}
                    onChange={(e) => setFormData({...formData, leaseId: e.target.value})}
                    disabled={!isEditing}
                    className={!isEditing ? 'bg-gray-50' : ''}
                  />
                </div>
                <div>
                  <Label htmlFor="status">Status</Label>
                  <Select 
                    value={formData.status} 
                    onValueChange={(value) => setFormData({...formData, status: value as Well['status']})}
                    disabled={!isEditing}
                  >
                    <SelectTrigger id="status" className={!isEditing ? 'bg-gray-50' : ''}>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="Active">Active</SelectItem>
                      <SelectItem value="Inactive">Inactive</SelectItem>
                      <SelectItem value="Suspended">Suspended</SelectItem>
                      <SelectItem value="Abandoned">Abandoned</SelectItem>
                      <SelectItem value="Drilling">Drilling</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div>
                  <Label htmlFor="type">Well Type</Label>
                  <Select 
                    value={formData.type} 
                    onValueChange={(value) => setFormData({...formData, type: value as Well['type']})}
                    disabled={!isEditing}
                  >
                    <SelectTrigger id="type" className={!isEditing ? 'bg-gray-50' : ''}>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="Oil">Oil</SelectItem>
                      <SelectItem value="Gas">Gas</SelectItem>
                      <SelectItem value="Oil & Gas">Oil & Gas</SelectItem>
                      <SelectItem value="Injection">Injection</SelectItem>
                      <SelectItem value="SWD">SWD</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div>
                  <Label htmlFor="operator">Operator</Label>
                  <Input
                    id="operator"
                    value={formData.operator || ''}
                    onChange={(e) => setFormData({...formData, operator: e.target.value})}
                    disabled={!isEditing}
                    className={!isEditing ? 'bg-gray-50' : ''}
                  />
                </div>
                <div>
                  <Label htmlFor="formation">Formation</Label>
                  <Input
                    id="formation"
                    value={formData.formation || ''}
                    onChange={(e) => setFormData({...formData, formation: e.target.value})}
                    disabled={!isEditing}
                    className={!isEditing ? 'bg-gray-50' : ''}
                  />
                </div>
              </div>
            </TabsContent>

            {/* Location Tab */}
            <TabsContent value="location" className="space-y-4 mt-6">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="fieldName">Field Name</Label>
                  <Input
                    id="fieldName"
                    value={formData.fieldName || ''}
                    onChange={(e) => setFormData({...formData, fieldName: e.target.value})}
                    disabled={!isEditing}
                    className={!isEditing ? 'bg-gray-50' : ''}
                  />
                </div>
                <div>
                  <Label htmlFor="county">County</Label>
                  <Input
                    id="county"
                    value={formData.county || ''}
                    onChange={(e) => setFormData({...formData, county: e.target.value})}
                    disabled={!isEditing}
                    className={!isEditing ? 'bg-gray-50' : ''}
                  />
                </div>
                <div>
                  <Label htmlFor="state">State</Label>
                  <Input
                    id="state"
                    value={formData.state || ''}
                    onChange={(e) => setFormData({...formData, state: e.target.value})}
                    disabled={!isEditing}
                    className={!isEditing ? 'bg-gray-50' : ''}
                    placeholder="TX"
                  />
                </div>
                <div>
                  <Label htmlFor="depth">Total Depth (ft)</Label>
                  <Input
                    id="depth"
                    type="number"
                    value={formData.depth || ''}
                    onChange={(e) => setFormData({...formData, depth: parseFloat(e.target.value)})}
                    disabled={!isEditing}
                    className={!isEditing ? 'bg-gray-50' : ''}
                  />
                </div>
                <div>
                  <Label htmlFor="latitude">Latitude</Label>
                  <Input
                    id="latitude"
                    type="number"
                    step="0.000001"
                    value={formData.latitude || ''}
                    onChange={(e) => setFormData({...formData, latitude: parseFloat(e.target.value)})}
                    disabled={!isEditing}
                    className={!isEditing ? 'bg-gray-50' : ''}
                  />
                </div>
                <div>
                  <Label htmlFor="longitude">Longitude</Label>
                  <Input
                    id="longitude"
                    type="number"
                    step="0.000001"
                    value={formData.longitude || ''}
                    onChange={(e) => setFormData({...formData, longitude: parseFloat(e.target.value)})}
                    disabled={!isEditing}
                    className={!isEditing ? 'bg-gray-50' : ''}
                  />
                </div>
              </div>
              
              {/* Map Placeholder */}
              <div className="mt-6 border rounded-lg p-4 bg-gray-50 h-64 flex items-center justify-center">
                <div className="text-center text-gray-500">
                  <MapPin className="h-12 w-12 mx-auto mb-2" />
                  <p>Map view will be displayed here</p>
                  {formData.latitude && formData.longitude && (
                    <p className="text-sm mt-2">
                      Coordinates: {formData.latitude}, {formData.longitude}
                    </p>
                  )}
                </div>
              </div>
            </TabsContent>

            {/* Production Tab */}
            <TabsContent value="production" className="space-y-4 mt-6">
              <div className="grid grid-cols-3 gap-4">
                <Card>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm flex items-center">
                      <Droplet className="h-4 w-4 mr-2 text-amber-600" />
                      Oil Production
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">
                      {selectedWell?.currentProduction?.oil || 0}
                    </div>
                    <p className="text-xs text-gray-500">BBL/Day</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm flex items-center">
                      <Activity className="h-4 w-4 mr-2 text-blue-600" />
                      Gas Production
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">
                      {selectedWell?.currentProduction?.gas || 0}
                    </div>
                    <p className="text-xs text-gray-500">MCF/Day</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm flex items-center">
                      <Droplet className="h-4 w-4 mr-2 text-gray-600" />
                      Water Production
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">
                      {selectedWell?.currentProduction?.water || 0}
                    </div>
                    <p className="text-xs text-gray-500">BBL/Day</p>
                  </CardContent>
                </Card>
              </div>

              {/* Production History Chart Placeholder */}
              <div className="mt-6 border rounded-lg p-4 bg-gray-50 h-64 flex items-center justify-center">
                <div className="text-center text-gray-500">
                  <BarChart3 className="h-12 w-12 mx-auto mb-2" />
                  <p>Production history chart will be displayed here</p>
                </div>
              </div>
            </TabsContent>

            {/* Ownership Tab */}
            <TabsContent value="ownership" className="space-y-4 mt-6">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="workingInterest">Working Interest (%)</Label>
                  <Input
                    id="workingInterest"
                    type="number"
                    step="0.01"
                    value={formData.workingInterest || ''}
                    onChange={(e) => setFormData({...formData, workingInterest: parseFloat(e.target.value)})}
                    disabled={!isEditing}
                    className={!isEditing ? 'bg-gray-50' : ''}
                  />
                </div>
                <div>
                  <Label htmlFor="netRevenueInterest">Net Revenue Interest (%)</Label>
                  <Input
                    id="netRevenueInterest"
                    type="number"
                    step="0.01"
                    value={formData.netRevenueInterest || ''}
                    onChange={(e) => setFormData({...formData, netRevenueInterest: parseFloat(e.target.value)})}
                    disabled={!isEditing}
                    className={!isEditing ? 'bg-gray-50' : ''}
                  />
                </div>
              </div>

              {/* Ownership Table Placeholder */}
              <div className="mt-6">
                <h3 className="text-sm font-semibold mb-3">Working Interest Owners</h3>
                <div className="border rounded-lg overflow-hidden">
                  <Table>
                    <TableHeader className="bg-gray-50">
                      <TableRow>
                        <TableHead>Owner ID</TableHead>
                        <TableHead>Owner Name</TableHead>
                        <TableHead className="text-right">WI %</TableHead>
                        <TableHead className="text-right">NRI %</TableHead>
                        <TableHead>Status</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      <TableRow>
                        <TableCell colSpan={5} className="text-center py-4 text-gray-500">
                          Owner information will be loaded from WELLINV.dbf
                        </TableCell>
                      </TableRow>
                    </TableBody>
                  </Table>
                </div>
              </div>
            </TabsContent>

            {/* Financial Tab */}
            <TabsContent value="financial" className="space-y-4 mt-6">
              <div className="grid grid-cols-2 gap-4">
                <Card>
                  <CardHeader>
                    <CardTitle className="text-sm">Revenue (Current Month)</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold text-green-600">$0.00</div>
                    <p className="text-xs text-gray-500 mt-1">Oil & Gas sales</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader>
                    <CardTitle className="text-sm">Expenses (Current Month)</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold text-red-600">$0.00</div>
                    <p className="text-xs text-gray-500 mt-1">Operating expenses</p>
                  </CardContent>
                </Card>
              </div>

              <div className="mt-6">
                <h3 className="text-sm font-semibold mb-3">Recent Transactions</h3>
                <div className="border rounded-lg p-4 text-center text-gray-500">
                  Financial transactions will be loaded from revenue and expense tables
                </div>
              </div>
            </TabsContent>

            {/* Documents Tab */}
            <TabsContent value="documents" className="space-y-4 mt-6">
              <div className="border-2 border-dashed border-gray-300 rounded-lg p-8 text-center">
                <FileText className="h-12 w-12 mx-auto text-gray-400 mb-3" />
                <p className="text-gray-600 mb-4">Upload well-related documents</p>
                <Button variant="outline">
                  <Upload className="h-4 w-4 mr-2" />
                  Upload Document
                </Button>
              </div>

              <div className="mt-6">
                <h3 className="text-sm font-semibold mb-3">Document List</h3>
                <div className="border rounded-lg p-4 text-center text-gray-500">
                  No documents uploaded yet
                </div>
              </div>
            </TabsContent>
          </Tabs>
        </div>
      </div>
    );
  }

  // Wells List View - Main screen
  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="bg-white rounded-lg shadow-sm p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <Droplet className="h-8 w-8 text-blue-600" />
            <div>
              <h1 className="text-2xl font-bold text-gray-900">Well Management</h1>
              <p className="text-sm text-gray-500">Manage wells, production data, and ownership information</p>
            </div>
          </div>
          <div className="flex items-center space-x-2">
            <Button
              onClick={handleAddNew}
              className="bg-blue-600 hover:bg-blue-700"
            >
              <Plus className="h-4 w-4 mr-2" />
              New Well
            </Button>
            <Button variant="outline">
              <Upload className="h-4 w-4 mr-2" />
              Import
            </Button>
            <Button variant="outline">
              <Download className="h-4 w-4 mr-2" />
              Export
            </Button>
          </div>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Total Wells</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{wells.length}</div>
            <p className="text-xs text-gray-500 mt-1">All statuses</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Active Wells</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              {wells.filter(w => w.status === 'Active').length}
            </div>
            <p className="text-xs text-gray-500 mt-1">Currently producing</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Avg Working Interest</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {wells.length > 0 
                ? (wells.reduce((sum, w) => sum + w.workingInterest, 0) / wells.length).toFixed(1)
                : 0}%
            </div>
            <p className="text-xs text-gray-500 mt-1">Across all wells</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Wells Drilling</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-blue-600">
              {wells.filter(w => w.status === 'Drilling').length}
            </div>
            <p className="text-xs text-gray-500 mt-1">In progress</p>
          </CardContent>
        </Card>
      </div>

      {/* Main Content */}
      <div className="bg-white rounded-lg shadow-sm">
        <div className="p-6">
          {/* Search and Filters */}
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center space-x-4 flex-1">
              <div className="relative flex-1 max-w-md">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
                <Input
                  placeholder="Search by Well ID, Name, or API Number..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10"
                />
              </div>
              <Select value={filterStatus} onValueChange={setFilterStatus}>
                <SelectTrigger className="w-40">
                  <SelectValue placeholder="All Statuses" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Statuses</SelectItem>
                  <SelectItem value="Active">Active</SelectItem>
                  <SelectItem value="Inactive">Inactive</SelectItem>
                  <SelectItem value="Suspended">Suspended</SelectItem>
                  <SelectItem value="Abandoned">Abandoned</SelectItem>
                  <SelectItem value="Drilling">Drilling</SelectItem>
                </SelectContent>
              </Select>
              <Select value={filterType} onValueChange={setFilterType}>
                <SelectTrigger className="w-40">
                  <SelectValue placeholder="All Types" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Types</SelectItem>
                  <SelectItem value="Oil">Oil</SelectItem>
                  <SelectItem value="Gas">Gas</SelectItem>
                  <SelectItem value="Oil & Gas">Oil & Gas</SelectItem>
                  <SelectItem value="Injection">Injection</SelectItem>
                  <SelectItem value="SWD">SWD</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Browser Mode Warning */}
          {!isWailsAvailable && (
            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-4">
              <div className="flex items-center">
                <AlertCircle className="h-5 w-5 text-yellow-600 mr-2" />
                <div>
                  <p className="text-sm font-medium text-yellow-800">Browser Mode - Using Mock Data</p>
                  <p className="text-xs text-yellow-600 mt-1">
                    Run with 'wails dev' from the desktop directory to connect to real DBF data.
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* Wells Table */}
          <div className="border rounded-lg overflow-hidden">
            <Table>
              <TableHeader className="bg-gray-50">
                <TableRow>
                  <TableHead>Well ID</TableHead>
                  <TableHead>Well Name</TableHead>
                  <TableHead>API Number</TableHead>
                  <TableHead>Field</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead className="text-right">WI %</TableHead>
                  <TableHead className="text-right">NRI %</TableHead>
                  <TableHead></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {loading ? (
                  <TableRow>
                    <TableCell colSpan={9} className="text-center py-8 text-gray-500">
                      Loading wells...
                    </TableCell>
                  </TableRow>
                ) : filteredWells.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={9} className="text-center py-8 text-gray-500">
                      No wells found
                    </TableCell>
                  </TableRow>
                ) : (
                  filteredWells.map((well) => (
                    <TableRow 
                      key={well.id}
                      className="cursor-pointer hover:bg-gray-50"
                      onClick={() => {
                        console.log('Well row clicked:', well.wellId);
                        setSelectedWell(well);
                        setFormData(well);
                        setShowDetails(true);
                        setIsAddingNew(false);
                        setIsEditing(false);
                        setActiveTab('general');
                        console.log('showDetails set to:', true, 'selectedWell:', well);
                      }}
                    >
                      <TableCell className="font-mono text-sm">{well.wellId}</TableCell>
                      <TableCell className="font-medium">{well.wellName}</TableCell>
                      <TableCell className="text-sm">{well.apiNumber}</TableCell>
                      <TableCell className="text-sm">{well.fieldName}</TableCell>
                      <TableCell>{getStatusBadge(well.status)}</TableCell>
                      <TableCell>{getTypeBadge(well.type)}</TableCell>
                      <TableCell className="text-right">{well.workingInterest.toFixed(2)}%</TableCell>
                      <TableCell className="text-right">{well.netRevenueInterest.toFixed(2)}%</TableCell>
                      <TableCell>
                        <div className="flex items-center space-x-1">
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={(e) => {
                              e.stopPropagation();
                              setSelectedWell(well);
                              setFormData(well);
                              setIsEditing(true);
                              setShowDetails(true);
                              setIsAddingNew(false);
                              setActiveTab('general');
                            }}
                          >
                            <Edit className="h-4 w-4" />
                          </Button>
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleDeleteWell(well.id);
                            }}
                            className="text-red-500 hover:text-red-700"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </div>
      </div>
    </div>
  );
}