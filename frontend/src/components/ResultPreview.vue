<template>
  <div class="result-preview" v-if="videoUrl || isSeries">
    <!-- SINGLE VIDEO MODE OR FIRST SERIES PART -->
    <template v-if="!isSeries">
      <div class="result-header">
        <div class="success-badge">✓ Video sẵn sàng!</div>
        <div v-if="savedPath" class="saved-path">
          <span class="folder-icon">📁</span>
          <span class="path-text">{{ savedPath }}</span>
        </div>
      </div>

      <div class="video-wrap" :class="{ portrait: isPortrait }">
        <video :src="videoUrl" controls class="video-el" />
      </div>

      <div class="action-row">
        <a :href="videoUrl" download class="btn-primary">⬇️ Tải video</a>
        <a v-if="jobId" :href="`/api/download-subtitle/${jobId}`" download class="btn-secondary">💬 Tải phụ đề</a>
        <button class="btn-ghost" @click="$emit('reset')">🔄 Tạo video mới</button>
      </div>

      <div class="copy-row">
        <button class="copy-btn" @click="copyLink(videoUrl)">{{ copied ? '✓ Đã copy!' : '🔗 Copy link' }}</button>
        <span v-if="copied" class="copied-msg">Link đã copy!</span>
      </div>
    </template>

    <!-- SERIES MODE (Multi-download) -->
    <template v-else>
      <div class="result-header">
        <div class="success-badge">🎬 Series đang được tạo</div>
        <button class="btn-ghost" @click="$emit('reset')" style="padding: 6px 12px; margin-top: 8px;">🔄 Tạo mới</button>
      </div>
      
      <div class="series-downloads">
        <div v-for="part in completedParts" :key="part.part_index" class="series-part-result">
          <div class="part-info">
            <span class="part-index">Tập {{ part.part_index + 1 }}</span>
            <span class="part-title">{{ part.title }}</span>
          </div>
          <div class="part-actions">
            <a :href="part.video_url" download class="btn-primary-mini">⬇️ Tải</a>
            <button class="btn-ghost-mini" @click="copyLink(part.video_url)">🔗 Copy</button>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  videoUrl: { type: String, default: null },
  jobId: { type: String, default: null },
  savedPath: { type: String, default: null },
  platform: { type: String, default: 'youtube' },
  isSeries: { type: Boolean, default: false },
  parts: { type: Array, default: () => [] }
})

defineEmits(['reset'])

const copied = ref(false)
const isPortrait = computed(() => props.platform === 'tiktok')

const completedParts = computed(() => {
  return (props.parts || []).filter(p => p.status === 'completed' && p.video_url)
})

const copyLink = async (url) => {
  if (!url) return
  try {
    await navigator.clipboard.writeText(window.location.origin + url)
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

/* SERIES MODE STYLES */
.series-downloads {
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-height: 400px;
  overflow-y: auto;
  padding-right: 6px;
}
.series-downloads::-webkit-scrollbar { width: 6px; }
.series-downloads::-webkit-scrollbar-track { background: rgba(255,255,255,0.02); border-radius: 10px; }
.series-downloads::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.15); border-radius: 10px; }

.series-part-result {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 10px;
  padding: 12px 14px;
}
.part-info { display: flex; flex-direction: column; gap: 4px; }
.part-index { font-size: 0.75rem; color: #63b3ff; font-weight: 700; }
.part-title { font-size: 0.85rem; font-weight: 600; color: #fff; max-width: 200px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.part-actions { display: flex; gap: 8px; }
.btn-primary-mini, .btn-ghost-mini {
  padding: 6px 10px;
  border-radius: 8px;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  text-decoration: none;
  border: none;
  transition: all 0.2s;
}
.btn-primary-mini { background: rgba(99,179,255,0.15); color: #63b3ff; }
.btn-primary-mini:hover { background: rgba(99,179,255,0.25); }
.btn-ghost-mini { background: rgba(255,255,255,0.05); color: rgba(255,255,255,0.6); }
.btn-ghost-mini:hover { background: rgba(255,255,255,0.15); color: #fff; }
</style>
