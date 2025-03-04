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
            path: '/editor/:id?',
            name: 'editor',
            component: () => import('../views/EditorView.vue')
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