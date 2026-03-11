<template>
  <div class="topic-input">
    <!-- Platform badge -->
    <div class="platform-badge" :class="platform">
      <span class="badge-icon">{{ platform === 'youtube' ? '🎬' : '⚡' }}</span>
      <span class="badge-text">{{ platform === 'youtube' ? 'ContentForge YouTube' : 'ViralCraft TikTok' }}</span>
    </div>

    <!-- Topic Input -->
    <div class="input-section">
      <label class="input-label">
        <span class="label-icon">💡</span>
        Bạn muốn tạo video về chủ đề gì?
      </label>
      <textarea
        v-model="localTopic"
        class="topic-textarea"
        :placeholder="placeholder"
        rows="4"
        maxlength="500"
      />
      <div class="char-count">{{ localTopic.length }} / 500</div>
    </div>

    <!-- Content Name (optional) -->
    <div class="input-section mt-3">
      <label class="input-label optional">
        <span class="label-icon">📁</span>
        Tên thư mục (tùy chọn)
        <span class="optional-tag">auto-generate nếu để trống</span>
      </label>
      <input
        v-model="localContentName"
        class="name-input"
        placeholder="vd: tai-chinh-ca-nhan-2024"
        maxlength="80"
      />
    </div>

    <!-- Series Options -->
    <div class="input-section mt-3 series-section">
      <label class="series-toggle-label">
        <input type="checkbox" v-model="localIsSeries" class="series-checkbox" />
        <span class="toggle-text">
          <span class="label-icon">🎬</span>
          Tạo chuỗi (Series) tự động liên kết nối tiếp
        </span>
      </label>
      
      <div v-if="localIsSeries" class="num-parts-input">
        <label class="input-label">Số tập ({{ localNumParts }} phần)</label>
        <div class="slider-wrap">
          <input type="range" v-model.number="localNumParts" min="2" max="20" class="range-slider" />
          <div class="slider-marks">
            <span>2</span>
            <span>10</span>
            <span>20</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Tips -->
    <div class="tips-grid">
      <div class="tip-card" v-for="tip in currentTips" :key="tip.title">
        <span class="tip-icon">{{ tip.icon }}</span>
        <div>
          <div class="tip-title">{{ tip.title }}</div>
          <div class="tip-desc">{{ tip.desc }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  topic: { type: String, default: '' },
  contentName: { type: String, default: '' },
  isSeries: { type: Boolean, default: false },
  numParts: { type: Number, default: 2 },
  platform: { type: String, default: 'youtube' }
})

const emit = defineEmits(['update:topic', 'update:contentName', 'update:isSeries', 'update:numParts'])

const localTopic = computed({
  get: () => props.topic,
  set: (v) => emit('update:topic', v)
})
const localContentName = computed({
  get: () => props.contentName,
  set: (v) => emit('update:contentName', v)
})
const localIsSeries = computed({
  get: () => props.isSeries,
  set: (v) => emit('update:isSeries', v)
})
const localNumParts = computed({
  get: () => props.numParts,
  set: (v) => emit('update:numParts', v)
})

const placeholder = computed(() => {
  if (props.platform === 'tiktok') {
    return 'vd: Tại sao người Việt không giàu được?\nhoặc: 3 thói quen phá hoại sự nghiệp của bạn\nhoặc: Bí mật về Bitcoin mà không ai nói với bạn'
  }
  return 'vd: Lỗ đen vũ trụ - Cánh cổng thời gian hay kẻ hủy diệt?\nhoặc: Tại sao chúng ta không bao giờ một mình trong vũ trụ\nhoặc: Nghịch lý Fermi giải thích như thế nào'
})

const youtubeTips = [
  { icon: '⏱️', title: '5–10 phút', desc: 'Video dài, nội dung sâu sắc' },
  { icon: '📚', title: 'Educational', desc: 'Kiến thức, khoa học, lịch sử' },
  { icon: '🎯', title: 'SEO-friendly', desc: 'Hook + sections rõ ràng' },
]
const tiktokTips = [
  { icon: '⚡', title: '30–60 giây', desc: 'Nhanh, punch, gây nghiện' },
  { icon: '🔥', title: 'Hook 3 giây', desc: 'Câu đầu phải gây sốc ngay' },
  { icon: '🌊', title: 'Viral format', desc: 'Tò mò → Value → CTA' },
]

const currentTips = computed(() => props.platform === 'youtube' ? youtubeTips : tiktokTips)
</script>

<style scoped>
.topic-input {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* ── Platform Badge ── */
.platform-badge {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 8px 16px;
  border-radius: var(--radius-full);
  font-size: 0.85rem;
  font-weight: 700;
  width: fit-content;
  letter-spacing: 0.02em;
  box-shadow: 0 4px 15px rgba(0,0,0,0.2);
  transition: var(--transition-smooth);
}

.platform-badge.youtube {
  background: rgba(255, 0, 0, 0.1);
  color: #ff5050;
  border: 1px solid rgba(255, 0, 0, 0.2);
}

.platform-badge.tiktok {
  background: linear-gradient(135deg, rgba(161,75,255,0.1), rgba(255,63,108,0.1));
  color: #d8b4fe;
  border: 1px solid rgba(161,75,255,0.2);
}

/* ── Input Sections ── */
.input-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.input-label {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text-main);
  letter-spacing: -0.01em;
}

