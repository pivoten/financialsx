import React, { useState } from 'react'
import { DashboardCard } from './DashboardCard'
import ChartOfAccountsReport from './ChartOfAccountsReport'
import OwnerStatements from './OwnerStatements'
import { FileText, TrendingUp, BarChart3, DollarSign, Receipt, BookOpen, Users } from 'lucide-react'

interface FinancialReportsProps {
  companyName: string
  currentUser?: any
}

const FinancialReports: React.FC<FinancialReportsProps> = ({ companyName, currentUser }) => {
  const [selectedReport, setSelectedReport] = useState<string | null>(null)

  const handleBack = () => {
    setSelectedReport(null)
  }

  if (selectedReport === 'chart-of-accounts') {
    return <ChartOfAccountsReport companyName={companyName} onBack={handleBack} />
  }
  
  if (selectedReport === 'owner-statements') {
    return <OwnerStatements companyName={companyName} onBack={handleBack} />
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h2 className="text-2xl font-bold text-gray-900">Financial Reports</h2>
        <p className="text-sm text-gray-500 mt-1">
          Generate financial statements and accounting reports
        </p>
      </div>

      {/* Report Cards Grid */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        <DashboardCard
          title="Chart of Accounts"
          subtitle="General Ledger"
          description="Complete listing of all GL accounts with sorting options"
          icon={BookOpen}
          onClick={() => setSelectedReport('chart-of-accounts')}
          accentColor="blue"
        />
        
        <DashboardCard
          title="Owner Distributions"
          subtitle="Statements"
          description="Generate owner distribution statements from DBF files"
          icon={Users}
          onClick={() => setSelectedReport('owner-statements')}
          accentColor="teal"
        />
        
        <DashboardCard
          title="Income Statement"
          subtitle="P&L Report"
          description="Revenue and expense summary for period"
          icon={TrendingUp}
          onClick={() => setSelectedReport('income-statement')}
          accentColor="green"
          disabled={true}
        />
        
        <DashboardCard
          title="Balance Sheet"
          subtitle="Financial Position"
          description="Assets, liabilities, and equity snapshot"
          icon={BarChart3}
          onClick={() => setSelectedReport('balance-sheet')}
          accentColor="purple"
          disabled={true}
        />
        
        <DashboardCard
          title="Cash Flow"
          subtitle="Cash Analysis"
          description="Operating, investing, and financing activities"
          icon={DollarSign}
          onClick={() => setSelectedReport('cash-flow')}
          accentColor="emerald"
          disabled={true}
        />
        
        <DashboardCard
          title="Trial Balance"
          subtitle="Account Balances"
          description="Debit and credit balances by account"
          icon={FileText}
          onClick={() => setSelectedReport('trial-balance')}
          accentColor="orange"
          disabled={true}
        />
        
        <DashboardCard
          title="General Ledger"
          subtitle="Transaction Detail"
          description="Detailed transaction listing by account"
          icon={Receipt}
          onClick={() => setSelectedReport('general-ledger')}
          accentColor="indigo"
          disabled={true}
        />
      </div>

      {/* Coming Soon Notice for disabled reports */}
      <div className="mt-8 p-4 bg-gray-50 rounded-lg border border-gray-200">
        <p className="text-sm text-gray-600">
          <span className="font-semibold">Note:</span> Additional financial reports are coming soon. 
          The Chart of Accounts report is currently available with PDF export functionality.
        </p>
      </div>
    </div>
  )
}

export default FinancialReports