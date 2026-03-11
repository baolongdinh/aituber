<template>
  <div class="script-input">
    <div class="input-header">
      <div class="input-label">
        <span class="label-icon">📄</span> Kịch bản văn bản (Long Form)
      </div>
      <div class="stats-row">
        <div class="stat-badge primary">
          {{ characterCount.toLocaleString() }} / 50,000 ký tự
        </div>
        <div class="stat-badge secondary">
          ~{{ estimatedWords.toLocaleString() }} từ
        </div>
      </div>
    </div>

    <div class="textarea-wrap">
      <textarea
        v-model="localScript"
        placeholder="Dán nội dung văn bản dài vào đây... (Tối đa 50,000 ký tự)"
        class="custom-textarea"
        maxlength="50000"
      ></textarea>
      
      <div class="textarea-glow"></div>
    </div>

    <div class="input-footer">
      <div class="tip-msg">
        <span class="tip-icon">💡</span>
        Hệ thống sẽ tự động tách đoạn và gán video footage phù hợp với nội dung của bạn.
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: { type: String, default: '' }
})

const emit = defineEmits(['update:modelValue'])

const localScript = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const characterCount = computed(() => localScript.value.length)
const estimatedWords = computed(() => {
  const words = localScript.value.trim().split(/\s+/)
  return words[0] === '' ? 0 : words.length
})
</script>

<style scoped>
.script-input {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.input-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
}

.input-label {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 1rem;
  font-weight: 800;
  color: #fff;
  letter-spacing: -0.01em;
}

.stats-row {
  display: flex;
  gap: 8px;
}

.stat-badge {
  font-size: 0.7rem;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: var(--radius-full);
}

.stat-badge.primary {
  background: rgba(99, 179, 255, 0.1);
  color: #63b3ff;
  border: 1px solid rgba(99, 179, 255, 0.15);
}

.stat-badge.secondary {
  background: rgba(167, 139, 250, 0.1);
  color: #a78bfa;
  border: 1px solid rgba(167, 139, 250, 0.15);
}

/* ── Textarea ── */
.textarea-wrap {
  position: relative;
  width: 100%;
}

.custom-textarea {
  width: 100%;
  height: 380px;
  background: rgba(0,0,0,0.3);
  border: 1.5px solid var(--card-border);
  border-radius: var(--radius-md);
  color: #fff;
  font-size: 1rem;
  padding: 20px;
  font-family: 'Inter', system-ui, sans-serif;
  line-height: 1.6;
  resize: vertical;
  transition: all var(--transition-fast);
  outline: none;
  position: relative;
  z-index: 2;
}

.custom-textarea:focus {
  background: rgba(0,0,0,0.4);
  border-color: rgba(255,255,255,0.2);
  box-shadow: 0 8px 30px rgba(0,0,0,0.3);
}

.textarea-glow {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  border-radius: var(--radius-md);
  background: linear-gradient(135deg, rgba(99, 179, 255, 0.1), rgba(167, 139, 250, 0.1));
  filter: blur(20px);
  opacity: 0;
  transition: opacity 0.4s ease;
  z-index: 1;
}

.custom-textarea:focus + .textarea-glow {
  opacity: 1;
}

.input-footer {
  padding: 12px 18px;
  background: rgba(255,255,255,0.02);
  border-radius: var(--radius-md);
  border: 1px solid var(--glass-border);
}

.tip-msg {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 0.8rem;
  color: var(--text-muted);
  font-weight: 500;
}

.tip-icon { font-size: 1.1rem; }
</style>