.input-label.optional { color: var(--text-muted); }

.optional-tag {
  font-size: 0.7rem;
  font-weight: 600;
  color: var(--text-dim);
  background: rgba(255,255,255,0.05);
  padding: 3px 10px;
  border-radius: var(--radius-full);
  margin-left: auto;
  text-transform: uppercase;
}

/* Fields */
.topic-textarea, .name-input {
  width: 100%;
  background: rgba(0,0,0,0.2);
  border: 1.5px solid var(--card-border);
  border-radius: var(--radius-md);
  color: #fff;
  font-size: 1rem;
  padding: 16px;
  font-family: inherit;
  line-height: 1.5;
  transition: all var(--transition-fast);
  box-shadow: inset 0 2px 8px rgba(0,0,0,0.1);
}

.topic-textarea:focus, .name-input:focus {
  outline: none;
  background: rgba(0,0,0,0.3);
  border-color: rgba(255,255,255,0.2);
  box-shadow: 0 0 0 4px rgba(255,255,255,0.03), inset 0 2px 8px rgba(0,0,0,0.2);
}

.app.youtube .topic-textarea:focus, .app.youtube .name-input:focus {
  border-color: rgba(255,0,0,0.4);
  box-shadow: 0 0 0 4px rgba(255,0,0,0.1), inset 0 2px 8px rgba(0,0,0,0.2);
}

.app.tiktok .topic-textarea:focus, .app.tiktok .name-input:focus {
  border-color: rgba(161,75,255,0.4);
  box-shadow: 0 0 0 4px rgba(161,75,255,0.1), inset 0 2px 8px rgba(0,0,0,0.2);
}

.topic-textarea::placeholder, .name-input::placeholder {
  color: var(--text-dim);
}

.char-count {
  text-align: right;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-dim);
  margin-top: -6px;
}

/* ── Tips Grid ── */
.tips-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-top: 8px;
}

.tip-card {
  background: rgba(255,255,255,0.02);
  border: 1px solid var(--card-border);
  border-radius: var(--radius-md);
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  transition: var(--transition-smooth);
}

.tip-card:hover {
  background: rgba(255,255,255,0.04);
  border-color: rgba(255,255,255,0.15);
  transform: translateY(-3px);
}

.tip-icon {
  font-size: 1.5rem;
  margin-bottom: 2px;
}

.tip-title {
  font-weight: 800;
  font-size: 0.85rem;
  color: #fff;
  letter-spacing: -0.01em;
}

.tip-desc {
  color: var(--text-muted);
  font-size: 0.75rem;
  line-height: 1.3;
}

/* ── Series Section ── */
.series-section {
  background: rgba(255,255,255,0.02);
  border: 1px solid var(--card-border);
  border-radius: var(--radius-md);
  padding: 18px;
  transition: var(--transition-smooth);
}

.series-section:has(.series-checkbox:checked) {
  border-color: rgba(99, 179, 255, 0.4);
  background: rgba(99, 179, 255, 0.03);
}

.series-toggle-label {
  display: flex;
  align-items: center;
  gap: 14px;
  cursor: pointer;
  user-select: none;
}

.series-checkbox {
  appearance: none;
  width: 22px;
  height: 22px;
  border: 2px solid var(--text-dim);
  border-radius: 6px;
  background: transparent;
  cursor: pointer;
  position: relative;
  transition: var(--transition-fast);
}

.series-checkbox:checked {
  background: #63b3ff;
  border-color: #63b3ff;
  box-shadow: 0 0 12px rgba(99, 179, 255, 0.4);
}

.series-checkbox:checked::after {
  content: '✓';
  position: absolute;
  top: 50%; left: 50%;
  transform: translate(-50%, -50%);
  color: #000;
  font-weight: 900;
  font-size: 14px;
}

.toggle-text {
  font-size: 0.95rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-muted);
  transition: var(--transition-fast);
}

.series-checkbox:checked + .toggle-text { color: #fff; }

.num-parts-input {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid var(--glass-border);
  animation: slideDown 0.4s var(--transition-smooth) backwards;
}

@keyframes slideDown {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}

.slider-wrap {
  margin-top: 12px;
}

.range-slider {
  width: 100%;
  height: 6px;
  appearance: none;
  background: rgba(255,255,255,0.1);
  border-radius: var(--radius-full);
  outline: none;
  cursor: pointer;
}

.range-slider::-webkit-slider-thumb {
  appearance: none;
  width: 20px;
  height: 20px;
  background: #fff;
  border: 3px solid #63b3ff;
  border-radius: 50%;
  box-shadow: 0 0 10px rgba(99, 179, 255, 0.5);
  transition: var(--transition-fast);
}

.range-slider::-webkit-slider-thumb:hover {
  transform: scale(1.15);
  box-shadow: 0 0 15px rgba(99, 179, 255, 0.7);
}

.slider-marks {
  display: flex;
  justify-content: space-between;
  margin-top: 10px;
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--text-dim);
}
</style>
