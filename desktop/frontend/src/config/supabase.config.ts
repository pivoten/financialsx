// Supabase Configuration
// Update these values with your Supabase project details

interface SupabaseFeatures {
  useSupabaseAuth: boolean
  enableSocialLogins: boolean
  enablePasswordReset: boolean
  enableEmailVerification: boolean
}

interface SupabaseConfig {
  url: string
  anonKey: string
  features: SupabaseFeatures
}

export const supabaseConfig: SupabaseConfig = {
  url: 'https://zzxndirkdzrvrqabhhfz.supabase.co',
  anonKey: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Inp6eG5kaXJrZHpydnJxYWJoaGZ6Iiwicm9sZSI6ImFub24iLCJpYXQiOjE3NDUyNjE4NzcsImV4cCI6MjA2MDgzNzg3N30.nWH7YNWczrNdC_bSuJWZKpsZvzCWKNHacDKwkGZ5rjY',
  
  // Optional: Feature flags
  features: {
    useSupabaseAuth: true,  // Set to false to use local SQLite auth
    enableSocialLogins: false,
    enablePasswordReset: true,
    enableEmailVerification: false
  }
}

// Example configuration (replace with your actual values):
// export const supabaseConfig = {
//   url: 'https://xyzcompany.supabase.co',
//   anonKey: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...',
//   features: {
//     useSupabaseAuth: true,
//     enableSocialLogins: false,
//     enablePasswordReset: true,
//     enableEmailVerification: false
//   }
// }