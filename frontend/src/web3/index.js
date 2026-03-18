import { createAppKit } from '@reown/appkit/vue'
import { EthersAdapter } from '@reown/appkit-adapter-ethers'
import { aiozMainnet, aiozTestnet } from '@/utils/chains'

const projectId = import.meta.env.VITE_REOWN_PROJECT_ID || 'c3dc626d03d01ee34468f9464670c226' // Placeholder or user env

const ethersAdapter = new EthersAdapter()

export const appKit = createAppKit({
    adapters: [ethersAdapter],
    networks: [aiozMainnet, aiozTestnet],
    defaultNetwork: aiozMainnet,
    projectId,
    metadata: {
        name: 'AITuber',
        description: 'AITuber - AI Video Series Generator',
        url: window.location.origin,
        icons: ['/logo.png'],
    },
    features: {
        analytics: false,
        email: true,
        socials: ['google', 'x', 'discord', 'apple'],
        emailShowWallets: true,
    },
    themeMode: 'dark',
})
