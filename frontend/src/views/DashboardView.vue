<script setup>
import { ref, onMounted, watch } from 'vue'
import { useUIStore } from '@/stores/ui.store'
import { useAuthStore } from '@/stores/auth.store'
import { useWalletStore } from '@/stores/wallet.store'
import { useAuth } from '@/composables/useAuth'
import { videoApi } from '@/api/video.api'
import { storeToRefs } from 'pinia'

const uiStore = useUIStore()
const { platform } = storeToRefs(uiStore)
const { user, login } = useAuth()
const walletStore = useWalletStore()
const { isConnected } = storeToRefs(walletStore)
const authStore = useAuthStore()
const { isAuthenticated } = storeToRefs(authStore)

const recentVideos = ref([])
const stats = ref({
  total: 0,
  storage: '0%'
})

const mockVideos = [
  { id: 'm1', title: 'Cách tạo video Viral với AI', status: 'completed', created_at: new Date(), thumbnail_url: 'https://images.unsplash.com/photo-1536240478700-b869070f9279?auto=format&fit=crop&q=80&w=400' },
  { id: 'm2', title: 'Tutorial: Từ ý tưởng đến sản phẩm', status: 'completed', created_at: new Date(), thumbnail_url: 'https://images.unsplash.com/photo-1492691523567-627395565cc5?auto=format&fit=crop&q=80&w=400' }
]

async function fetchDashboardData() {
  if (!isAuthenticated.value) {
    recentVideos.value = mockVideos
    stats.value = { total: 0, storage: '0%' }
    return
  }
  
  try {
    const res = await videoApi.getGallery({ limit: 4, platform: platform.value })
    recentVideos.value = res.data.data
    stats.value.total = res.data.total || 0
    stats.value.storage = '45%'
  } catch (e) {
    console.error('Failed to fetch dashboard data')
    recentVideos.value = mockVideos
  }
}

watch([isAuthenticated, platform], () => {
  fetchDashboardData()
})

onMounted(() => {
  fetchDashboardData()
})
</script>

<template>
  <div class="dashboard-view">
    <div class="welcome-section glass-card" :class="platform">
      <div class="welcome-text">
        <h1 v-if="isAuthenticated">Chào mừng trở lại, {{ user?.name || 'Creator' }}!</h1>
        <h1 v-else-if="isLoading">Đang xác thực...</h1>
        <h1 v-else-if="isConnected">Đang kết nối...</h1>
        <h1 v-else>Sẵn sàng tạo nội dung Viral?</h1>
        
        <p v-if="isAuthenticated">Hôm nay bạn muốn tạo video Viral nào?</p>
        <p v-else-if="isLoading">Vui lòng ký xác nhận trên ví của bạn để hoàn tất đăng nhập.</p>
        <p v-else-if="isConnected">Ví đã kết nối! Đang chuẩn bị môi trường sáng tạo cho bạn...</p>
        <p v-else>ViralCraft giúp bạn biến ý tưởng thành video triệu view bằng AI.</p>
      </div>
      <div class="welcome-actions">
        <!-- Khi đang load hoặc đã connect (đang chờ auto-sign), không cần hiện nút login thủ công nữa -->
        <div v-if="!isAuthenticated && (isConnected || isLoading)" class="loading-spinner-container">
          <div class="spinner"></div>
          <span>{{ isLoading ? 'Đang xác thực chữ ký...' : 'Đang khởi tạo ví...' }}</span>
        </div>
        <button v-else-if="!isAuthenticated" @click="login" class="action-btn primary">
          <span class="material-symbols-outlined">account_balance_wallet</span>
          Kết nối & Bắt đầu ngay
        </button>
        <router-link v-else to="/generator" class="action-btn primary">
          <span class="material-symbols-outlined">add_circle</span>
          Bắt đầu tạo Video mới
        </router-link>
      </div>
    </div>

    <div class="dashboard-grid">
      <!-- Recent Activity -->
      <section class="recent-section">
        <div class="section-header">
          <h3>{{ isAuthenticated ? 'Hoạt động gần đây' : 'Khám phá ViralCraft' }}</h3>
          <router-link to="/gallery" @click.prevent="!isAuthenticated && login()" class="view-all">Xem tất cả</router-link>
        </div>
        
        <div v-if="recentVideos.length === 0" class="empty-state glass-card">
          <span class="material-symbols-outlined emoji">movie_edit</span>
          <p>Chưa có video nào. Hãy bắt đầu tạo ngay!</p>
        </div>
        
        <div v-else class="video-grid">
          <div v-for="video in recentVideos" :key="video.id" class="video-card glass-card" @click="!isAuthenticated && login()">
            <div class="thumbnail">
              <img :src="video.thumbnail_url || '/placeholder.png'" alt="Thumbnail">
              <div class="badge" :class="video.status">{{ video.status === 'completed' ? 'Xong' : 'Đang xử lý' }}</div>
            </div>
            <div class="video-info">
              <h4>{{ video.title || 'Untitled Video' }}</h4>
              <span class="date">{{ new Date(video.created_at).toLocaleDateString() }}</span>
            </div>
          </div>
        </div>
      </section>

      <!-- Stats Sidebar -->
      <aside class="stats-sidebar">
        <div class="stats-card glass-card">
          <h3>Thống kê</h3>
          <div class="stat-item">
            <span class="label">Tổng số video</span>
            <span class="value">{{ stats.total }}</span>
          </div>
          <div class="stat-item">
            <span class="label">Dung lượng sử dụng</span>
            <div class="progress-bar">
              <div class="progress" :style="{ width: stats.storage }"></div>
            </div>
            <span class="value-sub">{{ stats.storage }}</span>
          </div>
          <div class="stat-item">
            <span class="label">Nền tảng chính</span>
            <span class="value capitalize">{{ platform }}</span>
          </div>
        </div>
        
        <!-- Guest Prompt Card -->
        <div v-if="!isAuthenticated" class="stats-card glass-card prompt-card" @click="login">
          <span class="material-symbols-outlined icon">auto_awesome</span>
          <h4>Unlock Pro Features</h4>
          <p>Đăng nhập để lưu lịch sử và quản lý video của bạn.</p>
        </div>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.dashboard-view {
  display: flex;
  flex-direction: column;
  gap: 32px;
  min-height: 100%;
}

