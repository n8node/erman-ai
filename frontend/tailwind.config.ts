import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./src/**/*.{js,ts,jsx,tsx,mdx}"],
  theme: {
    extend: {
      colors: {
        bg: "#ffffff",
        bg2: "#f8f8f7",
        bg3: "#f3f2ef",
        text: "#1a1a1a",
        text2: "#5f5e5a",
        text3: "#888780",
        border: "#e8e7e3",
        border2: "#d3d1c7",
        accent: "#185fa5",
        "accent-bg": "#e6f1fb",
      },
      fontFamily: {
        sans: [
          "Inter",
          "-apple-system",
          "BlinkMacSystemFont",
          "Segoe UI",
          "sans-serif",
        ],
      },
    },
  },
  plugins: [],
};

export default config;
