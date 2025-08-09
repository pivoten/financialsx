import React from 'react'
import { createRoot } from 'react-dom/client'
import './globals.css'
import './i18n' // Initialize i18n
import App from './App'

const container = document.getElementById('root') as HTMLElement
const root = createRoot(container)

root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
)
