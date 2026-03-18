import api from './index'

export const authApi = {
    getNonce: (address) => api.get(`/auth/nonce?address=${address}`),
    login: (address, signature) => api.post('/auth/login', {
        wallet_address: address,
        signature: signature,
    }),
    getMe: () => api.get('/me'),
    updateProfile: (name) => api.put('/me/profile', { name }),
}
