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
            :class="{ active: platform === 'tiktok' }"
            @click="setPlatform('tiktok')"
          >
            ⚡ TikTok
          </button>
          <button
            class="switch-btn"
            :class="{ active: platform === 'youtube' }"
            @click="setPlatform('youtube')"
          >
            🎬 YouTube
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
              v-model:isSeries="isSeriesInput"
              v-model:numParts="numPartsInput"
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
            <span>{{ generating ? 'Đang tạo video...' : (isSeriesInput ? `Tạo Series ${numPartsInput} phần` : 'Tạo Video') }}</span>
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
              :is-series="isSeries"
              :parts="seriesParts"
              @retry-part="retryPart"
            />
          </div>

          <div class="panel-card" v-if="videoUrl || (isSeries && hasCompletedParts)">
            <ResultPreview
              :video-url="videoUrl"
              :job-id="jobId"
              :saved-path="savedPath"
              :platform="platform"
              :is-series="isSeries"
              :parts="seriesParts"
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
const platform = ref('tiktok')
const setPlatform = (p) => {
  platform.value = p
  handleReset()
}

// Input state
const topic = ref('')
const contentName = ref('')
const isSeriesInput = ref(false)
const numPartsInput = ref(2)
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
  isSeries,
  seriesParts,
  seriesId,
  generateVideo,
  retryPart,
  reset,
} = useVideoGeneration()

// Computed
const canGenerate = computed(() => topic.value.trim().length > 3 && !generating.value)
const hasCompletedParts = computed(() => {
  return seriesParts.value && seriesParts.value.some(p => p.status === 'completed')
})

// Methods
const handleGenerate = async () => {
  await generateVideo(topic.value.trim(), contentName.value.trim(), platform.value, config.value, isSeriesInput.value, numPartsInput.value)
}

const handleReset = () => {
  reset()
  topic.value = ''
  contentName.value = ''
  isSeriesInput.value = false
  numPartsInput.value = 2
}
</script>

<style>
/* ── Design Tokens ── */
:root {
  --bg-deep: #08080c;
  --card-bg: rgba(255, 255, 255, 0.03);
  --card-border: rgba(255, 255, 255, 0.08);
  --glass-bg: rgba(13, 13, 18, 0.75);
  --glass-border: rgba(255, 255, 255, 0.06);
  
  --text-main: #ffffff;
  --text-muted: rgba(255, 255, 255, 0.45);
  --text-dim: rgba(255, 255, 255, 0.2);
  
  --accent-yt: #ff0000;
  --accent-yt-glow: rgba(255, 0, 0, 0.4);
  --accent-tk: #a14bff;
  --accent-tk-alt: #ff3f6c;
  --accent-tk-glow: rgba(161, 75, 255, 0.4);
  
  --radius-lg: 18px;
  --radius-md: 12px;
  --radius-full: 9999px;
  
  --transition-fast: 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  --transition-smooth: 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  --transition-bounce: 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
}

/* ── Reset & Base ── */
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

body {
  font-family: 'Outfit', 'Inter', system-ui, sans-serif;
  background: var(--bg-deep);
  color: var(--text-main);
  min-height: 100vh;
  -webkit-font-smoothing: antialiased;
  overflow-x: hidden;
}

/* ── App Shell ── */
.app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  position: relative;
  z-index: 1;
}

/* Dynamic Background Blobs */
.bg-blob {
  position: fixed;
  border-radius: 50%;
  pointer-events: none;
  filter: blur(100px);
  opacity: 0.15;
  z-index: -1;
  transition: all 1s ease;
  animation: blobFloat 20s infinite alternate cubic-bezier(0.45, 0, 0.55, 1);
}

@keyframes blobFloat {
  0% { transform: translate(0, 0) scale(1); }
  50% { transform: translate(50px, 30px) scale(1.1); }
  100% { transform: translate(-30px, 60px) scale(0.9); }
}

.blob-1 { width: 600px; height: 600px; top: -150px; left: -100px; }
.blob-2 { width: 500px; height: 500px; bottom: -100px; right: -50px; }

