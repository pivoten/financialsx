import 'react'

declare module 'react' {
  // Allow arbitrary CSS variables in style prop (e.g., { '--tw-ring-color': '...' })
  // This augments React.CSSProperties with an index signature for CSS custom properties.
  interface CSSProperties {
    [key: `--${string}`]: string | number | undefined
  }
}

export {}
