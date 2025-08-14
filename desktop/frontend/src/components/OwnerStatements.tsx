import React, { useState, useEffect } from 'react'
import { Button } from './ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card'
import { Alert, AlertDescription } from './ui/alert'
import { ArrowLeft, FileText, Download, RefreshCcw, AlertCircle, CheckCircle } from 'lucide-react'
import { CheckOwnerStatementFiles, GetOwnerStatementsList, GenerateOwnerStatementPDF } from '../../wailsjs/go/main/App'

interface OwnerStatementsProps {
  companyName: string
  onBack: () => void
}

const OwnerStatements: React.FC<OwnerStatementsProps> = ({ companyName, onBack }) => {
  const [loading, setLoading] = useState(true)
  const [hasFiles, setHasFiles] = useState(false)
  const [files, setFiles] = useState<any[]>([])
  const [error, setError] = useState<string>('')
  const [generating, setGenerating] = useState(false)
  const [generateMessage, setGenerateMessage] = useState<string>('')

  useEffect(() => {
    checkForFiles()
  }, [companyName])

  const checkForFiles = async () => {
    setLoading(true)
    setError('')
    
    try {
      // Check if owner statement files exist
      const result = await CheckOwnerStatementFiles(companyName)
      
      if (result.hasFiles) {
        setHasFiles(true)
        // Get the list of files with details
        const fileList = await GetOwnerStatementsList(companyName)
        setFiles(fileList || [])
      } else {
        setHasFiles(false)
        setError(result.error || 'No Owner Distribution Files Found')
      }
    } catch (err) {
      console.error('Error checking for owner statement files:', err)
      setError('Error accessing owner statement files')
    } finally {
      setLoading(false)
    }
  }

  const handleGeneratePDF = async (fileName: string) => {
    setGenerating(true)
    setGenerateMessage('')
    
    try {
      const result = await GenerateOwnerStatementPDF(companyName, fileName)
      setGenerateMessage(result)
    } catch (err) {
      console.error('Error generating PDF:', err)
      setGenerateMessage(`Error: ${err}`)
    } finally {
      setGenerating(false)
    }
  }

  return (
    <div className="space-y-6">
      {/* Header with Back Button */}
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="sm"
          onClick={onBack}
          className="hover:bg-gray-100"
        >
          <ArrowLeft className="h-4 w-4 mr-1" />
          Back
        </Button>
        <div className="flex-1">
          <h2 className="text-2xl font-bold text-gray-900">Owner Distribution Statements</h2>
          <p className="text-sm text-gray-500 mt-1">
            Generate PDF statements for owner distributions
          </p>
        </div>
      </div>

      {/* Main Content */}
      {loading ? (
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-center">
              <RefreshCcw className="h-6 w-6 animate-spin text-gray-400" />
              <span className="ml-2 text-gray-600">Checking for owner statement files...</span>
            </div>
          </CardContent>
        </Card>
      ) : !hasFiles ? (
        <Alert className="border-amber-200 bg-amber-50">
          <AlertCircle className="h-4 w-4 text-amber-600" />
          <AlertDescription className="text-amber-800">
            {error || 'No Owner Distribution Files Found'}
          </AlertDescription>
        </Alert>
      ) : (
        <>
          {/* File List */}
          <Card>
            <CardHeader>
              <CardTitle>Available Statement Files</CardTitle>
              <CardDescription>
                Select a file to generate PDF statements
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {files.map((file, index) => (
                  <div
                    key={index}
                    className="flex items-center justify-between p-4 border rounded-lg hover:bg-gray-50"
                  >
                    <div className="flex items-center gap-3">
                      <FileText className="h-5 w-5 text-blue-600" />
                      <div>
                        <p className="font-medium text-gray-900">{file.filename}</p>
                        <p className="text-sm text-gray-500">
                          Modified: {file.modified} | Size: {file.size} bytes
                          {file.hasFPT && ' | Has FPT'}
                        </p>
                      </div>
                    </div>
                    <Button
                      size="sm"
                      onClick={() => handleGeneratePDF(file.filename)}
                      disabled={generating}
                    >
                      <Download className="h-4 w-4 mr-1" />
                      Generate PDF
                    </Button>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>

          {/* Generate Message */}
          {generateMessage && (
            <Alert className={generateMessage.includes('Error') ? 'border-red-200 bg-red-50' : 'border-green-200 bg-green-50'}>
              {generateMessage.includes('Error') ? (
                <AlertCircle className="h-4 w-4 text-red-600" />
              ) : (
                <CheckCircle className="h-4 w-4 text-green-600" />
              )}
              <AlertDescription className={generateMessage.includes('Error') ? 'text-red-800' : 'text-green-800'}>
                {generateMessage}
              </AlertDescription>
            </Alert>
          )}

          {/* Refresh Button */}
          <div className="flex justify-end">
            <Button
              variant="outline"
              onClick={checkForFiles}
              disabled={loading}
            >
              <RefreshCcw className="h-4 w-4 mr-1" />
              Refresh Files
            </Button>
          </div>
        </>
      )}

      {/* Information Panel */}
      <Card className="bg-blue-50 border-blue-200">
        <CardHeader>
          <CardTitle className="text-blue-900">About Owner Distribution Statements</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-blue-800">
            Owner Distribution Statements are generated from DBF files located in the 
            <code className="mx-1 px-1 py-0.5 bg-blue-100 rounded">ownerstatements</code> 
            subdirectory of your company data folder. These statements provide detailed 
            distribution information for well owners and can be exported as PDF documents 
            for printing or electronic distribution.
          </p>
        </CardContent>
      </Card>
    </div>
  )
}

export default OwnerStatements