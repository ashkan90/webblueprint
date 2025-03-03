import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [vue()],
    resolve: {
        alias: {
            '@': path.resolve(__dirname, './src')
        }
    },
    server: {
        port: 8089,
        proxy: {
            // Proxy API requests
            '/api': {
                target: 'http://localhost:8089',
                changeOrigin: true,
            },
            // Proxy WebSocket requests
            '/ws': {
                target: 'ws://localhost:8089',
                ws: true,
                changeOrigin: true
            }
        }
    },
    build: {
        outDir: '../dist',
        emptyOutDir: true
    }
})