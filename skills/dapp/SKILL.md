# DApp / Smart Contract Skill

## TRIGGER
Read this file before doing any task related to blockchain: wallet connection, contract interaction, token/NFT operations, chain config.

---

## WEB3 STACK
- **Reown AppKit** (`@reown/appkit`) — wallet modal + session management (formerly WalletConnect Web3Modal)
- **ethers.js v6** — blockchain interaction
- **@reown/appkit-adapter-ethers** — bridge between AppKit and ethers.js
- **Chain**: AIOZ Network (EVM-compatible)

---
--- 

## ⚡ RAPID DEVELOPMENT (PRO-LEVEL) 

Use the automation script `scripts/web3-client-gen.js` to quickly generate a standardized Vue 3 Composable from a Smart Contract ABI. 

**Command:** 
```bash 
node scripts/web3-client-gen.js --abi=<path/to/abi.json> --name=<ContractName> 
``` 

**Benefits:** 
- Ensures consistent ethers.js v6 implementation. 
- Includes automatic handling of `isLoading` and `error` states. 
- Saves hours of manual ABI-to-JS plumbing. 


## AIOZ NETWORK CONFIG

```js
// src/utils/chains.js
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
```

---

## REOWN APPKIT SETUP

### Installation
```bash
npm install @reown/appkit @reown/appkit-adapter-ethers ethers
```

### Initialization
```js
// src/web3/index.js  ← init only once, import into main.js
import { createAppKit } from '@reown/appkit'
import { EthersAdapter } from '@reown/appkit-adapter-ethers'
import { aiozMainnet, aiozTestnet } from '@/utils/chains'

const projectId = import.meta.env.VITE_REOWN_PROJECT_ID // get from cloud.reown.com

const ethersAdapter = new EthersAdapter()

export const appKit = createAppKit({
  adapters: [ethersAdapter],
  networks: [aiozMainnet, aiozTestnet],  // aiozMainnet = default
  defaultNetwork: aiozMainnet,
  projectId,
  metadata: {
    name: import.meta.env.VITE_APP_NAME || 'AI Playground',
    description: 'AIOZ Network DApp',
    url: import.meta.env.VITE_APP_URL || 'http://localhost:3000',
    icons: ['/logo.png'],
  },
  features: {
    analytics: false,
    email: false,
    socials: [],
  },
  themeMode: 'dark',
})
```

```js
// src/main.js
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import '@/web3/index.js' // Init AppKit BEFORE mounting app

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
```

---

## WALLET STORE PATTERN (Reown AppKit)

```js
// src/stores/wallet.store.js
import { defineStore } from 'pinia'
import { ref, computed, markRaw } from 'vue'
import { useAppKitProvider, useAppKitAccount, useAppKitNetwork } from '@reown/appkit/vue'
import { ethers } from 'ethers'

export const useWalletStore = defineStore('wallet', () => {
  const signer = ref(null)

  // AppKit composables — đây là reactive state chính
  // Dùng trực tiếp trong component/composable thay vì lưu vào store
  // store chỉ lưu ethers signer (cần markRaw)

  async function initSigner() {
    const { walletProvider } = useAppKitProvider('eip155')
    if (!walletProvider.value) return
    const provider = new ethers.BrowserProvider(walletProvider.value)
    signer.value = markRaw(await provider.getSigner())
  }

  function clearSigner() {
    signer.value = null
  }

  return { signer, initSigner, clearSigner }
})
```

---

## WALLET COMPOSABLE PATTERN

```js
// src/composables/useWallet.js
import { computed, watch } from 'vue'
import { useAppKitAccount, useAppKitNetwork, useAppKit } from '@reown/appkit/vue'
import { useWalletStore } from '@/stores/wallet.store'
import { storeToRefs } from 'pinia'

export function useWallet() {
  const { open } = useAppKit()                    // mở modal connect/disconnect
  const { address, isConnected } = useAppKitAccount()  // reactive
  const { chainId, caipNetwork } = useAppKitNetwork()  // reactive
  const walletStore = useWalletStore()
  const { signer } = storeToRefs(walletStore)

  // Khi wallet connect → init ethers signer
  watch(isConnected, async (connected) => {
    if (connected) {
      await walletStore.initSigner()
    } else {
      walletStore.clearSigner()
    }
  })

  // Hiển thị địa chỉ rút gọn
  const shortAddress = computed(() => {
    if (!address.value) return ''
    return `${address.value.slice(0, 6)}...${address.value.slice(-4)}`
  })

  // Check đang ở đúng chain không
  const isCorrectChain = computed(() =>
    chainId.value === 168 || chainId.value === 4102
  )

  function openModal() { open() }
  function openNetworkModal() { open({ view: 'Networks' }) }

  return {
    address,
    isConnected,
    chainId,
    shortAddress,
    isCorrectChain,
    signer,
    openModal,
    openNetworkModal,
  }
}
```

---

## CONNECT WALLET BUTTON COMPONENT

