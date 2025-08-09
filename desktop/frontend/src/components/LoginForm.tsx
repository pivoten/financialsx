import React, { useState } from 'react'
import { Button } from './ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card'
import { Input } from './ui/input'
import pivotenLogo from '../assets/pivoten-logo.png'
import logger from '../services/logger'

interface LoginFormProps {
  onLogin: (e: React.FormEvent<HTMLFormElement>) => void
  onRegister: (e: React.FormEvent<HTMLFormElement>) => void
  username: string
  setUsername: (value: string) => void
  email: string
  setEmail: (value: string) => void
  password: string
  setPassword: (value: string) => void
  error: string
  isSubmitting: boolean
}

const LoginForm: React.FC<LoginFormProps> = ({
  onLogin,
  onRegister,
  username,
  setUsername,
  email,
  setEmail,
  password,
  setPassword,
  error,
  isSubmitting
}) => {
  const [showRegister, setShowRegister] = useState(false)

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    logger.debug('Form submitted', { showRegister })
    if (showRegister) {
      onRegister(e)
    } else {
      onLogin(e)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4">
            <img 
              src={pivotenLogo} 
              alt="Pivoten" 
              className="w-20 h-20 object-contain"
            />
          </div>
          <CardTitle className="text-2xl">Pivoten FinancialsX</CardTitle>
          <CardDescription>
            {showRegister ? 'Create your account' : 'Sign in to your account'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {/* Email or Username */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                {showRegister ? 'Email' : 'Email or Username'}
              </label>
              <Input
                type={showRegister ? 'email' : 'text'}
                value={showRegister ? email : username}
                onChange={(e) => {
                  if (showRegister) {
                    setEmail(e.target.value)
                  } else {
                    setUsername(e.target.value)
                  }
                }}
                placeholder={showRegister ? 'john@example.com' : 'Enter your email or username'}
                required
                disabled={isSubmitting}
              />
            </div>

            {/* Username for registration */}
            {showRegister && (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Username
                </label>
                <Input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="Choose a username"
                  required
                  disabled={isSubmitting}
                />
              </div>
            )}

            {/* Password */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Password
              </label>
              <Input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                required
                disabled={isSubmitting}
              />
            </div>

            {/* Error message */}
            {error && (
              <div className="text-sm text-red-600 bg-red-50 p-2 rounded">
                {error}
              </div>
            )}

            {/* Submit button */}
            <Button 
              type="submit" 
              className="w-full"
              disabled={isSubmitting}
            >
              {isSubmitting ? 'Please wait...' : (showRegister ? 'Create Account' : 'Sign In')}
            </Button>

            {/* Toggle between login and register */}
            <div className="text-center text-sm">
              <span className="text-gray-600">
                {showRegister ? 'Already have an account?' : "Don't have an account?"}
              </span>{' '}
              <button
                type="button"
                className="text-blue-600 hover:underline font-medium"
                onClick={() => setShowRegister(!showRegister)}
                disabled={isSubmitting}
              >
                {showRegister ? 'Sign in' : 'Create one'}
              </button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}

export default LoginForm