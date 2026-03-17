# Web3 Interaction Patterns (AIOZ Network + Reown)

## TRIGGER
Read this file when handling advanced Web3: network switching, gas estimation, ERC20 approval, event listening, backend signature verification with AIOZ Network.

---

## AIOZ NETWORK CHAIN IDS

| Network       | Chain ID | RPC URL                                  | Explorer                              |
|---------------|----------|------------------------------------------|---------------------------------------|
| Mainnet       | `168`    | `https://eth-dataseed.aioz.network`      | `https://explorer.aioz.network`       |
| Testnet       | `4102`   | `https://eth-ds.testnet.aioz.network`    | `https://testnet.explorer.aioz.network` |

> Native token: **AIOZ** (18 decimals) — used for gas fees and payments

---

## NETWORK SWITCHING VIA REOWN

Reown AppKit handles network switching via UI automatically. If need to switch programmatically:

```js
// src/composables/useNetwork.js
import { useAppKitNetwork } from '@reown/appkit/vue'
import { aiozMainnet, aiozTestnet } from '@/utils/chains'
import { computed } from 'vue'

export function useNetwork() {
  const { chainId, switchNetwork } = useAppKitNetwork()

  const isMainnet  = computed(() => chainId.value === 168)
  const isTestnet  = computed(() => chainId.value === 4102)
  const isAiozChain = computed(() => isMainnet.value || isTestnet.value)
  const networkLabel = computed(() => isMainnet.value ? 'Mainnet' : isTestnet.value ? 'Testnet' : 'Unknown')

  async function switchToMainnet() { await switchNetwork(aiozMainnet) }
  async function switchToTestnet() { await switchNetwork(aiozTestnet) }

  return { chainId, isMainnet, isTestnet, isAiozChain, networkLabel, switchToMainnet, switchToTestnet }
}
```

```vue
<!-- Component showing network badge and switch -->
<script setup>
import { useNetwork } from '@/composables/useNetwork'
const { networkLabel, isTestnet, switchToMainnet, switchToTestnet } = useNetwork()
</script>

<template>
  <div class="flex items-center gap-2">
    <span
      :class="isTestnet ? 'bg-yellow-500/20 text-yellow-400' : 'bg-green-500/20 text-green-400'"
      class="px-2 py-1 rounded-full text-xs font-medium"
    >
      {{ networkLabel }}
    </span>
    <!-- Reown built-in network selector -->
    <appkit-network-button />
  </div>
</template>
```

---

## GAS ESTIMATION PATTERN (AIOZ)

```js
async function estimateAndSend(contract, method, args, valueAioz = 0n) {
  try {
    const gasEstimate = await contract[method].estimateGas(...args, { value: valueAioz })
    const gasLimit = (gasEstimate * 120n) / 100n  // +20% buffer

    const tx = await contract[method](...args, { value: valueAioz, gasLimit })
    return await tx.wait()
  } catch (err) {
    if (err.code === 'ACTION_REJECTED') throw new Error('Transaction cancelled by user')
    if (err.code === 'INSUFFICIENT_FUNDS') throw new Error('Insufficient AIOZ balance')
    throw err
  }
}
```

---

## AIOZ TOKEN FORMAT HELPERS

```js
// src/utils/formatters.js
import { ethers } from 'ethers'

// Wei → AIOZ display
export function formatAioz(wei, decimals = 4) {
  return `${parseFloat(ethers.formatEther(wei)).toFixed(decimals)} AIOZ`
}

// AIOZ string → Wei (để gửi trong transaction)
export function parseAioz(aioz) {
  return ethers.parseEther(aioz.toString())
}

// Format số lớn (NFT token IDs)
export function formatTokenId(tokenId) {
  return `#${tokenId.toString()}`
}

// Short address
export function shortAddress(address, start = 6, end = 4) {
  if (!address) return ''
  return `${address.slice(0, start)}...${address.slice(-end)}`
}

// Link to AIOZ Explorer
export function explorerTxUrl(txHash, chainId = 168) {
  const base = chainId === 168
    ? 'https://explorer.aioz.network'
    : 'https://testnet.explorer.aioz.network'
  return `${base}/tx/${txHash}`
}

