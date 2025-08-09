import React, { useState } from 'react'
import { Login } from '../wailsjs/go/main/App'
import LoginForm from './components/LoginForm'
import { Card, CardContent, CardHeader, CardTitle } from './components/ui/card'
import logger from './services/logger'

interface User {
  id: number
  username: string
  email: string
  role_name: string
  is_root: boolean
  company_name: string
}

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [currentUser, setCurrentUser] = useState<User | null>(null)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleLogin = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setError('')
    setIsSubmitting(true)
    
    try {
      logger.debug('Login attempt', { username })
      // For now, just use local auth with a default company
      const user = await Login(username, password, 'apollo')
      setCurrentUser(user as User)
      setIsAuthenticated(true)
    } catch (err: any) {
      setError(err.message || 'Login failed')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleRegister = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setError('Registration not implemented yet')
  }

  const handleLogout = () => {
    setIsAuthenticated(false)
    setCurrentUser(null)
    setUsername('')
    setPassword('')
    setEmail('')
  }

  if (!isAuthenticated) {
    return (
      <LoginForm
        onLogin={handleLogin}
        onRegister={handleRegister}
        username={username}
        setUsername={setUsername}
        email={email}
        setEmail={setEmail}
        password={password}
        setPassword={setPassword}
        error={error}
        isSubmitting={isSubmitting}
      />
    )
  }

  return (
    <div className="min-h-screen bg-gray-100 p-8">
      <Card className="max-w-4xl mx-auto">
        <CardHeader>
          <CardTitle>Welcome to Pivoten FinancialsX</CardTitle>
        </CardHeader>
        <CardContent>
          <p>Logged in as: {currentUser?.username}</p>
          <p>Company: {currentUser?.company_name}</p>
          <button 
            onClick={handleLogout}
            className="mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
          >
            Logout
          </button>
        </CardContent>
      </Card>
    </div>
  )
}

export default App