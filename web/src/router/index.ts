import { createRouter, createWebHistory } from 'vue-router'
import WorkspaceView from '../views/WorkspaceView.vue'
import HomeView from '../views/HomeView.vue'

const router = createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: '/',
            name: 'workspace',
            component: WorkspaceView,
            children: [
                {
                    path: '',
                    name: 'home',
                    component: HomeView
                },
                {
                    path: '/content',
                    name: 'content',
                    component: () => import('../views/ContentBrowserView.vue')
                },
                {
                    path: '/about',
                    name: 'about',
                    component: () => import('../views/AboutView.vue')
                },
                {
                    path: '/user/settings',
                    name: 'userSettings',
                    component: () => import('../views/UserSettingsView.vue')
                }
            ]
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
            path: '/error-testing',
            name: 'errorTesting',
            component: () => import('../views/ErrorTestingView.vue')
        }
    ]
})

export default router