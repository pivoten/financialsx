import React, { useState, useEffect, useMemo } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Alert, AlertDescription } from './ui/alert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from './ui/tabs'
import { ScrollArea } from './ui/scroll-area'
import { Badge } from './ui/badge'
import { 
  FileText, 
  Users, 
  DollarSign, 
  FileSearch, 
  Printer,
  Calculator,
  CreditCard,
  UserPlus,
  Receipt,
  ClipboardList,
  Building,
  TrendingUp,
  FileBarChart,
  Settings,
  AlertCircle,
  ExternalLink,
  Search,
  Grid3X3,
  List,
  ChevronRight,
  Database,
  Activity,
  Wrench,
  Home,
  Layers,
  GripVertical
} from 'lucide-react'
import * as WailsApp from '../../wailsjs/go/main/App'
import { sherWareForms, quickAccessForms, type SherWareForm, type SherWareCategory } from '../data/sherware-forms'
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
} from '@dnd-kit/core'
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  rectSortingStrategy,
} from '@dnd-kit/sortable'
import {
  useSortable,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'

// SortableFormCard component for drag and drop
interface SortableFormCardProps {
  form: SherWareForm & { orderIndex?: number }
  isLaunching: boolean
  vfpEnabled: boolean
  onLaunch: (formName: string) => void
}

function SortableFormCard({ form, isLaunching, vfpEnabled, onLaunch }: SortableFormCardProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: form.id })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  return (
    <div ref={setNodeRef} style={style}>
      <Card
        className={`border border-gray-200 transition-all hover:shadow-md relative ${
          !vfpEnabled || isLaunching ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'
        } ${isDragging ? 'z-50' : ''}`}
      >
        <div
          className="absolute top-2 left-2 cursor-move z-10 p-1 hover:bg-gray-100 rounded"
          {...attributes}
          {...listeners}
        >
          <GripVertical className="h-4 w-4 text-gray-400" />
        </div>
        <div
          onClick={() => {
            if (vfpEnabled && !isLaunching) {
              onLaunch(form.formName)
            }
          }}
        >
          <CardHeader className="pb-2 pt-3 px-3 pl-10">
            <div className="flex items-start justify-between">
              <FileText className="h-5 w-5 text-blue-600" />
              <ExternalLink className="h-3 w-3 text-gray-400" />
            </div>
            <CardTitle className="text-sm mt-2 line-clamp-1">
              {form.title}
            </CardTitle>
          </CardHeader>
          <CardContent className="px-3 pb-3 pl-10">
            <p className="text-xs text-gray-500 line-clamp-2">
              {isLaunching ? 'Launching...' : form.description}
            </p>
            {'category' in form && (
              <Badge variant="outline" className="mt-2 text-xs">
                {(form as any).category}
              </Badge>
            )}
          </CardContent>
        </div>
      </Card>
    </div>
  )
}

// Icon mapping for categories
const categoryIcons: Record<string, React.ElementType> = {
  'Company Management': Building,
  'File Operations': FileText,
  'System Setup': Settings,
  'General Ledger': Calculator,
  'Accounts Receivable': DollarSign,
  'Accounts Payable': Receipt,
  'Cash Management': CreditCard,
  'Oil & Gas - Wells': Activity,
  'Oil & Gas - Revenue': TrendingUp,
  'Oil & Gas - Expenses': FileBarChart,
  'default': FileText
}

