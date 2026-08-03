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
          50: token('primary-50', '255 241 244'),
          100: token('primary-100', '255 228 234'),
          200: token('primary-200', '254 205 216'),
          300: token('primary-300', '252 163 183'),
          400: token('primary-400', '244 94 127'),
          500: token('primary-500', '210 31 75'),
          600: token('primary-600', '184 23 63'),
          700: token('primary-700', '154 21 55'),
          800: token('primary-800', '128 23 51'),
          900: token('primary-900', '110 24 48'),
          950: token('primary-950', '61 7 22')
        },
        accent: {
          50: token('accent-50', '255 251 235'),
          100: token('accent-100', '255 244 194'),
          200: token('accent-200', '255 231 128'),
          300: token('accent-300', '251 207 75'),
          400: token('accent-400', '247 196 59'),
          500: token('accent-500', '244 189 56'),
          600: token('accent-600', '211 148 24'),
          700: token('accent-700', '173 106 14'),
          800: token('accent-800', '142 82 18'),
          900: token('accent-900', '117 66 19'),
          950: token('accent-950', '67 33 5')
        },
        dark: {
          50: token('ink-50', '250 247 240'),
          100: token('ink-100', '238 237 233'),
          200: token('ink-200', '217 221 225'),
          300: token('ink-300', '170 178 188'),
          400: token('ink-400', '139 149 160'),
          500: token('ink-500', '104 115 128'),
          600: token('ink-600', '74 85 98'),
          700: token('ink-700', '48 57 69'),
          800: token('ink-800', '23 29 36'),
          900: token('ink-900', '13 17 22'),
          950: token('ink-950', '3 5 7')
        },
        canvas: token('canvas', '245 247 248'),
        surface: token('surface', '255 255 255'),
        'surface-muted': token('surface-muted', '237 241 243'),
        'theme-border': token('theme-border', '217 224 228'),
        'theme-text': token('theme-text', '32 42 49'),
        'theme-muted': token('theme-muted', '99 113 122'),
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
        glow: '0 0 20px rgb(var(--color-primary-500, 210 31 75) / 0.25)',
        'glow-lg': '0 0 40px rgb(var(--color-primary-500, 210 31 75) / 0.35)',
        card: '0 1px 3px rgba(0, 0, 0, 0.04), 0 1px 2px rgba(0, 0, 0, 0.06)',
        'card-hover': '0 10px 40px rgba(0, 0, 0, 0.08)',
        'inner-glow': 'inset 0 1px 0 rgba(255, 255, 255, 0.1)'
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-primary':
          'linear-gradient(135deg, rgb(var(--color-primary-500, 210 31 75)) 0%, rgb(var(--color-primary-600, 184 23 63)) 100%)',
        'gradient-dark':
          'linear-gradient(135deg, rgb(var(--color-ink-800, 23 29 36)) 0%, rgb(var(--color-ink-900, 13 17 22)) 100%)',
        'gradient-glass':
          'linear-gradient(135deg, rgba(255,255,255,0.1) 0%, rgba(255,255,255,0.05) 100%)',
        'mesh-gradient':
          'radial-gradient(at 40% 20%, rgb(var(--color-primary-500, 210 31 75) / 0.1) 0px, transparent 50%), radial-gradient(at 80% 0%, rgb(var(--color-accent-500, 244 189 56) / 0.06) 0px, transparent 50%), radial-gradient(at 0% 50%, rgb(var(--color-primary-500, 210 31 75) / 0.06) 0px, transparent 50%)'
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
            boxShadow: '0 0 20px rgb(var(--color-primary-500, 210 31 75) / 0.25)'
          },
          '100%': {
            boxShadow: '0 0 30px rgb(var(--color-primary-500, 210 31 75) / 0.4)'
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
