// astro.config.mjs
import { defineConfig } from 'astro/config';
import tailwind from "@tailwindcss/vite";

// https://astro.build/config
export default defineConfig({
    integrations: [tailwind()],
    output: 'server', // Für Server-Side Rendering und API-Routen
    server: {
        port: 4321,
        host: true
    },
    vite: {
        ssr: {
            noExternal: ['chart.js']
        }
    }
});