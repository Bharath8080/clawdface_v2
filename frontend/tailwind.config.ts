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
        inter: ["var(--font-inter)", "sans-serif"],
        outfit: ["var(--font-outfit)", "sans-serif"],
        mono: ["var(--font-jetbrains-mono)", "monospace"],
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
        // Background scale (darkest → lightest)
        canvas:  "#050505",
        surface: {
          DEFAULT:   "#0d0d0d",
          secondary: "#0a0a0a",
          card:      "#111111",
          elevated:  "#1a1a1a",
          overlay:   "#1c1c1e",
        },
      },
    },
  },
  plugins: [
    require('@tailwindcss/typography'),
  ],
};
export default config;
