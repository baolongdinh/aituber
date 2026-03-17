# State Management (Pinia) Skill

## TRIGGER
Read this file when writing Pinia stores or managing complex state.

---

## PINIA STORE TEMPLATE (FULL)

```js
// src/stores/auth.store.js
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth.api'

export const useAuthStore = defineStore('auth', () => {
  // ─── State ───────────────────────────────────────────
  const user = ref(null)          // null = not logged in
  const token = ref(localStorage.getItem('access_token') || null)
  const isLoading = ref(false)
  const error = ref(null)

  // ─── Getters (computed) ───────────────────────────────
  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const walletAddress = computed(() => user.value?.walletAddress ?? null)

  // ─── Actions ─────────────────────────────────────────
  async function login(walletAddress, signature) {
    isLoading.value = true
    error.value = null
    try {
      const res = await authApi.login({ wallet_address: walletAddress, signature })
      token.value = res.data.data.access_token
      user.value  = res.data.data.user
      localStorage.setItem('access_token', token.value)
    } catch (err) {
      error.value = err.response?.data?.error || 'Login failed'
      throw err   // re-throw so composable/component can handle
    } finally {
      isLoading.value = false
    }
  }

  async function fetchMe() {
    if (!token.value) return
    try {
      const res = await authApi.getMe()
      user.value = res.data.data
    } catch {
      logout()
    }
  }

  function logout() {
    user.value  = null
    token.value = null
    localStorage.removeItem('access_token')
  }

  // ─── Init (run on app start) ──────────────────────────
  function init() {
    if (token.value) fetchMe()
  }

  return {
    // State
    user, token, isLoading, error,
    // Getters
    isAuthenticated, walletAddress,
    // Actions
    login, logout, fetchMe, init,
  }
}, {
  persist: false, // use localStorage manually for more control
})
```

---

## STORE NAMING CONVENTIONS

| Domain  | Store file         | Store ID   | Usage                |
|---------|--------------------|------------|----------------------|
| Auth    | `auth.store.js`    | `'auth'`   | `useAuthStore()`     |
| NFT     | `nft.store.js`     | `'nft'`    | `useNftStore()`      |
| User    | `user.store.js`    | `'user'`   | `useUserStore()`     |
| Wallet  | `wallet.store.js`  | `'wallet'` | `useWalletStore()`   |
| UI      | `ui.store.js`      | `'ui'`     | `useUIStore()` (modals, toasts) |

---

## UI STORE (global toasts / modals)

```js
// src/stores/ui.store.js
export const useUIStore = defineStore('ui', () => {
  const toasts = ref([])
  const modal = ref({ open: false, component: null, props: {} })

  function showToast({ message, type = 'info', duration = 3000 }) {
    const id = Date.now()
    toasts.value.push({ id, message, type })
    setTimeout(() => {
      toasts.value = toasts.value.filter(t => t.id !== id)
    }, duration)
  }

  function showSuccess(message) { showToast({ message, type: 'success' }) }
  function showError(message)   { showToast({ message, type: 'error' }) }

  function openModal(component, props = {}) {
    modal.value = { open: true, component, props }
  }
  function closeModal() {
    modal.value = { open: false, component: null, props: {} }
  }

  return { toasts, modal, showSuccess, showError, openModal, closeModal }
})
```

---

## PATTERN: Reactive Destructuring

```js
// ✅ CORRECT: Use storeToRefs to destructure state/getters
import { storeToRefs } from 'pinia'
const store = useNftStore()
const { nfts, isLoading, error } = storeToRefs(store) // reactive refs
const { fetchNfts, buyNft } = store // actions don't need storeToRefs

// ❌ WRONG: Doing this loses reactivity
const { nfts } = useNftStore() // nfts is plain value, not reactive
```

---

## DO / DON'T

✅ **DO**
- One store per domain (nft, auth, user...)
- Actions are `async` functions
- Re-throw errors in actions so composables can handle
- Use `storeToRefs` when destructuring in composables/components
- Reset store on logout: `$reset()` or manual reset

❌ **DON'T**
- NEVER call API directly in component — must go through store action
- NEVER store sensitive data in store without clear plan to clear
- NEVER use `$patch` for complex logic — write clearer action instead
