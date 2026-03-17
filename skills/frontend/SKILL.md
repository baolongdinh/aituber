# Frontend Vue 3 Skill

## TRIGGER
Read this file before writing ANY Vue 3 code: components, composables, stores, views, API calls.

---
name: vue3-frontend
description: >
  Production-grade Vue 3 frontend architecture with Composition API, Pinia, and Tailwind CSS.
  Use this skill for writing, reviewing, refactoring, or optimizing ANY Vue 3 frontend code — including
  components, composables, stores, views, API calls, routing, state management, or UI patterns.
  Trigger on any Vue file creation or edit, component design question, or when the
  user asks to "clean up", "restructure", "optimize", or "add a feature" to the Vue 3 frontend.
---

## PROJECT STACK
- **Vue 3** (Composition API + `<script setup>`)
- **Vite** build tool
- **Tailwind CSS** styling
- **Pinia** state management
- **Vue Router 4** routing
- **Axios** HTTP client
- **ethers.js v6** Web3

## ENHANCED ARCHITECTURE: STRICT N-LAYER

```
API Layer (src/api/) — HTTP calls only
    ↓
Pinia Store (src/stores/) — State + actions
    ↓
Composable (src/composables/) — Business logic orchestration
    ↓
View / Component (src/views/, src/components/) — UI only
```

**Strict Rules:**
- **Views** call **Composables** only — never Store/API directly
- **Composables** orchestrate Stores and business logic
- **Stores** call **API** functions and manage state
- **API** functions make HTTP calls only
- **No cross-layer violations** — each layer has single responsibility

## ENHANCED COMPONENT STRUCTURE TEMPLATE

