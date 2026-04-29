import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans:   ["var(--font-inter)", "sans-serif"],
        inter:  ["var(--font-inter)", "sans-serif"],
        outfit: ["var(--font-inter)", "sans-serif"],
        mono:   ["var(--font-jetbrains-mono)", "monospace"],
      },
      colors: {
        // Brand
        brand: {
          DEFAULT: "#00E3AA",
          hover:   "#00cfA0",
          muted:   "#82e8b2",
          subtle:  "#1a2d24",
        },
        // Semantic
        danger:     "#FF4747",
        enterprise: "#a78bfa",
        body:       "#848483",
        // Background scale (darkest → lightest)
        canvas:  "#060d09",
        surface: {
          DEFAULT:   "#0d1510",
          secondary: "#080f0a",
          card:      "#111813",
          elevated:  "#1a211c",
          overlay:   "#1c231e",
        },
      },
    },
  },
  plugins: [
    require('@tailwindcss/typography'),
  ],
};
export default config;
