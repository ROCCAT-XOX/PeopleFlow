// tailwind.config.cjs
/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ['./src/**/*.{astro,html,js,jsx,md,mdx,svelte,ts,tsx,vue}'],
    theme: {
        extend: {
            colors: {
                primary: {
                    DEFAULT: '#22C55E',
                    dark: '#15803D',
                    light: '#D9FBE5',
                    lighter: '#F7FDE6',
                },
                secondary: {
                    DEFAULT: '#0F151C',
                }
            },
        },
    },
    plugins: [],
}