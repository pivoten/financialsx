import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from './ui/tabs';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui/table';
import { RadioGroup, RadioGroupItem } from './ui/radio-group';
import { Checkbox } from './ui/checkbox';
import { Calendar } from './ui/calendar';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover';
import { CalendarIcon, Plus, Trash2, Copy, RotateCcw, Search, Save, X, FileText, DollarSign } from 'lucide-react';
import { format } from 'date-fns';
import { cn } from '../lib/utils';

interface BillLineItem {
  id: string;
  wellId: string;
  expCode: string;
  cls: string;
  deck: string;
  description: string;
  account: string;
  afeNo: string;
  dept: string;
  year: string;
  period: string;
  allocateTo: string;
  amount: number;
}

interface BillEntryProps {
  companyName?: string;
  currentUser?: any;
}

export default function BillEntry({ companyName, currentUser }: BillEntryProps) {
  const [vendorId, setVendorId] = useState('');
  const [vendorName, setVendorName] = useState('');
  const [invoiceNo, setInvoiceNo] = useState('');
  const [reference, setReference] = useState('');
  const [terms, setTerms] = useState('NONE');
  const [invoiceDate, setInvoiceDate] = useState<Date | undefined>(new Date());
  const [postDate, setPostDate] = useState<Date | undefined>(new Date());
  const [dueDate, setDueDate] = useState<Date | undefined>();
  const [discDate, setDiscDate] = useState<Date | undefined>();
  const [approvedToPay, setApprovedToPay] = useState(false);
  const [isRecurring, setIsRecurring] = useState(false);
  const [billSource, setBillSource] = useState('manual');
  
  // Line item entry fields
  const [currentItem, setCurrentItem] = useState<Partial<BillLineItem>>({
    amount: 0
  });
  const [lineItems, setLineItems] = useState<BillLineItem[]>([]);
  
  // Totals
  const [invoiceTotal, setInvoiceTotal] = useState(0);
  const [invoiceBalance, setInvoiceBalance] = useState(0);

  // Calculate totals when line items change
  useEffect(() => {
    const total = lineItems.reduce((sum, item) => sum + item.amount, 0);
    setInvoiceTotal(total);
    setInvoiceBalance(total); // Will be adjusted when payments are applied
  }, [lineItems]);

  const handleAddLineItem = () => {
    // Validation
    const errors: string[] = [];
    
    if (!currentItem.description) {
      errors.push('Description is required');
    }
    
    if (!currentItem.account) {
      errors.push('Account is required');
    }
    
    if (!currentItem.amount || currentItem.amount <= 0) {
      errors.push('Amount must be greater than 0');
    }
    
    if (currentItem.wellId && !currentItem.expCode) {
      errors.push('Expense code is required when a well is specified');
    }
    
    if (errors.length > 0) {
      // TODO: Show error message
      console.error('Validation errors:', errors);
      return;
    }
    
    const newItem: BillLineItem = {
      id: Date.now().toString(),
      wellId: currentItem.wellId || '',
      expCode: currentItem.expCode || '',
      cls: currentItem.cls || '0',
      deck: currentItem.deck || '',
      description: currentItem.description || '',
      account: currentItem.account || '',
      afeNo: currentItem.afeNo || '',
      dept: currentItem.dept || '',
      year: currentItem.year || new Date().getFullYear().toString(),
      period: currentItem.period || (new Date().getMonth() + 1).toString().padStart(2, '0'),
      allocateTo: currentItem.allocateTo || '',
      amount: currentItem.amount || 0
    };
    
    setLineItems([...lineItems, newItem]);
    
    // Reset form but keep some fields for convenience
    setCurrentItem({ 
      amount: 0,
      year: currentItem.year || new Date().getFullYear().toString(),
      period: currentItem.period || (new Date().getMonth() + 1).toString().padStart(2, '0'),
      cls: currentItem.cls || '0'
    });
  };

  const handleDeleteLineItem = (id: string) => {
    setLineItems(lineItems.filter(item => item.id !== id));
  };

  // Calculate due date based on terms
  useEffect(() => {
    if (invoiceDate && terms && terms !== 'NONE') {
      const invoice = new Date(invoiceDate);
      let due = new Date(invoice);
      let disc = null;
      
      switch(terms) {
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
          setDiscDate(disc);
          due.setDate(due.getDate() + 30);
          break;
      }
      setDueDate(due);
    }
  }, [invoiceDate, terms]);

  const validateBill = (): { isValid: boolean; errors: string[] } => {
    const errors: string[] = [];
    
    if (!vendorId) {
      errors.push('Vendor ID is required');
    }
    
    if (!invoiceNo) {
      errors.push('Invoice number is required');
    }
    
    if (!invoiceDate) {
      errors.push('Invoice date is required');
    }
    
    if (lineItems.length === 0) {
      errors.push('At least one line item is required');
    }
    
    if (invoiceTotal <= 0) {
      errors.push('Invoice total must be greater than 0');
    }
    
    return {
      isValid: errors.length === 0,
      errors
    };
  };

  const handleSaveBill = async () => {
    const validation = validateBill();
    
    if (!validation.isValid) {
      console.error('Bill validation failed:', validation.errors);
      // TODO: Show error message to user
      return;
    }
    
    const billData = {
      vendorId,
      vendorName,
      invoiceNo,
      reference,
      terms,
      invoiceDate,
      postDate,
      dueDate,
      discDate,
      approvedToPay,
      isRecurring,
      billSource,
      lineItems,
      invoiceTotal,
      invoiceBalance
    };
    
    try {
      // TODO: Call backend API to save bill
      console.log('Saving bill:', billData);
      // const result = await SaveBill(companyName, billData);
      // if (result.success) {
      //   // Navigate back or show success message
      // }
    } catch (error) {
      console.error('Error saving bill:', error);
    }
  };

  const handleDuplicateBill = () => {
    // TODO: Implement duplicate functionality
    console.log('Duplicating bill');
  };

  const handleReverseBill = () => {
    // TODO: Implement reverse functionality
    console.log('Reversing bill');
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="bg-white rounded-lg shadow-sm p-6">
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center space-x-4">
            <FileText className="h-8 w-8 text-blue-600" />
            <div>
              <h1 className="text-2xl font-bold text-gray-900">Bill Entry</h1>
              <p className="text-sm text-gray-500">Enter new vendor bills and expenses</p>
            </div>
          </div>
          <div className="flex items-center space-x-2">
            <Button onClick={handleSaveBill} className="bg-blue-600 hover:bg-blue-700">
              <Save className="h-4 w-4 mr-2" />
              Save Bill
            </Button>
            <Button variant="outline" onClick={() => window.history.back()}>
              <X className="h-4 w-4 mr-2" />
              Cancel
            </Button>
          </div>
        </div>

        {/* Status Badge */}
        <div className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-yellow-100 text-yellow-800">
          NEW BILL
        </div>
      </div>

      {/* Main Form */}
      <div className="bg-white rounded-lg shadow-sm">
        <div className="p-6">
          {/* Vendor Information */}
          <div className="grid grid-cols-12 gap-4 mb-6">
            <div className="col-span-3">
              <Label htmlFor="vendorId">Vendor ID</Label>
              <div className="flex space-x-2">
                <Input
                  id="vendorId"
                  value={vendorId}
                  onChange={(e) => setVendorId(e.target.value)}
                  placeholder="Enter vendor ID"
                />
                <Button size="sm" variant="outline">
                  <Search className="h-4 w-4" />
                </Button>
              </div>
            </div>
            <div className="col-span-5">
              <Label htmlFor="vendorName">Vendor Name</Label>
              <Input
                id="vendorName"
                value={vendorName}
                onChange={(e) => setVendorName(e.target.value)}
                placeholder="Vendor name"
                disabled
                className="bg-gray-50"
              />
            </div>
            <div className="col-span-2">
              <Label htmlFor="postDate">Post Date</Label>
              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    className={cn(
                      "w-full justify-start text-left font-normal",
                      !postDate && "text-muted-foreground"
                    )}
                  >
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {postDate ? format(postDate, "MM/dd/yyyy") : "Select date"}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0">
                  <Calendar
                    mode="single"
                    selected={postDate}
                    onSelect={setPostDate}
                    initialFocus
                  />
                </PopoverContent>
              </Popover>
            </div>
            <div className="col-span-2 flex items-end">
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="approved"
                  checked={approvedToPay}
                  onCheckedChange={(checked) => setApprovedToPay(checked as boolean)}
                />
                <Label htmlFor="approved" className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                  Approved to Pay
                </Label>
              </div>
            </div>
          </div>

          {/* Invoice Information */}
          <div className="grid grid-cols-12 gap-4 mb-6">
            <div className="col-span-3">
              <Label htmlFor="invoiceNo">Invoice No</Label>
              <Input
                id="invoiceNo"
                value={invoiceNo}
                onChange={(e) => setInvoiceNo(e.target.value)}
                placeholder="Enter invoice number"
              />
            </div>
            <div className="col-span-3">
              <Label htmlFor="reference">Ref:</Label>
              <Input
                id="reference"
                value={reference}
                onChange={(e) => setReference(e.target.value)}
                placeholder="Reference"
              />
            </div>
            <div className="col-span-2">
              <Label htmlFor="terms">Terms</Label>
              <Select value={terms} onValueChange={setTerms}>
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
            </div>
            <div className="col-span-2">
              <Label htmlFor="invoiceDate">Invoice Date</Label>
              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    className={cn(
                      "w-full justify-start text-left font-normal",
                      !invoiceDate && "text-muted-foreground"
                    )}
                  >
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {invoiceDate ? format(invoiceDate, "MM/dd/yyyy") : "Select date"}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0">
                  <Calendar
                    mode="single"
                    selected={invoiceDate}
                    onSelect={setInvoiceDate}
                    initialFocus
                  />
                </PopoverContent>
              </Popover>
            </div>
            <div className="col-span-2 flex items-end">
              <Button
                variant="outline"
                className="w-full"
                onClick={() => setIsRecurring(!isRecurring)}
              >
                {isRecurring ? 'Recurring Bill ✓' : 'Recurring Bills'}
              </Button>
            </div>
          </div>

          {/* Due and Discount Dates */}
          <div className="grid grid-cols-12 gap-4 mb-6">
            <div className="col-span-3">
              <Label htmlFor="dueDate">Due Date</Label>
              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    className={cn(
                      "w-full justify-start text-left font-normal",
                      !dueDate && "text-muted-foreground"
                    )}
                  >
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {dueDate ? format(dueDate, "MM/dd/yyyy") : "//"}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0">
                  <Calendar
                    mode="single"
                    selected={dueDate}
                    onSelect={setDueDate}
                    initialFocus
                  />
                </PopoverContent>
              </Popover>
            </div>
            <div className="col-span-3">
              <Label htmlFor="discDate">Disc Date</Label>
              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    className={cn(
                      "w-full justify-start text-left font-normal",
                      !discDate && "text-muted-foreground"
                    )}
                  >
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {discDate ? format(discDate, "MM/dd/yyyy") : "//"}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0">
                  <Calendar
                    mode="single"
                    selected={discDate}
                    onSelect={setDiscDate}
                    initialFocus
                  />
                </PopoverContent>
              </Popover>
            </div>
          </div>
        </div>

        {/* Line Item Entry */}
        <div className="border-t border-gray-200 p-6">
          <div className="mb-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Line Items</h3>
            
            {/* Entry Form */}
            <div className="bg-gray-50 rounded-lg p-4 mb-4">
              <div className="grid grid-cols-12 gap-3">
                <div className="col-span-2">
                  <Label htmlFor="wellId" className="text-xs">Well ID</Label>
                  <div className="flex space-x-1">
                    <Input
                      id="wellId"
                      value={currentItem.wellId || ''}
                      onChange={(e) => setCurrentItem({...currentItem, wellId: e.target.value})}
                      placeholder="Well/Lease"
                      className="text-sm"
                    />
                    <Button size="sm" variant="outline" className="px-2">
                      <Search className="h-3 w-3" />
                    </Button>
                  </div>
                </div>
                <div className="col-span-1">
                  <Label htmlFor="expCode" className="text-xs">Exp Code</Label>
                  <Input
                    id="expCode"
                    value={currentItem.expCode || ''}
                    onChange={(e) => setCurrentItem({...currentItem, expCode: e.target.value})}
                    placeholder="DOI"
                    className="text-sm"
                  />
                </div>
                <div className="col-span-1">
                  <Label htmlFor="cls" className="text-xs">Class</Label>
                  <Select value={currentItem.cls || '0'} onValueChange={(value) => setCurrentItem({...currentItem, cls: value})}>
                    <SelectTrigger id="cls" className="text-sm">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="0">0</SelectItem>
                      <SelectItem value="1">1</SelectItem>
                      <SelectItem value="2">2</SelectItem>
                      <SelectItem value="P">P</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="col-span-1">
                  <Label htmlFor="deck" className="text-xs">Deck</Label>
                  <Input
                    id="deck"
                    value={currentItem.deck || ''}
                    onChange={(e) => setCurrentItem({...currentItem, deck: e.target.value})}
                    placeholder=""
                    className="text-sm"
                  />
                </div>
                <div className="col-span-3">
                  <Label htmlFor="description" className="text-xs">Description</Label>
                  <Input
                    id="description"
                    value={currentItem.description || ''}
                    onChange={(e) => setCurrentItem({...currentItem, description: e.target.value})}
                    placeholder="Enter description"
                    className="text-sm"
                  />
                </div>
                <div className="col-span-2">
                  <Label htmlFor="account" className="text-xs">Account</Label>
                  <div className="flex space-x-1">
                    <Input
                      id="account"
                      value={currentItem.account || ''}
                      onChange={(e) => setCurrentItem({...currentItem, account: e.target.value})}
                      placeholder="Account"
                      className="text-sm"
                    />
                    <Button size="sm" variant="outline" className="px-2">
                      <Search className="h-3 w-3" />
                    </Button>
                  </div>
                </div>
                <div className="col-span-1">
                  <Label htmlFor="afeNo" className="text-xs">AFE No</Label>
                  <Input
                    id="afeNo"
                    value={currentItem.afeNo || ''}
                    onChange={(e) => setCurrentItem({...currentItem, afeNo: e.target.value})}
                    placeholder=""
                    className="text-sm"
                  />
                </div>
                <div className="col-span-1">
                  <Label htmlFor="dept" className="text-xs">Dept</Label>
                  <Input
                    id="dept"
                    value={currentItem.dept || ''}
                    onChange={(e) => setCurrentItem({...currentItem, dept: e.target.value})}
                    placeholder=""
                    className="text-sm"
                  />
                </div>
              </div>
              
              <div className="grid grid-cols-12 gap-3 mt-3">
                <div className="col-span-2">
                  <Label htmlFor="prodPeriod" className="text-xs">Prod Period</Label>
                  <div className="flex space-x-2">
                    <Input
                      id="year"
                      value={currentItem.year || ''}
                      onChange={(e) => setCurrentItem({...currentItem, year: e.target.value})}
                      placeholder="Year"
                      className="text-sm w-20"
                    />
                    <span className="self-center">/</span>
                    <Input
                      id="period"
                      value={currentItem.period || ''}
                      onChange={(e) => setCurrentItem({...currentItem, period: e.target.value})}
                      placeholder="Period"
                      className="text-sm w-16"
                    />
                  </div>
                </div>
                <div className="col-span-2">
                  <Label htmlFor="allocateTo" className="text-xs">Allocate To</Label>
                  <Input
                    id="allocateTo"
                    value={currentItem.allocateTo || ''}
                    onChange={(e) => setCurrentItem({...currentItem, allocateTo: e.target.value})}
                    placeholder="Owner ID"
                    className="text-sm"
                  />
                </div>
                <div className="col-span-2">
                  <Label htmlFor="amount" className="text-xs">Amount</Label>
                  <div className="relative">
                    <DollarSign className="absolute left-2 top-2.5 h-4 w-4 text-gray-400" />
                    <Input
                      id="amount"
                      type="number"
                      step="0.01"
                      value={currentItem.amount || ''}
                      onChange={(e) => setCurrentItem({...currentItem, amount: parseFloat(e.target.value) || 0})}
                      placeholder="0.00"
                      className="text-sm pl-8 text-right"
                    />
                  </div>
                </div>
                <div className="col-span-2 flex items-end">
                  <Button 
                    onClick={handleAddLineItem}
                    className="w-full bg-green-600 hover:bg-green-700"
                  >
                    <Plus className="h-4 w-4 mr-2" />
                    Add Line
                  </Button>
                </div>
                <div className="col-span-4 flex items-end justify-end">
                  <Button variant="outline" size="sm">
                    Allocate All To
                  </Button>
                </div>
              </div>
            </div>

            {/* Line Items Table */}
            <div className="border rounded-lg overflow-hidden">
              <Table>
                <TableHeader className="bg-gray-50">
                  <TableRow>
                    <TableHead className="w-[100px]">Well ID</TableHead>
                    <TableHead className="w-[80px]">Exp Code</TableHead>
                    <TableHead className="w-[50px]">Cls</TableHead>
                    <TableHead className="w-[80px]">Deck</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead className="w-[120px]">Account</TableHead>
                    <TableHead className="w-[80px]">AFE No</TableHead>
                    <TableHead className="w-[60px]">Dept</TableHead>
                    <TableHead className="w-[60px]">Year</TableHead>
                    <TableHead className="w-[60px]">Period</TableHead>
                    <TableHead className="w-[100px]">Allocate To</TableHead>
                    <TableHead className="w-[100px] text-right">Amount</TableHead>
                    <TableHead className="w-[50px]"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {lineItems.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={13} className="text-center text-gray-500 py-8">
                        No line items added yet
                      </TableCell>
                    </TableRow>
                  ) : (
                    lineItems.map((item) => (
                      <TableRow key={item.id}>
                        <TableCell className="font-mono text-xs">{item.wellId}</TableCell>
                        <TableCell className="text-xs">{item.expCode}</TableCell>
                        <TableCell className="text-xs text-center">{item.cls}</TableCell>
                        <TableCell className="text-xs">{item.deck}</TableCell>
                        <TableCell className="text-sm">{item.description}</TableCell>
                        <TableCell className="font-mono text-xs">{item.account}</TableCell>
                        <TableCell className="text-xs">{item.afeNo}</TableCell>
                        <TableCell className="text-xs text-center">{item.dept}</TableCell>
                        <TableCell className="text-xs text-center">{item.year}</TableCell>
                        <TableCell className="text-xs text-center">{item.period}</TableCell>
                        <TableCell className="text-xs">{item.allocateTo}</TableCell>
                        <TableCell className="text-right font-mono text-sm">
                          ${item.amount.toFixed(2)}
                        </TableCell>
                        <TableCell>
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => handleDeleteLineItem(item.id)}
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
                <Button variant="outline" size="sm" onClick={() => handleDeleteLineItem(lineItems[lineItems.length - 1]?.id)}>
                  Delete Row
                </Button>
                <Button variant="outline" size="sm" onClick={handleReverseBill}>
                  <RotateCcw className="h-4 w-4 mr-2" />
                  Reverse Bill
                </Button>
                <Button variant="outline" size="sm" onClick={handleDuplicateBill}>
                  <Copy className="h-4 w-4 mr-2" />
                  Duplicate Bill
                </Button>
              </div>
              
              {/* Source Radio Options */}
              <div className="flex items-center space-x-4">
                <span className="text-sm font-medium">Source:</span>
                <RadioGroup value={billSource} onValueChange={setBillSource} className="flex space-x-4">
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
              </div>
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
                <span className="font-mono font-semibold">${invoiceBalance.toFixed(2)}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}