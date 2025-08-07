# Supabase Authentication Setup for FinancialsX Desktop

## Overview
This desktop application can use Supabase for cloud-based authentication while maintaining local data operations through the Go backend. This allows for centralized user management with Row Level Security (RLS) while keeping sensitive financial data local.

## Architecture
- **Frontend (React)**: Handles authentication directly with Supabase
- **Backend (Go/Wails)**: Manages local DBF files and SQLite operations
- **Hybrid Approach**: Cloud auth + local data = best of both worlds

## Setup Instructions

### 1. Configure Your Existing Supabase Project

Since you already have a Supabase project, you just need to update the configuration file.

Edit `frontend/src/config/supabase.config.js`:
```javascript
export const supabaseConfig = {
  url: 'https://your-project.supabase.co',  // Your Supabase URL
  anonKey: 'your-anon-key-here',           // Your Supabase anon key
  
  features: {
    useSupabaseAuth: true,     // Enable Supabase auth
    enableSocialLogins: false,
    enablePasswordReset: true,
    enableEmailVerification: false
  }
}
```

### 2. How It Works

#### Frontend Authentication Flow:
1. User enters email/password on login screen
2. Frontend calls Supabase directly for authentication
3. Supabase returns JWT token and user data
4. Token is stored in React context and used for subsequent API calls

#### Backend Integration:
1. Frontend can pass JWT token to Go backend functions
2. Go backend can optionally validate the JWT
3. Local operations (DBF files, SQLite) continue to work normally

### 3. User Data Storage

With Supabase auth enabled:
- **User credentials**: Stored in Supabase
- **User profile/metadata**: Stored in Supabase user metadata
- **Company data**: Remains local in DBF files
- **Reconciliations**: Stored in local SQLite

### 4. Authentication Modes

The app supports dual authentication:

#### Supabase Mode (when configured):
- Users authenticate against Supabase
- Centralized user management
- Support for password reset, email verification
- Can add social logins later

#### Local Mode (fallback):
- Uses SQLite for user storage
- Works offline/air-gapped
- No external dependencies

### 5. API Call Pattern

When making authenticated API calls from frontend to Go backend:

```javascript
// Without authentication wrapper (current)
const data = await GetBankAccounts(companyName)

// With authentication wrapper (when needed for RLS)
import { useAuthenticatedAPI } from '../hooks/useAuthenticatedAPI'

const { callAPI } = useAuthenticatedAPI()
const data = await callAPI(GetBankAccounts, companyName)
```

### 6. Row Level Security (RLS)

If you want to use Supabase for data storage later:
1. Users are automatically associated with companies via metadata
2. RLS policies can restrict data access based on user's company
3. JWT token contains user info for authorization

### 7. Building for Distribution

When building the desktop app for distribution:
1. Update `supabase.config.js` with production Supabase credentials
2. Build the app: `wails build`
3. The compiled app will include the Supabase configuration
4. Each client installation will connect to the same Supabase project

### 8. Security Considerations

- **Anon Key**: Safe to include in frontend (protected by RLS)
- **Service Key**: Never include in frontend code
- **JWT Tokens**: Automatically refreshed by Supabase client
- **Local Data**: Remains secure on user's machine

## Testing the Integration

### 1. Test Supabase Connection:
```javascript
// In browser console (with app running)
import { supabase, isSupabaseConfigured } from './lib/supabase'
console.log('Supabase configured:', isSupabaseConfigured())
```

### 2. Test Authentication:
1. Set `useSupabaseAuth: true` in config
2. Run the app: `wails dev`
3. Try to register a new user
4. Check Supabase dashboard for new user

### 3. Test Fallback:
1. Set `useSupabaseAuth: false` in config
2. App should use local SQLite auth
3. No external calls to Supabase

## Benefits of This Approach

1. **Centralized User Management**: Manage all users from Supabase dashboard
2. **Local Data Security**: Financial data never leaves the client's machine
3. **Scalability**: Can serve unlimited desktop clients with one Supabase project
4. **Flexibility**: Can gradually migrate features to cloud as needed
5. **Offline Capability**: App can work offline with cached credentials

## Troubleshooting

### "Supabase is not configured"
- Check that URL and anon key are set correctly in config
- Ensure `useSupabaseAuth` is set to `true`

### "Authentication failed"
- Verify user exists in Supabase
- Check network connection
- Ensure Supabase project is not paused

### "Company mismatch"
- User's company metadata should match selected company
- Check user's metadata in Supabase dashboard

## Next Steps

1. Update the configuration with your Supabase credentials
2. Test the authentication flow
3. Optionally add JWT validation in Go backend
4. Consider adding social login providers
5. Implement password reset flow if needed

## Optional: JWT Validation in Go

If you want the Go backend to validate Supabase JWTs:

```go
// Install JWT library
go get github.com/golang-jwt/jwt/v5

// Validate token in Go
func ValidateSupabaseToken(tokenString string) (*jwt.Token, error) {
    // Parse and validate JWT
    // Check claims for user info and permissions
}
```

This would add an extra security layer for sensitive operations.