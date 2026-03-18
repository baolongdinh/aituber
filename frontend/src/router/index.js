import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth.store'

const routes = [
    {
        path: '/',
        name: 'dashboard',
        component: () => import('@/views/DashboardView.vue'),
        meta: { requiresAuth: false }
    },
    {
        path: '/explore',
        name: 'explore',
        component: () => import('@/views/GalleryView.vue'),
        meta: { requiresAuth: false }
    },
    {
        path: '/generator',
        name: 'generator',
        component: () => import('@/views/GeneratorView.vue'),
        meta: { requiresAuth: true }
    },
    {
        path: '/job/:id',
        name: 'job-detail',
        component: () => import('@/views/JobDetailView.vue'),
        meta: { requiresAuth: true }
    },
    {
        path: '/gallery',
        name: 'gallery',
        component: () => import('@/views/GalleryView.vue'),
        meta: { requiresAuth: true }
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

router.beforeEach((to, from) => {
    const authStore = useAuthStore()

    if (to.meta.requiresAuth && !authStore.isAuthenticated) {
        return { name: 'dashboard' }
    }
})

export default router
