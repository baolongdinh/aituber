<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useUIStore } from '@/stores/ui.store'
import { useAuthStore } from '@/stores/auth.store'
import { useWalletStore } from '@/stores/wallet.store'
import { useAuth } from '@/composables/useAuth'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'

const uiStore = useUIStore()
const authStore = useAuthStore()
const walletStore = useWalletStore()
const { platform } = storeToRefs(uiStore)
const { isAuthenticated, user: authUser } = storeToRefs(authStore)
const { isConnected } = storeToRefs(walletStore)
const { login, logout, isLoading: isLoggingIn } = useAuth()
const router = useRouter()

const isEditingName = ref(false)
const newName = ref('')
const isDropdownOpen = ref(false)

const avatarUrl = computed(() => {
  if (authUser.value?.avatar_url) return authUser.value.avatar_url
  const seed = authUser.value?.wallet_address || 'default'
  return `https://api.dicebear.com/7.x/identicon/svg?seed=${seed}`
})

function toggleDropdown(e) {
  e.stopPropagation()
  isDropdownOpen.value = !isDropdownOpen.value
}

function closeDropdown() {
  isDropdownOpen.value = false
  isEditingName.value = false
}

function startEditing() {
  newName.value = authUser.value?.name || ''
  isEditingName.value = true
}

async function handleUpdateProfile() {
  if (!newName.value.trim()) return
  try {
    await authStore.updateProfile(newName.value)
    isEditingName.value = false
  } catch (err) {
    console.error('Update profile failed:', err)
  }
}

function handleLogout() {
  logout()
  isDropdownOpen.value = false
  router.push('/')
}

onMounted(() => {
  window.addEventListener('click', closeDropdown)
})

onUnmounted(() => {
  window.removeEventListener('click', closeDropdown)
})

function setPlatform(p) {
  uiStore.setPlatform(p)
}
</script>

<template>
  <div class="main-layout" :class="'theme-' + platform">
    <!-- Global Glow -->
    <div class="bg-radial-glow"></div>

    <!-- Global Sidebar Navigation -->
    <aside class="side-navigation" :class="platform">
      <div class="sidebar-top">
        <div class="logo-v2">
          <div class="logo-circle" :class="platform">
            <span class="material-symbols-outlined">auto_videocam</span>
          </div>
          <h2 class="logo-text-v2">ViralCraft</h2>
        </div>
      </div>

      <nav class="sidebar-links">
        <router-link to="/" class="sidebar-link">
          <span class="material-symbols-outlined">dashboard</span>
          <span>Dashboard</span>
        </router-link>
        <router-link to="/explore" class="sidebar-link">
          <span class="material-symbols-outlined">explore</span>
          <span>Explore</span>
        </router-link>
        <router-link to="/generator" class="sidebar-link" @click.prevent="!isAuthenticated && login()">
          <span class="material-symbols-outlined">bolt</span>
          <span>Generator</span>
        </router-link>
        <router-link to="/gallery" class="sidebar-link" @click.prevent="!isAuthenticated && login()">
          <span class="material-symbols-outlined">library_books</span>
          <span>My Library</span>
        </router-link>
      </nav>

      <div class="sidebar-footer">
        <!-- Platform switcher in sidebar as well for easy access -->
        <div class="sidebar-platform-info">
          <p class="opacity-30 text-[10px] uppercase tracking-widest font-bold mb-3">Active Engine</p>
          <div class="platform-mini-pill" :class="platform">
            {{ platform.toUpperCase() }} STUDIO
          </div>
        </div>
      </div>
    </aside>

    <div class="content-wrapper">
      <!-- Top Header (Clean) -->
      <header class="app-header-v2">
        <div class="header-right-controls">
          <!-- Platform Switcher -->
          <div class="platform-switcher-v2">
            <button 
              :class="{ active: platform === 'tiktok' }" 
              @click="setPlatform('tiktok')"
              class="switch-btn-v2 tiktok"
            >
              <span class="material-symbols-outlined">filter_frames</span>
              TikTok
            </button>
            <button 
              :class="{ active: platform === 'youtube' }" 
              @click="setPlatform('youtube')"
              class="switch-btn-v2 youtube"
            >
              <span class="material-symbols-outlined">play_circle</span>
              YouTube
            </button>
          </div>

          <!-- Gemini/User Block -->
          <div class="user-block-v2">
            <div class="gemini-btn" :class="platform">
              <span class="material-symbols-outlined">bolt</span>
              GEMINI AI
            </div>

            <div v-if="isAuthenticated" class="profile-area-v2" :class="{ 'is-open': isDropdownOpen }">
              <div class="avatar-trigger-v2" @click="toggleDropdown">
                <img :src="avatarUrl" alt="Avatar" class="avatar-v2">
                <span class="material-symbols-outlined text-lg opacity-40">expand_more</span>
              </div>

              <div v-if="isDropdownOpen" class="dropdown-content glass-card" @click.stop>
                <div class="user-info">
                  <div v-if="isEditingName" class="edit-name-group">
                    <input v-model="newName" type="text" class="edit-input" placeholder="Nhập tên mới..." @keyup.enter="handleUpdateProfile">
                    <div class="edit-actions">
                      <button @click="handleUpdateProfile" class="save-btn"><span class="material-symbols-outlined">done</span></button>
                      <button @click="isEditingName = false" class="cancel-btn"><span class="material-symbols-outlined">close</span></button>
                    </div>
                  </div>
                  <template v-else>
                    <div class="name-row">
                      <span class="name">{{ authUser?.name || 'Creator' }}</span>
                      <button @click="startEditing" class="edit-btn"><span class="material-symbols-outlined">edit</span></button>
                    </div>
                    <span class="address">{{ authUser?.wallet_address ? authUser.wallet_address.slice(0, 6) + '...' + authUser.wallet_address.slice(-4) : '' }}</span>
                  </template>
                </div>
                <div class="divider"></div>
                <button @click="handleLogout" class="logout-btn">
                  <span class="material-symbols-outlined">logout</span>
                  Logout
                </button>
              </div>
            </div>

            <button v-else @click="login" :disabled="isLoggingIn" class="header-login-btn" :class="platform">
              <span v-if="isLoggingIn" class="material-symbols-outlined spin">sync</span>
              <span v-else class="material-symbols-outlined">account_balance_wallet</span>
              {{ isLoggingIn ? 'Đang đăng nhập' : (isConnected ? 'Sign to Login' : 'Connect & Login') }}
            </button>
          </div>
        </div>
      </header>

      <!-- Main Content -->
      <main class="page-container-v2">
        <slot></slot>
      </main>
    </div>
  </div>