```vue
<script setup>
// 1. Imports (Vue, router, stores, composables, components, utils)
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useNftStore } from '@/stores/nft.store'
import { useNft } from '@/composables/useNft'
import { useAuth } from '@/composables/useAuth'
import { formatPrice, formatDate } from '@/utils/formatters'
import NftCard from '@/components/nft/NftCard.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseModal from '@/components/ui/BaseModal.vue'

// 2. Props & Emits (with TypeScript-style validation)
const props = defineProps({
  categoryId: { 
    type: Number, 
    required: false,
    default: null 
  },
  showFilters: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits({
  'nft-selected': (nft) => typeof nft === 'object' && nft !== null,
  'filter-change': (filters) => typeof filters === 'object',
  'load-more': () => true
})

// 3. Router & Route
const router = useRouter()
const route = useRoute()

// 4. Reactive state (local component state)
const searchQuery = ref('')
const isLoading = ref(false)
const error = ref(null)
const selectedNft = ref(null)
const showDetailModal = ref(false)
const filters = ref({
  priceRange: [0, 1000],
  category: null,
  sortBy: 'created_at'
})

// 5. Composables / Stores (destructured with storeToRefs)
const { nfts, totalCount, isLoading: nftsLoading } = storeToRefs(useNftStore())
const { fetchNfts, buyNft, toggleFavorite } = useNft()
const { isAuthenticated, user } = storeToRefs(useAuthStore())

// 6. Computed properties (derived state)
const filteredNfts = computed(() => {
  let filtered = nfts.value
  
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(n => 
      n.name.toLowerCase().includes(query) ||
      n.description.toLowerCase().includes(query)
    )
  }
  
  if (props.categoryId) {
    filtered = filtered.filter(n => n.categoryId === props.categoryId)
  }
  
  return filtered
})

const hasMoreNfts = computed(() => {
  return filteredNfts.value.length < totalCount.value
})

// 7. Watchers (reactive side effects)
watch(
  () => filters.value,
  (newFilters) => {
    emit('filter-change', newFilters)
    // Reset pagination when filters change
    fetchNfts({ ...newFilters, page: 1 })
  },
  { deep: true }
)

watch(
  () => route.query.category,
  (newCategory) => {
    if (newCategory) {
      filters.value.category = parseInt(newCategory)
    }
  },
  { immediate: true }
)

// 8. Lifecycle hooks
onMounted(async () => {
  await fetchNfts({ ...filters.value, page: 1 })
})

// 9. Methods / Event handlers
async function handleSelectNft(nft) {
  selectedNft.value = nft
  showDetailModal.value = true
  emit('nft-selected', nft)
}

async function handleBuyNft(nft) {
  if (!isAuthenticated.value) {
    router.push('/auth/login')
    return
  }
  
  try {
    isLoading.value = true
    await buyNft(nft.id)
    // Show success notification
    showSuccessToast(`Successfully purchased ${nft.name}!`)
  } catch (err) {
    error.value = err.message
    showErrorToast('Failed to purchase NFT')
  } finally {
    isLoading.value = false
  }
}

function handleLoadMore() {
  emit('load-more')
}

function updateFilter(key, value) {
  filters.value[key] = value
}

// 10. Utility functions
function showSuccessToast(message) {
  // Implement toast notification
}

function showErrorToast(message) {
  // Implement error notification
}
</script>

<template>
  <div class="container mx-auto px-4 py-8">
    <!-- Header with search and filters -->
    <header class="mb-8">
      <div class="flex flex-col md:flex-row gap-4 items-center justify-between">
        <div class="flex-1 max-w-md">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search NFTs..."
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          />
        </div>
        
        <BaseButton
          v-if="showFilters"
          @click="showFilterPanel = !showFilterPanel"
          variant="outline"
        >
          Filters
        </BaseButton>
      </div>
      
      <!-- Filter panel -->
      <div v-if="showFilterPanel" class="mt-4 p-4 bg-gray-50 rounded-lg">
        <FilterPanel
          :filters="filters"
          @update="updateFilter"
        />
      </div>
    </header>

    <!-- Loading state -->
    <div v-if="isLoading || nftsLoading" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="text-center py-12">
      <div class="text-red-600 mb-4">{{ error }}</div>
      <BaseButton @click="fetchNfts(filters)" variant="outline">
        Retry
      </BaseButton>
    </div>

    <!-- Empty state -->
    <div v-else-if="filteredNfts.length === 0" class="text-center py-12">
      <div class="text-gray-500 mb-4">No NFTs found</div>
      <BaseButton @click="resetFilters" variant="outline">
        Clear Filters
      </BaseButton>
    </div>

    <!-- NFT Grid -->
    <main v-else>
      <div class="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 gap-6">
        <NftCard
          v-for="nft in filteredNfts"
          :key="nft.id"
          :nft="nft"
          @click="handleSelectNft"
          @buy="handleBuyNft"
          @favorite="toggleFavorite"
        />
      </div>
      
      <!-- Load more button -->
      <div v-if="hasMoreNfts" class="text-center mt-8">
        <BaseButton
          @click="handleLoadMore"
          :loading="isLoading"
          variant="outline"
          size="lg"
        >
          Load More
        </BaseButton>
      </div>
    </main>

    <!-- NFT Detail Modal -->
    <BaseModal
      v-if="showDetailModal"
      @close="showDetailModal = false"
      size="xl"
    >
      <NftDetail
        :nft="selectedNft"
        @buy="handleBuyNft"
        @close="showDetailModal = false"
      />
    </BaseModal>
  </div>
</template>

<style scoped>
/* Component-specific styles if needed */
</style>
```

--- 

## ⚡ RAPID DEVELOPMENT (PRO-LEVEL) 

Use the automation script `scripts/vue-scaffold.js` to quickly generate a new Vue component and its corresponding Pinia store following the strict folder structure. 

**Command:** 
```bash 
node scripts/vue-scaffold.js --name=<ComponentName> 
``` 

**Benefits:** 
- Ensures component and store are created in the correct directories (`src/components/`, `src/stores/`). 
- Includes standard imports and Composition API setup. 
- Enforces consistent naming conventions. 

---

## ENHANCED API LAYER PATTERN

