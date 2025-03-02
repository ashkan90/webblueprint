import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [vue()],
    resolve: {
        alias: {
            '@': path.resolve(__dirname, './web/src')
        }
    },
    root: './web',
    build: {
        outDir: '../dist',
        emptyOutDir: true
    },
    server: {
        proxy: {
            '/api': {
                target: 'http://localhost:8089',
                changeOrigin: true
            },
            '/ws': {
                target: 'ws://localhost:8089',
                ws: true
            }
        }
    }
})