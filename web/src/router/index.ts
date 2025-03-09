import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'

const router = createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: '/',
            name: 'home',
            component: HomeView
        },
        {
            path: '/content',
            name: 'content',
            component: () => import('../views/ContentBrowserView.vue')
        },
        {
            path: '/editor/:id?',
            name: 'editor',
            component: () => import('../views/EditorView.vue')
        },
        {
            path: '/editor2/:id?',
            name: 'editor2',
            component: () => import('../views/EnhancedEditorView.vue')
        },
        {
            path: '/debug/:executionId',
            name: 'debug',
            component: () => import('../views/DebugView.vue')
        },
        {
            path: '/about',
            name: 'about',
            component: () => import('../views/AboutView.vue')
        }
    ]
})

export default router