```js
// src/api/index.js — Enhanced Axios instance
import axios from 'axios'
import { toast } from 'vue3-toastify'

// Create axios instance with defaults
const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor: attach JWT and request ID
api.interceptors.request.use(
  (config) => {
    // Add request ID for tracing
    config.headers['X-Request-ID'] = generateRequestId()
    
    // Add auth token if available
    const token = localStorage.getItem('access_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    
    // Log request in development
    if (import.meta.env.DEV) {
      console.log(`🚀 API Request: ${config.method?.toUpperCase()} ${config.url}`, config.data)
    }
    
    return config
  },
  (error) => {
    console.error('Request interceptor error:', error)
    return Promise.reject(error)
  }
)

// Response interceptor: handle common errors and responses
api.interceptors.response.use(
  (response) => {
    // Log successful response in development
    if (import.meta.env.DEV) {
      console.log(`✅ API Response: ${response.config.method?.toUpperCase()} ${response.config.url}`, response.data)
    }
    
    return response
  },
  async (error) => {
    const originalRequest = error.config
    
    // Handle 401 Unauthorized
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true
      
      try {
        // Try to refresh token
        const refreshResponse = await api.post('/auth/refresh', {
          refresh_token: localStorage.getItem('refresh_token')
        })
        
        const { access_token } = refreshResponse.data.data
        localStorage.setItem('access_token', access_token)
        
        // Retry original request with new token
        originalRequest.headers.Authorization = `Bearer ${access_token}`
        return api(originalRequest)
      } catch (refreshError) {
        // Refresh failed, logout user
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
        window.location.href = '/auth/login'
        return Promise.reject(refreshError)
      }
    }
    
    // Handle network errors
    if (!error.response) {
      toast.error('Network error. Please check your connection.')
      return Promise.reject(error)
    }
    
    // Handle rate limiting
    if (error.response?.status === 429) {
      const retryAfter = error.response.headers['retry-after']
      toast.error(`Rate limited. Please try again in ${retryAfter || 60} seconds.`)
      return Promise.reject(error)
    }
    
    // Handle server errors
    if (error.response?.status >= 500) {
      toast.error('Server error. Please try again later.')
      return Promise.reject(error)
    }
    
    return Promise.reject(error)
  }
)

// Utility function to generate request IDs
function generateRequestId() {
  return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
}

export default api
```

```js
// src/api/nft.api.js — Enhanced NFT API module
import api from './index'

export const nftApi = {
  // List NFTs with pagination and filters
  list: (params = {}) => {
    const queryParams = new URLSearchParams()
    
    Object.entries(params).forEach(([key, value]) => {
      if (value !== null && value !== undefined) {
        if (Array.isArray(value)) {
          value.forEach(v => queryParams.append(key, v))
        } else {
          queryParams.append(key, value)
        }
      }
    })
    
    return api.get(`/nfts?${queryParams}`)
  },
  
  // Get NFT by ID
  getById: (id) => api.get(`/nfts/${id}`),
  
  // Get NFT owners/transaction history
  getOwners: (id) => api.get(`/nfts/${id}/owners`),
  
  // Create new NFT (with file upload support)
  create: (formData) => {
    return api.post('/nfts', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
  },
  
  // Buy NFT
  buy: (id, paymentData) => api.post(`/nfts/${id}/buy`, paymentData),
  
  // Transfer NFT
  transfer: (id, transferData) => api.post(`/nfts/${id}/transfer`, transferData),
  
  // List NFT for sale
  listForSale: (id, saleData) => api.post(`/nfts/${id}/list`, saleData),
  
  // Cancel listing
  cancelListing: (id) => api.delete(`/nfts/${id}/listing`),
  
  // Toggle favorite
  toggleFavorite: (id) => api.post(`/nfts/${id}/favorite`),
  
  // Get user's NFTs
  getMyNfts: (params = {}) => api.get('/my-nfts', { params }),
  
  // Get NFT analytics
  getAnalytics: (id) => api.get(`/nfts/${id}/analytics`),
}
```

---

## ENHANCED PINIA STORE PATTERN

