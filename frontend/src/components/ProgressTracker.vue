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
.progress-tracker {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* ── Status Header ── */
.status-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.status-chip {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 8px 18px;
  border-radius: var(--radius-full);
  font-size: 0.9rem;
  font-weight: 800;
  letter-spacing: 0.02em;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  transition: var(--transition-smooth);
}

.status-chip.idle { background: rgba(255,255,255,0.05); color: var(--text-muted); border: 1px solid var(--glass-border); }
.status-chip.processing { background: rgba(99,179,255,0.15); color: #63b3ff; border: 1px solid rgba(99,179,255,0.2); }
.status-chip.completed { background: rgba(72,199,142,0.15); color: #48c78e; border: 1px solid rgba(72,199,142,0.2); }
.status-chip.failed { background: rgba(255,99,99,0.1); color: #ff6363; border: 1px solid rgba(255,99,99,0.2); }
.status-chip.partial_failed { background: rgba(250,186,85,0.1); color: #faba55; border: 1px solid rgba(250,186,85,0.2); }

.status-icon { font-size: 1.1rem; }

.progress-pct {
  font-size: 1.6rem;
  font-weight: 900;
  color: #fff;
  letter-spacing: -0.05em;
  background: linear-gradient(135deg, #63b3ff, #a78bfa);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

/* ── Progress Bar ── */
.progress-bar-wrap {
  height: 10px;
  background: rgba(0,0,0,0.3);
  border-radius: var(--radius-full);
  overflow: hidden;
  border: 1px solid var(--glass-border);
  position: relative;
}

.progress-bar-fill {
  height: 100%;
  border-radius: var(--radius-full);
  transition: width 0.8s cubic-bezier(0.34, 1.56, 0.64, 1);
  position: relative;
}

/* Shimmer effect */
.progress-bar-fill::after {
  content: '';
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: linear-gradient(
    90deg,
    transparent,
    rgba(255, 255, 255, 0.2),
    transparent
  );
  animation: shimmer 2s infinite;
}

@keyframes shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

.progress-bar-fill.processing { background: linear-gradient(90deg, #63b3ff, #a78bfa, #63b3ff); background-size: 200% 100%; }
.progress-bar-fill.completed { background: linear-gradient(90deg, #48c78e, #06d6a0); }
.progress-bar-fill.partial_failed { background: linear-gradient(90deg, #48c78e, #faba55); }
.progress-bar-fill.failed { background: #ff6363; }

/* ── Timeline ── */
.timeline {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 10px;
}

.timeline-item {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 10px 16px;
  border-radius: var(--radius-md);
  background: rgba(255,255,255,0.01);
  border: 1px solid transparent;
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--text-dim);
  transition: all var(--transition-smooth);
}

.timeline-item.completed {
  color: #48c78e;
  background: rgba(72,199,142,0.03);
}

.timeline-item.current {
  color: #fff;
  background: rgba(255,255,255,0.04);
  border-color: rgba(255,255,255,0.1);
  box-shadow: 0 4px 15px rgba(0,0,0,0.1);
  transform: translateX(4px);
}

.timeline-dot {
  width: 22px;
  height: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.9rem;
  border-radius: 50%;
  background: rgba(255,255,255,0.05);
  flex-shrink: 0;
  transition: var(--transition-fast);
}

.timeline-item.completed .timeline-dot { background: rgba(72,199,142,0.2); }
.timeline-item.current .timeline-dot { background: rgba(99,179,255,0.2); color: #63b3ff; }

.timeline-label { flex: 1; }

/* ── Error Box ── */
.error-box {
  background: rgba(255,99,99,0.08);
  border: 1px solid rgba(255,99,99,0.2);
  border-radius: var(--radius-md);
  padding: 16px;
  font-size: 0.9rem;
  font-weight: 600;
  color: #ff6363;
  display: flex;
  gap: 12px;
  animation: shake 0.5s cubic-bezier(.36,.07,.19,.97) both;
}

@keyframes shake {
  10%, 90% { transform: translate3d(-1px, 0, 0); }
  20%, 80% { transform: translate3d(2px, 0, 0); }
  30%, 50%, 70% { transform: translate3d(-4px, 0, 0); }
  40%, 60% { transform: translate3d(4px, 0, 0); }
}

/* ── Series Parts Tracker ── */
.series-parts {
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-top: 10px;
  padding-top: 20px;
  border-top: 1px solid var(--glass-border);
}

.parts-title {
  font-size: 0.75rem;
  font-weight: 800;
  color: var(--text-dim);
  text-transform: uppercase;
  letter-spacing: 0.1em;
  margin-bottom: 4px;
}

.part-item {
  background: rgba(0,0,0,0.2);
  border: 1px solid var(--card-border);
  border-radius: var(--radius-md);
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  transition: var(--transition-smooth);
}

.part-item:hover { border-color: rgba(255,255,255,0.12); background: rgba(0,0,0,0.3); }

.part-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.part-name {
  font-size: 0.9rem;
  font-weight: 800;
  color: #fff;
  flex: 1;
  letter-spacing: -0.01em;
}

.part-status-chip {
  font-size: 0.65rem;
  font-weight: 800;
  padding: 4px 10px;
  border-radius: var(--radius-full);
  text-transform: uppercase;
}

.part-status-chip.queued { background: rgba(255,255,255,0.05); color: var(--text-dim); }
.part-status-chip.processing { background: rgba(99,179,255,0.15); color: #63b3ff; }
.part-status-chip.completed { background: rgba(72,199,142,0.15); color: #48c78e; }
.part-status-chip.failed { background: rgba(255,99,99,0.1); color: #ff6363; }

.mini-bar-wrap {
  height: 5px;
  background: rgba(255,255,255,0.05);
  border-radius: var(--radius-full);
  overflow: hidden;
}
.mini-bar-fill {
  height: 100%;
  border-radius: var(--radius-full);
  transition: width 0.4s ease;
}
.mini-bar-fill.processing { background: #63b3ff; }
.mini-bar-fill.completed { background: #48c78e; }
.mini-bar-fill.failed { background: #ff6363; }

.spinning { display: inline-block; animation: spin 1s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
</style>
