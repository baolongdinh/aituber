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
  platform: { type: String, default: 'youtube' }
})

const emit = defineEmits(['update:topic', 'update:contentName'])

const localTopic = computed({
  get: () => props.topic,
  set: (v) => emit('update:topic', v)
})
const localContentName = computed({
  get: () => props.contentName,
  set: (v) => emit('update:contentName', v)
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
  gap: 0px;
}

.platform-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 600;
  margin-bottom: 16px;
  width: fit-content;
}
.platform-badge.youtube {
  background: rgba(255, 0, 0, 0.12);
  color: #ff4444;
  border: 1px solid rgba(255, 68, 68, 0.3);
}
.platform-badge.tiktok {
  background: rgba(161, 75, 255, 0.12);
  color: #a14bff;
  border: 1px solid rgba(161, 75, 255, 0.3);
}

.input-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.9rem;
  font-weight: 600;
  color: rgba(255,255,255,0.9);
  margin-bottom: 10px;
}
.input-label.optional { color: rgba(255,255,255,0.7); }
.optional-tag {
  font-size: 0.72rem;
  font-weight: 400;
  color: rgba(255,255,255,0.4);
  background: rgba(255,255,255,0.06);
  padding: 2px 8px;
  border-radius: 10px;
}

.topic-textarea, .name-input {
  width: 100%;
  background: rgba(255,255,255,0.05);
  border: 1.5px solid rgba(255,255,255,0.1);
  border-radius: 12px;
  color: #fff;
  font-size: 0.95rem;
  padding: 14px;
  font-family: inherit;
  resize: vertical;
  transition: border-color 0.2s;
  box-sizing: border-box;
}
.topic-textarea:focus, .name-input:focus {
  outline: none;
  border-color: rgba(255,255,255,0.3);
  background: rgba(255,255,255,0.08);
}
.topic-textarea::placeholder, .name-input::placeholder {
  color: rgba(255,255,255,0.3);
}
.name-input { padding: 12px 14px; border-radius: 10px; }

.char-count {
  text-align: right;
  font-size: 0.75rem;
  color: rgba(255,255,255,0.35);
  margin-top: 4px;
}

.mt-3 { margin-top: 16px; }

.tips-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  margin-top: 20px;
}
.tip-card {
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 10px;
  padding: 12px;
  display: flex;
  align-items: flex-start;
  gap: 10px;
  font-size: 0.82rem;
}
.tip-icon { font-size: 1.2rem; flex-shrink: 0; }
.tip-title { font-weight: 600; color: rgba(255,255,255,0.85); margin-bottom: 2px; }
.tip-desc { color: rgba(255,255,255,0.45); font-size: 0.78rem; }
</style>