```js
// src/stores/nft.store.js — Production-grade NFT store
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { nftApi } from '@/api/nft.api'
import { toast } from 'vue3-toastify'

export const useNftStore = defineStore('nft', () => {
  // State
  const nfts = ref([])
  const currentNft = ref(null)
  const myNfts = ref([])
  const favorites = ref(new Set())
  const isLoading = ref(false)
  const error = ref(null)
  
  // Pagination state
  const pagination = ref({
    page: 1,
    limit: 20,
    total: 0,
    totalPages: 0,
    hasNext: false,
    hasPrev: false,
  })
  
  // Filters state
  const filters = ref({
    search: '',
    category: null,
    priceRange: [0, 10000],
    sortBy: 'created_at',
    sortDir: 'desc',
    status: 'all',
  })
  
  // Computed properties
  const filteredNfts = computed(() => {
    let filtered = nfts.value
    
    if (filters.value.search) {
      const search = filters.value.search.toLowerCase()
      filtered = filtered.filter(nft => 
        nft.name.toLowerCase().includes(search) ||
        nft.description.toLowerCase().includes(search) ||
        nft.creator.username.toLowerCase().includes(search)
      )
    }
    
    if (filters.value.category) {
      filtered = filtered.filter(nft => nft.categoryId === filters.value.category)
    }
    
    if (filters.value.status !== 'all') {
      filtered = filtered.filter(nft => nft.status === filters.value.status)
    }
    
    // Price range filter
    filtered = filtered.filter(nft => 
      nft.price >= filters.value.priceRange[0] && 
      nft.price <= filters.value.priceRange[1]
    )
    
    // Sorting
    filtered.sort((a, b) => {
      const field = filters.value.sortBy
      const direction = filters.value.sortDir === 'asc' ? 1 : -1
      
      if (field === 'price') {
        return (a.price - b.price) * direction
      } else if (field === 'created_at') {
        return (new Date(a.createdAt) - new Date(b.createdAt)) * direction
      } else if (field === 'name') {
        return a.name.localeCompare(b.name) * direction
      }
      
      return 0
    })
    
    return filtered
  })
  
  const favoriteNfts = computed(() => {
    return nfts.value.filter(nft => favorites.value.has(nft.id))
  })
  
  const hasMoreNfts = computed(() => pagination.value.hasNext)
  
  // Actions
  async function fetchNfts(params = {}) {
    isLoading.value = true
    error.value = null
    
    try {
      const requestParams = {
        page: params.page || pagination.value.page,
        limit: params.limit || pagination.value.limit,
        search: filters.value.search,
        category: filters.value.category,
        sort_by: filters.value.sortBy,
        sort_dir: filters.value.sortDir,
        ...params
      }
      
      const response = await nftApi.list(requestParams)
      const { data, meta } = response.data
      
      if (params.page === 1) {
        // First page, replace all nfts
        nfts.value = data
      } else {
        // Append for pagination
        nfts.value.push(...data)
      }
      
      // Update pagination
      pagination.value = {
        ...pagination.value,
        ...meta,
      }
      
      // Update favorites set
      data.forEach(nft => {
        if (nft.isFavorite) {
          favorites.value.add(nft.id)
        }
      })
      
    } catch (err) {
      error.value = err.response?.data?.error?.message || 'Failed to load NFTs'
      toast.error(error.value)
      throw err
    } finally {
      isLoading.value = false
    }
  }
  
  async function fetchNftById(id) {
    isLoading.value = true
    error.value = null
    
    try {
      const response = await nftApi.getById(id)
      currentNft.value = response.data.data
      return currentNft.value
    } catch (err) {
      error.value = err.response?.data?.error?.message || 'Failed to load NFT'
      toast.error(error.value)
      throw err
    } finally {
      isLoading.value = false
    }
  }
  
  async function fetchMyNfts(params = {}) {
    isLoading.value = true
    
    try {
      const response = await nftApi.getMyNfts(params)
      myNfts.value = response.data.data
      return myNfts.value
    } catch (err) {
      error.value = err.response?.data?.error?.message || 'Failed to load your NFTs'
      toast.error(error.value)
      throw err
    } finally {
      isLoading.value = false
    }
  }
  
  async function buyNft(id, paymentData) {
    isLoading.value = true
    
    try {
      const response = await nftApi.buy(id, paymentData)
      
      // Update local state
      const nftIndex = nfts.value.findIndex(nft => nft.id === id)
      if (nftIndex !== -1) {
        nfts.value[nftIndex] = response.data.data
      }
      
      // Add to my NFTs
      myNfts.value.unshift(response.data.data)
      
      toast.success('NFT purchased successfully!')
      return response.data.data
    } catch (err) {
      const errorMessage = err.response?.data?.error?.message || 'Failed to purchase NFT'
      toast.error(errorMessage)
      throw err
    } finally {
      isLoading.value = false
    }
  }
  
  async function toggleFavorite(id) {
    try {
      await nftApi.toggleFavorite(id)
      
      // Update favorites set
      if (favorites.value.has(id)) {
        favorites.value.delete(id)
      } else {
        favorites.value.add(id)
      }
      
      // Update nft in list
      const nftIndex = nfts.value.findIndex(nft => nft.id === id)
      if (nftIndex !== -1) {
        nfts.value[nftIndex].isFavorite = favorites.value.has(id)
      }
      
      // Update current nft if it's the same
      if (currentNft.value?.id === id) {
        currentNft.value.isFavorite = favorites.value.has(id)
      }
      
    } catch (err) {
      const errorMessage = err.response?.data?.error?.message || 'Failed to update favorite'
      toast.error(errorMessage)
      throw err
    }
  }
  
  function updateFilters(newFilters) {
    filters.value = { ...filters.value, ...newFilters }
    // Reset to first page when filters change
    pagination.value.page = 1
  }
  
  function resetFilters() {
    filters.value = {
      search: '',
      category: null,
      priceRange: [0, 10000],
      sortBy: 'created_at',
      sortDir: 'desc',
      status: 'all',
    }
    pagination.value.page = 1
  }
  
  function loadMore() {
    if (hasMoreNfts.value && !isLoading.value) {
      fetchNfts({ page: pagination.value.page + 1 })
    }
  }
  
  function $reset() {
    nfts.value = []
    currentNft.value = null
    myNfts.value = []
    favorites.value.clear()
    isLoading.value = false
    error.value = null
    pagination.value = {
      page: 1,
      limit: 20,
      total: 0,
      totalPages: 0,
      hasNext: false,
      hasPrev: false,
    }
    resetFilters()
  }
  
  return {
    // State
    nfts,
    currentNft,
    myNfts,
    favorites,
    isLoading,
    error,
    pagination,
    filters,
    
    // Computed
    filteredNfts,
    favoriteNfts,
    hasMoreNfts,
    
    // Actions
    fetchNfts,
    fetchNftById,
    fetchMyNfts,
    buyNft,
    toggleFavorite,
    updateFilters,
    resetFilters,
    loadMore,
    $reset,
  }
})
```

