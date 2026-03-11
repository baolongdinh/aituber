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

    <!-- Speaking Speed -->
    <div class="config-group">
      <label class="group-label">Tốc độ & Cảm xúc</label>
      <div class="speed-controls">
        <div class="speed-slider-wrap">
          <input
            type="range"
            v-model.number="localConfig.speaking_speed"
            min="0.5"
            max="2.0"
            step="0.1"
            class="range-slider"
            @input="emitUpdate"
          />
        </div>
        <div class="speed-value-card">
          <span class="speed-num">{{ localConfig.speaking_speed.toFixed(1) }}</span>
          <span class="speed-unit">x</span>
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
      speaking_speed: 1.0,
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

/* ── Speed Controls ── */
.speed-controls {
  background: rgba(0,0,0,0.2);
  border: 1px solid var(--card-border);
  border-radius: var(--radius-md);
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 20px;
}

.speed-slider-wrap { flex: 1; }

.range-slider {
  width: 100%;
  height: 6px;
  appearance: none;
  background: rgba(255,255,255,0.1);
  border-radius: var(--radius-full);
  outline: none;
}

.range-slider::-webkit-slider-thumb {
  appearance: none;
  width: 18px;
  height: 18px;
  background: #fff;
  border: 3px solid #63b3ff;
  border-radius: 50%;
  box-shadow: 0 0 10px rgba(99, 179, 255, 0.4);
  cursor: pointer;
  transition: var(--transition-fast);
}

.range-slider::-webkit-slider-thumb:hover { transform: scale(1.1); }

.speed-value-card {
  background: rgba(255,255,255,0.05);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  padding: 6px 12px;
  min-width: 60px;
  text-align: center;
}

.speed-num { font-weight: 800; font-size: 1rem; color: #fff; }
.speed-unit { font-size: 0.7rem; font-weight: 700; color: var(--text-dim); margin-left: 2px; }

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
