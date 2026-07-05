/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        darkBg: "#0B0F19",
        darkCard: "#151B2C",
        darkBorder: "#222D44",
        accentBlue: "#3B82F6",
      }
    },
  },
  plugins: [],
}