.welcome-section {
  padding: 60px 40px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  position: relative;
  overflow: hidden;
  border-radius: 24px;
}

.welcome-section.tiktok { background: linear-gradient(135deg, rgba(161, 75, 255, 0.2), rgba(255, 63, 108, 0.2)); }
.welcome-section.youtube { background: linear-gradient(135deg, rgba(255, 0, 0, 0.15), rgba(179, 0, 0, 0.15)); }

.welcome-text h1 { font-size: 2rem; font-weight: 800; margin-bottom: 8px; }
.welcome-text p { color: rgba(255, 255, 255, 0.6); }

.action-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 24px;
  border-radius: 12px;
  font-weight: 700;
  text-decoration: none;
  transition: transform 0.2s;
}

.action-btn.primary { background: #fff; color: #000; }
.action-btn:hover { transform: translateY(-2px); }

.dashboard-grid {
  display: grid;
  grid-template-columns: 1fr 300px;
  gap: 32px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.section-header h3 { font-size: 1.1rem; font-weight: 700; }
.view-all { font-size: 0.85rem; color: var(--tiktok-primary); text-decoration: none; }
.theme-youtube .view-all { color: var(--youtube-primary); }

.video-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 20px;
}

.video-card {
  padding: 0;
  overflow: hidden;
  transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.video-card:hover { transform: scale(1.02); }

.thumbnail { position: relative; aspect-ratio: 9/16; background: #252529; }
.theme-youtube .thumbnail { aspect-ratio: 16/9; }
.thumbnail img { width: 100%; height: 100%; object-fit: cover; }

.badge {
  position: absolute;
  top: 12px;
  right: 12px;
  padding: 4px 8px;
  border-radius: 6px;
  font-size: 0.65rem;
  font-weight: 800;
  text-transform: uppercase;
}
.badge.completed { background: #10b981; color: #fff; }
.badge.processing { background: #40baf7; color: #fff; }

.video-info { padding: 16px; }
.video-info h4 { font-size: 0.95rem; font-weight: 700; margin-bottom: 4px; line-clamp: 1; -webkit-line-clamp: 1; display: -webkit-box; -webkit-box-orient: vertical; overflow: hidden; }
.video-info .date { font-size: 0.75rem; color: rgba(255, 255, 255, 0.4); }

.stats-card { padding: 24px; display: flex; flex-direction: column; gap: 20px; }
.stats-card h3 { font-size: 1rem; margin-bottom: 4px; }

.stat-item { display: flex; flex-direction: column; gap: 8px; }
.stat-item .label { font-size: 0.75rem; color: rgba(255, 255, 255, 0.5); font-weight: 600; text-transform: uppercase; }
.stat-item .value { font-size: 1.25rem; font-weight: 800; }

.progress-bar { height: 6px; background: rgba(255, 255, 255, 0.1); border-radius: 3px; overflow: hidden; }
.progress { height: 100%; background: linear-gradient(to right, #a14bff, #ff3f6c); transition: width 1s ease; }
.theme-youtube .progress { background: linear-gradient(to right, #ff0000, #b30000); }

.stat-item .value-sub { font-size: 0.75rem; color: rgba(255, 255, 255, 0.4); align-self: flex-end; }

.empty-state { padding: 64px; text-align: center; color: rgba(255, 255, 255, 0.3); }
.empty-state .emoji { font-size: 3rem; margin-bottom: 16px; opacity: 0.2; }

.prompt-card {
  background: linear-gradient(135deg, rgba(161, 75, 255, 0.1), rgba(255, 63, 108, 0.1));
  border: 1px dashed rgba(161, 75, 255, 0.3);
  cursor: pointer;
  text-align: center;
  transition: all 0.2s;
}

.theme-youtube .prompt-card {
  background: linear-gradient(135deg, rgba(255, 0, 0, 0.1), rgba(179, 0, 0, 0.1));
  border-color: rgba(255, 0, 0, 0.3);
}

.prompt-card:hover { transform: scale(1.02); filter: brightness(1.2); }
.prompt-card .icon { font-size: 2.5rem; color: var(--tiktok-primary); margin-bottom: 12px; }
.theme-youtube .prompt-card .icon { color: var(--youtube-primary); }
.prompt-card h4 { margin-bottom: 8px; color: #fff; }
.prompt-card p { font-size: 0.8rem; color: rgba(255, 255, 255, 0.5); line-height: 1.4; }

.loading-spinner-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: rgba(255, 255, 255, 0.7);
  font-size: 0.9rem;
  font-weight: 500;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid rgba(255, 255, 255, 0.1);
  border-top-color: var(--tiktok-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@keyframes bg-pulse {
  0% { box-shadow: 0 0 0 0 rgba(161, 75, 255, 0.4); }
  70% { box-shadow: 0 0 0 10px rgba(161, 75, 255, 0); }
  100% { box-shadow: 0 0 0 0 rgba(161, 75, 255, 0); }
}
</style>
