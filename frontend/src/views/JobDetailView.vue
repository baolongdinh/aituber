<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { videoApi } from '@/api/video.api'
import { useUIStore } from '@/stores/ui.store'
import { storeToRefs } from 'pinia'

const route = useRoute()
const router = useRouter()
const uiStore = useUIStore()
const { platform } = storeToRefs(uiStore)

const job = ref(null)
const isLoading = ref(true)

async function fetchJobDetail() {
  try {
    const res = await videoApi.getStatus(route.params.id)
    job.value = res.data.data
  } catch (e) {
    console.error('Failed to fetch job detail')
  } finally {
    isLoading.value = false
  }
}

function handleDownload() {
  if (job.value?.video_url) {
    window.open(job.value.video_url, '_blank')
  }
}

onMounted(() => {
  fetchJobDetail()
})
</script>

<template>
  <div class="job-detail-view">
    <div v-if="isLoading" class="loading-state">
      <div class="spinner"></div>
      <p>Đang tải thông tin video...</p>
    </div>

    <div v-else-if="job" class="detail-container">
      <!-- Player Section -->
      <section class="player-section glass-card" :class="platform">
        <div class="video-wrapper" :class="platform">
          <video 
            v-if="job.video_url" 
            controls 
            :src="job.video_url" 
            class="main-video"
          ></video>
          <div v-else class="no-video">
            <span class="material-symbols-outlined emoji">videocam_off</span>
            <p>Video chưa sẵn sàng hoặc có lỗi xảy ra.</p>
          </div>
        </div>
        
        <div class="player-controls">
          <div class="job-meta">
            <h1 class="text-xl font-bold">{{ job.title || 'Untitled Video' }}</h1>
            <p class="text-xs text-slate-400">ID: {{ job.job_id }} • {{ new Date(job.created_at).toLocaleString() }}</p>
          </div>
          <div class="actions">
            <button @click="handleDownload" class="icon-btn download">
              <span class="material-symbols-outlined">download</span>
              Tải xuống
            </button>
            <button class="icon-btn share">
              <span class="material-symbols-outlined">share</span>
              Chia sẻ
            </button>
          </div>
        </div>
      </section>

      <!-- Info Sidebar -->
      <aside class="info-sidebar">
        <div class="info-card glass-card">
          <h3>Thông số Video</h3>
          <div class="info-item">
            <span class="label">Chủ đề</span>
            <p class="val">{{ job.topic }}</p>
          </div>
          <div class="info-item">
            <span class="label">Nền tảng</span>
            <span class="val capitalize">{{ job.platform || platform }}</span>
          </div>
          <div class="info-item">
            <span class="label">Giọng đọc</span>
            <span class="val">{{ job.voice || 'Mặc định' }}</span>
          </div>
          <div class="info-item">
            <span class="label">Trạng thái</span>
            <div class="status-badge" :class="job.status">
              {{ job.status === 'completed' ? 'Thành công' : 'Thất bại' }}
            </div>
          </div>
        </div>

        <div class="ai-insight-card glass-card">
          <div class="insight-header">
            <span class="material-symbols-outlined icon">bolt</span>
            <h4>Gemini AI Insight</h4>
          </div>
          <p>Video này được tối ưu hóa cho thuật toán {{ platform === 'tiktok' ? 'TikTok FYP' : 'YouTube Shorts' }} với các từ khóa trending được tích hợp tự động.</p>
        </div>
      </aside>
    </div>

    <div v-else class="error-state glass-card">
      <span class="material-symbols-outlined emoji">error</span>
      <h3>Không tìm thấy công việc</h3>
      <p>Yêu cầu không tồn tại hoặc bạn không có quyền truy cập.</p>
      <router-link to="/" class="back-link">Quay lại Dashboard</router-link>
    </div>
  </div>
</template>

<style scoped>
.job-detail-view { display: flex; flex-direction: column; gap: 32px; }

