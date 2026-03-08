<template>
  <div class="app" :class="platform">
    <!-- Background blobs -->
    <div class="bg-blob blob-1" />
    <div class="bg-blob blob-2" />

    <!-- Header -->
    <header class="app-header">
      <div class="header-inner">
        <!-- Logo -->
        <div class="logo">
          <span class="logo-icon">{{ platform === 'youtube' ? '🎬' : '⚡' }}</span>
          <div class="logo-text">
            <span class="logo-main">{{ platform === 'youtube' ? 'ContentForge' : 'ViralCraft' }}</span>
            <span class="logo-sub">{{ platform === 'youtube' ? 'YouTube Studio' : 'TikTok Creator' }}</span>
          </div>
          <div class="logo-badge">Powered by Gemini AI</div>
        </div>

        <!-- Platform switch -->
        <div class="platform-switch">
          <button
            class="switch-btn"
            :class="{ active: platform === 'youtube' }"
            @click="setPlatform('youtube')"
          >
            🎬 YouTube
          </button>
          <button
            class="switch-btn"
            :class="{ active: platform === 'tiktok' }"
            @click="setPlatform('tiktok')"
          >
            ⚡ TikTok
          </button>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="main-content">
      <div class="content-grid">
        <!-- Left Panel -->
        <div class="panel left-panel">
          <div class="panel-card">
            <TopicInput
              v-model:topic="topic"
              v-model:contentName="contentName"
              :platform="platform"
            />
          </div>

          <div class="panel-card">
            <ConfigPanel v-model="config" />
          </div>

          <!-- Generate Button -->
          <button
            class="generate-btn"
            :class="{ loading: generating, disabled: !canGenerate }"
            :disabled="!canGenerate"
            @click="handleGenerate"
          >
            <span v-if="generating" class="btn-spinner">⟳</span>
            <span v-else class="btn-icon">{{ platform === 'youtube' ? '🎬' : '⚡' }}</span>
            <span>{{ generating ? 'Đang tạo video...' : 'Tạo Video' }}</span>
          </button>
        </div>

        <!-- Right Panel -->
        <div class="panel right-panel">
          <div class="panel-card">
            <ProgressTracker
              :status="jobStatus"
              :progress="progress"
              :current-step="currentStep"
              :error="error"
            />
          </div>

          <div class="panel-card" v-if="videoUrl">
            <ResultPreview
              :video-url="videoUrl"
              :job-id="jobId"
              :saved-path="savedPath"
              :platform="platform"
              @reset="handleReset"
            />
          </div>
        </div>
      </div>
    </main>

    <!-- Footer -->
    <footer class="app-footer">
      <span>{{ platform === 'youtube' ? '🎬 ContentForge' : '⚡ ViralCraft' }}</span>
      <span class="sep">·</span>
      <span>AI Video Generator</span>
      <span class="sep">·</span>
      <span>{{ platform === 'youtube' ? 'YouTube 16:9' : 'TikTok 9:16' }}</span>
    </footer>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import TopicInput from './components/TopicInput.vue'
import ConfigPanel from './components/ConfigPanel.vue'
import ProgressTracker from './components/ProgressTracker.vue'
import ResultPreview from './components/ResultPreview.vue'
import { useVideoGeneration } from './composables/useVideoGeneration'

// Platform state
const platform = ref('youtube')
const setPlatform = (p) => {
  platform.value = p
  handleReset()
}

// Input state
const topic = ref('')
const contentName = ref('')
const config = ref({
  voice: 'banmai',
  speaking_speed: 1.0,
})

// Video generation composable
const {
  generating,
  progress,
  currentStep,
  jobStatus,
  videoUrl,
  savedPath,
  error,
  jobId,
  generateVideo,
  reset,
} = useVideoGeneration()

// Computed
const canGenerate = computed(() => topic.value.trim().length > 3 && !generating.value)

// Methods
const handleGenerate = async () => {
  await generateVideo(topic.value.trim(), contentName.value.trim(), platform.value, config.value)
}

const handleReset = () => {
  reset()
  topic.value = ''
  contentName.value = ''
}
</script>

<style>
/* ── Reset & Base ── */
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

body {
  font-family: 'Inter', 'Segoe UI', system-ui, sans-serif;
  background: #0d0d12;
  color: #fff;
  min-height: 100vh;
}

/* ── App shell ── */
.app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  position: relative;
  overflow-x: hidden;
  transition: --accent 0.4s;
}

