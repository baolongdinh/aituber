<template>
  <div class="config-panel">
    <div class="panel-title">
      <span class="title-icon">🎙️</span> Cài đặt giọng đọc
    </div>

    <!-- Voice Selector -->
    <div class="config-group">
      <label class="group-label">Chọn giọng AI</label>
      <div class="voice-grid">
        <div
          v-for="v in voiceOptions"
          :key="v.value"
          class="voice-card"
          :class="{ active: localConfig.voice === v.value }"
          @click="setVoice(v.value)"
        >
          <div class="voice-row">
            <span class="voice-avatar">{{ v.gender }}</span>
            <span class="voice-name">{{ v.name }}</span>
          </div>
          <span class="gender-badge">{{ v.region }}</span>
        </div>
      </div>
    </div>


    <!-- Summary -->
    <div class="config-summary">
      <div class="summary-item">
        <span class="item-label">Nền tảng</span>
        <span class="item-val">Gemini 1.5 Pro</span>
      </div>
      <div class="summary-item">
        <span class="item-label">Xử lý</span>
        <span class="item-val">Parallel TTS</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, watch } from 'vue'

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({
      voice: 'banmai',
      tts_provider: 'fpt',
    })
  }
})

const emit = defineEmits(['update:modelValue'])

const localConfig = reactive({ ...props.modelValue })

watch(() => props.modelValue, (val) => {
  Object.assign(localConfig, val)
}, { deep: true })

const emitUpdate = () => emit('update:modelValue', { ...localConfig })

const setVoice = (v) => {
  localConfig.voice = v
  emitUpdate()
}

const setProvider = (p) => {
  localConfig.tts_provider = p
  emitUpdate()
}

const voiceOptions = [
  { value: 'banmai', name: 'Ban Mai', gender: '👩', region: 'Bắc' },
  { value: 'leminh', name: 'Lê Minh', gender: '👩', region: 'Nam' },
  { value: 'minhquang', name: 'Minh Quang', gender: '👨', region: 'Bắc' },
  { value: 'giahuy', name: 'Gia Huy', gender: '👨', region: 'Nam' },
]


</script>

<style scoped>
.config-panel {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.9rem;
  font-weight: 700;
  color: rgba(255,255,255,0.7);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding-bottom: 4px;
  border-bottom: 1px solid rgba(255,255,255,0.08);
}

.field-group { display: flex; flex-direction: column; gap: 10px; }

.provider-toggle {
  display: flex;
  background: rgba(255,255,255,0.04);
  border: 1.5px solid rgba(255,255,255,0.08);
  border-radius: 10px;
  padding: 4px;
  gap: 4px;
}

.provider-btn {
  flex: 1;
  padding: 8px;
  border: none;
  background: transparent;
  color: rgba(255,255,255,0.5);
  font-size: 0.8rem;
  font-weight: 600;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.provider-btn.active {
  background: rgba(255,255,255,0.1);
  color: #fff;
}

.model-select {
  background: rgba(255,255,255,0.04);
  border: 1.5px solid rgba(255,255,255,0.08);
  border-radius: 10px;
  padding: 10px;
  color: #fff;
  font-size: 0.85rem;
  outline: none;
  cursor: pointer;
  transition: all 0.2s;
}

/* ── Voice Grid ── */
.voice-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}

.voice-card {
  background: rgba(255,255,255,0.02);
  border: 1px solid var(--card-border);
  border-radius: var(--radius-md);
  padding: 12px;
  cursor: pointer;
  transition: var(--transition-smooth);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  position: relative;
  overflow: hidden;
}

.voice-card:hover {
  background: rgba(255,255,255,0.05);
  border-color: rgba(255,255,255,0.15);
  transform: translateY(-2px);
}

.voice-card.active {
  background: rgba(255, 255, 255, 0.08);
  border-color: #63b3ff;
  box-shadow: 0 4px 15px rgba(99, 179, 255, 0.2);
}

.voice-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.voice-avatar { font-size: 1.6rem; }
.voice-name { font-weight: 700; font-size: 0.85rem; color: var(--text-muted); transition: var(--transition-fast); }
.voice-card.active .voice-name { color: #fff; }

.gender-badge {
  font-size: 0.65rem;
  font-weight: 800;
  padding: 2px 8px;
  border-radius: var(--radius-full);
  background: rgba(255,255,255,0.05);
  color: var(--text-dim);
}

.voice-card.active .gender-badge {
  background: rgba(99, 179, 255, 0.2);
  color: #63b3ff;
}


/* ── Config Summary ── */
.config-summary {
  display: flex;
  gap: 12px;
  margin-top: 4px;
}

.summary-item {
  flex: 1;
  background: rgba(255,255,255,0.02);
  border: 1px solid var(--card-border);
  border-radius: var(--radius-md);
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.item-label { font-size: 0.65rem; font-weight: 800; color: var(--text-dim); text-transform: uppercase; }
.item-val { font-size: 0.85rem; font-weight: 700; color: var(--text-muted); }
</style>