.loading-state, .error-state { padding: 80px; text-align: center; display: flex; flex-direction: column; align-items: center; gap: 20px; }
.spinner { width: 40px; height: 40px; border: 4px solid rgba(255, 255, 255, 0.1); border-top-color: var(--tiktok-primary); border-radius: 50%; animation: spin 1s linear infinite; }

.detail-container { display: grid; grid-template-columns: 1fr 320px; gap: 32px; align-items: start; }

.player-section { padding: 24px; display: flex; flex-direction: column; gap: 24px; }
.player-section.tiktok { background: linear-gradient(135deg, rgba(161, 75, 255, 0.1), rgba(10, 10, 12, 0)); }
.player-section.youtube { background: linear-gradient(135deg, rgba(255, 0, 0, 0.1), rgba(10, 10, 12, 0)); }

.video-wrapper { 
  width: 100%; 
  background: #000; 
  border-radius: 12px; 
  overflow: hidden; 
  box-shadow: 0 20px 50px rgba(0,0,0,0.5); 
  display: flex;
  align-items: center;
  justify-content: center;
}
.video-wrapper.tiktok { aspect-ratio: 9/16; max-width: 400px; margin: 0 auto; }
.video-wrapper.youtube { aspect-ratio: 16/9; }

.main-video { width: 100%; height: 100%; }

.no-video { text-align: center; color: rgba(255, 255, 255, 0.2); padding: 40px; }
.no-video .emoji { font-size: 3rem; margin-bottom: 12px; }

.player-controls { display: flex; justify-content: space-between; align-items: center; }

.actions { display: flex; gap: 12px; }
.icon-btn { 
  display: flex; align-items: center; gap: 8px; padding: 10px 20px; 
  border-radius: 10px; border: 1px solid rgba(255, 255, 255, 0.1); 
  background: rgba(255, 255, 255, 0.05); color: #fff; cursor: pointer; 
  font-size: 0.85rem; font-weight: 700; transition: all 0.2s;
}
.icon-btn:hover { background: rgba(255, 255, 255, 0.1); border-color: rgba(255, 255, 255, 0.2); }
.icon-btn.download { background: var(--tiktok-primary); border: none; }
.theme-youtube .icon-btn.download { background: var(--youtube-primary); }

.info-sidebar { display: flex; flex-direction: column; gap: 24px; }
.info-card { padding: 24px; display: flex; flex-direction: column; gap: 20px; }
.info-card h3 { font-size: 1rem; font-weight: 700; }

.info-item { display: flex; flex-direction: column; gap: 4px; }
.info-item .label { font-size: 0.7rem; color: rgba(255, 255, 255, 0.4); text-transform: uppercase; font-weight: 700; }
.info-item .val { font-size: 0.9rem; font-weight: 600; }

.status-badge { 
  display: inline-block; padding: 4px 10px; border-radius: 6px; 
  font-size: 0.7rem; font-weight: 800; text-transform: uppercase; 
}
.status-badge.completed { background: rgba(16, 185, 129, 0.15); color: #10b981; }
.status-badge.failed { background: rgba(239, 68, 68, 0.15); color: #ef4444; }

.ai-insight-card { 
  padding: 24px; background: linear-gradient(135deg, rgba(161, 75, 255, 0.1), rgba(255, 63, 108, 0.1)); 
  border-color: rgba(161, 75, 255, 0.2); 
}
.insight-header { display: flex; align-items: center; gap: 8px; margin-bottom: 12px; color: var(--tiktok-primary); }
.insight-header h4 { font-size: 0.85rem; font-weight: 800; text-transform: uppercase; }
.insight-header .icon { font-size: 1.25rem; }
.ai-insight-card p { font-size: 0.8rem; color: rgba(255, 255, 255, 0.6); line-height: 1.6; }

@keyframes spin { to { transform: rotate(360deg); } }

.back-link { margin-top: 12px; color: var(--tiktok-primary); text-decoration: none; font-weight: 700; }
</style>
