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
.result-preview {
  display: flex;
  flex-direction: column;
  gap: 24px;
  animation: fadeIn 0.5s var(--transition-smooth);
}

@keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

.result-header {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.success-badge {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 8px 18px;
  background: rgba(72,199,142,0.1);
  border: 1px solid rgba(72,199,142,0.2);
  border-radius: var(--radius-full);
  color: #48c78e;
  font-weight: 800;
  font-size: 0.95rem;
  width: fit-content;
  box-shadow: 0 4px 15px rgba(72,199,142,0.1);
}

.saved-path {
  display: flex;
  align-items: center;
  gap: 12px;
  background: rgba(0,0,0,0.2);
  border: 1px solid var(--card-border);
  border-radius: var(--radius-md);
  padding: 12px 16px;
  font-size: 0.8rem;
}

.folder-icon { font-size: 1.2rem; }
.path-text {
  color: var(--text-muted);
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  word-break: break-all;
  line-height: 1.4;
}

/* ── Video Player ── */
.video-wrap {
  width: 100%;
  background: #000;
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: 0 12px 40px rgba(0,0,0,0.4);
  border: 1px solid var(--card-border);
  position: relative;
  aspect-ratio: 16 / 9;
}

.video-wrap.portrait {
  max-width: 320px;
  aspect-ratio: 9 / 16;
  margin: 0 auto;
}

.video-el {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: contain;
}

/* ── Actions ── */
.action-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.btn-primary, .btn-secondary, .btn-ghost {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 14px 22px;
  border-radius: var(--radius-md);
  font-size: 0.95rem;
  font-weight: 700;
  cursor: pointer;
  text-decoration: none;
  border: 1px solid transparent;
  transition: var(--transition-bounce);
  flex: 1;
  min-width: 140px;
}

.btn-primary {
  background: linear-gradient(135deg, #63b3ff, #448aff);
  color: #fff;
  box-shadow: 0 8px 20px rgba(99, 179, 255, 0.3);
}

.btn-secondary {
  background: rgba(167, 139, 250, 0.1);
  color: #a78bfa;
  border-color: rgba(167, 139, 250, 0.2);
}

.btn-ghost {
  background: rgba(255,255,255,0.05);
  color: var(--text-muted);
  border-color: var(--glass-border);
}

.btn-primary:hover { transform: translateY(-3px); box-shadow: 0 12px 25px rgba(99, 179, 255, 0.4); }
.btn-secondary:hover { background: rgba(167, 139, 250, 0.15); border-color: rgba(167, 139, 250, 0.4); }
.btn-ghost:hover { background: rgba(255,255,255,0.08); color: #fff; }

.copy-row {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 8px 16px;
  background: rgba(255,255,255,0.02);
  border-radius: var(--radius-md);
  border: 1px dashed var(--glass-border);
}

.copy-btn {
  background: none;
  border: none;
  color: var(--text-dim);
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  padding: 6px 0;
  transition: var(--transition-fast);
}

.copy-btn:hover { color: #fff; }
.copied-msg { font-size: 0.8rem; font-weight: 700; color: #48c78e; animation: pulse 1s infinite alternate; }

@keyframes pulse { from { opacity: 0.6; } to { opacity: 1; } }

/* ── Series List ── */
.series-downloads {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 450px;
  overflow-y: auto;
  padding-right: 8px;
}

.series-downloads::-webkit-scrollbar { width: 6px; }
.series-downloads::-webkit-scrollbar-track { background: transparent; }
.series-downloads::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.1); border-radius: 10px; }

.series-part-result {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(0,0,0,0.2);
  border: 1px solid var(--card-border);
  border-radius: var(--radius-md);
  padding: 16px;
  transition: var(--transition-smooth);
}

.series-part-result:hover { border-color: rgba(255,255,255,0.12); transform: translateX(4px); }

.part-info { display: flex; flex-direction: column; gap: 4px; }
.part-index { font-size: 0.75rem; color: #63b3ff; font-weight: 800; text-transform: uppercase; letter-spacing: 0.05em; }
.part-title { font-size: 0.95rem; font-weight: 700; color: #fff; max-width: 240px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.part-actions { display: flex; gap: 10px; }

.btn-primary-mini, .btn-ghost-mini {
  padding: 8px 14px;
  border-radius: 8px;
  font-size: 0.8rem;
  font-weight: 700;
  cursor: pointer;
  text-decoration: none;
  transition: var(--transition-fast);
  border: none;
}

.btn-primary-mini { background: rgba(99,179,255,0.15); color: #63b3ff; }
.btn-primary-mini:hover { background: #63b3ff; color: #000; }

.btn-ghost-mini { background: rgba(255,255,255,0.05); color: var(--text-dim); }
.btn-ghost-mini:hover { background: rgba(255,255,255,0.1); color: #fff; }
</style>
