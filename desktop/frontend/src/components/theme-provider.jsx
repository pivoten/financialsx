import { createContext, useContext, useEffect, useState } from 'react'

const ThemeProviderContext = createContext({
  theme: 'light',
  colorScheme: 'default',
  setTheme: () => null,
  setColorScheme: () => null,
})

export function ThemeProvider({
  children,
  defaultTheme = 'light',
  defaultColorScheme = 'default',
  storageKey = 'ui-theme',
  colorStorageKey = 'ui-color-scheme',
  ...props
}) {
  const [theme, setTheme] = useState(defaultTheme)
  const [colorScheme, setColorScheme] = useState(defaultColorScheme)

  useEffect(() => {
    const storedTheme = localStorage.getItem(storageKey)
    const storedColorScheme = localStorage.getItem(colorStorageKey)
    
    if (storedTheme) {
      setTheme(storedTheme)
    }
    if (storedColorScheme) {
      setColorScheme(storedColorScheme)
    }
  }, [storageKey, colorStorageKey])

  useEffect(() => {
    const root = window.document.documentElement

    // Remove existing theme classes
    root.classList.remove('light', 'dark')
    root.classList.remove('theme-default', 'theme-blue', 'theme-green', 'theme-violet', 'theme-rose', 'theme-orange', 'theme-yellow')

    // Add new theme classes
    root.classList.add(theme)
    if (colorScheme !== 'default') {
      root.classList.add(`theme-${colorScheme}`)
    }

    // Store in localStorage
    localStorage.setItem(storageKey, theme)
    localStorage.setItem(colorStorageKey, colorScheme)
  }, [theme, colorScheme, storageKey, colorStorageKey])

  const value = {
    theme,
    colorScheme,
    setTheme: (theme) => {
      setTheme(theme)
    },
    setColorScheme: (colorScheme) => {
      setColorScheme(colorScheme)
    },
  }

  return (
    <ThemeProviderContext.Provider {...props} value={value}>
      {children}
    </ThemeProviderContext.Provider>
  )
}

export const useTheme = () => {
  const context = useContext(ThemeProviderContext)

  if (context === undefined)
    throw new Error('useTheme must be used within a ThemeProvider')

  return context
}