export default function SherWareLegacy() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [vfpEnabled, setVfpEnabled] = useState(false)
  const [launching, setLaunching] = useState<string | null>(null)
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedCategory, setSelectedCategory] = useState<string>('quick-access')
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')
  const [formOrder, setFormOrder] = useState<string[]>([])
  
  // DnD sensors
  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  )

  useEffect(() => {
    checkVFPStatus()
    loadFormOrder()
  }, [])
  
  useEffect(() => {
    // Save form order whenever it changes
    if (formOrder.length > 0) {
      saveFormOrder()
    }
  }, [formOrder])

  const checkVFPStatus = async () => {
    try {
      const settings = await WailsApp.GetVFPSettings()
      setVfpEnabled(settings?.enabled || false)
    } catch (err) {
      console.error('Failed to check VFP status:', err)
      setVfpEnabled(false)
    }
  }
  
  const loadFormOrder = () => {
    const key = `sherware-forms-order-${selectedCategory}`
    const saved = localStorage.getItem(key)
    if (saved) {
      try {
        const order = JSON.parse(saved)
        setFormOrder(order)
      } catch (e) {
        console.error('Failed to load form order:', e)
        setFormOrder([])
      }
    }
  }
  
  const saveFormOrder = () => {
    const key = `sherware-forms-order-${selectedCategory}`
    localStorage.setItem(key, JSON.stringify(formOrder))
  }
  
  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event

    if (active.id !== over?.id) {
      setFormOrder((items) => {
        const oldIndex = items.indexOf(active.id as string)
        const newIndex = items.indexOf(over?.id as string)
        return arrayMove(items, oldIndex, newIndex)
      })
    }
  }

  const launchForm = async (formName: string, argument?: string) => {
    setError(null)
    setLaunching(formName)
    
    try {
      const result = await WailsApp.LaunchVFPForm(formName, argument || '')
      
      if (!result.success) {
        setError(result.message || 'Failed to launch form')
      }
    } catch (err) {
      console.error(`Failed to launch ${formName}:`, err)
      setError(`Failed to launch ${formName}: ${err.message || 'Unknown error'}`)
    } finally {
      setLaunching(null)
    }
  }

  // Load form order when category changes
  useEffect(() => {
    loadFormOrder()
  }, [selectedCategory])
  
  // Filter and order forms based on search term and saved order
  const filteredForms = useMemo(() => {
    let forms: SherWareForm[] = []
    
    // If there's a search term, search across ALL categories
    if (searchTerm) {
      const searchLower = searchTerm.toLowerCase()
      const allMatchingForms: (SherWareForm & { category?: string })[] = []
      
      // Search in quick access forms
      quickAccessForms.forEach(form => {
        if (
          form.title.toLowerCase().includes(searchLower) ||
          form.description.toLowerCase().includes(searchLower) ||
          form.formName.toLowerCase().includes(searchLower)
        ) {
          allMatchingForms.push({ ...form })
        }
      })
      
      // Search in all category forms
      sherWareForms.forEach(category => {
        category.forms.forEach(form => {
          if (
            form.title.toLowerCase().includes(searchLower) ||
            form.description.toLowerCase().includes(searchLower) ||
            form.formName.toLowerCase().includes(searchLower)
          ) {
            // Add category info to the form for display
            allMatchingForms.push({ ...form, category: category.category })
          }
        })
      })
      
      // Remove duplicates (forms that appear in both quick access and categories)
      const uniqueForms = Array.from(
        new Map(allMatchingForms.map(form => [form.id, form])).values()
      )
      
      forms = uniqueForms
    } else {
      // No search term - use selected category
      if (selectedCategory === 'quick-access') {
        forms = quickAccessForms
      } else if (selectedCategory === 'all') {
        const allForms: SherWareForm[] = []
        sherWareForms.forEach(category => {
          category.forms.forEach(form => {
            allForms.push(form)
          })
        })
        forms = allForms
      } else {
        const category = sherWareForms.find(cat => cat.category === selectedCategory)
        if (category) {
          forms = category.forms
        }
      }
    }
    
    // Initialize formOrder if empty
    if (formOrder.length === 0 && forms.length > 0) {
      setFormOrder(forms.map(f => f.id))
      return forms
    }
    
    // Apply saved order
    if (formOrder.length > 0) {
      const orderedForms = [...forms].sort((a, b) => {
        const indexA = formOrder.indexOf(a.id)
        const indexB = formOrder.indexOf(b.id)
        if (indexA === -1) return 1
        if (indexB === -1) return -1
        return indexA - indexB
      })
      return orderedForms
    }
    
    return forms
  }, [selectedCategory, searchTerm, formOrder])

  const getCategoryIcon = (category: string) => {
    const Icon = categoryIcons[category] || categoryIcons.default
    return Icon
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-semibold text-gray-900">SherWare Legacy Forms</h2>
          <p className="text-sm text-gray-500 mt-1">
            Launch Visual FoxPro forms directly from FinancialsX
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <Button
            variant={viewMode === 'grid' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setViewMode('grid')}
          >
            <Grid3X3 className="h-4 w-4" />
          </Button>
          <Button
            variant={viewMode === 'list' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setViewMode('list')}
          >
            <List className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Status Alert */}
      {!vfpEnabled && (
        <Alert className="border-yellow-200 bg-yellow-50">
          <AlertCircle className="h-4 w-4 text-yellow-600" />
          <AlertDescription className="text-yellow-800">
            VFP integration is disabled. Go to Settings → Legacy Integration to enable it.
          </AlertDescription>
        </Alert>
      )}

      {error && (
        <Alert className="border-red-200 bg-red-50">
          <AlertCircle className="h-4 w-4 text-red-600" />
          <AlertDescription className="text-red-800">
            {error}
          </AlertDescription>
        </Alert>
      )}

      {/* Search Bar */}
      <div className="space-y-2">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
          <Input
            type="text"
            placeholder="Search across all forms..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10"
          />
        </div>
        {searchTerm && (
          <div className="flex items-center justify-between px-1">
            <p className="text-sm text-gray-500">
              Searching across all categories • {filteredForms.length} result{filteredForms.length !== 1 ? 's' : ''} found
            </p>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setSearchTerm('')}
              className="h-6 px-2 text-xs"
            >
              Clear search
            </Button>
          </div>
        )}
      </div>

      {/* Main Content with Sidebar */}
      <div className="flex gap-6">
        {/* Category Sidebar */}
        <div className={`w-64 flex-shrink-0 ${searchTerm ? 'opacity-50' : ''}`}>
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm">
                Categories {searchTerm && <span className="text-xs font-normal text-gray-500">(searching all)</span>}
              </CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              <ScrollArea className="h-[600px]">
                <div className="space-y-1 p-3">
                  {/* Quick Access */}
                  <Button
                    variant={selectedCategory === 'quick-access' ? 'secondary' : 'ghost'}
                    className="w-full justify-start"
                    onClick={() => setSelectedCategory('quick-access')}
                  >
                    <Home className="mr-2 h-4 w-4" />
                    Quick Access
                    <Badge className="ml-auto" variant="secondary">
                      {quickAccessForms.length}
                    </Badge>
                  </Button>

                  {/* All Forms */}
                  <Button
                    variant={selectedCategory === 'all' ? 'secondary' : 'ghost'}
                    className="w-full justify-start"
                    onClick={() => setSelectedCategory('all')}
                  >
                    <Layers className="mr-2 h-4 w-4" />
                    All Forms
                    <Badge className="ml-auto" variant="secondary">
                      {sherWareForms.reduce((acc, cat) => acc + cat.forms.length, 0)}
                    </Badge>
                  </Button>

                  <div className="my-2 border-t" />

                  {/* Categories */}
                  {sherWareForms.slice(0, 12).map((category) => {
                    const Icon = getCategoryIcon(category.category)
                    return (
                      <Button
                        key={category.category}
                        variant={selectedCategory === category.category ? 'secondary' : 'ghost'}
                        className="w-full justify-start text-left"
                        onClick={() => setSelectedCategory(category.category)}
                      >
                        <Icon className="mr-2 h-4 w-4 flex-shrink-0" />
                        <span className="truncate">{category.category}</span>
                        <Badge className="ml-auto" variant="outline">
                          {category.forms.length}
                        </Badge>
                      </Button>
                    )
                  })}
                </div>
              </ScrollArea>
            </CardContent>
          </Card>
        </div>

        {/* Forms Display */}
        <div className="flex-1">
          <Card>
            <CardHeader>
              <CardTitle>
                {searchTerm ? (
                  <span>Search Results</span>
                ) : (
                  selectedCategory === 'quick-access' ? 'Quick Access Forms' :
                  selectedCategory === 'all' ? 'All Forms' :
                  selectedCategory
                )}
              </CardTitle>
              <CardDescription>
                {searchTerm ? (
                  <span>Found {filteredForms.length} form{filteredForms.length !== 1 ? 's' : ''} matching "{searchTerm}"</span>
                ) : (
                  <span>{filteredForms.length} form{filteredForms.length !== 1 ? 's' : ''} available</span>
                )}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <ScrollArea className="h-[550px]">
                {viewMode === 'grid' ? (
                  <DndContext
                    sensors={sensors}
                    collisionDetection={closestCenter}
                    onDragEnd={handleDragEnd}
                  >
                    <SortableContext
                      items={filteredForms.map(f => f.id)}
                      strategy={rectSortingStrategy}
                    >
                      <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                        {filteredForms.map((form) => {
                          const isLaunching = launching === form.formName
                          
                          return (
                            <SortableFormCard
                              key={form.id}
                              form={form}
                              isLaunching={isLaunching}
                              vfpEnabled={vfpEnabled}
                              onLaunch={launchForm}
                            />
                          )
                        })}
                      </div>
                    </SortableContext>
                  </DndContext>
                ) : (
                  <div className="space-y-2">
                    {filteredForms.map((form) => {
                      const isLaunching = launching === form.formName
                      
                      return (
                        <div
                          key={form.id}
                          className={`flex items-center justify-between p-3 border rounded-lg transition-all cursor-pointer hover:bg-gray-50 ${
                            !vfpEnabled || isLaunching ? 'opacity-50 cursor-not-allowed' : ''
                          }`}
                          onClick={() => {
                            if (vfpEnabled && !isLaunching) {
                              launchForm(form.formName)
                            }
                          }}
                        >
                          <div className="flex items-center space-x-3">
                            <FileText className="h-5 w-5 text-blue-600" />
                            <div>
                              <p className="font-medium text-sm">{form.title}</p>
                              <p className="text-xs text-gray-500">
                                {isLaunching ? 'Launching...' : form.description}
                              </p>
                            </div>
                          </div>
                          <div className="flex items-center space-x-2">
                            {'category' in form && (
                              <Badge variant="outline" className="text-xs">
                                {(form as any).category}
                              </Badge>
                            )}
                            <ChevronRight className="h-4 w-4 text-gray-400" />
                          </div>
                        </div>
                      )
                    })}
                  </div>
                )}

                {filteredForms.length === 0 && (
                  <div className="text-center py-12">
                    <FileSearch className="h-12 w-12 text-gray-300 mx-auto mb-3" />
                    <p className="text-gray-500">No forms found matching your search</p>
                  </div>
                )}
              </ScrollArea>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}