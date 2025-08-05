import { useTheme } from './theme-provider'
import { Button } from './ui/button'
import { Moon, Sun, Palette } from 'lucide-react'

export function ThemeSwitcher() {
  const { theme, colorScheme, setTheme, setColorScheme } = useTheme()

  const themes = [
    { value: 'default', label: 'Default', color: 'bg-slate-900' },
    { value: 'blue', label: 'Blue', color: 'bg-blue-600' },
    { value: 'green', label: 'Green', color: 'bg-green-600' },
    { value: 'violet', label: 'Violet', color: 'bg-violet-600' },
    { value: 'rose', label: 'Rose', color: 'bg-rose-600' },
    { value: 'orange', label: 'Orange', color: 'bg-orange-600' },
    { value: 'yellow', label: 'Yellow', color: 'bg-yellow-500' },
  ]

  return (
    <div className="flex items-center gap-2">
      {/* Dark/Light Mode Toggle */}
      <Button
        variant="ghost"
        size="sm"
        onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}
        className="h-8 w-8 p-0"
      >
        {theme === 'light' ? (
          <Moon className="h-4 w-4" />
        ) : (
          <Sun className="h-4 w-4" />
        )}
        <span className="sr-only">Toggle theme</span>
      </Button>

      {/* Color Scheme Selector */}
      <div className="flex items-center gap-1 px-2 py-1 rounded-md border bg-background">
        <Palette className="h-3 w-3 text-muted-foreground" />
        <div className="flex gap-1">
          {themes.map((colorTheme) => (
            <button
              key={colorTheme.value}
              onClick={() => setColorScheme(colorTheme.value)}
              className={`w-4 h-4 rounded-full border-2 transition-all ${
                colorScheme === colorTheme.value
                  ? 'border-foreground scale-110'
                  : 'border-muted-foreground/20 hover:border-muted-foreground/40'
              } ${colorTheme.color}`}
              title={colorTheme.label}
            />
          ))}
        </div>
      </div>
    </div>
  )
}