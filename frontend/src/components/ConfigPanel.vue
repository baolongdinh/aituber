<template>
  <div class="config-panel">
    <div class="section-title">
      <span>🎙️</span> Cài đặt giọng đọc
    </div>

    <!-- TTS Provider Selector -->
    <div class="field-group">
      <label class="field-label">Công nghệ TTS</label>
      <div class="provider-toggle">
        <button 
          class="provider-btn" 
          :class="{ active: localConfig.tts_provider !== 'elevenlabs' }"
          @click="setProvider('fpt')"
        >
          FPT.AI (Standard)
        </button>
        <button 
          class="provider-btn" 
          :class="{ active: localConfig.tts_provider === 'elevenlabs' }"
          @click="setProvider('elevenlabs')"
        >
          ElevenLabs (Pro)
        </button>
      </div>
    </div>

    <!-- Voice Selector -->
    <div class="field-group">
      <label class="field-label">Giọng đọc</label>
      <div class="voice-grid">
        <button
          v-for="v in voiceOptions"
          :key="v.value"
          class="voice-btn"
          :class="{ active: localConfig.voice === v.value }"
          @click="setVoice(v.value)"
        >
          <span class="voice-gender">{{ v.gender }}</span>
          <span class="voice-name">{{ v.name }}</span>
          <span class="voice-region">{{ v.region }}</span>
        </button>
      </div>
    </div>

    <!-- Speaking Speed -->
    <div class="field-group">
      <label class="field-label">
        Tốc độ đọc
        <span class="speed-badge">{{ localConfig.speaking_speed.toFixed(1) }}x</span>
      </label>
      <input
        type="range"
        v-model.number="localConfig.speaking_speed"
        min="0.5"
        max="2.0"
        step="0.1"
        class="speed-slider"
        @input="emitUpdate"
      />
      <div class="speed-labels">
        <span>Chậm 0.5x</span>
        <span>Bình thường 1.0x</span>
        <span>Nhanh 2.0x</span>
      </div>
    </div>

    <!-- AI Powered note -->
    <div class="ai-note">
      <span class="ai-icon">✨</span>
      <div>
        <div class="ai-title">Powered by Gemini AI</div>
        <div class="ai-desc">Script tự động sinh từ topic của bạn. Stock video từ Pexels.</div>
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

.model-select:focus {
  border-color: rgba(255,255,255,0.2);
  background: rgba(255,255,255,0.08);
}

.model-select option {
  background: #1a1a1a;
  color: #fff;
}

.field-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.88rem;
  font-weight: 600;
  color: rgba(255,255,255,0.8);
}

.speed-badge {
  background: rgba(255,255,255,0.1);
  padding: 2px 10px;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 700;
  color: #fff;
}

.voice-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}

.voice-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 10px 8px;
  background: rgba(255,255,255,0.04);
  border: 1.5px solid rgba(255,255,255,0.08);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  color: rgba(255,255,255,0.7);
}
.voice-btn:hover { background: rgba(255,255,255,0.08); border-color: rgba(255,255,255,0.2); }
.voice-btn.active {
  background: rgba(255,255,255,0.12);
  border-color: rgba(255,255,255,0.4);
  color: #fff;
}
.voice-gender { font-size: 1.2rem; }
.voice-name { font-size: 0.8rem; font-weight: 600; }
.voice-region {
  font-size: 0.7rem;
  background: rgba(255,255,255,0.08);
  padding: 1px 6px;
  border-radius: 6px;
  color: rgba(255,255,255,0.5);
}

.speed-slider {
  width: 100%;
  -webkit-appearance: none;
  appearance: none;
  height: 4px;
  border-radius: 4px;
  background: rgba(255,255,255,0.15);
  outline: none;
}
.speed-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: #fff;
  cursor: pointer;
  box-shadow: 0 0 8px rgba(0,0,0,0.4);
}

.speed-labels {
  display: flex;
  justify-content: space-between;
  font-size: 0.72rem;
  color: rgba(255,255,255,0.35);
}

.ai-note {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 12px;
  padding: 14px;
  margin-top: 4px;
}
.ai-icon { font-size: 1.3rem; flex-shrink: 0; }
.ai-title { font-size: 0.85rem; font-weight: 600; color: rgba(255,255,255,0.85); margin-bottom: 2px; }
.ai-desc { font-size: 0.78rem; color: rgba(255,255,255,0.45); }
</style>
