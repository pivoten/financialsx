import React, { useState } from 'react'
import { Button } from './ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card'
import { Checkbox } from './ui/checkbox'
import { Label } from './ui/label'
import logger from '../services/logger'
import { InitializeLogging, SetDebugMode, GetDebugMode, LogMessage } from '../../wailsjs/go/main/App'

const LoggingTest: React.FC = () => {
  const [debugMode, setDebugModeState] = useState(false)
  const [logCount, setLogCount] = useState(0)
  const [status, setStatus] = useState('')

  React.useEffect(() => {
    // Check current debug mode on mount
    checkDebugMode()
  }, [])

  const checkDebugMode = async () => {
    try {
      const isDebug = await GetDebugMode()
      setDebugModeState(isDebug)
      setStatus(`Debug mode is ${isDebug ? 'enabled' : 'disabled'}`)
    } catch (error) {
      setStatus('Error checking debug mode')
    }
  }

  const handleToggleDebug = async (checked: boolean) => {
    try {
      const result = await SetDebugMode(checked)
      if (result.success) {
        setDebugModeState(checked)
        logger.setDebugMode(checked)
        setStatus(`Debug mode ${checked ? 'enabled' : 'disabled'} successfully`)
      } else {
        setStatus(`Failed to set debug mode: ${result.error}`)
      }
    } catch (error) {
      setStatus(`Error: ${error}`)
    }
  }

  const testFrontendLogging = () => {
    const testNum = logCount + 1
    setLogCount(testNum)
    
    // Test different log levels
    logger.debug(`Frontend debug test #${testNum}`, { component: 'LoggingTest', timestamp: new Date().toISOString() })
    logger.info(`Frontend info test #${testNum}`, { testData: 'some info' })
    logger.warn(`Frontend warning test #${testNum}`, { warningCode: 'TEST_WARN' })
    logger.error(`Frontend error test #${testNum}`, { errorCode: 'TEST_ERROR' })
    
    setStatus(`Sent 4 frontend log messages (test #${testNum})`)
  }

  const testBackendLogging = async () => {
    const testNum = logCount + 1
    setLogCount(testNum)
    
    try {
      // Test different log levels via backend
      await LogMessage('DEBUG', `Backend debug test #${testNum}`, 'LoggingTest', { test: true })
      await LogMessage('INFO', `Backend info test #${testNum}`, 'LoggingTest', { test: true })
      await LogMessage('WARN', `Backend warning test #${testNum}`, 'LoggingTest', { test: true })
      await LogMessage('ERROR', `Backend error test #${testNum}`, 'LoggingTest', { test: true })
      
      setStatus(`Sent 4 backend log messages (test #${testNum})`)
    } catch (error) {
      setStatus(`Error sending backend logs: ${error}`)
    }
  }

  const testLargeLogEntry = () => {
    const largeData = {
      users: Array(100).fill(null).map((_, i) => ({
        id: i,
        name: `User ${i}`,
        email: `user${i}@example.com`,
        metadata: { created: new Date().toISOString(), active: true }
      })),
      timestamp: new Date().toISOString(),
      description: 'This is a large log entry to test file rotation and size limits'
    }
    
    logger.debug('Large data log test', largeData)
    setStatus('Sent large log entry')
  }

  return (
    <Card className="w-full max-w-2xl mx-auto">
      <CardHeader>
        <CardTitle>Logging System Test</CardTitle>
        <CardDescription>Test the file-based logging system with debug mode</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center space-x-2">
          <Checkbox
            id="debug-mode"
            checked={debugMode}
            onCheckedChange={handleToggleDebug}
          />
          <Label htmlFor="debug-mode" className="cursor-pointer">Debug Mode</Label>
        </div>

        <div className="space-y-2">
          <Button onClick={testFrontendLogging} className="w-full">
            Test Frontend Logging
          </Button>
          <Button onClick={testBackendLogging} variant="secondary" className="w-full">
            Test Backend Logging
          </Button>
          <Button onClick={testLargeLogEntry} variant="outline" className="w-full">
            Test Large Log Entry
          </Button>
        </div>

        {status && (
          <div className="p-3 bg-muted rounded-md">
            <p className="text-sm">{status}</p>
          </div>
        )}

        <div className="text-xs text-muted-foreground">
          <p>When debug mode is enabled, logs are written to:</p>
          <p className="font-mono">~/.financialsx/logs/financialsx_YYYY-MM-DD.log</p>
          <p className="mt-2">Log files rotate when they exceed 10MB</p>
        </div>
      </CardContent>
    </Card>
  )
}

export default LoggingTest