export function explorerAddressUrl(address, chainId = 168) {
  const base = chainId === 168
    ? 'https://explorer.aioz.network'
    : 'https://testnet.explorer.aioz.network'
  return `${base}/address/${address}`
}
```

---

## ERC20 APPROVAL PATTERN (AIOZ-based token)

```js
// Dùng khi contract cần transferFrom một ERC20 token thay vì native AIOZ
async function approveERC20AndExecute(tokenAddress, spender, amount, actionFn) {
  const { signer } = useWallet()
  if (!signer.value) throw new Error('Wallet not connected')

  const erc20Abi = [
    'function approve(address spender, uint256 amount) returns (bool)',
    'function allowance(address owner, address spender) view returns (uint256)',
  ]
  const token = new ethers.Contract(tokenAddress, erc20Abi, signer.value)
  const owner = await signer.value.getAddress()

  const allowance = await token.allowance(owner, spender)
  if (allowance < amount) {
    const tx = await token.approve(spender, amount)
    await tx.wait()
  }

  return await actionFn()
}
```

---

## LISTEN TO CONTRACT EVENTS

```js
// src/composables/useContractEvents.js
import { onUnmounted } from 'vue'
import { ethers } from 'ethers'
import { aiozMainnet, aiozTestnet } from '@/utils/chains'
import { CONTRACTS } from '@/utils/contracts'

export function useContractEvents() {
  const chainId = parseInt(import.meta.env.VITE_CHAIN_ID)
  const chain = chainId === 168 ? aiozMainnet : aiozTestnet
  const provider = new ethers.JsonRpcProvider(chain.rpcUrls.default.http[0])
  const contract = new ethers.Contract(CONTRACTS.NFT.address, CONTRACTS.NFT.abi, provider)

  function onNftMinted(callback) {
    contract.on('Transfer', (from, to, tokenId) => {
      if (from === ethers.ZeroAddress) {
        callback({ to, tokenId: tokenId.toString() })
      }
    })
  }

  function onNftSold(callback) {
    contract.on('Sale', (tokenId, seller, buyer, price) => {
      callback({ tokenId: tokenId.toString(), seller, buyer, price })
    })
  }

  // QUAN TRỌNG: cleanup
  onUnmounted(() => contract.removeAllListeners())

  return { onNftMinted, onNftSold }
}
```

---

## BACKEND: VERIFY SIGNATURE (Go)

```go
// pkg/web3/signature.go
// Dùng để verify SIWE (Sign-In with Ethereum) trên backend Go
package web3

import (
    "strings"
    "github.com/ethereum/go-ethereum/accounts"
    "github.com/ethereum/go-ethereum/common/hexutil"
    "github.com/ethereum/go-ethereum/crypto"
)

// VerifySignature kiểm tra signature từ frontend SIWE flow.
// message phải giống hệt message đã ký ở frontend.
// address là địa chỉ ví người dùng (case-insensitive).
func VerifySignature(address, message, signature string) (bool, error) {
    msgHash := accounts.TextHash([]byte(message))

    sigBytes, err := hexutil.Decode(signature)
    if err != nil {
        return false, err
    }
    // Normalize recovery bit (MetaMask/Reown thêm 27)
    if sigBytes[64] >= 27 {
        sigBytes[64] -= 27
    }

    pubKey, err := crypto.SigToPub(msgHash, sigBytes)
    if err != nil {
        return false, err
    }

    recovered := crypto.PubkeyToAddress(*pubKey)
    return strings.EqualFold(recovered.Hex(), address), nil
}
```

```go
// go.mod dependency cần thêm:
// require github.com/ethereum/go-ethereum v1.13+
```

---

## CHAIN-AWARE ENVIRONMENT

> Mọi feature liên quan đến contract phải check `VITE_CHAIN_ID` và dùng đúng RPC / explorer URL tương ứng.

```js
// src/utils/env.js
export const activeChainId = parseInt(import.meta.env.VITE_CHAIN_ID)
export const isTestnet = activeChainId === 4102
export const isMainnet = activeChainId === 168
```

---

## DO / DON'T

✅ **DO**
- Dùng `useAppKitNetwork()` → `switchNetwork()` để switch chain qua Reown
- Dùng `<appkit-network-button />` là cách đơn nhất để show network chooser
- Render explorer links trỏ đúng network (mainnet vs testnet)
- Luôn thêm gas buffer 20% khi estimate

❌ **DON'T**
- KHÔNG dùng `wallet_switchEthereumChain` thủ công — Reown đã wrap sẵn
- KHÔNG mix native AIOZ (wei) với ERC20 token amounts khi không format rõ ràng
- KHÔNG hardcode explorer URL — dùng helper `explorerTxUrl(hash, chainId)`
