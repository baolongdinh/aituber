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

    <!-- Top Navigation -->
    <header class="app-header">
      <div class="header-inner">
        <div class="left-section">
          <div class="logo">
            <div class="logo-icon" :class="platform">
              <span class="material-symbols-outlined">auto_videocam</span>
            </div>
            <h2 class="logo-text" :class="'text-gradient-' + platform">ViralCraft</h2>
            <span class="platform-label">{{ platform === 'tiktok' ? 'TIKTOK CREATOR' : 'YOUTUBE STUDIO' }}</span>
          </div>
          
          <nav class="desktop-nav">
            <router-link to="/">Dashboard</router-link>
            <router-link to="/explore">Explore</router-link>
            <router-link to="/generator" @click.prevent="!isAuthenticated && login()">Generator</router-link>
            <router-link to="/gallery" @click.prevent="!isAuthenticated && login()">My Library</router-link>
          </nav>
        </div>

        <div class="right-section">
          <!-- Platform Switcher -->
          <div class="platform-switcher">
            <button 
              :class="{ active: platform === 'tiktok' }" 
              @click="setPlatform('tiktok')"
              class="switch-btn tiktok"
            >
              <span class="material-symbols-outlined">filter_frames</span>
              TikTok
            </button>
            <button 
              :class="{ active: platform === 'youtube' }" 
              @click="setPlatform('youtube')"
              class="switch-btn youtube"
            >
              <span class="material-symbols-outlined">play_circle</span>
              YouTube
            </button>
          </div>

          <!-- User Profile -->
          <!-- User Profile / Auth Action -->
          <div class="user-block">
            <div class="gemini-badge">
              <span class="material-symbols-outlined">bolt</span>
              <span>Gemini AI</span>
            </div>
            
            <div v-if="isAuthenticated" class="profile-area" :class="{ 'is-open': isDropdownOpen }">
              <div class="avatar-trigger" @click="toggleDropdown">
                <div class="avatar">
                  <img :src="avatarUrl" alt="Avatar">
                </div>
                <span class="material-symbols-outlined expand-icon">expand_more</span>
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
      </div>
    </header>

    <!-- Main Content -->
    <main class="page-container">
      <slot></slot>
    </main>
  </div>
</template>

<style scoped>
.main-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  position: relative;
  z-index: 1;
}