```vue
<!-- src/components/wallet/ConnectWalletButton.vue -->
<script setup>
import { useWallet } from '@/composables/useWallet'

const { address, isConnected, shortAddress, openModal } = useWallet()
</script>

<template>
  <!-- Reown AppKit button — tự handle UI connect/disconnect/network -->
  <appkit-button />

  <!-- Hoặc dùng button tự custom: -->
  <button @click="openModal">
    {{ isConnected ? shortAddress : 'Connect Wallet' }}
  </button>
</template>
```

> ✅ **Reown AppKit cung cấp** `<appkit-button>` Web Component built-in — thêm vào bất kỳ đâu, tự xử lý connect/disconnect/network switching.

---

## CONTRACT INTERACTION PATTERN

```js
// src/composables/useNftContract.js
import { ethers } from 'ethers'
import { useWallet } from '@/composables/useWallet'
import { CONTRACTS } from '@/utils/contracts'
import { aiozMainnet, aiozTestnet } from '@/utils/chains'

export function useNftContract() {
  const { signer } = useWallet()

  // Read-only contract — dùng public RPC của AIOZ
  function getReadContract() {
    const activeChainId = parseInt(import.meta.env.VITE_CHAIN_ID)
    const chain = activeChainId === 168 ? aiozMainnet : aiozTestnet
    const provider = new ethers.JsonRpcProvider(chain.rpcUrls.default.http[0])
    return new ethers.Contract(CONTRACTS.NFT.address, CONTRACTS.NFT.abi, provider)
  }

  // Write contract — cần signer từ Reown
  function getWriteContract() {
    if (!signer.value) throw new Error('Wallet not connected')
    return new ethers.Contract(CONTRACTS.NFT.address, CONTRACTS.NFT.abi, signer.value)
  }

  async function mintNft(tokenURI) {
    const contract = getWriteContract()
    const tx = await contract.mint(tokenURI)
    await tx.wait()
    return tx.hash
  }

  async function buyNft(tokenId, priceWei) {
    const contract = getWriteContract()
    const tx = await contract.buy(tokenId, { value: priceWei })
    return await tx.wait()
  }

  return { mintNft, buyNft, getReadContract }
}
```

---

## CONTRACT CONFIG

```js
// src/utils/contracts.js
export const CONTRACTS = {
  NFT: {
    address: import.meta.env.VITE_NFT_CONTRACT_ADDRESS,
    abi: [], // import từ artifacts/NFT.json sau khi compile contract
  },
}
```

---

## SIGN-IN WITH ETHEREUM (SIWE) VIA REOWN

```js
// src/composables/useAuth.js
// Reown AppKit hỗ trợ SIWE built-in — dùng để authenticate với backend
import { useAppKitAccount } from '@reown/appkit/vue'
import { useWallet } from '@/composables/useWallet'
import { useAuthStore } from '@/stores/auth.store'
import { authApi } from '@/api/auth.api'

export function useAuth() {
  const { address, isConnected } = useAppKitAccount()
  const { signer } = useWallet()
  const authStore = useAuthStore()

  async function loginWithWallet() {
    if (!isConnected.value || !signer.value) throw new Error('Wallet not connected')

    // 1. Get nonce từ backend
    const nonceRes = await authApi.getNonce(address.value)
    const { nonce } = nonceRes.data.data

    // 2. Ký message bằng wallet (Reown đã handle provider)
    const message = `Sign in to AIOZ DApp\nAddress: ${address.value}\nNonce: ${nonce}`
    const signature = await signer.value.signMessage(message)

    // 3. Verify với backend
    await authStore.login(address.value, signature)
  }

  return { loginWithWallet, isConnected, address }
}
```

---

## .ENV.EXAMPLE TEMPLATE

```env
# Reown (cloud.reown.com)
VITE_REOWN_PROJECT_ID=your_project_id_here

# App
VITE_APP_NAME=AI Playground
VITE_APP_URL=http://localhost:3000

# Chain (168 = mainnet, 4102 = testnet)
VITE_CHAIN_ID=4102
VITE_NFT_CONTRACT_ADDRESS=0x...

# Backend API
VITE_API_URL=http://localhost:8080/api/v1
```

---

## DO / DON'T

✅ **DO**
- Dùng `useAppKitAccount()` từ `@reown/appkit/vue` để lấy address/isConnected (reactive)
- Dùng `<appkit-button />` built-in cho nhanh; custom button nếu cần branding riêng
- `markRaw()` cho ethers signer/provider khi lưu vào Pinia store
- Thêm cả mainnet lẫn testnet vào `networks[]` để user có thể switch
- Lấy `VITE_REOWN_PROJECT_ID` từ [cloud.reown.com](https://cloud.reown.com)

❌ **DON'T**
- KHÔNG dùng `window.ethereum` trực tiếp — để Reown AppKit quản lý provider
- KHÔNG hardcode chain ID hay contract address trong code
- KHÔNG bỏ `markRaw()` khi store ethers objects trong Pinia
- KHÔNG quên `await tx.wait()` sau write transaction
- KHÔNG dùng ethers v5 `BigNumber` — ethers v6 dùng native `BigInt`
