import { createContext, useContext, useEffect, useState, ReactNode } from 'react'

type ThemeProviderValue = {
  theme: string
  colorScheme: string
  setTheme: (t: string) => void
  setColorScheme: (c: string) => void
}

const ThemeProviderContext = createContext<ThemeProviderValue>({
  theme: 'light',
  colorScheme: 'default',
  setTheme: () => null,
  setColorScheme: () => null,
})

type ThemeProviderProps = {
  children: ReactNode
  defaultTheme?: string
  defaultColorScheme?: string
  storageKey?: string
  colorStorageKey?: string
}

export function ThemeProvider({
  children,
  defaultTheme = 'light',
  defaultColorScheme = 'default',
  storageKey = 'ui-theme',
  colorStorageKey = 'ui-color-scheme',
}: ThemeProviderProps) {
  const [theme, setTheme] = useState(defaultTheme)
  const [colorScheme, setColorScheme] = useState(defaultColorScheme)

  useEffect(() => {
    const storedTheme = localStorage.getItem(storageKey)
    const storedColorScheme = localStorage.getItem(colorStorageKey)
    
    if (storedTheme) setTheme(storedTheme)
    if (storedColorScheme) setColorScheme(storedColorScheme)
  }, [storageKey, colorStorageKey])

  useEffect(() => {
    const root = window.document.documentElement
    root.classList.remove('light', 'dark')
    root.classList.remove('theme-default', 'theme-blue', 'theme-green', 'theme-violet', 'theme-rose', 'theme-orange', 'theme-yellow')
    root.classList.add(theme)
    if (colorScheme !== 'default') root.classList.add(`theme-${colorScheme}`)
    localStorage.setItem(storageKey, theme)
    localStorage.setItem(colorStorageKey, colorScheme)
  }, [theme, colorScheme, storageKey, colorStorageKey])

  const value: ThemeProviderValue = {
    theme,
    colorScheme,
    setTheme: (t) => setTheme(t),
    setColorScheme: (c) => setColorScheme(c),
  }

  return (
  <ThemeProviderContext.Provider value={value}>
      {children}
    </ThemeProviderContext.Provider>
  )
}

export const useTheme = (): ThemeProviderValue => {
  const context = useContext(ThemeProviderContext)
  if (context === undefined) throw new Error('useTheme must be used within a ThemeProvider')
  return context
}
