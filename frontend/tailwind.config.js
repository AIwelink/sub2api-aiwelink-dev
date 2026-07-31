const token = (name, fallback) =>
  `rgb(var(--color-${name}, ${fallback}) / <alpha-value>)`

/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        primary: {
          50: token('primary-50', '255 241 243'),
          100: token('primary-100', '255 228 232'),
          200: token('primary-200', '253 203 212'),
          300: token('primary-300', '249 168 184'),
          400: token('primary-400', '232 108 132'),
          500: token('primary-500', '186 54 80'),
          600: token('primary-600', '165 45 69'),
          700: token('primary-700', '135 38 59'),
          800: token('primary-800', '112 35 55'),
          900: token('primary-900', '96 33 51'),
          950: token('primary-950', '53 14 25')
        },
        accent: {
          50: token('accent-50', '255 249 235'),
          100: token('accent-100', '255 240 199'),
          200: token('accent-200', '255 225 138'),
          300: token('accent-300', '247 213 143'),
          400: token('accent-400', '240 202 135'),
          500: token('accent-500', '233 190 115'),
          600: token('accent-600', '213 163 82'),
          700: token('accent-700', '185 129 49'),
          800: token('accent-800', '147 97 35'),
          900: token('accent-900', '101 63 32'),
          950: token('accent-950', '58 32 14')
        },
        dark: {
          50: token('ink-50', '248 250 250'),
          100: token('ink-100', '237 241 241'),
          200: token('ink-200', '220 228 230'),
          300: token('ink-300', '189 201 205'),
          400: token('ink-400', '154 172 181'),
          500: token('ink-500', '113 134 144'),
          600: token('ink-600', '75 98 109'),
          700: token('ink-700', '38 55 64'),
          800: token('ink-800', '13 29 39'),
          900: token('ink-900', '11 24 35'),
          950: token('ink-950', '7 18 26')
        },
        canvas: token('canvas', '245 247 248'),
        surface: token('surface', '255 255 255'),
        'surface-muted': token('surface-muted', '240 243 244'),
        'theme-border': token('theme-border', '228 232 234'),
        'theme-text': token('theme-text', '37 51 59'),
        'theme-muted': token('theme-muted', '101 114 122'),
        'on-primary': token('on-primary', '255 255 255')
      },
      fontFamily: {
        sans: [
          'system-ui',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'Roboto',
          'Helvetica Neue',
          'Arial',
          'PingFang SC',
          'Hiragino Sans GB',
          'Microsoft YaHei',
          'sans-serif'
        ],
        mono: ['ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'monospace']
      },
      boxShadow: {
        glass: '0 8px 32px rgba(0, 0, 0, 0.08)',
        'glass-sm': '0 4px 16px rgba(0, 0, 0, 0.06)',
        glow: '0 0 20px rgb(var(--color-primary-500, 186 54 80) / 0.25)',
        'glow-lg': '0 0 40px rgb(var(--color-primary-500, 186 54 80) / 0.35)',
        card: '0 1px 3px rgba(0, 0, 0, 0.04), 0 1px 2px rgba(0, 0, 0, 0.06)',
        'card-hover': '0 10px 40px rgba(0, 0, 0, 0.08)',
        'inner-glow': 'inset 0 1px 0 rgba(255, 255, 255, 0.1)'
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-primary':
          'linear-gradient(135deg, rgb(var(--color-primary-500, 186 54 80)) 0%, rgb(var(--color-primary-600, 165 45 69)) 100%)',
        'gradient-dark':
          'linear-gradient(135deg, rgb(var(--color-ink-800, 13 29 39)) 0%, rgb(var(--color-ink-900, 11 24 35)) 100%)',
        'gradient-glass':
          'linear-gradient(135deg, rgba(255,255,255,0.1) 0%, rgba(255,255,255,0.05) 100%)',
        'mesh-gradient':
          'radial-gradient(at 40% 20%, rgb(var(--color-primary-500, 186 54 80) / 0.1) 0px, transparent 50%), radial-gradient(at 80% 0%, rgb(var(--color-accent-500, 233 190 115) / 0.06) 0px, transparent 50%), radial-gradient(at 0% 50%, rgb(var(--color-primary-500, 186 54 80) / 0.06) 0px, transparent 50%)'
      },
      animation: {
        'fade-in': 'fadeIn 0.3s ease-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'slide-down': 'slideDown 0.3s ease-out',
        'slide-in-right': 'slideInRight 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        shimmer: 'shimmer 2s linear infinite',
        glow: 'glow 2s ease-in-out infinite alternate'
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' }
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideDown: {
          '0%': { opacity: '0', transform: 'translateY(-10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideInRight: {
          '0%': { opacity: '0', transform: 'translateX(20px)' },
          '100%': { opacity: '1', transform: 'translateX(0)' }
        },
        scaleIn: {
          '0%': { opacity: '0', transform: 'scale(0.95)' },
          '100%': { opacity: '1', transform: 'scale(1)' }
        },
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' }
        },
        glow: {
          '0%': {
            boxShadow: '0 0 20px rgb(var(--color-primary-500, 186 54 80) / 0.25)'
          },
          '100%': {
            boxShadow: '0 0 30px rgb(var(--color-primary-500, 186 54 80) / 0.4)'
          }
        }
      },
      backdropBlur: {
        xs: '2px'
      },
      borderRadius: {
        '4xl': '2rem'
      }
    }
  },
  plugins: []
}
