import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    rollupOptions: {
      // Don't externalize supabase - include it in the bundle
      external: []
    }
  },
  optimizeDeps: {
    include: ['@supabase/supabase-js']
  }
})
