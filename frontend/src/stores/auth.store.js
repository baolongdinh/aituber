import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth.api'

export const useAuthStore = defineStore('auth', () => {
    const user = ref(JSON.parse(localStorage.getItem('user')) || null)
    const token = ref(localStorage.getItem('access_token') || null)
    const isLoading = ref(false)
    const error = ref(null)

    const isAuthenticated = computed(() => !!token.value && !!user.value)

    async function login(address, signature) {
        console.log('Starting backend login for:', address)
        isLoading.value = true
        error.value = null
        try {
            const response = await authApi.login(address, signature)
            console.log('Backend login success')
            const { token: jwt, user: userData } = response.data.data
            // ...

            token.value = jwt
            user.value = userData

            localStorage.setItem('access_token', jwt)
            localStorage.setItem('user', JSON.stringify(userData))

            return userData
        } catch (err) {
            error.value = err.response?.data?.message || 'Login failed'
            throw err
        } finally {
            isLoading.value = false
        }
    }

    function logout() {
        token.value = null
        user.value = null
        localStorage.removeItem('access_token')
        localStorage.removeItem('user')
    }

    async function fetchMe() {
        if (!token.value) return
        try {
            const response = await authApi.getMe()
            user.value = response.data.data
            localStorage.setItem('user', JSON.stringify(user.value))
        } catch (err) {
            logout()
        }
    }

    async function updateProfile(name) {
        isLoading.value = true
        error.value = null
        try {
            await authApi.updateProfile(name)
            user.value = { ...user.value, name }
            localStorage.setItem('user', JSON.stringify(user.value))
            return true
        } catch (err) {
            error.value = err.response?.data?.message || 'Update profile failed'
            throw err
        } finally {
            isLoading.value = false
        }
    }

    return {
        user,
        token,
        isLoading,
        error,
        isAuthenticated,
        login,
        logout,
        fetchMe,
        updateProfile,
    }
})
