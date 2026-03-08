<template>
  <div class="result-preview" v-if="videoUrl">
    <!-- Header -->
    <div class="result-header">
      <div class="success-badge">✓ Video sẵn sàng!</div>
      <div v-if="savedPath" class="saved-path">
        <span class="folder-icon">📁</span>
        <span class="path-text">{{ savedPath }}</span>
      </div>
    </div>

    <!-- Video player -->
    <div class="video-wrap" :class="{ portrait: isPortrait }">
      <video :src="videoUrl" controls class="video-el" />
    </div>

    <!-- Actions -->
    <div class="action-row">
      <a :href="videoUrl" download class="btn-primary">
        ⬇️ Tải video
      </a>
      <a v-if="jobId" :href="`/api/download-subtitle/${jobId}`" download class="btn-secondary">
        💬 Tải phụ đề
      </a>
      <button class="btn-ghost" @click="$emit('reset')">
        🔄 Tạo video mới
      </button>
    </div>

    <!-- Copy link -->
    <div class="copy-row">
      <button class="copy-btn" @click="copyLink">
        {{ copied ? '✓ Đã copy!' : '🔗 Copy link' }}
      </button>
      <span v-if="copied" class="copied-msg">Link đã được copy vào clipboard</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  videoUrl: { type: String, default: null },
  jobId: { type: String, default: null },
  savedPath: { type: String, default: null },
  platform: { type: String, default: 'youtube' }
})

defineEmits(['reset'])

const copied = ref(false)
const isPortrait = computed(() => props.platform === 'tiktok')

const copyLink = async () => {
  try {
    await navigator.clipboard.writeText(window.location.origin + props.videoUrl)
    copied.value = true
    setTimeout(() => { copied.value = false }, 3000)
  } catch (err) {
    console.error('Failed to copy:', err)
  }
}
</script>

<style scoped>
.result-preview { display: flex; flex-direction: column; gap: 16px; }

.result-header { display: flex; flex-direction: column; gap: 8px; }

.success-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 16px;
  background: rgba(72,199,142,0.15);
  border: 1px solid rgba(72,199,142,0.3);
  border-radius: 20px;
  color: #48c78e;
  font-weight: 700;
  font-size: 0.9rem;
  width: fit-content;
}

.saved-path {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 8px;
  padding: 8px 12px;
  font-size: 0.8rem;
}
.folder-icon { flex-shrink: 0; }
.path-text {
  color: rgba(255,255,255,0.6);
  font-family: 'Courier New', monospace;
  word-break: break-all;
}

.video-wrap {
  width: 100%;
  background: #000;
  border-radius: 12px;
  overflow: hidden;
}
.video-wrap.portrait {
  max-width: 280px;
  margin: 0 auto;
}
.video-el { width: 100%; height: auto; display: block; }

.action-row { display: flex; flex-wrap: wrap; gap: 10px; }

.btn-primary, .btn-secondary, .btn-ghost {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 18px;
  border-radius: 10px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  text-decoration: none;
  border: none;
  transition: all 0.2s;
}
.btn-primary { background: rgba(99,179,255,0.2); color: #63b3ff; border: 1px solid rgba(99,179,255,0.3); }
.btn-primary:hover { background: rgba(99,179,255,0.3); }
.btn-secondary { background: rgba(167,139,250,0.15); color: #a78bfa; border: 1px solid rgba(167,139,250,0.25); }
.btn-secondary:hover { background: rgba(167,139,250,0.25); }
.btn-ghost { background: rgba(255,255,255,0.05); color: rgba(255,255,255,0.6); border: 1px solid rgba(255,255,255,0.1); }
.btn-ghost:hover { background: rgba(255,255,255,0.1); }

.copy-row { display: flex; align-items: center; gap: 12px; }
.copy-btn {
  padding: 8px 14px;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 8px;
  color: rgba(255,255,255,0.55);
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s;
}
.copy-btn:hover { background: rgba(255,255,255,0.1); color: #fff; }
.copied-msg { font-size: 0.78rem; color: #48c78e; }
</style>
