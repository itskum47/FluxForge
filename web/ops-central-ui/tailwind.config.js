/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{js,ts,jsx,tsx}",
    ],
    darkMode: "class",
    theme: {
        extend: {
            colors: {
                "primary": "#6324eb",
                "primary-dark": "#4f1bc4",
                "background-light": "#f6f6f8",
                "background-dark": "#0B1220", // Unified dark background base
                "ops-bg": "#0B1220",
                "ops-card": "#11161F",
                "ops-success": "#10b981",
                "ops-warning": "#f59e0b",
                "ops-danger": "#ef4444",
                "surface-dark": "#18181b",
                "surface-darker": "#110d1a",
                "critical": "#FF3B3B",
                "critical-dark": "#8B0000",
                "recovery": {
                    400: "#2dd4bf",
                    500: "#14b8a6",
                    600: "#0d9488",
                    900: "#134e4a",
                },
                "neutral-slate": {
                    800: "#1e1b29",
                    700: "#2d2a3d",
                    600: "#454159",
                    400: "#9ca3af",
                }
            },
            fontFamily: {
                "display": ["Inter", "sans-serif"],
                "mono": ["JetBrains Mono", "monospace"],
            },
            borderRadius: {
                "DEFAULT": "0.5rem",
                "lg": "0.75rem",
                "xl": "1.5rem",
                "full": "9999px"
            },
            animation: {
                'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
                'shimmer': 'shimmer 1s infinite',
            },
            keyframes: {
                shimmer: {
                    '0%': { transform: 'translateX(-100%)' },
                    '100%': { transform: 'translateX(100%)' },
                }
            }
        },
    },
    plugins: [],
}
