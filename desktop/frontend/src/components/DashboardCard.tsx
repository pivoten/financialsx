import React from 'react'
import { Card, CardContent } from './ui/card'
import { ChevronRight } from 'lucide-react'

interface DashboardCardProps {
  title: string
  subtitle: string
  description: string
  icon: React.ComponentType<{ className?: string }>
  onClick: () => void
  accentColor?: string
}

export const DashboardCard: React.FC<DashboardCardProps> = ({
  title,
  subtitle,
  description,
  icon: Icon,
  onClick,
  accentColor = 'orange'
}) => {
  // Define color classes for different accent colors
  const colorClasses = {
    orange: {
      icon: 'text-orange-600 bg-orange-50',
      hover: 'hover:border-orange-200',
      arrow: 'text-orange-600'
    },
    blue: {
      icon: 'text-blue-600 bg-blue-50',
      hover: 'hover:border-blue-200',
      arrow: 'text-blue-600'
    },
    green: {
      icon: 'text-green-600 bg-green-50',
      hover: 'hover:border-green-200',
      arrow: 'text-green-600'
    },
    purple: {
      icon: 'text-purple-600 bg-purple-50',
      hover: 'hover:border-purple-200',
      arrow: 'text-purple-600'
    },
    gray: {
      icon: 'text-gray-600 bg-gray-50',
      hover: 'hover:border-gray-300',
      arrow: 'text-gray-600'
    }
  }

  const colors = colorClasses[accentColor as keyof typeof colorClasses] || colorClasses.orange

  return (
    <Card 
      className={`cursor-pointer transition-all duration-200 hover:shadow-lg hover:-translate-y-0.5 border border-gray-200 bg-white ${colors.hover}`}
      onClick={onClick}
    >
      <CardContent className="p-6 pt-8">
        <div className="flex items-start justify-between mb-4">
          <div className="flex-1">
            <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">
              {subtitle}
            </p>
            <h3 className="text-xl font-bold text-gray-900 mb-3">{title}</h3>
            <p className="text-sm text-gray-600 leading-relaxed">{description}</p>
          </div>
          <div className={`p-3 rounded-lg ${colors.icon}`}>
            <Icon className="w-6 h-6" />
          </div>
        </div>
        <div className={`flex items-center ${colors.arrow}`}>
          <span className="text-xs font-medium">Open</span>
          <ChevronRight className="w-4 h-4 ml-1" />
        </div>
      </CardContent>
    </Card>
  )
}