import { createClient, SupabaseClient, User, Session, AuthChangeEvent } from '@supabase/supabase-js'
import { supabaseConfig } from '../config/supabase.config'
import logger from '../services/logger'

// Create a single supabase client for interacting with your database
// Only create if Supabase auth is enabled and credentials are provided
logger.debug('Supabase config loaded', {
  url: supabaseConfig.url,
  hasKey: !!supabaseConfig.anonKey,
  useSupabaseAuth: supabaseConfig.features.useSupabaseAuth
})

export const supabase: SupabaseClient | null = supabaseConfig.features.useSupabaseAuth && 
                        supabaseConfig.url !== 'YOUR_SUPABASE_PROJECT_URL'
  ? createClient(supabaseConfig.url, supabaseConfig.anonKey)
  : null

// Check if Supabase is configured
export const isSupabaseConfigured = (): boolean => {
  const configured = supabase !== null
  logger.debug('Supabase configuration check', { configured })
  return configured
}

// Type definitions for auth functions
interface AuthResponse<T> {
  data: T | null
  error: Error | null
}

interface SignUpMetadata {
  [key: string]: any
}

// Helper functions for auth
export const signUp = async (
  email: string, 
  password: string, 
  metadata: SignUpMetadata = {}
): Promise<AuthResponse<{ user: User | null; session: Session | null }>> => {
  if (!supabase) {
    return { data: null, error: new Error('Supabase not configured') }
  }
  
  const { data, error } = await supabase.auth.signUp({
    email,
    password,
    options: {
      data: metadata
    }
  })
  return { data, error }
}

export const signIn = async (
  email: string, 
  password: string
): Promise<AuthResponse<{ user: User | null; session: Session | null }>> => {
  if (!supabase) {
    return { data: null, error: new Error('Supabase not configured') }
  }
  
  const { data, error } = await supabase.auth.signInWithPassword({
    email,
    password
  })
  return { data, error }
}

export const signOut = async (): Promise<{ error: Error | null }> => {
  if (!supabase) {
    return { error: new Error('Supabase not configured') }
  }
  
  const { error } = await supabase.auth.signOut()
  return { error }
}

export const getCurrentUser = async (): Promise<{ user: User | null; error: Error | null }> => {
  if (!supabase) {
    return { user: null, error: new Error('Supabase not configured') }
  }
  
  const { data: { user }, error } = await supabase.auth.getUser()
  return { user, error }
}

export const getSession = async (): Promise<{ session: Session | null; error: Error | null }> => {
  if (!supabase) {
    return { session: null, error: new Error('Supabase not configured') }
  }
  
  const { data: { session }, error } = await supabase.auth.getSession()
  return { session, error }
}

// Fetch user account data from Supabase
export const getUserAccountData = async (): Promise<{ account: any | null; error: Error | null }> => {
  if (!supabase) {
    return { account: null, error: new Error('Supabase not configured') }
  }
  
  try {
    // Get current user
    const { data: { user }, error: userError } = await supabase.auth.getUser()
    if (userError || !user) {
      return { account: null, error: userError || new Error('No user logged in') }
    }
    
    // Fetch user's account data from the user_accounts view
    const { data, error } = await supabase
      .from('user_accounts')
      .select('id, name, picture_url, slug, role')
      .eq('is_personal_account', false)
      .limit(1)
      .single()
    
    if (error) {
      logger.error('Failed to fetch user account data', error)
      // Try fetching from accounts table directly if view doesn't exist
      const { data: accountData, error: accountError } = await supabase
        .from('accounts')
        .select('id, name, picture_url, slug')
        .eq('is_personal_account', false)
        .limit(1)
        .single()
      
      if (accountError) {
        return { account: null, error: accountError }
      }
      return { account: accountData, error: null }
    }
    
    return { account: data, error: null }
  } catch (err) {
    logger.error('Error fetching user account data', err)
    return { account: null, error: err as Error }
  }
}

// Subscribe to auth state changes
export const onAuthStateChange = (
  callback: (event: AuthChangeEvent, session: Session | null) => void
) => {
  if (!supabase) {
    logger.warn('Supabase not configured, auth state changes not available')
    return null
  }
  
  return supabase.auth.onAuthStateChange(callback)
}