import React, { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Switch } from './ui/switch'
import { Alert, AlertDescription } from './ui/alert'
import { Loader2, CheckCircle, XCircle, ExternalLink, ArrowLeft } from 'lucide-react'
import * as WailsApp from '../../wailsjs/go/main/App'

interface VFPSettings {
  host: string
  port: number
  enabled: boolean
  timeout: number
  updated_at?: string
}

interface LegacyIntegrationProps {
  onBack?: () => void
}

export default function LegacyIntegration({ onBack }: LegacyIntegrationProps) {
  const [settings, setSettings] = useState<VFPSettings>({
    host: 'localhost',
    port: 23456,
    enabled: false,
    timeout: 5
  })
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [saveSuccess, setSaveSuccess] = useState(false)

  // Load settings on mount
  useEffect(() => {
    loadSettings()
  }, [])

  const loadSettings = async () => {
    setLoading(true)
    setError(null)
    try {
      const result = await WailsApp.GetVFPSettings()
      if (result) {
        setSettings({
          host: result.host || 'localhost',
          port: result.port || 23456,
          enabled: result.enabled || false,
          timeout: result.timeout || 5,
          updated_at: result.updated_at
        })
      }
    } catch (err) {
      console.error('Failed to load VFP settings:', err)
      setError('Failed to load settings. Using defaults.')
    } finally {
      setLoading(false)
    }
  }

  const handleSave = async () => {
    setSaving(true)
    setError(null)
    setSaveSuccess(false)
    setTestResult(null)
    
    try {
      await WailsApp.SaveVFPSettings(
        settings.host,
        settings.port,
        settings.enabled,
        settings.timeout
      )
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    } catch (err) {
      console.error('Failed to save VFP settings:', err)
      setError('Failed to save settings. Please try again.')
    } finally {
      setSaving(false)
    }
  }

  const handleTest = async () => {
    setTesting(true)
    setTestResult(null)
    setError(null)
    
    try {
      const result = await WailsApp.TestVFPConnection()
      setTestResult(result)
      if (!result.success && result.message.includes('disabled')) {
        setError('Please enable the integration first')
      }
    } catch (err) {
      console.error('Failed to test VFP connection:', err)
      setTestResult({
        success: false,
        message: err.message || 'Connection test failed'
      })
    } finally {
      setTesting(false)
    }
  }

  const handleInputChange = (field: keyof VFPSettings, value: any) => {
    setSettings(prev => ({
      ...prev,
      [field]: value
    }))
    // Clear test result when settings change
    setTestResult(null)
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-gray-500" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header with back button */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          {onBack && (
            <Button
              variant="ghost"
              size="sm"
              onClick={onBack}
              className="hover:bg-gray-100"
            >
              <ArrowLeft className="h-4 w-4 mr-1" />
              Back
            </Button>
          )}
          <div>
            <h2 className="text-2xl font-semibold text-gray-900">Legacy Integration</h2>
            <p className="text-sm text-gray-500 mt-1">Configure Visual FoxPro integration settings</p>
          </div>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>VFP Connection Settings</CardTitle>
          <CardDescription>
            Configure the connection to your Visual FoxPro application's Winsock listener
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Enable/Disable Toggle */}
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="enabled">Enable Integration</Label>
              <p className="text-sm text-gray-500">
                Allow FinancialsX to communicate with Visual FoxPro
              </p>
            </div>
            <Switch
              id="enabled"
              checked={settings.enabled}
              onCheckedChange={(checked) => handleInputChange('enabled', checked)}
            />
          </div>

          {/* Connection Settings */}
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="host">Host Address</Label>
              <Input
                id="host"
                type="text"
                value={settings.host}
                onChange={(e) => handleInputChange('host', e.target.value)}
                placeholder="localhost"
                disabled={!settings.enabled}
              />
              <p className="text-xs text-gray-500">
                Usually 'localhost' for local VFP application
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="port">Port Number</Label>
              <Input
                id="port"
                type="number"
                value={settings.port}
                onChange={(e) => handleInputChange('port', parseInt(e.target.value) || 23456)}
                placeholder="23456"
                min="1"
                max="65535"
                disabled={!settings.enabled}
              />
              <p className="text-xs text-gray-500">
                Default port is 23456
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="timeout">Connection Timeout (seconds)</Label>
              <Input
                id="timeout"
                type="number"
                value={settings.timeout}
                onChange={(e) => handleInputChange('timeout', parseInt(e.target.value) || 5)}
                placeholder="5"
                min="1"
                max="30"
                disabled={!settings.enabled}
              />
              <p className="text-xs text-gray-500">
                How long to wait for VFP to respond
              </p>
            </div>
          </div>

          {/* Status Messages */}
          {error && (
            <Alert className="border-red-200 bg-red-50">
              <XCircle className="h-4 w-4 text-red-600" />
              <AlertDescription className="text-red-800">
                {error}
              </AlertDescription>
            </Alert>
          )}

          {saveSuccess && (
            <Alert className="border-green-200 bg-green-50">
              <CheckCircle className="h-4 w-4 text-green-600" />
              <AlertDescription className="text-green-800">
                Settings saved successfully
              </AlertDescription>
            </Alert>
          )}

          {testResult && (
            <Alert className={testResult.success ? "border-green-200 bg-green-50" : "border-yellow-200 bg-yellow-50"}>
              {testResult.success ? (
                <CheckCircle className="h-4 w-4 text-green-600" />
              ) : (
                <XCircle className="h-4 w-4 text-yellow-600" />
              )}
              <AlertDescription className={testResult.success ? "text-green-800" : "text-yellow-800"}>
                {testResult.message}
              </AlertDescription>
            </Alert>
          )}

          {/* Action Buttons */}
          <div className="flex items-center justify-between pt-4">
            <div className="text-sm text-gray-500">
              {settings.updated_at && (
                <span>Last updated: {new Date(settings.updated_at).toLocaleString()}</span>
              )}
            </div>
            <div className="flex space-x-3">
              <Button
                variant="outline"
                onClick={handleTest}
                disabled={testing || saving || !settings.enabled}
              >
                {testing ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Testing...
                  </>
                ) : (
                  'Test Connection'
                )}
              </Button>
              <Button
                onClick={handleSave}
                disabled={saving || testing}
              >
                {saving ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Saving...
                  </>
                ) : (
                  'Save Settings'
                )}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Information Card */}
      <Card>
        <CardHeader>
          <CardTitle>How It Works</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <h4 className="font-medium text-sm">Prerequisites</h4>
            <ul className="text-sm text-gray-600 space-y-1 list-disc list-inside">
              <li>Visual FoxPro 9 SP2 with Winsock listener form running</li>
              <li>Microsoft Winsock Control 6.0 ActiveX installed</li>
              <li>VFP application configured to listen on the specified port</li>
            </ul>
          </div>
          
          <div className="space-y-2">
            <h4 className="font-medium text-sm">Available Features</h4>
            <ul className="text-sm text-gray-600 space-y-1 list-disc list-inside">
              <li>Launch VFP forms directly from FinancialsX</li>
              <li>Pass parameters to VFP forms (customer ID, invoice number, etc.)</li>
              <li>Seamless navigation between modern and legacy interfaces</li>
            </ul>
          </div>

          <div className="pt-2">
            <Button variant="outline" size="sm" className="text-blue-600 hover:text-blue-700">
              <ExternalLink className="h-4 w-4 mr-2" />
              View Integration Guide
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}