import axios from 'axios'

const api = axios.create({
    baseURL: import.meta.env.VITE_API_URL || '/api/v1',
    timeout: 30000,
    headers: {
        'Content-Type': 'application/json',
    },
})

// Request interceptor: attach JWT
api.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem('access_token')
        if (token) {
            config.headers.Authorization = `Bearer ${token}`
        }
        return config
    },
    (error) => Promise.reject(error)
)

// Response interceptor: handle token expiration
api.interceptors.response.use(
    (response) => response,
    async (error) => {
        const originalRequest = error.config

        // If error is 401 and we haven't retried yet
        if (error.response?.status === 401 && !originalRequest._retry) {
            originalRequest._retry = true

            // In a real app, you might try to refresh the token here
            // For now, we'll just logout on 401
            localStorage.removeItem('access_token')
            localStorage.removeItem('user')
            window.dispatchEvent(new Event('auth-logout'))
        }

        return Promise.reject(error)
    }
)

export default api
