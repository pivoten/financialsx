import { createClient } from '@supabase/supabase-js'
import { supabaseConfig } from '../config/supabase.config'

// Create a single supabase client for interacting with your database
// Only create if Supabase auth is enabled and credentials are provided
console.log('Supabase config:', {
  url: supabaseConfig.url,
  hasKey: !!supabaseConfig.anonKey,
  useSupabaseAuth: supabaseConfig.features.useSupabaseAuth
})

export const supabase = supabaseConfig.features.useSupabaseAuth && 
                        supabaseConfig.url !== 'YOUR_SUPABASE_PROJECT_URL'
  ? createClient(supabaseConfig.url, supabaseConfig.anonKey)
  : null

// Check if Supabase is configured
export const isSupabaseConfigured = () => {
  const configured = supabase !== null
  console.log('Is Supabase configured?', configured)
  return configured
}

// Helper functions for auth
export const signUp = async (email, password, metadata = {}) => {
  const { data, error } = await supabase.auth.signUp({
    email,
    password,
    options: {
      data: metadata
    }
  })
  return { data, error }
}

export const signIn = async (email, password) => {
  const { data, error } = await supabase.auth.signInWithPassword({
    email,
    password
  })
  return { data, error }
}

export const signOut = async () => {
  const { error } = await supabase.auth.signOut()
  return { error }
}

export const getCurrentUser = async () => {
  const { data: { user }, error } = await supabase.auth.getUser()
  return { user, error }
}

export const getSession = async () => {
  const { data: { session }, error } = await supabase.auth.getSession()
  return { session, error }
}

// Subscribe to auth state changes
export const onAuthStateChange = (callback) => {
  return supabase.auth.onAuthStateChange(callback)
}