/* Background blobs */
.bg-blob {
  position: fixed;
  border-radius: 50%;
  pointer-events: none;
  filter: blur(80px);
  opacity: 0.12;
  z-index: 0;
}
.blob-1 {
  width: 500px; height: 500px;
  top: -100px; left: -100px;
}
.blob-2 {
  width: 400px; height: 400px;
  bottom: -80px; right: -80px;
}

/* Platform color themes */
.app.youtube .blob-1 { background: #ff0000; }
.app.youtube .blob-2 { background: #cc0000; }
.app.tiktok .blob-1 { background: #a14bff; }
.app.tiktok .blob-2 { background: #ff3f6c; }

/* ── Header ── */
.app-header {
  position: sticky;
  top: 0;
  z-index: 10;
  background: rgba(13,13,18,0.85);
  backdrop-filter: blur(16px);
  border-bottom: 1px solid rgba(255,255,255,0.06);
}
.header-inner {
  max-width: 1200px;
  margin: 0 auto;
  padding: 14px 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.logo { display: flex; align-items: center; gap: 12px; }
.logo-icon { font-size: 1.8rem; }
.logo-text { display: flex; flex-direction: column; }
.logo-main { font-size: 1.2rem; font-weight: 800; letter-spacing: -0.02em; }
.logo-sub { font-size: 0.72rem; color: rgba(255,255,255,0.4); margin-top: -2px; }
.logo-badge {
  font-size: 0.68rem;
  background: rgba(255,255,255,0.07);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 20px;
  padding: 3px 10px;
  color: rgba(255,255,255,0.45);
  white-space: nowrap;
}

.platform-switch {
  display: flex;
  gap: 6px;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 12px;
  padding: 4px;
}
.switch-btn {
  padding: 8px 18px;
  border-radius: 9px;
  border: none;
  background: transparent;
  color: rgba(255,255,255,0.45);
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  font-family: inherit;
}
.switch-btn.active {
  background: rgba(255,255,255,0.1);
  color: #fff;
}
.app.youtube .switch-btn.active { background: rgba(255,0,0,0.15); color: #ff6060; }
.app.tiktok .switch-btn.active { background: rgba(161,75,255,0.2); color: #c084fc; }

/* ── Main ── */
.main-content {
  flex: 1;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
  padding: 32px 24px;
  position: relative;
  z-index: 1;
}

.content-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}
@media (max-width: 900px) {
  .content-grid { grid-template-columns: 1fr; }
}

.panel { display: flex; flex-direction: column; gap: 16px; }

.panel-card {
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.07);
  border-radius: 16px;
  padding: 20px;
  backdrop-filter: blur(8px);
}

/* ── Generate button ── */
.generate-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  width: 100%;
  padding: 16px 24px;
  border: none;
  border-radius: 14px;
  font-size: 1rem;
  font-weight: 700;
  cursor: pointer;
  font-family: inherit;
  transition: all 0.25s;
  position: relative;
  overflow: hidden;
}

.app.youtube .generate-btn {
  background: linear-gradient(135deg, #ff4444, #cc0000);
  color: #fff;
  box-shadow: 0 4px 24px rgba(255,68,68,0.35);
}
.app.youtube .generate-btn:hover:not(.disabled) {
  box-shadow: 0 6px 32px rgba(255,68,68,0.5);
  transform: translateY(-2px);
}

.app.tiktok .generate-btn {
  background: linear-gradient(135deg, #a14bff, #ff3f6c);
  color: #fff;
  box-shadow: 0 4px 24px rgba(161,75,255,0.4);
}
.app.tiktok .generate-btn:hover:not(.disabled) {
  box-shadow: 0 6px 32px rgba(161,75,255,0.6);
  transform: translateY(-2px);
}

.generate-btn.disabled {
  opacity: 0.4;
  cursor: not-allowed;
  transform: none !important;
}
.generate-btn.loading { opacity: 0.7; cursor: wait; }

.btn-icon { font-size: 1.2rem; }
.btn-spinner {
  font-size: 1.1rem;
  display: inline-block;
  animation: spin 1s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* ── Footer ── */
.app-footer {
  text-align: center;
  padding: 16px 24px;
  font-size: 0.75rem;
  color: rgba(255,255,255,0.2);
  display: flex;
  justify-content: center;
  gap: 10px;
  border-top: 1px solid rgba(255,255,255,0.04);
  position: relative;
  z-index: 1;
}
.sep { opacity: 0.4; }
</style>
