<template>
  <div class="progress-tracker">
    <!-- Status -->
    <div class="status-header">
      <div class="status-chip" :class="status">
        <span class="status-icon">{{ statusIcon }}</span>
        <span>{{ statusText }}</span>
      </div>
      <div v-if="status === 'processing' || status === 'partial_failed'" class="progress-pct">
        {{ progress }}%
      </div>
    </div>

    <!-- Overall Progress bar -->
    <div v-if="status !== 'idle'" class="progress-bar-wrap">
      <div
        class="progress-bar-fill"
        :class="status"
        :style="{ width: progress + '%' }"
      />
    </div>

    <!-- Error -->
    <div v-if="error" class="error-box">
      <span>⚠️</span> {{ error }}
    </div>

    <!-- SINGLE VIDEO TIMELINE -->
    <template v-if="!isSeries">
      <!-- Current step -->
      <div v-if="currentStep && status === 'processing'" class="current-step">
        <span class="step-dot spinning">⟳</span>
        {{ currentStep }}
      </div>

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
    </template>

    <!-- SERIES PARTS TRACKER -->
    <template v-else>
      <div v-if="parts && parts.length > 0" class="series-parts">
        <div class="parts-title">TIẾN ĐỘ TỪNG PHẦN</div>
        <div v-for="part in parts" :key="part.part_index" class="part-item">
          <!-- Part Header -->
          <div class="part-header">
            <span class="part-name">{{ part.title || `Tập ${part.part_index + 1}` }}</span>
            <span class="part-status-chip" :class="part.status">
              {{ partStatusLabels[part.status] || part.status }}
            </span>
            <button 
              v-if="part.status === 'failed'" 
              class="retry-btn" 
              @click="handleRetry(part.part_index)"
              title="Thử lại phần này"
            >
              🔄
            </button>
          </div>

          <!-- Part current step -->
          <div class="part-step" v-if="part.status === 'processing'">
            {{ part.current_step || 'Đang chờ...' }}
          </div>

          <div class="part-error" v-if="part.error">
             ⚠️ {{ part.error }}
          </div>

          <!-- Mini progress bar -->
          <div class="mini-bar-wrap">
            <div class="mini-bar-fill" :class="part.status" :style="{ width: part.progress + '%' }" />
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  status: { type: String, default: 'idle' },
  progress: { type: Number, default: 0 },
  currentStep: { type: String, default: '' },
  error: { type: String, default: null },
  isSeries: { type: Boolean, default: false },
  parts: { type: Array, default: () => [] }
})

const emit = defineEmits(['retry-part'])

const handleRetry = (partIndex) => {
  emit('retry-part', partIndex)
}

const statusIcon = computed(() => ({
  processing: '⟳',
  completed: '✓',
  partial_failed: '⚠️',
  failed: '✕',
  idle: '◇'
}[props.status] || '◇'))

const statusText = computed(() => ({
  processing: 'Đang xử lý...',
  completed: 'Hoàn thành!',
  partial_failed: 'Xong (có lỗi)',
  failed: 'Thất bại',
  idle: 'Sẵn sàng'
}[props.status] || 'Sẵn sàng'))

const partStatusLabels = {
  queued: 'Chờ',
  processing: 'Đang chạy',
  completed: 'Xong ✓',
  failed: 'Lỗi ✕'
}

const steps = computed(() => {
  if (props.isSeries) return [] // Hide timeline for series
  const allSteps = [
    { name: 'Khởi tạo', threshold: 3 },
    { name: 'Gemini AI viết kịch bản', threshold: 8 },
    { name: 'Tách văn bản', threshold: 12 },
    { name: 'Tạo audio', threshold: 20 },
    { name: 'Tạo phụ đề', threshold: 32 },
    { name: 'Ghép audio', threshold: 42 },
    { name: 'Tải stock video', threshold: 50 },
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
.status-chip.partial_failed { background: rgba(250,186,85,0.15); color: #faba55; }
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
.progress-bar-fill.partial_failed { background: linear-gradient(90deg, #48c78e, #faba55); }
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

/* ── Series Parts Tracker ── */
.series-parts {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 8px;
  padding-top: 16px;
  border-top: 1px dashed rgba(255,255,255,0.1);
}
.parts-title {
  font-size: 0.75rem;
  letter-spacing: 0.05em;
  color: rgba(255,255,255,0.4);
  font-weight: 700;
}
.part-item {
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.06);
  border-radius: 10px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.part-header {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 12px;
}
.part-name {
  font-size: 0.85rem;
  font-weight: 600;
  color: rgba(255,255,255,0.9);
  flex: 1;
}
.part-status-chip {
  font-size: 0.7rem;
  padding: 2px 8px;
  border-radius: 12px;
  font-weight: 600;
}
.retry-btn {
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 4px;
  cursor: pointer;
  padding: 2px 6px;
  font-size: 0.8rem;
  transition: all 0.2s;
}
.retry-btn:hover {
  background: rgba(255,255,255,0.1);
  transform: scale(1.1);
}
.part-status-chip.queued { background: rgba(255,255,255,0.1); color: rgba(255,255,255,0.6); }
.part-status-chip.processing { background: rgba(99,179,255,0.2); color: #63b3ff; }
.part-status-chip.completed { background: rgba(72,199,142,0.2); color: #48c78e; }
.part-status-chip.failed { background: rgba(255,99,99,0.2); color: #ff6363; }

.part-step {
  font-size: 0.75rem;
  color: rgba(255,255,255,0.5);
}
.part-error {
  font-size: 0.75rem;
  color: #ff6363;
  background: rgba(255,99,99,0.1);
  padding: 4px 8px;
  border-radius: 4px;
}

.mini-bar-wrap {
  height: 4px;
  background: rgba(255,255,255,0.08);
  border-radius: 4px;
  overflow: hidden;
}
.mini-bar-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.3s;
}
.mini-bar-fill.queued { background: rgba(255,255,255,0.2); }
.mini-bar-fill.processing { background: #63b3ff; }
.mini-bar-fill.completed { background: #48c78e; }
.mini-bar-fill.failed { background: #ff6363; }
</style>
