import { storeToRefs } from 'pinia'
import { useAuthStore } from '@/stores/auth.store'
import { useWalletStore } from '@/stores/wallet.store'
import { authApi } from '@/api/auth.api'
import { appKit } from '@/web3'
import { watch } from 'vue'

export function useAuth() {
    const authStore = useAuthStore()
    const walletStore = useWalletStore()
    const { user, token, isAuthenticated, isLoading, error } = storeToRefs(authStore)
    const { isConnected, address, signer } = storeToRefs(walletStore)

    // Local guard for concurrent calls
    let loginInProgress = false

    // AUTO-LOGIN: Trigger SIWE automatically when connected and signer is ready
    watch([isConnected, signer, isAuthenticated], async ([connected, currentSigner, authenticated]) => {
        console.log('[useAuth] Watch triggered - connected:', connected, 'hasSigner:', !!currentSigner, 'auth:', authenticated, 'loading:', isLoading.value)

        if (connected && currentSigner && !authenticated && !isLoading.value && !loginInProgress) {
            console.log('[useAuth] Auto-triggering SIWE login...')
            try {
                await login()
            } catch (err) {
                console.error('[useAuth] Auto-login failed:', err)
            }
        }
    }, { immediate: true })

    async function login() {
        if (isLoading.value || isAuthenticated.value || loginInProgress) return

        console.log('[useAuth] Login starting. isConnected:', isConnected.value, 'hasSigner:', !!signer.value)

        if (!isConnected.value) {
            await appKit.open()
            return
        }

        try {
            loginInProgress = true
            isLoading.value = true // Set loading early

            // Wait for signer if not yet ready
            if (!signer.value) {
                console.log('[useAuth] Signer missing, attempting to init...')
                await walletStore.initSigner()

                // Wait slightly for reactive update from store
                let attempts = 0
                while (!signer.value && attempts < 10) {
                    await new Promise(r => setTimeout(r, 500))
                    attempts++
                }
            }

            if (!signer.value) {
                console.error('[useAuth] Signer initialization timed out')
                throw new Error('Đang chuẩn bị ví, vui lòng đợi giây lát hoặc thử lại.')
            }

            // 1. Get Nonce
            const nonceRes = await authApi.getNonce(address.value)
            console.log('[useAuth] Nonce Response:', nonceRes.data)
            const nonce = nonceRes.data.data.nonce

            // 2. Sign Message (EIP-191)
            const message = `Chào mừng bạn đến với ViralCraft!\n\nĐịa chỉ ví: ${address.value.toLowerCase()}\nNonce: ${nonce}`
            const signature = await signer.value.signMessage(message)

            // 3. Verify & Login
            await authStore.login(address.value, signature)

            return true
        } catch (err) {
            console.error('[useAuth] SIWE Login failed:', err)
            throw err
        } finally {
            isLoading.value = false
            loginInProgress = false
        }
    }

    function logout() {
        authStore.logout()
    }

    return {
        user,
        token,
        isAuthenticated,
        isLoading,
        error,
        address,
        isConnected,
        login,
        logout
    }
}
