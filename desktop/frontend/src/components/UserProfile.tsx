import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from './ui/card';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Avatar, AvatarFallback, AvatarImage } from './ui/avatar';
import { Tabs, TabsContent, TabsList, TabsTrigger } from './ui/tabs';
import { 
  User, 
  Mail, 
  Building2, 
  Shield, 
  Key, 
  Save, 
  Camera,
  Bell,
  Lock,
  Palette,
  Globe
} from 'lucide-react';

interface UserProfileProps {
  currentUser: any;
  companyName?: string;
}

export default function UserProfile({ currentUser, companyName }: UserProfileProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [profileData, setProfileData] = useState({
    username: currentUser?.username || '',
    email: currentUser?.email || '',
    fullName: currentUser?.full_name || '',
    phone: currentUser?.phone || '',
    department: currentUser?.department || '',
    title: currentUser?.title || ''
  });

  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map(word => word[0])
      .join('')
      .toUpperCase()
      .slice(0, 2);
  };

  const handleSave = async () => {
    // TODO: Implement save to Supabase
    console.log('Saving profile:', profileData);
    setIsEditing(false);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="bg-white rounded-lg shadow-sm p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <User className="h-8 w-8 text-blue-600" />
            <div>
              <h1 className="text-2xl font-bold text-gray-900">User Profile</h1>
              <p className="text-sm text-gray-500">Manage your account settings and preferences</p>
            </div>
          </div>
        </div>
      </div>

      {/* Profile Card */}
      <div className="bg-white rounded-lg shadow-sm">
        <div className="p-6">
          <div className="flex items-start space-x-6">
            {/* Avatar Section */}
            <div className="flex flex-col items-center space-y-3">
              <Avatar className="h-24 w-24">
                <AvatarImage src={currentUser?.avatar_url} />
                <AvatarFallback className="bg-blue-100 text-blue-600 text-xl font-semibold">
                  {getInitials(profileData.fullName || profileData.username)}
                </AvatarFallback>
              </Avatar>
              <Button variant="outline" size="sm" className="text-xs">
                <Camera className="h-3 w-3 mr-1" />
                Change Photo
              </Button>
            </div>

            {/* Profile Info */}
            <div className="flex-1">
              <div className="flex items-center justify-between mb-6">
                <div>
                  <h2 className="text-xl font-semibold text-gray-900">
                    {profileData.fullName || profileData.username}
                  </h2>
                  <div className="flex items-center space-x-4 text-sm text-gray-500 mt-1">
                    <span className="flex items-center">
                      <Mail className="h-3 w-3 mr-1" />
                      {profileData.email}
                    </span>
                    <span className="flex items-center">
                      <Building2 className="h-3 w-3 mr-1" />
                      {companyName}
                    </span>
                    <span className="flex items-center">
                      <Shield className="h-3 w-3 mr-1" />
                      {currentUser?.role_name || 'User'}
                    </span>
                  </div>
                </div>
                <Button
                  variant={isEditing ? "default" : "outline"}
                  onClick={() => isEditing ? handleSave() : setIsEditing(true)}
                >
                  {isEditing ? (
                    <>
                      <Save className="h-4 w-4 mr-2" />
                      Save Changes
                    </>
                  ) : (
                    'Edit Profile'
                  )}
                </Button>
              </div>

              {/* Tabs */}
              <Tabs defaultValue="personal" className="w-full">
                <TabsList className="grid w-full grid-cols-4">
                  <TabsTrigger value="personal">Personal</TabsTrigger>
                  <TabsTrigger value="security">Security</TabsTrigger>
                  <TabsTrigger value="notifications">Notifications</TabsTrigger>
                  <TabsTrigger value="preferences">Preferences</TabsTrigger>
                </TabsList>

                {/* Personal Tab */}
                <TabsContent value="personal" className="space-y-4 mt-6">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="username">Username</Label>
                      <Input
                        id="username"
                        value={profileData.username}
                        onChange={(e) => setProfileData({...profileData, username: e.target.value})}
                        disabled={!isEditing}
                        className={!isEditing ? 'bg-gray-50' : ''}
                      />
                    </div>
                    <div>
                      <Label htmlFor="email">Email Address</Label>
                      <Input
                        id="email"
                        type="email"
                        value={profileData.email}
                        onChange={(e) => setProfileData({...profileData, email: e.target.value})}
                        disabled={!isEditing}
                        className={!isEditing ? 'bg-gray-50' : ''}
                      />
                    </div>
                    <div>
                      <Label htmlFor="fullName">Full Name</Label>
                      <Input
                        id="fullName"
                        value={profileData.fullName}
                        onChange={(e) => setProfileData({...profileData, fullName: e.target.value})}
                        disabled={!isEditing}
                        className={!isEditing ? 'bg-gray-50' : ''}
                        placeholder="Enter your full name"
                      />
                    </div>
                    <div>
                      <Label htmlFor="phone">Phone Number</Label>
                      <Input
                        id="phone"
                        value={profileData.phone}
                        onChange={(e) => setProfileData({...profileData, phone: e.target.value})}
                        disabled={!isEditing}
                        className={!isEditing ? 'bg-gray-50' : ''}
                        placeholder="(555) 123-4567"
                      />
                    </div>
                    <div>
                      <Label htmlFor="department">Department</Label>
                      <Input
                        id="department"
                        value={profileData.department}
                        onChange={(e) => setProfileData({...profileData, department: e.target.value})}
                        disabled={!isEditing}
                        className={!isEditing ? 'bg-gray-50' : ''}
                        placeholder="e.g., Accounting, Operations"
                      />
                    </div>
                    <div>
                      <Label htmlFor="title">Job Title</Label>
                      <Input
                        id="title"
                        value={profileData.title}
                        onChange={(e) => setProfileData({...profileData, title: e.target.value})}
                        disabled={!isEditing}
                        className={!isEditing ? 'bg-gray-50' : ''}
                        placeholder="e.g., Revenue Accountant"
                      />
                    </div>
                  </div>
                </TabsContent>

                {/* Security Tab */}
                <TabsContent value="security" className="space-y-4 mt-6">
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base flex items-center">
                        <Key className="h-4 w-4 mr-2" />
                        Change Password
                      </CardTitle>
                      <CardDescription>
                        Update your password to keep your account secure
                      </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-3">
                      <div>
                        <Label htmlFor="currentPassword">Current Password</Label>
                        <Input id="currentPassword" type="password" disabled={!isEditing} />
                      </div>
                      <div>
                        <Label htmlFor="newPassword">New Password</Label>
                        <Input id="newPassword" type="password" disabled={!isEditing} />
                      </div>
                      <div>
                        <Label htmlFor="confirmPassword">Confirm New Password</Label>
                        <Input id="confirmPassword" type="password" disabled={!isEditing} />
                      </div>
                      <Button disabled={!isEditing} className="w-full">
                        <Lock className="h-4 w-4 mr-2" />
                        Update Password
                      </Button>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base">Two-Factor Authentication</CardTitle>
                      <CardDescription>
                        Add an extra layer of security to your account
                      </CardDescription>
                    </CardHeader>
                    <CardContent>
                      <Button variant="outline" className="w-full">
                        Enable Two-Factor Authentication
                      </Button>
                    </CardContent>
                  </Card>
                </TabsContent>

                {/* Notifications Tab */}
                <TabsContent value="notifications" className="space-y-4 mt-6">
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base flex items-center">
                        <Bell className="h-4 w-4 mr-2" />
                        Email Notifications
                      </CardTitle>
                      <CardDescription>
                        Choose what notifications you want to receive
                      </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="font-medium text-sm">Bill Approvals</p>
                          <p className="text-xs text-gray-500">Receive notifications when bills need approval</p>
                        </div>
                        <input type="checkbox" className="h-4 w-4" defaultChecked />
                      </div>
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="font-medium text-sm">Payment Reminders</p>
                          <p className="text-xs text-gray-500">Get reminded about upcoming payments</p>
                        </div>
                        <input type="checkbox" className="h-4 w-4" defaultChecked />
                      </div>
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="font-medium text-sm">Reconciliation Alerts</p>
                          <p className="text-xs text-gray-500">Notifications for bank reconciliation issues</p>
                        </div>
                        <input type="checkbox" className="h-4 w-4" />
                      </div>
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="font-medium text-sm">System Updates</p>
                          <p className="text-xs text-gray-500">Important system and feature updates</p>
                        </div>
                        <input type="checkbox" className="h-4 w-4" />
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>

                {/* Preferences Tab */}
                <TabsContent value="preferences" className="space-y-4 mt-6">
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base flex items-center">
                        <Palette className="h-4 w-4 mr-2" />
                        Display Preferences
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-4">
                      <div>
                        <Label htmlFor="theme">Theme</Label>
                        <select 
                          id="theme" 
                          className="w-full px-3 py-2 border rounded-md"
                          disabled={!isEditing}
                        >
                          <option value="light">Light</option>
                          <option value="dark">Dark</option>
                          <option value="system">System</option>
                        </select>
                      </div>
                      <div>
                        <Label htmlFor="dateFormat">Date Format</Label>
                        <select 
                          id="dateFormat" 
                          className="w-full px-3 py-2 border rounded-md"
                          disabled={!isEditing}
                        >
                          <option value="MM/DD/YYYY">MM/DD/YYYY</option>
                          <option value="DD/MM/YYYY">DD/MM/YYYY</option>
                          <option value="YYYY-MM-DD">YYYY-MM-DD</option>
                        </select>
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base flex items-center">
                        <Globe className="h-4 w-4 mr-2" />
                        Regional Settings
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-4">
                      <div>
                        <Label htmlFor="timezone">Timezone</Label>
                        <select 
                          id="timezone" 
                          className="w-full px-3 py-2 border rounded-md"
                          disabled={!isEditing}
                        >
                          <option value="America/Chicago">Central Time (CT)</option>
                          <option value="America/New_York">Eastern Time (ET)</option>
                          <option value="America/Denver">Mountain Time (MT)</option>
                          <option value="America/Los_Angeles">Pacific Time (PT)</option>
                        </select>
                      </div>
                      <div>
                        <Label htmlFor="currency">Currency</Label>
                        <select 
                          id="currency" 
                          className="w-full px-3 py-2 border rounded-md"
                          disabled={!isEditing}
                        >
                          <option value="USD">USD ($)</option>
                          <option value="CAD">CAD ($)</option>
                          <option value="EUR">EUR (€)</option>
                        </select>
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
              </Tabs>
            </div>
          </div>
        </div>
      </div>

      {/* Account Info Card */}
      <div className="bg-white rounded-lg shadow-sm p-6">
        <h3 className="text-lg font-semibold mb-4">Account Information</h3>
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-gray-500">Account ID:</span>
            <span className="ml-2 font-medium">{currentUser?.id || 'N/A'}</span>
          </div>
          <div>
            <span className="text-gray-500">Role:</span>
            <span className="ml-2 font-medium">{currentUser?.role_name || 'User'}</span>
          </div>
          <div>
            <span className="text-gray-500">Company:</span>
            <span className="ml-2 font-medium">{companyName || 'N/A'}</span>
          </div>
          <div>
            <span className="text-gray-500">Status:</span>
            <span className="ml-2 font-medium text-green-600">Active</span>
          </div>
          <div>
            <span className="text-gray-500">Member Since:</span>
            <span className="ml-2 font-medium">
              {currentUser?.created_at ? new Date(currentUser.created_at).toLocaleDateString() : 'N/A'}
            </span>
          </div>
          <div>
            <span className="text-gray-500">Last Login:</span>
            <span className="ml-2 font-medium">
              {currentUser?.last_login ? new Date(currentUser.last_login).toLocaleDateString() : 'Today'}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}