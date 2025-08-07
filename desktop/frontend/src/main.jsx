console.log('=== MAIN.JSX STARTING ===')

import React from 'react'
import {createRoot} from 'react-dom/client'
import './style.css'
import App from './App'

console.log('Imports complete, looking for root element...')

const container = document.getElementById('root')
console.log('Root element found:', !!container)

const root = createRoot(container)

console.log('About to render App component...')

root.render(
    <React.StrictMode>
        <App/>
    </React.StrictMode>
)

console.log('=== APP RENDERED ===')
