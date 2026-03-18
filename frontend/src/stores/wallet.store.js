import { defineStore } from 'pinia'
import { ref, markRaw } from 'vue'
import { BrowserProvider } from 'ethers'
import { appKit } from '@/web3'

export const useWalletStore = defineStore('wallet', () => {
    const isConnected = ref(false)
    const address = ref('')
    const signer = ref(null)
    const provider = ref(null)

    // Helper to initialize signer from a raw provider
    async function _initSigner(walletProvider) {
        if (!walletProvider) return
        try {
            console.log('[WalletStore] Initializing Signer...')
            const browserProvider = new BrowserProvider(walletProvider)
            const s = await browserProvider.getSigner()

            provider.value = markRaw(browserProvider)
            signer.value = markRaw(s)
            console.log('[WalletStore] Signer Ready for:', address.value)
        } catch (error) {
            console.error('[WalletStore] Signer Init failed:', error)
            signer.value = null
            provider.value = null
        }
    }

    // Sync account state
    appKit.subscribeAccount(async (state) => {
        console.log('[WalletStore] Account Update:', state)
        isConnected.value = state.isConnected || false
        address.value = state.address || ''

        if (state.isConnected) {
            // Check if we already have a provider
            const wp = appKit.getWalletProvider()
            if (wp) {
                await _initSigner(wp)
            }
        } else {
            signer.value = null
            provider.value = null
        }
    })

    // Listen for provider changes
    // subscribeProviders expects a callback(providers) where providers is an object
    appKit.subscribeProviders(async (providers) => {
        console.log('[WalletStore] Providers Update:', providers)
        const wp = providers['eip155']
        if (wp && isConnected.value) {
            await _initSigner(wp)
        }
    })

    async function initSigner() {
        console.log('[WalletStore] Manual initSigner called')
        const wp = appKit.getWalletProvider()
        if (wp) {
            await _initSigner(wp)
        } else {
            console.warn('[WalletStore] No provider available yet')
        }
    }

    return {
        isConnected,
        address,
        signer,
        provider,
        initSigner
    }
})
