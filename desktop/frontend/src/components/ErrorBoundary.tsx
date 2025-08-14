
import React from 'react'
import { Card, CardHeader, CardTitle, CardContent } from './ui/card'
import { AlertCircle } from 'lucide-react'
import logger from '../services/logger'

class ErrorBoundary extends React.Component<any, any> {
  constructor(props: any) {
    super(props)
    this.state = { hasError: false, error: null, errorInfo: null }
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true }
  }

  componentDidCatch(error: Error, errorInfo: any) {
    logger.error('Error caught by boundary', { error: error.message, errorInfo })
    this.setState({ error, errorInfo })
    try {
      if ((window as any).LogError) {
        ;(window as any).LogError(error.toString(), errorInfo.componentStack)
      }
    } catch (e) {
      logger.error('Failed to log to backend', { error: e.message })
    }
  }

  render() {
    if (this.state.hasError) {
      return (
        <Card className="m-4 border-red-200 bg-red-50">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-red-800">
              <AlertCircle className="w-5 h-5" />
              Something went wrong
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-red-700 mb-2">
              An error occurred in the {this.props.name || 'application'}. 
              The error has been logged.
            </p>
            {this.state.error && (
              <details className="mt-4">
                <summary className="cursor-pointer text-sm text-red-600 hover:text-red-800">
                  Show error details
                </summary>
                <pre className="mt-2 p-3 bg-red-100 rounded text-xs overflow-auto text-red-800">
                  {this.state.error.toString()}
                  {this.state.errorInfo && this.state.errorInfo.componentStack}
                </pre>
              </details>
            )}
            <button
              onClick={() => window.location.reload()}
              className="mt-4 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
            >
              Reload Application
            </button>
          </CardContent>
        </Card>
      )
    }

    return this.props.children
  }
}

export default ErrorBoundary
