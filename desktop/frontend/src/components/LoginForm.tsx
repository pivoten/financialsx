import React, { useState } from 'react'
import { Button } from './ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { User, Mail, Lock, Eye, EyeOff, AlertCircle } from 'lucide-react'
import pivotenLogo from '../assets/pivoten-logo.png'
import logger from '../services/logger'

interface LoginFormProps {
  onLogin: (e: React.FormEvent<HTMLFormElement>) => void
  onRequestLogin: () => void
  onResetPassword: () => void
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
  onRequestLogin,
  onResetPassword,
  username,
  setUsername,
  email,
  setEmail,
  password,
  setPassword,
  error,
  isSubmitting
}) => {
  const [showPassword, setShowPassword] = useState(false)

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    logger.debug('Login form submitted')
    onLogin(e)
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-background via-muted to-background flex items-center justify-center p-6">
      <Card className="w-full max-w-md shadow-lg border-border/60">
        <CardHeader className="text-center">
          <div className="mx-auto mb-3">
            <img
              src={pivotenLogo}
              alt="Pivoten"
              style={{ width: '60px', height: '60px', objectFit: 'contain' }}
            />
          </div>
          <CardTitle className="text-2xl tracking-tight">Pivoten FinancialsX</CardTitle>
          <CardDescription>
            Sign in to access your financial data
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-5">
            <div className="space-y-2">
              <Label htmlFor="auth-identifier">Email</Label>
              <div className="relative">
                <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">
                  <Mail className="h-4 w-4" />
                </span>
                <Input
                  id="auth-identifier"
                  type="email"
                  value={email || username}
                  onChange={(e) => {
                    setEmail(e.target.value)
                    setUsername(e.target.value)
                  }}
                  placeholder="name@company.com"
                  autoComplete="email"
                  required
                  disabled={isSubmitting}
                  className="w-full pl-10"
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="auth-password">Password</Label>
              <div className="relative">
                <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">
                  <Lock className="h-4 w-4" />
                </span>
                <Input
                  id="auth-password"
                  type={showPassword ? 'text' : 'password'}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••"
                  autoComplete="current-password"
                  required
                  disabled={isSubmitting}
                  className="w-full pl-10 pr-10"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword((v) => !v)}
                  aria-label={showPassword ? 'Hide password' : 'Show password'}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                  disabled={isSubmitting}
                >
                  {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
            </div>

            {error && (
              <div className="flex items-start gap-2 rounded-md border border-destructive/30 bg-destructive/10 p-2 text-sm text-destructive">
                <AlertCircle className="mt-0.5 h-4 w-4" />
                <span>{error}</span>
              </div>
            )}

            <Button
              type="submit"
              className="w-full"
              disabled={isSubmitting}
            >
              {isSubmitting ? 'Please wait...' : 'Sign In'}
            </Button>

            <div className="flex justify-between text-sm">
              <button
                type="button"
                className="text-primary hover:underline font-medium"
                onClick={onResetPassword}
                disabled={isSubmitting}
              >
                Reset Password
              </button>
              <button
                type="button"
                className="text-primary hover:underline font-medium"
                onClick={onRequestLogin}
                disabled={isSubmitting}
              >
                Request Login
              </button>
            </div>

            <div className="text-center text-xs text-muted-foreground">Powered by Pivoten</div>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}

export default LoginForm