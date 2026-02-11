/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['var(--font-interface)', 'sans-serif'],
        mono: ['var(--font-text)', 'monospace'], // Using 'mono' for editor/text font
      },
      colors: {
        // Theme Backgrounds
        primary: 'var(--background-primary)',
        'primary-alt': 'var(--background-primary-alt)',
        secondary: 'var(--background-secondary)',
        'secondary-alt': 'var(--background-secondary-alt)',
        
        // Modifiers
        'modifier-hover': 'var(--background-modifier-hover)',
        'modifier-border': 'var(--background-modifier-border)',
        'modifier-border-focus': 'var(--background-modifier-border-focus)',
        
        // Text Colors
        normal: 'var(--text-normal)',
        muted: 'var(--text-muted)',
        faint: 'var(--text-faint)',
        // accent is reserved for text-accent in this context
        'text-accent': 'var(--text-accent)',
        
        // Obsidian Accents
        'obsidian-red': 'var(--color-red)',
        'obsidian-orange': 'var(--color-orange)',
        'obsidian-yellow': 'var(--color-yellow)',
        'obsidian-green': 'var(--color-green)',
        'obsidian-cyan': 'var(--color-cyan)',
        'obsidian-blue': 'var(--color-blue)',
        'obsidian-purple': 'var(--color-purple)',
        'obsidian-pink': 'var(--color-pink)',
      },
      animation: {
        'fade-in': 'fadeIn 0.2s ease-in-out',
        'slide-in': 'slideIn 0.2s ease-out',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideIn: {
          '0%': { transform: 'translateY(-10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
      },
      typography: (theme) => ({
        DEFAULT: {
          css: {
            color: 'var(--text-normal)',
            '[class~="lead"]': {
              color: 'var(--text-muted)',
            },
            a: {
              color: 'var(--color-blue)', // Match Editor link color
              '&:hover': {
                color: 'var(--text-accent)',
              },
            },
            strong: {
              color: 'var(--color-orange)', // Match Editor strong color
              fontWeight: '700',
            },
            'ol > li::marker': {
              color: 'var(--text-muted)',
            },
            'ul > li::marker': {
              color: 'var(--text-muted)',
            },
            hr: {
              borderColor: 'var(--background-modifier-border)',
            },
            blockquote: {
              color: 'var(--text-muted)',
              borderLeftColor: 'var(--background-modifier-border)',
            },
            h1: {
              color: 'var(--text-accent)',
              fontWeight: '700',
            },
            h2: {
              color: 'var(--text-accent)',
              fontWeight: '700',
            },
            h3: {
              color: 'var(--text-accent)',
              fontWeight: '600',
            },
            h4: {
              color: 'var(--text-accent)',
              fontWeight: '600',
            },
            code: {
              color: 'var(--color-pink)',
              backgroundColor: 'var(--background-primary-alt)',
              borderRadius: '0.25rem',
              paddingLeft: '0.25rem',
              paddingRight: '0.25rem',
              paddingTop: '0.125rem',
              paddingBottom: '0.125rem',
              fontWeight: '400',
            },
            'code::before': {
              content: '""',
            },
            'code::after': {
              content: '""',
            },
            pre: {
              backgroundColor: 'var(--background-primary-alt)',
              color: 'var(--text-normal)',
            },
            thead: {
              color: 'var(--text-normal)',
              borderBottomColor: 'var(--background-modifier-border)',
            },
            'tbody tr': {
              borderBottomColor: 'var(--background-modifier-border)',
            },
          },
        },
      }),
    },
  },
  plugins: [
    require('@tailwindcss/typography'),
  ],
}
