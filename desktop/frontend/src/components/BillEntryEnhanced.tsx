import React, { useEffect } from 'react';
import { useForm, useFieldArray, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui/table';
import { RadioGroup, RadioGroupItem } from './ui/radio-group';
import { Checkbox } from './ui/checkbox';
import { Calendar } from './ui/calendar';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover';
import { Alert, AlertDescription } from './ui/alert';
import { CalendarIcon, Plus, Trash2, Copy, RotateCcw, Search, Save, X, FileText, DollarSign, Loader2 } from 'lucide-react';
import { format } from 'date-fns';
import { cn } from '../lib/utils';

// Zod schema for validation
const lineItemSchema = z.object({
  wellId: z.string().optional(),
  expCode: z.string().optional(),
  cls: z.string().default('0'),
  deck: z.string().optional(),
  description: z.string().min(1, 'Description is required'),
  account: z.string().min(1, 'Account is required'),
  afeNo: z.string().optional(),
  dept: z.string().optional(),
  year: z.string().default(new Date().getFullYear().toString()),
  period: z.string().default((new Date().getMonth() + 1).toString().padStart(2, '0')),
  allocateTo: z.string().optional(),
  amount: z.number().positive('Amount must be greater than 0')
});

const billSchema = z.object({
  vendorId: z.string().min(1, 'Vendor ID is required'),
  vendorName: z.string().optional(),
  invoiceNo: z.string().min(1, 'Invoice number is required'),
  reference: z.string().optional(),
  terms: z.string().default('NONE'),
  invoiceDate: z.date({
    required_error: 'Invoice date is required',
  }),
  postDate: z.date().default(new Date()),
  dueDate: z.date().optional(),
  discDate: z.date().optional(),
  approvedToPay: z.boolean().default(false),
  isRecurring: z.boolean().default(false),
  billSource: z.enum(['manual', 'imported', 'energylink']).default('manual'),
  lineItems: z.array(lineItemSchema).min(1, 'At least one line item is required')
});

type BillFormData = z.infer<typeof billSchema>;

interface BillEntryEnhancedProps {
  companyName?: string;
  currentUser?: any;
  billId?: string; // For editing existing bills
}

// API functions (to be connected to Go backend)
const fetchBill = async (billId: string): Promise<BillFormData> => {
  // TODO: Replace with actual API call
  // const response = await fetch(`/api/apbill/${billId}`);
  // return response.json();
  throw new Error('Not implemented');
};

const saveBill = async (data: BillFormData): Promise<{ success: boolean; id: string }> => {
  // TODO: Replace with actual API call
  // const response = await fetch('/api/apbill', {
  //   method: 'POST',
  //   headers: { 'Content-Type': 'application/json' },
  //   body: JSON.stringify(data)
  // });
  // return response.json();
  console.log('Saving bill:', data);
  return { success: true, id: 'mock-id' };
};

const duplicateBill = async (billId: string): Promise<{ success: boolean; newId: string }> => {
  // TODO: Replace with actual API call
  console.log('Duplicating bill:', billId);
  return { success: true, newId: 'mock-new-id' };
};

const reverseBill = async (billId: string): Promise<{ success: boolean }> => {
  // TODO: Replace with actual API call
  console.log('Reversing bill:', billId);
  return { success: true };
};

export default function BillEntryEnhanced({ companyName, currentUser, billId }: BillEntryEnhancedProps) {
  const queryClient = useQueryClient();
  const isEditMode = !!billId;

  // React Hook Form setup
  const {
    control,
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors, isSubmitting },
    reset
  } = useForm<BillFormData>({
    resolver: zodResolver(billSchema),
    defaultValues: {
      vendorId: '',
      vendorName: '',
      invoiceNo: '',
      reference: '',
      terms: 'NONE',
      invoiceDate: new Date(),
      postDate: new Date(),
      approvedToPay: false,
      isRecurring: false,
      billSource: 'manual',
      lineItems: []
    }
  });

  const { fields, append, remove } = useFieldArray({
    control,
    name: 'lineItems'
  });

  // Watch for changes that affect calculations
  const watchLineItems = watch('lineItems');
  const watchInvoiceDate = watch('invoiceDate');
  const watchTerms = watch('terms');

  // Calculate totals
  const invoiceTotal = watchLineItems?.reduce((sum, item) => sum + (item.amount || 0), 0) || 0;

  // Auto-calculate due date based on terms
  useEffect(() => {
    if (watchInvoiceDate && watchTerms && watchTerms !== 'NONE') {
      const invoice = new Date(watchInvoiceDate);
      let due = new Date(invoice);
      let disc = null;
      
      switch(watchTerms) {
        case 'NET30':
          due.setDate(due.getDate() + 30);
          break;
        case 'NET60':
          due.setDate(due.getDate() + 60);
          break;
        case 'NET90':
          due.setDate(due.getDate() + 90);
          break;
        case '210NET30':
          disc = new Date(invoice);
          disc.setDate(disc.getDate() + 10);
          setValue('discDate', disc);
          due.setDate(due.getDate() + 30);
          break;
      }
      setValue('dueDate', due);
    }
  }, [watchInvoiceDate, watchTerms, setValue]);

  // Fetch existing bill data if in edit mode
  const { data: existingBill, isLoading: isLoadingBill } = useQuery({
    queryKey: ['bill', billId],
    queryFn: () => fetchBill(billId!),
    enabled: isEditMode,
    onSuccess: (data) => {
      reset(data);
    }
  });

  // Mutations
  const saveMutation = useMutation({
    mutationFn: saveBill,
    onSuccess: (result) => {
      queryClient.invalidateQueries(['bills']);
      // TODO: Navigate to bill list or show success message
      console.log('Bill saved successfully:', result.id);
    },
    onError: (error) => {
      console.error('Error saving bill:', error);
    }
  });

  const duplicateMutation = useMutation({
    mutationFn: () => duplicateBill(billId!),
    onSuccess: (result) => {
      // TODO: Navigate to new bill
      console.log('Bill duplicated:', result.newId);
    }
  });

  const reverseMutation = useMutation({
    mutationFn: () => reverseBill(billId!),
    onSuccess: () => {
      queryClient.invalidateQueries(['bills']);
      console.log('Bill reversed successfully');
    }
  });

  const onSubmit = (data: BillFormData) => {
    saveMutation.mutate(data);
  };

  const handleAddLineItem = () => {
    append({
      wellId: '',
      expCode: '',
      cls: '0',
      deck: '',
      description: '',
      account: '',
      afeNo: '',
      dept: '',
      year: new Date().getFullYear().toString(),
      period: (new Date().getMonth() + 1).toString().padStart(2, '0'),
      allocateTo: '',
      amount: 0
    });
  };

  if (isLoadingBill) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-gray-500" />
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Header */}
      <div className="bg-white rounded-lg shadow-sm p-6">
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center space-x-4">
            <FileText className="h-8 w-8 text-blue-600" />
            <div>
              <h1 className="text-2xl font-bold text-gray-900">
                {isEditMode ? 'Edit Bill' : 'New Bill Entry'}
              </h1>
              <p className="text-sm text-gray-500">Enter vendor bills and expenses</p>
            </div>
          </div>
          <div className="flex items-center space-x-2">
            <Button 
              type="submit" 
              disabled={isSubmitting || saveMutation.isLoading}
              className="bg-blue-600 hover:bg-blue-700"
            >
              {(isSubmitting || saveMutation.isLoading) ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Saving...
                </>
              ) : (
                <>
                  <Save className="h-4 w-4 mr-2" />
                  Save Bill
                </>
              )}
            </Button>
            <Button type="button" variant="outline" onClick={() => window.history.back()}>
              <X className="h-4 w-4 mr-2" />
              Cancel
            </Button>
          </div>
        </div>

        {/* Status Badge */}
        <div className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-yellow-100 text-yellow-800">
          {isEditMode ? `BILL #${billId}` : 'NEW BILL'}
        </div>
      </div>

      {/* Display validation errors */}
      {Object.keys(errors).length > 0 && (
        <Alert variant="destructive">
          <AlertDescription>
            Please correct the following errors:
            <ul className="list-disc list-inside mt-2">
              {errors.vendorId && <li>{errors.vendorId.message}</li>}
              {errors.invoiceNo && <li>{errors.invoiceNo.message}</li>}
              {errors.invoiceDate && <li>{errors.invoiceDate.message}</li>}
              {errors.lineItems && <li>{errors.lineItems.message || 'Line items have errors'}</li>}
            </ul>
          </AlertDescription>
        </Alert>
      )}

      {/* Main Form */}
      <div className="bg-white rounded-lg shadow-sm">
        <div className="p-6">
          {/* Vendor Information */}
          <div className="grid grid-cols-12 gap-4 mb-6">
            <div className="col-span-3">
              <Label htmlFor="vendorId">Vendor ID *</Label>
              <div className="flex space-x-2">
                <Input
                  id="vendorId"
                  {...register('vendorId')}
                  placeholder="Enter vendor ID"
                  className={errors.vendorId ? 'border-red-500' : ''}
                />
                <Button size="sm" variant="outline" type="button">
                  <Search className="h-4 w-4" />
                </Button>
              </div>
              {errors.vendorId && (
                <p className="text-red-500 text-xs mt-1">{errors.vendorId.message}</p>
              )}
            </div>
            <div className="col-span-5">
              <Label htmlFor="vendorName">Vendor Name</Label>
              <Input
                id="vendorName"
                {...register('vendorName')}
                placeholder="Vendor name"
                disabled
                className="bg-gray-50"
              />
            </div>
            <div className="col-span-2">
              <Label htmlFor="postDate">Post Date</Label>
              <Controller
                name="postDate"
                control={control}
                render={({ field }) => (
                  <Popover>
                    <PopoverTrigger asChild>
                      <Button
                        type="button"
                        variant="outline"
                        className={cn(
                          "w-full justify-start text-left font-normal",
                          !field.value && "text-muted-foreground"
                        )}
                      >
                        <CalendarIcon className="mr-2 h-4 w-4" />
                        {field.value ? format(field.value, "MM/dd/yyyy") : "Select date"}
                      </Button>
                    </PopoverTrigger>
                    <PopoverContent className="w-auto p-0">
                      <Calendar
                        mode="single"
                        selected={field.value}
                        onSelect={field.onChange}
                        initialFocus
                      />
                    </PopoverContent>
                  </Popover>
                )}
              />
            </div>
            <div className="col-span-2 flex items-end">
              <Controller
                name="approvedToPay"
                control={control}
                render={({ field }) => (
                  <div className="flex items-center space-x-2">
                    <Checkbox
                      id="approved"
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                    <Label htmlFor="approved" className="text-sm font-medium leading-none">
                      Approved to Pay
                    </Label>
                  </div>
                )}
              />
            </div>
          </div>

          {/* Invoice Information */}
          <div className="grid grid-cols-12 gap-4 mb-6">
            <div className="col-span-3">
              <Label htmlFor="invoiceNo">Invoice No *</Label>
              <Input
                id="invoiceNo"
                {...register('invoiceNo')}
                placeholder="Enter invoice number"
                className={errors.invoiceNo ? 'border-red-500' : ''}
              />
              {errors.invoiceNo && (
                <p className="text-red-500 text-xs mt-1">{errors.invoiceNo.message}</p>
              )}
            </div>
            <div className="col-span-3">
              <Label htmlFor="reference">Ref:</Label>
              <Input
                id="reference"
                {...register('reference')}
                placeholder="Reference"
              />
            </div>
            <div className="col-span-2">
              <Label htmlFor="terms">Terms</Label>
              <Controller
                name="terms"
                control={control}
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id="terms">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="NONE">NONE</SelectItem>
                      <SelectItem value="NET30">NET 30</SelectItem>
                      <SelectItem value="NET60">NET 60</SelectItem>
                      <SelectItem value="NET90">NET 90</SelectItem>
                      <SelectItem value="210NET30">2/10 NET 30</SelectItem>
                    </SelectContent>
                  </Select>
                )}
              />
            </div>
            <div className="col-span-2">
              <Label htmlFor="invoiceDate">Invoice Date *</Label>
              <Controller
                name="invoiceDate"
                control={control}
                render={({ field }) => (
                  <Popover>
                    <PopoverTrigger asChild>
                      <Button
                        type="button"
                        variant="outline"
                        className={cn(
                          "w-full justify-start text-left font-normal",
                          !field.value && "text-muted-foreground",
                          errors.invoiceDate && "border-red-500"
                        )}
                      >
                        <CalendarIcon className="mr-2 h-4 w-4" />
                        {field.value ? format(field.value, "MM/dd/yyyy") : "Select date"}
                      </Button>
                    </PopoverTrigger>
                    <PopoverContent className="w-auto p-0">
                      <Calendar
                        mode="single"
                        selected={field.value}
                        onSelect={field.onChange}
                        initialFocus
                      />
                    </PopoverContent>
                  </Popover>
                )}
              />
              {errors.invoiceDate && (
                <p className="text-red-500 text-xs mt-1">{errors.invoiceDate.message}</p>
              )}
            </div>
            <div className="col-span-2 flex items-end">
              <Controller
                name="isRecurring"
                control={control}
                render={({ field }) => (
                  <Button
                    type="button"
                    variant="outline"
                    className="w-full"
                    onClick={() => field.onChange(!field.value)}
                  >
                    {field.value ? 'Recurring Bill ✓' : 'Recurring Bills'}
                  </Button>
                )}
              />
            </div>
          </div>
        </div>

        {/* Line Items Section */}
        <div className="border-t border-gray-200 p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-gray-900">Line Items</h3>
            <Button type="button" onClick={handleAddLineItem} className="bg-green-600 hover:bg-green-700">
              <Plus className="h-4 w-4 mr-2" />
              Add Line Item
            </Button>
          </div>

          {/* Line Items Table */}
          <div className="border rounded-lg overflow-hidden">
            <Table>
              <TableHeader className="bg-gray-50">
                <TableRow>
                  <TableHead className="w-[100px]">Well ID</TableHead>
                  <TableHead className="w-[80px]">Exp Code</TableHead>
                  <TableHead className="w-[50px]">Cls</TableHead>
                  <TableHead>Description *</TableHead>
                  <TableHead className="w-[120px]">Account *</TableHead>
                  <TableHead className="w-[100px] text-right">Amount *</TableHead>
                  <TableHead className="w-[50px]"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {fields.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center text-gray-500 py-8">
                      No line items added yet. Click "Add Line Item" to begin.
                    </TableCell>
                  </TableRow>
                ) : (
                  fields.map((field, index) => (
                    <TableRow key={field.id}>
                      <TableCell>
                        <Input
                          {...register(`lineItems.${index}.wellId`)}
                          placeholder="Well ID"
                          className="text-xs"
                        />
                      </TableCell>
                      <TableCell>
                        <Input
                          {...register(`lineItems.${index}.expCode`)}
                          placeholder="Code"
                          className="text-xs"
                        />
                      </TableCell>
                      <TableCell>
                        <Controller
                          name={`lineItems.${index}.cls`}
                          control={control}
                          render={({ field }) => (
                            <Select value={field.value} onValueChange={field.onChange}>
                              <SelectTrigger className="text-xs">
                                <SelectValue />
                              </SelectTrigger>
                              <SelectContent>
                                <SelectItem value="0">0</SelectItem>
                                <SelectItem value="1">1</SelectItem>
                                <SelectItem value="2">2</SelectItem>
                                <SelectItem value="P">P</SelectItem>
                              </SelectContent>
                            </Select>
                          )}
                        />
                      </TableCell>
                      <TableCell>
                        <Input
                          {...register(`lineItems.${index}.description`)}
                          placeholder="Description"
                          className={cn(
                            "text-sm",
                            errors.lineItems?.[index]?.description && "border-red-500"
                          )}
                        />
                      </TableCell>
                      <TableCell>
                        <Input
                          {...register(`lineItems.${index}.account`)}
                          placeholder="Account"
                          className={cn(
                            "text-xs font-mono",
                            errors.lineItems?.[index]?.account && "border-red-500"
                          )}
                        />
                      </TableCell>
                      <TableCell>
                        <Input
                          type="number"
                          step="0.01"
                          {...register(`lineItems.${index}.amount`, { valueAsNumber: true })}
                          placeholder="0.00"
                          className={cn(
                            "text-right font-mono text-sm",
                            errors.lineItems?.[index]?.amount && "border-red-500"
                          )}
                        />
                      </TableCell>
                      <TableCell>
                        <Button
                          type="button"
                          size="sm"
                          variant="ghost"
                          onClick={() => remove(index)}
                          className="h-8 w-8 p-0 text-red-500 hover:text-red-700"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>

          {/* Action Buttons */}
          <div className="flex items-center justify-between mt-4">
            <div className="flex space-x-2">
              {isEditMode && (
                <>
                  <Button 
                    type="button" 
                    variant="outline" 
                    size="sm" 
                    onClick={() => reverseMutation.mutate()}
                    disabled={reverseMutation.isLoading}
                  >
                    <RotateCcw className="h-4 w-4 mr-2" />
                    Reverse Bill
                  </Button>
                  <Button 
                    type="button" 
                    variant="outline" 
                    size="sm" 
                    onClick={() => duplicateMutation.mutate()}
                    disabled={duplicateMutation.isLoading}
                  >
                    <Copy className="h-4 w-4 mr-2" />
                    Duplicate Bill
                  </Button>
                </>
              )}
            </div>
            
            {/* Source Radio Options */}
            <div className="flex items-center space-x-4">
              <span className="text-sm font-medium">Source:</span>
              <Controller
                name="billSource"
                control={control}
                render={({ field }) => (
                  <RadioGroup value={field.value} onValueChange={field.onChange} className="flex space-x-4">
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="manual" id="manual" />
                      <Label htmlFor="manual" className="text-sm">Manual</Label>
                    </div>
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="imported" id="imported" />
                      <Label htmlFor="imported" className="text-sm">Imported</Label>
                    </div>
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="energylink" id="energylink" />
                      <Label htmlFor="energylink" className="text-sm">EnergyLink</Label>
                    </div>
                  </RadioGroup>
                )}
              />
            </div>
          </div>
        </div>

        {/* Totals Section */}
        <div className="border-t border-gray-200 bg-gray-50 p-6">
          <div className="flex justify-end">
            <div className="w-64 space-y-2">
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Invoice Total:</span>
                <span className="font-mono font-semibold">${invoiceTotal.toFixed(2)}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Invoice Balance:</span>
                <span className="font-mono font-semibold">${invoiceTotal.toFixed(2)}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </form>
  );
}