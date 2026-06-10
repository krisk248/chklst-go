/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        background: '#2b2b2b',
        foreground: '#ffffff',
        success: '#4caf50',
        muted: '#404040',
        border: '#555555',
        surface: {
          DEFAULT: '#353535',
          light: '#404040',
          lighter: '#4a4a4a',
          border: '#555555',
          deep: '#2f2f2f',
          deeper: '#2b2b2b',
          deepest: '#2a2a2a',
        },
        accent: {
          DEFAULT: '#4a9eff',
        },
      },
      fontSize: {
        xs: ['0.75rem', { lineHeight: '1rem' }],
        sm: ['0.875rem', { lineHeight: '1.25rem' }],
        base: ['1rem', { lineHeight: '1.5rem' }],
        lg: ['1.125rem', { lineHeight: '1.75rem' }],
        xl: ['1.25rem', { lineHeight: '1.75rem' }],
        '2xl': ['1.5rem', { lineHeight: '2rem' }],
      },
    },
  },
  plugins: [],
}