.app.youtube .blob-1 { background: var(--accent-yt); opacity: 0.12; }
.app.youtube .blob-2 { background: #600000; opacity: 0.08; }
.app.tiktok .blob-1 { background: var(--accent-tk); opacity: 0.12; }
.app.tiktok .blob-2 { background: var(--accent-tk-alt); opacity: 0.1; }

/* ── Header ── */
.app-header {
  position: sticky;
  top: 0;
  z-index: 100;
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-bottom: 1px solid var(--glass-border);
  padding: 12px 0;
}

.header-inner {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.logo {
  display: flex;
  align-items: center;
  gap: 14px;
  cursor: pointer;
  transition: var(--transition-fast);
}

.logo:hover { transform: scale(1.02); }

.logo-icon {
  font-size: 2rem;
  filter: drop-shadow(0 0 8px rgba(255,255,255,0.3));
}

.logo-text { display: flex; flex-direction: column; line-height: 1.1; }
.logo-main {
  font-size: 1.35rem;
  font-weight: 800;
  letter-spacing: -0.03em;
  background: linear-gradient(to right, #fff, #a0a0a0);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.logo-sub {
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--text-muted);
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.logo-badge {
  font-size: 0.65rem;
  font-weight: 700;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: var(--radius-full);
  padding: 4px 10px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.02em;
}

/* Platform Switcher */
.platform-switch {
  display: flex;
  background: rgba(0,0,0,0.3);
  border: 1px solid var(--glass-border);
  border-radius: 14px;
  padding: 4px;
  gap: 4px;
}

.switch-btn {
  padding: 8px 18px;
  border-radius: 10px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition-fast);
  display: flex;
  align-items: center;
  gap: 8px;
}

.switch-btn:hover { color: #fff; background: rgba(255,255,255,0.03); }

.switch-btn.active {
  color: #fff;
  background: rgba(255,255,255,0.08);
  box-shadow: 0 4px 12px rgba(0,0,0,0.2);
}

.app.youtube .switch-btn.active {
  background: rgba(255, 0, 0, 0.15);
  color: #ff5050;
  border: 1px solid rgba(255,0,0,0.2);
}

.app.tiktok .switch-btn.active {
  background: linear-gradient(135deg, rgba(161,75,255,0.2), rgba(255,63,108,0.2));
  color: #d8b4fe;
  border: 1px solid rgba(161,75,255,0.2);
}

/* ── Main Content ── */
.main-content {
  flex: 1;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
  padding: 40px 24px;
}

.content-grid {
  display: grid;
  grid-template-columns: 520px 1fr;
  gap: 32px;
  align-items: start;
}

@media (max-width: 1000px) {
  .content-grid { grid-template-columns: 1fr; max-width: 600px; margin: 0 auto; }
}

.panel { display: flex; flex-direction: column; gap: 24px; }

.panel-card {
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: var(--radius-lg);
  padding: 24px;
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  box-shadow: 0 8px 32px rgba(0,0,0,0.2);
  transition: var(--transition-smooth);
  animation: cardEntry 0.6s var(--transition-bounce) backwards;
}

@keyframes cardEntry {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.left-panel .panel-card:nth-child(1) { animation-delay: 0.1s; }
.left-panel .panel-card:nth-child(2) { animation-delay: 0.2s; }
.right-panel .panel-card:nth-child(1) { animation-delay: 0.3s; }
.right-panel .panel-card:nth-child(2) { animation-delay: 0.4s; }

.panel-card:hover {
  border-color: rgba(255,255,255,0.15);
  transform: translateY(-2px);
  box-shadow: 0 12px 40px rgba(0,0,0,0.3);
}

/* ── Primary Action Button ── */
.generate-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  width: 100%;
  padding: 18px 24px;
  border: none;
  border-radius: var(--radius-lg);
  font-size: 1.05rem;
  font-weight: 800;
  cursor: pointer;
  transition: var(--transition-bounce);
  position: relative;
  overflow: hidden;
  letter-spacing: 0.02em;
}

.generate-btn::before {
  content: '';
  position: absolute;
  top: 0; left: -100%;
  width: 100%; height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.2), transparent);
  transition: 0.5s;
}

.generate-btn:hover::before { left: 100%; }

.app.youtube .generate-btn {
  background: linear-gradient(135deg, #ff4d4d, #b30000);
  color: #fff;
  box-shadow: 0 8px 25px var(--accent-yt-glow);
}

.app.tiktok .generate-btn {
  background: linear-gradient(135deg, var(--accent-tk), var(--accent-tk-alt));
  color: #fff;
  box-shadow: 0 8px 25px var(--accent-tk-glow);
}

.generate-btn:hover:not(.disabled) {
  transform: translateY(-3px) scale(1.02);
  filter: brightness(1.1);
}

.generate-btn:active:not(.disabled) { transform: translateY(0) scale(0.98); }

.generate-btn.disabled {
  background: #252529 !important;
  color: var(--text-dim) !important;
  box-shadow: none !important;
  cursor: not-allowed;
  opacity: 0.5;
}

.btn-icon { font-size: 1.4rem; }

/* ── Footer ── */
.app-footer {
  text-align: center;
  padding: 30px 24px;
  font-size: 0.8rem;
  color: var(--text-dim);
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 12px;
  border-top: 1px solid var(--glass-border);
  margin-top: auto;
}

.sep { width: 4px; height: 4px; border-radius: 50%; background: var(--text-dim); }

/* ── Custom Utilities ── */
.mt-3 { margin-top: 16px; }
.mb-2 { margin-bottom: 12px; }

/* Shared Component Styles Override (to avoid scoping issues) */
input[type="text"], textarea {
  transition: var(--transition-fast) !important;
}
</style>
