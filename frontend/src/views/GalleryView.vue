<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { videoApi } from '@/api/video.api'
import { useUIStore } from '@/stores/ui.store'
import { useAuthStore } from '@/stores/auth.store'
import { storeToRefs } from 'pinia'
import { useRoute } from 'vue-router'

const uiStore = useUIStore()
const { platform } = storeToRefs(uiStore)
const authStore = useAuthStore()
const { isAuthenticated } = storeToRefs(authStore)
const route = useRoute()

const videos = ref([])
const isLoading = ref(true)

const isExploreMode = computed(() => route.name === 'explore')

async function fetchData() {
  isLoading.value = true
  try {
    const params = { limit: 20, platform: platform.value }
    let res
    if (isExploreMode.value) {
      res = await videoApi.getExplore(params)
    } else {
      if (!isAuthenticated.value) return
      res = await videoApi.getGallery(params)
    }
    videos.value = res.data.data
  } catch (e) {
    console.error('Failed to fetch data')
  } finally {
    isLoading.value = false
  }
}

async function handleTogglePublic(e, video) {
  e.preventDefault()
  e.stopPropagation()
  try {
    const res = await videoApi.togglePublic(video.id || video.job_id)
    video.is_public = res.data.data.is_public
  } catch (err) {
    console.error('Failed to toggle public status')
  }
}

watch([() => route.name, isAuthenticated, platform], () => {
  fetchData()
})

onMounted(() => {
  fetchData()
})
</script>

<template>
  <div class="gallery-view">
    <div class="header-section">
      <h1 class="text-3xl font-black">{{ isExploreMode ? 'Khám phá cộng đồng' : 'Thư viện của tôi' }}</h1>
      <p class="text-slate-400">{{ isExploreMode ? 'Những nội dung Viral đỉnh nhất từ ViralCraft' : 'Tất cả các video Viral bạn đã tạo' }}</p>
    </div>

    <div v-if="isLoading" class="loading-grid">
      <div v-for="i in 8" :key="i" class="skeleton-card glass-card"></div>
    </div>

    <div v-else-if="videos.length === 0" class="empty-state glass-card">
      <span class="material-symbols-outlined emoji">movie_filter</span>
      <h3>Chưa có video nào</h3>
      <p>Bắt đầu tạo video đầu tiên của bạn ngay!</p>
      <router-link to="/generator" class="btn">Tạo Video ngay</router-link>
    </div>

    <div v-else class="video-grid">
      <router-link 
        v-for="video in videos" 
        :key="video.job_id" 
        :to="'/job/' + (video.job_id || video.id)"
        class="video-card glass-card"
      >
        <div class="thumbnail" :class="platform">
          <img :src="video.thumbnail_url || '/placeholder.png'" alt="Thumbnail">
          <div class="status-badge" :class="video.status">
            {{ video.status === 'completed' ? 'Xong' : 'Đang xử lý' }}
          </div>
          <button 
            v-if="!isExploreMode && video.status === 'completed'" 
            class="publish-toggle"
            :class="{ 'is-public': video.is_public }"
            @click="handleTogglePublic($event, video)"
            :title="video.is_public ? 'Make Private' : 'Make Public'"
          >
            <span class="material-symbols-outlined">{{ video.is_public ? 'public' : 'public_off' }}</span>
          </button>
        </div>
        <div class="video-info">
          <h4>{{ video.title || 'Untitled Video' }}</h4>
          <div class="meta">
            <span>{{ new Date(video.created_at).toLocaleDateString() }}</span>
            <span class="dot">·</span>
            <span class="capitalize">{{ video.platform || 'General' }}</span>
          </div>
        </div>
      </router-link>
    </div>
  </div>
</template>

<style scoped>
.gallery-view { display: flex; flex-direction: column; gap: 32px; }
.header-section { margin-bottom: 8px; }

.video-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 24px;
}

.video-card { padding: 0; overflow: hidden; text-decoration: none; color: inherit; transition: transform 0.3s; }
.video-card:hover { transform: translateY(-8px); }

.thumbnail { position: relative; aspect-ratio: 9/16; background: #000; overflow: hidden; }
.thumbnail.youtube { aspect-ratio: 16/9; }
.thumbnail img { width: 100%; height: 100%; object-fit: cover; transition: transform 0.5s; }
.video-card:hover img { transform: scale(1.1); }

.status-badge {
  position: absolute; top: 12px; right: 12px; padding: 4px 10px;
  border-radius: 6px; font-size: 0.65rem; font-weight: 800; text-transform: uppercase;
}
.status-badge.completed { background: #10b981; color: #fff; }
.status-badge.processing { background: #3b82f6; color: #fff; }

.publish-toggle {
  position: absolute; bottom: 12px; right: 12px;
  width: 36px; height: 36px; border-radius: 50%;
  background: rgba(0, 0, 0, 0.5); border: 1px solid rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.5); display: flex; align-items: center; justify-content: center;
  cursor: pointer; transition: all 0.2s; z-index: 10;
}
.publish-toggle:hover { background: rgba(0, 0, 0, 0.7); box-shadow: 0 0 10px rgba(0,0,0,0.5); }
.publish-toggle.is-public { color: #10b981; border-color: #10b981; background: rgba(16, 185, 129, 0.1); }
.publish-toggle .material-symbols-outlined { font-size: 20px; }

.video-info { padding: 20px; }
.video-info h4 { font-size: 1rem; font-weight: 700; margin-bottom: 8px; line-clamp: 2; -webkit-line-clamp: 2; display: -webkit-box; -webkit-box-orient: vertical; overflow: hidden; }
.meta { display: flex; align-items: center; gap: 8px; font-size: 0.75rem; color: rgba(255, 255, 255, 0.4); font-weight: 600; }
.dot { font-weight: 900; }

.loading-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 24px; }
.skeleton-card { height: 400px; animation: pulse 1.5s infinite; }

.empty-state { padding: 80px; text-align: center; border: 2px dashed rgba(255, 255, 255, 0.05); }
.empty-state .emoji { font-size: 4rem; opacity: 0.2; margin-bottom: 24px; }

.btn { 
  display: inline-block; margin-top: 24px; padding: 12px 24px; 
  background: #fff; color: #000; border-radius: 10px; 
  font-weight: 700; text-decoration: none; 
  transition: all 0.3s ease;
  box-shadow: 0 4px 12px rgba(0,0,0,0.2);
}

/* Platform-specific button colors */
.theme-tiktok .btn {
  background: linear-gradient(135deg, #a14bff 0%, #ff0050 100%);
  color: #fff;
  box-shadow: 0 4px 12px rgba(161, 75, 255, 0.3);
}

.theme-youtube .btn {
  background: linear-gradient(135deg, #ff0000 0%, #cc0000 100%);
  color: #fff;
  box-shadow: 0 4px 12px rgba(255, 0, 0, 0.3);
}

.btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0,0,0,0.3);
}

.theme-tiktok .btn:hover {
  box-shadow: 0 6px 16px rgba(161, 75, 255, 0.5);
}

.theme-youtube .btn:hover {
  box-shadow: 0 6px 16px rgba(255, 0, 0, 0.5);
}

@keyframes pulse { 0% { opacity: 0.6; } 50% { opacity: 0.3; } 100% { opacity: 0.6; } }
</style>
