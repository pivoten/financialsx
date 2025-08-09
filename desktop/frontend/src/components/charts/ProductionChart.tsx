
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'

const data = [
  { well: 'Well-001', oil: 120, gas: 85 },
  { well: 'Well-002', oil: 98, gas: 67 },
  { well: 'Well-003', oil: 145, gas: 102 },
  { well: 'Well-004', oil: 87, gas: 58 },
  { well: 'Well-005', oil: 132, gas: 94 },
  { well: 'Well-006', oil: 156, gas: 111 },
]

export function ProductionChart() {
  return (
    <ResponsiveContainer width="100%" height={300}>
      <BarChart data={data}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="well" />
        <YAxis />
        <Tooltip formatter={(value, name) => [value, name === 'oil' ? 'Oil (bbl)' : 'Gas (mcf)']} />
        <Bar dataKey="oil" fill="hsl(var(--chart-1))" />
        <Bar dataKey="gas" fill="hsl(var(--chart-2))" />
      </BarChart>
    </ResponsiveContainer>
  )
}