.app-header {
  height: 64px;
  background: rgba(10, 10, 12, 0.6);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-inner {
  max-width: 1440px;
  margin: 0 auto;
  height: 100%;
  padding: 0 32px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.left-section, .right-section {
  display: flex;
  align-items: center;
  gap: 32px;
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}

.logo-icon.tiktok { background: linear-gradient(135deg, #a14bff, #ff3f6c); }
.logo-icon.youtube { background: linear-gradient(135deg, #ff0000, #b30000); }

.logo-text { font-size: 1.25rem; font-weight: 800; letter-spacing: -0.02em; }
.platform-label { font-size: 0.65rem; color: rgba(255, 255, 255, 0.3); font-weight: 700; margin-left: -4px; border-left: 1px solid rgba(255, 255, 255, 0.1); padding-left: 8px; }

.desktop-nav {
  display: flex;
  gap: 24px;
}

.desktop-nav a {
  color: #94a3b8;
  font-size: 0.875rem;
  font-weight: 500;
  text-decoration: none;
  transition: color 0.2s;
}

.desktop-nav a:hover, .desktop-nav a.router-link-active { color: #fff; }

.platform-switcher {
  display: flex;
  background: #121214;
  padding: 4px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.switch-btn {
  padding: 6px 16px;
  border-radius: 999px;
  border: none;
  font-size: 0.75rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  background: transparent;
  color: #64748b;
  transition: all 0.2s;
}

.switch-btn:hover { color: #fff; }

.switch-btn.active.tiktok { background: #a14bff; color: #fff; }
.switch-btn.active.youtube { background: #ff0000; color: #fff; }

.user-block {
  display: flex;
  align-items: center;
  gap: 16px;
}

.gemini-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: rgba(161, 75, 255, 0.1);
  border: 1px solid rgba(161, 75, 255, 0.2);
  border-radius: 8px;
  color: #a14bff;
  font-size: 0.65rem;
  font-weight: 800;
  text-transform: uppercase;
}

.theme-youtube .gemini-badge { background: rgba(255, 0, 0, 0.1); border-color: rgba(255, 0, 0, 0.2); color: #ff0000; }

.avatar-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px;
  padding-right: 12px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 999px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid rgba(255, 255, 255, 0.05);
}
.avatar-trigger:hover { background: rgba(255, 255, 255, 0.1); border-color: rgba(255, 255, 255, 0.1); }
.is-open .avatar-trigger { background: rgba(255, 255, 255, 0.1); border-color: var(--tiktok-primary); }
.theme-youtube .is-open .avatar-trigger { border-color: var(--youtube-primary); }

.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  overflow: hidden;
  border: 2px solid rgba(161, 75, 255, 0.2);
}

.theme-youtube .avatar { border-color: rgba(255, 0, 0, 0.2); }

.expand-icon { font-size: 18px; color: rgba(255, 255, 255, 0.3); transition: transform 0.3s; }
.is-open .expand-icon { transform: rotate(180deg); color: #fff; }

.profile-area { position: relative; }

.dropdown-content {
  position: absolute;
  top: calc(100% + 12px);
  right: 0;
  width: 200px;
  padding: 16px;
  z-index: 1000;
  opacity: 0;
  transform: translateY(10px);
  pointer-events: none;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
.is-open .dropdown-content { opacity: 1; transform: translateY(0); pointer-events: auto; }

.page-container {
  flex: 1;
  max-width: 1440px;
  margin: 0 auto;
  width: 100%;
  padding: 32px;
}

.icon-btn.download { background: var(--tiktok-primary); border: none; }
.theme-youtube .icon-btn.download { background: var(--youtube-primary); }

.header-login-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-radius: 999px;
  border: none;
  background: var(--tiktok-primary);
  color: #fff;
  font-size: 0.85rem;
  font-weight: 700;
  cursor: pointer;
  transition: transform 0.2s, filter 0.2s;
}

.theme-youtube .header-login-btn { background: var(--youtube-primary); }
.header-login-btn:hover { transform: translateY(-2px); filter: brightness(1.1); }
.header-login-btn:disabled { opacity: 0.7; cursor: not-allowed; }

.spin { animation: spin 1s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }

.user-info { display: flex; flex-direction: column; gap: 2px; max-width: 120px; }
.user-info .name { font-size: 0.85rem; font-weight: 700; color: #fff; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.user-info .address { font-size: 0.7rem; color: rgba(255, 255, 255, 0.4); font-family: monospace; }

.name-row { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.edit-btn { background: transparent; border: none; color: rgba(255, 255, 255, 0.4); cursor: pointer; padding: 2px; border-radius: 4px; display: flex; align-items: center; }
.edit-btn:hover { color: #fff; background: rgba(255, 255, 255, 0.1); }
.edit-btn .material-symbols-outlined { font-size: 14px; }

.edit-name-group { display: flex; flex-direction: column; gap: 8px; }
.edit-input { background: rgba(0, 0, 0, 0.2); border: 1px solid rgba(255, 255, 255, 0.1); border-radius: 6px; padding: 6px 10px; color: #fff; font-size: 0.8rem; width: 100%; outline: none; }
.edit-input:focus { border-color: var(--tiktok-primary); }
.edit-actions { display: flex; gap: 4px; justify-content: flex-end; }
.edit-actions button { background: transparent; border: none; cursor: pointer; padding: 4px; border-radius: 4px; display: flex; align-items: center; }
.save-btn { color: #10b981; }
.cancel-btn { color: #ef4444; }
.edit-actions button:hover { background: rgba(255, 255, 255, 0.1); }
.edit-actions .material-symbols-outlined { font-size: 16px; }

.divider { height: 1px; background: rgba(255, 255, 255, 0.05); margin: 12px 0; }
.logout-btn { 
  width: 100%; display: flex; align-items: center; gap: 8px; 
  padding: 8px; border: none; background: transparent; 
  color: #ef4444; font-size: 0.85rem; font-weight: 600; cursor: pointer; 
  border-radius: 8px; transition: background 0.2s;
}
.logout-btn:hover { background: rgba(239, 68, 68, 0.1); }
</style>
