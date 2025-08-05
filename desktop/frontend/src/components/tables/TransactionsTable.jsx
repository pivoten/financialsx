import { useState } from 'react'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../ui/table'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Search, Filter } from 'lucide-react'

const transactions = [
  { id: 1, date: '2024-07-30', description: 'Oil Sales - Well 001', amount: 15420.50, type: 'Income', status: 'Completed' },
  { id: 2, date: '2024-07-30', description: 'Equipment Maintenance', amount: -2850.00, type: 'Expense', status: 'Completed' },
  { id: 3, date: '2024-07-29', description: 'Gas Sales - Well 003', amount: 8760.25, type: 'Income', status: 'Completed' },
  { id: 4, date: '2024-07-29', description: 'Truck Fuel', amount: -450.75, type: 'Expense', status: 'Pending' },
  { id: 5, date: '2024-07-28', description: 'Oil Sales - Well 002', amount: 12350.00, type: 'Income', status: 'Completed' },
  { id: 6, date: '2024-07-28', description: 'Contractor Services', amount: -5200.00, type: 'Expense', status: 'Completed' },
  { id: 7, date: '2024-07-27', description: 'Royalty Payment', amount: -3250.50, type: 'Expense', status: 'Completed' },
  { id: 8, date: '2024-07-27', description: 'Gas Sales - Well 004', amount: 6890.75, type: 'Income', status: 'Completed' },
]

export function TransactionsTable() {
  const [searchTerm, setSearchTerm] = useState('')
  const [filter, setFilter] = useState('all')

  const filteredTransactions = transactions.filter(transaction => {
    const matchesSearch = transaction.description.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesFilter = filter === 'all' || transaction.type.toLowerCase() === filter
    return matchesSearch && matchesFilter
  })

  const formatCurrency = (amount) => {
    const isNegative = amount < 0
    const formatted = Math.abs(amount).toLocaleString('en-US', { 
      style: 'currency', 
      currency: 'USD' 
    })
    return isNegative ? `-${formatted}` : formatted
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search transactions..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-8"
          />
        </div>
        <div className="flex gap-2">
          <Button
            variant={filter === 'all' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setFilter('all')}
          >
            All
          </Button>
          <Button
            variant={filter === 'income' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setFilter('income')}
          >
            Income
          </Button>
          <Button
            variant={filter === 'expense' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setFilter('expense')}
          >
            Expenses
          </Button>
        </div>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Date</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Type</TableHead>
              <TableHead className="text-right">Amount</TableHead>
              <TableHead>Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredTransactions.map((transaction) => (
              <TableRow key={transaction.id}>
                <TableCell className="font-medium">
                  {new Date(transaction.date).toLocaleDateString()}
                </TableCell>
                <TableCell>{transaction.description}</TableCell>
                <TableCell>
                  <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                    transaction.type === 'Income' 
                      ? 'bg-green-100 text-green-800' 
                      : 'bg-red-100 text-red-800'
                  }`}>
                    {transaction.type}
                  </span>
                </TableCell>
                <TableCell className={`text-right font-medium ${
                  transaction.amount > 0 ? 'text-green-600' : 'text-red-600'
                }`}>
                  {formatCurrency(transaction.amount)}
                </TableCell>
                <TableCell>
                  <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                    transaction.status === 'Completed' 
                      ? 'bg-blue-100 text-blue-800' 
                      : 'bg-yellow-100 text-yellow-800'
                  }`}>
                    {transaction.status}
                  </span>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}