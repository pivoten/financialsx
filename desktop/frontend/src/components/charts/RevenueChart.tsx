
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'

const data = [
  { month: 'Jan', revenue: 45000, expenses: 28000 },
  { month: 'Feb', revenue: 52000, expenses: 31000 },
  { month: 'Mar', revenue: 48000, expenses: 29000 },
  { month: 'Apr', revenue: 61000, expenses: 33000 },
  { month: 'May', revenue: 55000, expenses: 32000 },
  { month: 'Jun', revenue: 67000, expenses: 35000 },
]

export function RevenueChart() {
  return (
    <ResponsiveContainer width="100%" height={300}>
      <LineChart data={data}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="month" />
        <YAxis tickFormatter={(value) => `$${(value / 1000).toFixed(0)}k`} />
        <Tooltip formatter={(value) => [`$${value.toLocaleString()}`, value === data[0]?.revenue ? 'Revenue' : 'Expenses']} />
        <Line 
          type="monotone" 
          dataKey="revenue" 
          stroke="hsl(var(--primary))" 
          strokeWidth={2}
          dot={{ fill: 'hsl(var(--primary))' }}
        />
        <Line 
          type="monotone" 
          dataKey="expenses" 
          stroke="hsl(var(--destructive))" 
          strokeWidth={2}
          dot={{ fill: 'hsl(var(--destructive))' }}
        />
      </LineChart>
    </ResponsiveContainer>
  )
}
