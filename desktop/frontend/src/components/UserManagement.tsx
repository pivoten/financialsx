
/**
 * USER MANAGEMENT COMPONENT - LOCAL SQLITE ONLY
 * 
 * IMPORTANT: This component manages users in the local SQLite database only.
 * It is NOT connected to Supabase authentication.
 * 
 * This component will be DEPRECATED once Supabase integration is complete.
 * It exists only for local testing and development purposes.
 * 
 * TODO: Remove this component once Supabase auth is fully integrated.
 *       All user management should be done through Supabase dashboard.
 */

import { useState, useEffect } from 'react'
import { GetAllUsers, GetAllRoles, UpdateUserRole, UpdateUserStatus, CreateUser } from '../../wailsjs/go/main/App'
import { Button } from './ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './ui/card'
import { Input } from './ui/input'
import { Table, TableHeader, TableRow, TableHead, TableBody, TableCell } from './ui/table'
import { Badge } from './ui/badge'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from './ui/dialog'
import { Label } from './ui/label'
import { UserPlus, Shield, Users, AlertCircle, CheckCircle, XCircle } from 'lucide-react'
import { User, Role } from '../types'

export function UserManagement({ currentUser }: { currentUser: User | null }) {
  const [users, setUsers] = useState<User[]>([])
  const [roles, setRoles] = useState<Role[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string>('')
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [updating, setUpdating] = useState<Record<string, boolean>>({})

  const [newUser, setNewUser] = useState({
    username: '',
    password: '',
    email: '',
    roleId: ''
  })

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    setError('')
    try {
      const [usersData, rolesData] = await Promise.all([
        GetAllUsers(),
        GetAllRoles()
      ])
      setUsers((usersData || []) as User[])
      setRoles((rolesData || []) as Role[])
    } catch (err) {
      setError(err.message || 'Failed to load user data')
    } finally {
      setLoading(false)
    }
  }

  const handleRoleChange = async (userId: number, newRoleId: string) => {
    const key = `role-${userId}`
    setUpdating(prev => ({ ...prev, [key]: true }))
    try {
      await UpdateUserRole(userId, parseInt(newRoleId))
      await loadData()
    } catch (err) {
      setError(err.message || 'Failed to update user role')
    } finally {
      setUpdating(prev => ({ ...prev, [key]: false }))
    }
  }

  const handleStatusToggle = async (userId: number, currentStatus: boolean) => {
    const key = `status-${userId}`
    setUpdating(prev => ({ ...prev, [key]: true }))
    try {
      await UpdateUserStatus(userId, !currentStatus)
      await loadData()
    } catch (err) {
      setError(err.message || 'Failed to update user status')
    } finally {
      setUpdating(prev => ({ ...prev, [key]: false }))
    }
  }

  const handleCreateUser = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    if (!newUser.username || !newUser.password || !newUser.email || !newUser.roleId) {
      setError('Please fill in all fields')
      return
    }
    try {
      await CreateUser(newUser.username, newUser.password, newUser.email, parseInt(newUser.roleId))
      setShowCreateDialog(false)
      setNewUser({ username: '', password: '', email: '', roleId: '' })
      await loadData()
    } catch (err) {
      setError(err.message || 'Failed to create user')
    }
  }

  const getRoleById = (roleId: number) => roles.find(role => role.id === roleId)

  const getRoleBadgeVariant = (roleName: string) => {
    switch (roleName) {
      case 'root': return 'destructive'
      case 'admin': return 'default'
      case 'readonly': return 'secondary'
      default: return 'outline'
    }
  }

  const canManageUser = (targetUser: User) => {
    if (!currentUser) return false
    if (currentUser.is_root) return !targetUser.is_root || targetUser.id === currentUser.id
    if (currentUser.role_name === 'admin') return targetUser.role_name === 'readonly' || targetUser.id === currentUser.id
    return false
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-lg">Loading users...</div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">User Management</h2>
          <p className="text-muted-foreground">Manage user accounts and permissions</p>
        </div>
        {currentUser?.permissions?.includes('users.create') && (
          <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
            <DialogTrigger asChild>
              <Button>
                <UserPlus className="w-4 h-4 mr-2" />
                Add User
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-md">
              <DialogHeader>
                <DialogTitle>Create New User</DialogTitle>
                <DialogDescription>Add a new user to the system with specified role and permissions.</DialogDescription>
              </DialogHeader>
              <form onSubmit={handleCreateUser} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="username">Username</Label>
                  <Input id="username" value={newUser.username} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewUser(prev => ({ ...prev, username: e.target.value }))} placeholder="Enter username" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="email">Email</Label>
                  <Input id="email" type="email" value={newUser.email} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewUser(prev => ({ ...prev, email: e.target.value }))} placeholder="Enter email address" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="password">Password</Label>
                  <Input id="password" type="password" value={newUser.password} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewUser(prev => ({ ...prev, password: e.target.value }))} placeholder="Enter password" />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="role">Role</Label>
                  <select 
                    id="role" 
                    value={newUser.roleId} 
                    onChange={(e: React.ChangeEvent<HTMLSelectElement>) => setNewUser(prev => ({ ...prev, roleId: e.target.value }))}
                    className="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm"
                  >
                    <option value="">Select a role</option>
                    {roles.map((role) => (
                      <option key={role.id} value={role.id.toString()}>
                        {role.role_name}
                      </option>
                    ))}
                  </select>
                </div>
                <DialogFooter>
                  <Button type="button" variant="outline" onClick={() => setShowCreateDialog(false)}>Cancel</Button>
                  <Button type="submit">Create User</Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>
        )}
      </div>

      {error && (
        <Card className="border-red-200 bg-red-50">
          <CardContent className="pt-6">
            <div className="flex items-center gap-2 text-red-600">
              <AlertCircle className="w-4 h-4" />
              {error}
            </div>
          </CardContent>
        </Card>
      )}

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-2">
              <Users className="w-5 h-5 text-blue-600" />
              <div>
                <p className="text-2xl font-bold">{users.length}</p>
                <p className="text-sm text-muted-foreground">Total Users</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-2">
              <CheckCircle className="w-5 h-5 text-green-600" />
              <div>
                <p className="text-2xl font-bold">{users.filter(u => u.is_active).length}</p>
                <p className="text-sm text-muted-foreground">Active Users</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-2">
              <Shield className="w-5 h-5 text-purple-600" />
              <div>
                <p className="text-2xl font-bold">{users.filter(u => u.role_name === 'admin' || u.is_root).length}</p>
                <p className="text-sm text-muted-foreground">Administrators</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>All Users</CardTitle>
          <CardDescription>Manage user accounts, roles, and permissions</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>User</TableHead>
                  <TableHead>Role</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead>Last Login</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {users.map((user) => {
                  const role = getRoleById(user.role_id)
                  const canManage = canManageUser(user)
                  return (
                    <TableRow key={user.id}>
                      <TableCell>
                        <div>
                          <div className="font-medium flex items-center gap-2">
                            {user.username}
                            {user.is_root && (
                              <Badge variant="destructive" className="text-xs">ROOT</Badge>
                            )}
                            {user.id === currentUser?.id && (
                              <Badge variant="outline" className="text-xs">YOU</Badge>
                            )}
                          </div>
                          <div className="text-sm text-muted-foreground">{user.email}</div>
                        </div>
                      </TableCell>
                      <TableCell>
                        {canManage && currentUser?.permissions?.includes('users.manage_roles') ? (
                          <select 
                            value={user.role_id?.toString() || ''} 
                            onChange={(e: React.ChangeEvent<HTMLSelectElement>) => handleRoleChange(user.id, e.target.value)}
                            disabled={updating[`role-${user.id}`]}
                            className="w-32 p-1 border rounded text-sm"
                          >
                            {roles.map((role) => (
                              <option key={role.id} value={role.id.toString()}>
                                {role.role_name}
                              </option>
                            ))}
                          </select>
                        ) : (
                          <Badge variant={getRoleBadgeVariant(user.role_name)}>
                            {role?.role_name || user.role_name}
                          </Badge>
                        )}
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          {user.is_active ? (
                            <CheckCircle className="w-4 h-4 text-green-600" />
                          ) : (
                            <XCircle className="w-4 h-4 text-red-600" />
                          )}
                          <span className={user.is_active ? 'text-green-600' : 'text-red-600'}>
                            {user.is_active ? 'Active' : 'Inactive'}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {new Date(user.created_at).toLocaleDateString()}
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {user.last_login ? new Date(user.last_login).toLocaleDateString() : 'Never'}
                      </TableCell>
                      <TableCell>
                        {canManage && currentUser?.permissions?.includes('users.update') && user.id !== currentUser?.id && (
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleStatusToggle(user.id, user.is_active)}
                            disabled={updating[`status-${user.id}`]}
                          >
                            {user.is_active ? 'Deactivate' : 'Activate'}
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Role Descriptions</CardTitle>
          <CardDescription>Understanding user roles and their permissions</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {roles.map((role) => (
              <div key={role.id} className="flex items-start gap-3 p-3 rounded-lg border bg-muted/50">
                <Shield className="w-5 h-5 text-muted-foreground mt-0.5" />
                <div>
                  <div className="flex items-center gap-2">
                    <h4 className="font-medium">{role.display_name}</h4>
                    <Badge variant={getRoleBadgeVariant(role.name)}>{role.name}</Badge>
                  </div>
                  <p className="text-sm text-muted-foreground mt-1">{role.description}</p>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
