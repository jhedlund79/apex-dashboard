import type { Config } from 'tailwindcss'

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        bg: '#0a0a0f',
        surface: {
          DEFAULT: '#12121a',
          2: '#1a1a26',
          3: '#22222f',
        },
        accent: { DEFAULT: '#00ff88', 2: '#00cc6a' },
        danger: '#ff3366',
        warn: '#ffaa00',
        brand: {
          blue: '#4488ff',
          purple: '#8855ff',
          cyan: '#00ccff',
        },
        content: {
          primary: '#e8e8f0',
          secondary: '#8888a0',
          muted: '#555570',
        },
      },
      fontFamily: {
        sans: ["'Instrument Sans'", 'sans-serif'],
        mono: ["'DM Mono'", 'monospace'],
      },
      borderRadius: {
        card: '12px',
      },
      keyframes: {
        fadeIn: {
          from: { opacity: '0', transform: 'translateY(8px)' },
          to: { opacity: '1', transform: 'translateY(0)' },
        },
        pulse: {
          '0%, 100%': { boxShadow: '0 0 0 0 rgba(0,255,136,.4)' },
          '50%': { boxShadow: '0 0 0 8px rgba(0,255,136,0)' },
        },
        shimmer: {
          '0%, 100%': { opacity: '0.4' },
          '50%': { opacity: '0.8' },
        },
      },
      animation: {
        fadeIn: 'fadeIn 0.5s ease forwards',
        pulse: 'pulse 2s infinite',
        shimmer: 'shimmer 1.5s ease-in-out infinite',
      },
    },
  },
  plugins: [],
} satisfies Config
