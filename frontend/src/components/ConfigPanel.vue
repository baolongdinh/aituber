<template>
  <div class="config-panel">
    <div class="panel-title">
      <span class="title-icon">🎙️</span> Cài đặt giọng đọc
    </div>

    <!-- Provider Selection -->
    <div class="config-group">
      <label class="group-label">Chọn nhà cung cấp TTS</label>
      <div class="provider-toggle">
        <button
          class="provider-btn"
          :class="{ active: localConfig.tts_provider === 'fpt' }"
          @click="setProvider('fpt')"
        >
          <span class="provider-icon">🌐</span>
          FPT.AI
        </button>
        <button
          class="provider-btn"
          :class="{ active: localConfig.tts_provider === 'hub' }"
          @click="setProvider('hub')"
        >
          <span class="provider-icon">🚀</span>
          Hub
        </button>
      </div>
    </div>

    <!-- Gender Selection -->
    <div class="config-group">
      <label class="group-label">Chọn giới tính</label>
      <div class="gender-toggle">
        <button
          class="gender-btn"
          :class="{ active: selectedGender === 'male' }"
          @click="setGender('male')"
        >
          <span class="gender-icon">👨</span>
          Nam
        </button>
        <button
          class="gender-btn"
          :class="{ active: selectedGender === 'female' }"
          @click="setGender('female')"
        >
          <span class="gender-icon">👩</span>
          Nữ
        </button>
      </div>
    </div>

    <!-- Voice Selector -->
    <div class="config-group">
      <label class="group-label">Chọn giọng AI</label>
      <div class="voice-grid">
        <div
          v-for="v in filteredVoices"
          :key="v.key"
          class="voice-card"
          :class="{ active: localConfig.voice === v.key }"
          @click="setVoice(v.key)"
        >
          <div class="voice-row">
            <span class="voice-avatar">{{ v.gender === 'male' ? '👨' : '👩' }}</span>
            <span class="voice-name">{{ v.label }}</span>
          </div>
          <div class="voice-badges">
            <span class="gender-badge">{{ v.gender === 'male' ? 'Nam' : 'Nữ' }}</span>
            <span class="provider-badge">{{ localConfig.tts_provider.toUpperCase() }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Current Selection Summary -->
    <div class="config-summary">
      <div class="summary-item">
        <div class="item-label">Nhà cung cấp</div>
        <div class="item-val">{{ localConfig.tts_provider?.toUpperCase() || 'FPT' }}</div>
      </div>
      <div class="summary-item">
        <div class="item-label">Giới tính</div>
        <div class="item-val">{{ selectedGender === 'male' ? 'Nam' : 'Nữ' }}</div>
      </div>
      <div class="summary-item">
        <div class="item-label">Giọng đọc</div>
        <div class="item-val">{{ currentVoiceLabel }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, watch, computed, ref, onMounted } from 'vue'
import { api } from '@/utils/api'

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
const selectedGender = ref('female')
const voiceCatalog = ref([])

watch(() => props.modelValue, (val) => {
  Object.assign(localConfig, val)
}, { deep: true })

const emitUpdate = () => emit('update:modelValue', { ...localConfig })

// Load voice catalog from API
const loadVoiceCatalog = async () => {
  try {
    const response = await api.get('/api/v1/voices/catalog')
    voiceCatalog.value = response.data.all_voices || []
  } catch (error) {
    console.error('Failed to load voice catalog:', error)
    // Fallback to hardcoded voices
    voiceCatalog.value = [
      { key: 'banmai', label: 'Ban Mai', gender: 'female', providerSupport: ['fpt', 'hub'] },
      { key: 'leminh', label: 'Lê Minh', gender: 'female', providerSupport: ['fpt', 'hub'] },
      { key: 'minhquang', label: 'Minh Quang', gender: 'male', providerSupport: ['fpt', 'hub'] },
      { key: 'giahuy', label: 'Gia Huy', gender: 'male', providerSupport: ['fpt', 'hub'] },
    ]
  }
}

// Filter voices based on provider and gender
const filteredVoices = computed(() => {
  return voiceCatalog.value.filter(voice => {
    const supportsProvider = voice.providerSupport?.includes(localConfig.tts_provider)
    const matchesGender = voice.gender === selectedGender.value
    return supportsProvider && matchesGender
  })
})

// Get current voice label for summary
const currentVoiceLabel = computed(() => {
  const voice = voiceCatalog.value.find(v => v.key === localConfig.voice)
  return voice?.label || localConfig.voice
})

const setVoice = (voiceKey) => {
  localConfig.voice = voiceKey
  emitUpdate()
}

const setProvider = (provider) => {
  localConfig.tts_provider = provider
  // Reset voice selection if current voice doesn't support new provider
  const currentVoice = voiceCatalog.value.find(v => v.key === localConfig.voice)
  if (!currentVoice?.providerSupport?.includes(provider)) {
    // Find first available voice for new provider and gender
    const availableVoice = filteredVoices.value[0]
    if (availableVoice) {
      localConfig.voice = availableVoice.key
    }
  }
  emitUpdate()
}

const setGender = (gender) => {
  selectedGender.value = gender
  // Reset voice selection if no voices available for current gender
  if (filteredVoices.value.length === 0) {
    // Try to find a voice for the other gender
    const otherGender = gender === 'male' ? 'female' : 'male'
    selectedGender.value = otherGender
  }
  // Select first available voice
  const availableVoice = filteredVoices.value[0]
  if (availableVoice) {
    localConfig.voice = availableVoice.key
  }
  emitUpdate()
}

// Initialize on mount
onMounted(() => {
  loadVoiceCatalog()
  // Set initial gender based on current voice
  const currentVoice = voiceCatalog.value.find(v => v.key === localConfig.voice)
  if (currentVoice) {
    selectedGender.value = currentVoice.gender
  }
})
</script>

<style scoped>
.config-panel {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.panel-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1.1rem;
  font-weight: 700;
  color: #fff;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(255,255,255,0.1);
}

.title-icon {
  font-size: 1.3rem;
}

.config-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.group-label {
  font-size: 0.9rem;
  font-weight: 600;
  color: rgba(255,255,255,0.8);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

/* Provider Toggle */
.provider-toggle {
  display: flex;
  background: rgba(255,255,255,0.04);
  border: 1.5px solid rgba(255,255,255,0.08);
  border-radius: 12px;
  padding: 4px;
  gap: 4px;
}

.provider-btn {
  flex: 1;
  padding: 12px 16px;
  border: none;
  background: transparent;
  color: rgba(255,255,255,0.6);
  font-size: 0.85rem;
  font-weight: 600;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.provider-btn:hover {
  background: rgba(255,255,255,0.05);
  color: rgba(255,255,255,0.8);
}

.provider-btn.active {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.3);
}

.provider-icon {
  font-size: 1.1rem;
}

/* Gender Toggle */
.gender-toggle {
  display: flex;
  background: rgba(255,255,255,0.04);
  border: 1.5px solid rgba(255,255,255,0.08);
  border-radius: 12px;
  padding: 4px;
  gap: 4px;
}

.gender-btn {
  flex: 1;
  padding: 10px 16px;
  border: none;
  background: transparent;
  color: rgba(255,255,255,0.6);
  font-size: 0.85rem;
  font-weight: 600;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.gender-btn:hover {
  background: rgba(255,255,255,0.05);
  color: rgba(255,255,255,0.8);
}

.gender-btn.active {
  background: rgba(255,255,255,0.1);
  color: #fff;
  border-color: rgba(255,255,255,0.2);
}

.gender-icon {
  font-size: 1.1rem;
}

/* Voice Grid */
.voice-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 12px;
}

.voice-card {
  background: rgba(255,255,255,0.02);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 12px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  position: relative;
  overflow: hidden;
}

.voice-card:hover {
  background: rgba(255,255,255,0.05);
  border-color: rgba(255,255,255,0.15);
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0,0,0,0.2);
}

.voice-card.active {
  background: rgba(255, 255, 255, 0.08);
  border-color: #63b3ff;
  box-shadow: 0 4px 20px rgba(99, 179, 255, 0.3);
}

.voice-row {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  justify-content: center;
}

.voice-avatar { 
  font-size: 1.8rem; 
}

.voice-name { 
  font-weight: 700; 
  font-size: 0.85rem; 
  color: var(--text-muted); 
  transition: var(--transition-fast);
  text-align: center;
}

.voice-card.active .voice-name { 
  color: #fff; 
}

.voice-badges {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  justify-content: center;
}

.gender-badge, .provider-badge {
  font-size: 0.6rem;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 20px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.gender-badge {
  background: rgba(255,255,255,0.05);
  color: var(--text-dim);
}

.provider-badge {
  background: rgba(99, 179, 255, 0.1);
  color: #63b3ff;
}

.voice-card.active .gender-badge {
  background: rgba(99, 179, 255, 0.2);
  color: #63b3ff;
}

.voice-card.active .provider-badge {
  background: rgba(99, 179, 255, 0.3);
  color: #80c7ff;
}

/* Config Summary */
.config-summary {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-top: 8px;
  padding: 16px;
  background: rgba(255,255,255,0.02);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 12px;
}

.summary-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  text-align: center;
}

.item-label { 
  font-size: 0.65rem; 
  font-weight: 800; 
  color: var(--text-dim); 
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.item-val { 
  font-size: 0.85rem; 
  font-weight: 700; 
  color: var(--text-muted); 
}

/* Responsive */
@media (max-width: 768px) {
  .voice-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .config-summary {
    grid-template-columns: 1fr;
    gap: 8px;
  }
  
  .summary-item {
    flex-direction: row;
    justify-content: space-between;
    text-align: left;
  }
}
</style>
