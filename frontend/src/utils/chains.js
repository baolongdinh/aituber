import { defineChain } from '@reown/appkit/networks'

export const aiozMainnet = defineChain({
    id: 168,
    name: 'AIOZ Network',
    nativeCurrency: { name: 'AIOZ', symbol: 'AIOZ', decimals: 18 },
    rpcUrls: {
        default: { http: ['https://eth-dataseed.aioz.network'] },
    },
    blockExplorers: {
        default: { name: 'AIOZ Explorer', url: 'https://explorer.aioz.network' },
    },
})

export const aiozTestnet = defineChain({
    id: 4102,
    name: 'AIOZ Network Testnet',
    nativeCurrency: { name: 'AIOZ', symbol: 'AIOZ', decimals: 18 },
    rpcUrls: {
        default: { http: ['https://eth-ds.testnet.aioz.network'] },
    },
    blockExplorers: {
        default: { name: 'AIOZ Testnet Explorer', url: 'https://testnet.explorer.aioz.network' },
    },
    testnet: true,
})

// Active chains (order = default first)
export const SUPPORTED_CHAINS = [aiozMainnet, aiozTestnet]
