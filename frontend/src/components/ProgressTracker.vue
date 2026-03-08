<template>
  <div class="progress-tracker">
    <!-- Status -->
    <div class="status-header">
      <div class="status-chip" :class="status">
        <span class="status-icon">{{ statusIcon }}</span>
        <span>{{ statusText }}</span>
      </div>
      <div v-if="status === 'processing'" class="progress-pct">{{ progress }}%</div>
    </div>

    <!-- Progress bar -->
    <div v-if="status !== 'idle'" class="progress-bar-wrap">
      <div
        class="progress-bar-fill"
        :class="status"
        :style="{ width: progress + '%' }"
      />
    </div>

    <!-- Current step -->
    <div v-if="currentStep && status === 'processing'" class="current-step">
      <span class="step-dot spinning">⟳</span>
      {{ currentStep }}
    </div>

    <!-- Error -->
    <div v-if="error" class="error-box">
      <span>⚠️</span> {{ error }}
    </div>

    <!-- Timeline -->
    <div v-if="status !== 'idle'" class="timeline">
      <div
        v-for="step in steps"
        :key="step.name"
        class="timeline-item"
        :class="{ completed: step.completed, current: step.current }"
      >
        <div class="timeline-dot">
          <span v-if="step.completed">✓</span>
          <span v-else-if="step.current" class="spinning">◌</span>
          <span v-else>○</span>
        </div>
        <span class="timeline-label">{{ step.name }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  status: { type: String, default: 'idle' },
  progress: { type: Number, default: 0 },
  currentStep: { type: String, default: '' },
  error: { type: String, default: null }
})

const statusIcon = computed(() => ({
  processing: '⟳',
  completed: '✓',
  failed: '✕',
  idle: '◇'
}[props.status] || '◇'))

const statusText = computed(() => ({
  processing: 'Đang xử lý...',
  completed: 'Hoàn thành!',
  failed: 'Lỗi',
  idle: 'Sẵn sàng'
}[props.status] || 'Sẵn sàng'))

const steps = computed(() => {
  const allSteps = [
    { name: 'Khởi tạo', threshold: 3 },
    { name: 'Gemini AI viết kịch bản', threshold: 8 },
    { name: 'Tách văn bản audio', threshold: 12 },
    { name: 'Tạo audio chunks (TTS)', threshold: 20 },
    { name: 'Tạo phụ đề', threshold: 32 },
    { name: 'Ghép audio', threshold: 42 },
    { name: 'Tải stock video (Pexels)', threshold: 50 },
    { name: 'Ghép video segments', threshold: 82 },
    { name: 'Compose video + audio', threshold: 90 },
    { name: 'Lưu vào thư mục', threshold: 98 },
    { name: 'Hoàn thành', threshold: 100 },
  ]
  return allSteps.map((step, i) => ({
    ...step,
    completed: props.progress > step.threshold,
    current: props.progress >= step.threshold && props.progress < (allSteps[i + 1]?.threshold ?? 101)
  }))
})
</script>

<style scoped>
.progress-tracker { display: flex; flex-direction: column; gap: 14px; }

.status-header { display: flex; align-items: center; justify-content: space-between; }

.status-chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 16px;
  border-radius: 20px;
  font-size: 0.9rem;
  font-weight: 600;
}
.status-chip.idle { background: rgba(255,255,255,0.07); color: rgba(255,255,255,0.5); }
.status-chip.processing { background: rgba(99,179,255,0.15); color: #63b3ff; }
.status-chip.completed { background: rgba(72,199,142,0.15); color: #48c78e; }
.status-chip.failed { background: rgba(255,99,99,0.15); color: #ff6363; }
.status-icon { font-size: 1rem; }

.progress-pct { font-size: 1.3rem; font-weight: 700; color: #63b3ff; }

.progress-bar-wrap {
  height: 6px;
  background: rgba(255,255,255,0.08);
  border-radius: 6px;
  overflow: hidden;
}
.progress-bar-fill {
  height: 100%;
  border-radius: 6px;
  transition: width 0.5s ease;
  background: rgba(255,255,255,0.3);
}
.progress-bar-fill.processing { background: linear-gradient(90deg, #63b3ff, #a78bfa); }
.progress-bar-fill.completed { background: linear-gradient(90deg, #48c78e, #06d6a0); }
.progress-bar-fill.failed { background: #ff6363; }

.current-step {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.85rem;
  color: rgba(255,255,255,0.7);
  background: rgba(255,255,255,0.04);
  padding: 10px 14px;
  border-radius: 8px;
  border-left: 3px solid #63b3ff;
}

.error-box {
  background: rgba(255,99,99,0.1);
  border: 1px solid rgba(255,99,99,0.2);
  border-radius: 10px;
  padding: 12px 14px;
  font-size: 0.85rem;
  color: #ff6363;
  display: flex;
  gap: 8px;
}

.timeline { display: flex; flex-direction: column; gap: 6px; }
.timeline-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 0;
  font-size: 0.82rem;
  color: rgba(255,255,255,0.3);
  transition: color 0.2s;
}
.timeline-item.completed { color: rgba(72,199,142,0.8); }
.timeline-item.current { color: rgba(255,255,255,0.85); }

.timeline-dot {
  width: 18px;
  height: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8rem;
  flex-shrink: 0;
}
.timeline-label { flex: 1; }

.spinning { display: inline-block; animation: spin 1s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
</style>