---

## COMPOSABLE PATTERN

```js
// src/composables/useNft.js
import { storeToRefs } from 'pinia'
import { useNftStore } from '@/stores/nft.store'
import { useRouter } from 'vue-router'

export function useNft() {
  const store = useNftStore()
  const { nfts, currentNft, isLoading, error } = storeToRefs(store)
  const router = useRouter()

  async function fetchNfts(params) {
    await store.fetchNfts(params)
  }

  async function buyNft(id) {
    try {
      await store.buyNft(id)
      router.push('/my-nfts')
    } catch (err) {
      console.error('Buy NFT failed:', err)
    }
  }

  return { nfts, currentNft, isLoading, error, fetchNfts, buyNft }
}
```

---

## ROUTER PATTERN

```js
// src/router/index.js
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth.store'

const routes = [
  {
    path: '/',
    component: () => import('@/layouts/DefaultLayout.vue'),
    children: [
      { path: '', name: 'home', component: () => import('@/views/HomeView.vue') },
      { path: 'nfts', name: 'nfts', component: () => import('@/views/nft/NftListView.vue') },
      { path: 'nfts/:id', name: 'nft-detail', component: () => import('@/views/nft/NftDetailView.vue') },
    ],
  },
  {
    path: '/auth',
    component: () => import('@/layouts/AuthLayout.vue'),
    children: [
      { path: 'login', name: 'login', component: () => import('@/views/auth/LoginView.vue') },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Navigation guard
router.beforeEach((to) => {
  const authStore = useAuthStore()
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return { name: 'login' }
  }
})

export default router
```

---

## VITE CONFIG TEMPLATE

```js
// vite.config.js
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': { target: 'http://localhost:8080', changeOrigin: true },
    },
  },
})
```

---

## .ENV.EXAMPLE TEMPLATE

```env
VITE_API_URL=http://localhost:8080/api/v1
VITE_CONTRACT_ADDRESS=0x...
VITE_CHAIN_ID=1
VITE_RPC_URL=https://mainnet.infura.io/v3/...
```

---

## DO / DON'T

✅ **DO**
- Luôn dùng `<script setup>` và Composition API
- Dùng `storeToRefs` khi destructure store để giữ reactivity
- Lazy-load route components với `() => import(...)`
- Prefix boolean ref: `isLoading`, `hasError`, `canEdit`
- Dùng `@/` alias cho tất cả imports

❌ **DON'T**
- KHÔNG gọi API trực tiếp trong component (phải qua composable/store)
- KHÔNG dùng Options API
- KHÔNG destructure store trực tiếp (mất reactivity): `const { nfts } = store` ❌
- KHÔNG dùng `localStorage` trực tiếp trong component (vào store/composable)
