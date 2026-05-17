/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        gray: {
          950: '#0a0a0f',
          900: '#111118',
          850: '#161620',
          800: '#1c1c28',
          750: '#22222f',
          700: '#2a2a3a',
          600: '#3a3a50',
          500: '#5a5a78',
          400: '#8080a0',
          300: '#a0a0c0',
          200: '#c0c0d8',
          100: '#e0e0f0',
        },
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
  ],
}