</template>

<style scoped>
.main-layout {
  min-height: 100vh;
  display: flex;
  background: #000;
  color: #fff;
  position: relative;
  overflow: hidden;
}

/* SIDEBAR NAVIGATION */
.side-navigation {
  width: 260px;
  height: 100vh;
  background: #0a0a0b;
  border-right: 1px solid rgba(255,255,255,0.03);
  display: flex;
  flex-direction: column;
  padding: 32px 20px;
  position: fixed;
  left: 0;
  top: 0;
  z-index: 100;
  transition: 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.logo-v2 { display: flex; align-items: center; gap: 14px; margin-bottom: 60px; padding-left: 12px; }
.logo-circle { width: 36px; height: 36px; border-radius: 10px; display: flex; align-items: center; justify-content: center; }
.logo-circle.tiktok { background: linear-gradient(135deg, #a14bff, #ff3f6c); }
.logo-circle.youtube { background: linear-gradient(135deg, #ff0000, #b30000); }
.logo-text-v2 { font-size: 1.4rem; font-weight: 800; letter-spacing: -0.04em; }

.sidebar-links { display: flex; flex-direction: column; gap: 8px; flex: 1; }
.sidebar-link {
  display: flex; align-items: center; gap: 16px; padding: 14px 16px; border-radius: 12px;
  color: #64748b; text-decoration: none; font-weight: 600; font-size: 0.95rem; transition: 0.2s;
}
.sidebar-link:hover { background: rgba(255,255,255,0.03); color: #fff; }
.sidebar-link.router-link-active { background: rgba(255,255,255,0.05); color: #fff; font-weight: 700; box-shadow: 0 4px 20px rgba(0,0,0,0.2); }
.side-navigation.tiktok .sidebar-link.router-link-active { border-right: 3px solid #a14bff; }
.side-navigation.youtube .sidebar-link.router-link-active { border-right: 3px solid #ff0000; }
.sidebar-link .material-symbols-outlined { font-size: 22px; }

.sidebar-footer { margin-top: auto; padding: 20px 12px; border-top: 1px solid rgba(255,255,255,0.03); }
.platform-mini-pill {
  padding: 8px 12px; border-radius: 8px; font-size: 0.65rem; font-weight: 800; letter-spacing: 0.1em;
  background: rgba(255,255,255,0.03); color: rgba(255,255,255,0.4); text-align: center;
}
.platform-mini-pill.tiktok { color: #e5b4ff; border: 1px solid rgba(161, 75, 255, 0.1); }
.platform-mini-pill.youtube { color: #ff0000; border: 1px solid rgba(255, 0, 0, 0.1); }

/* CONTENT WRAPPER */
.content-wrapper {
  flex: 1;
  margin-left: 260px;
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  position: relative;
}

.app-header-v2 {
  height: 80px;
  padding: 0 40px;
  display: flex;
  align-items: center;
  justify-content: flex-end; /* Push to right */
  position: sticky;
  top: 0;
  z-index: 90;
}

.header-right-controls { display: flex; align-items: center; gap: 40px; }

.platform-switcher-v2 {
  display: flex; background: #0a0a0b; padding: 4px; border-radius: 999px; border: 1px solid rgba(255,255,255,0.05);
}
.switch-btn-v2 {
  padding: 8px 20px; border-radius: 999px; border: none; font-size: 0.75rem; font-weight: 700;
  display: flex; align-items: center; gap: 8px; cursor: pointer; background: transparent; color: #64748b; transition: 0.3s;
}
.switch-btn-v2.active.tiktok { background: #fff; color: #000; box-shadow: 0 4px 15px rgba(161, 75, 255, 0.2); }
.switch-btn-v2.active.youtube { background: #ff0000; color: #fff; box-shadow: 0 4px 15px rgba(255, 0, 0, 0.2); }

.user-block-v2 { display: flex; align-items: center; gap: 24px; }
.gemini-btn {
  display: flex; align-items: center; gap: 8px; padding: 10px 18px; border-radius: 12px;
  background: #121214; border: 1px solid rgba(255,255,255,0.05); color: #fff;
  font-size: 0.75rem; font-weight: 800; cursor: pointer; transition: 0.3s;
}
.gemini-btn:hover { background: #1c1c1f; border-color: rgba(255,255,255,0.1); }
.gemini-btn .material-symbols-outlined { font-size: 18px; color: #a14bff; }
.theme-youtube .gemini-btn .material-symbols-outlined { color: #ff0000; }

.profile-area-v2 { position: relative; }
.avatar-trigger-v2 {
  display: flex; align-items: center; gap: 10px; cursor: pointer; padding: 4px; border-radius: 12px; transition: 0.2s;
}
.avatar-trigger-v2:hover { background: rgba(255,255,255,0.05); }
.avatar-v2 { width: 36px; height: 36px; border-radius: 50%; border: 2px solid rgba(255,255,255,0.1); }

.dropdown-content {
  position: absolute; top: calc(100% + 12px); right: 0; width: 220px; padding: 20px;
  background: #121214; border: 1px solid rgba(255,255,255,0.05); border-radius: 16px;
  z-index: 1000; opacity: 0; transform: translateY(10px); pointer-events: none;
  transition: 0.3s cubic-bezier(0.4, 0, 0.2, 1); backdrop-filter: blur(20px);
}
.is-open .dropdown-content { opacity: 1; transform: translateY(0); pointer-events: auto; }

.page-container-v2 { 
  flex: 1; 
  padding: 0 60px 40px 60px; /* Increased left padding for better spacing from sidebar */
}

.user-info { display: flex; flex-direction: column; gap: 4px; border-bottom: 1px solid rgba(255,255,255,0.05); padding-bottom: 16px; margin-bottom: 16px; }
.name { font-weight: 700; font-size: 1rem; color: #fff; }
.address { font-size: 0.75rem; color: rgba(255,255,255,0.3); font-family: monospace; }

.logout-btn {
  width: 100%; display: flex; align-items: center; gap: 10px; padding: 12px;
  border: none; background: transparent; color: #ef4444; font-weight: 700; font-size: 0.9rem;
  cursor: pointer; border-radius: 10px; transition: 0.2s;
}
.logout-btn:hover { background: rgba(239, 68, 68, 0.05); }

/* SPIN ANIMATION */
.spin { animation: spin 1s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }

.header-login-btn {
  padding: 10px 24px; border-radius: 999px; border: none; font-weight: 800; font-size: 0.85rem;
  background: #fff; color: #000; cursor: pointer; transition: 0.3s;
}
.header-login-btn:hover { transform: translateY(-2px); box-shadow: 0 8px 20px rgba(255,255,255,0.15); }
.header-login-btn.tiktok { background: #a14bff; color: #fff; }
.header-login-btn.youtube { background: #ff0000; color: #fff; }
</style>
