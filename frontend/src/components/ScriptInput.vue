<template>
  <v-card>
    <v-card-title class="text-h5">
      <v-icon left>mdi-text-box</v-icon>
      Text Script
    </v-card-title>
    <v-card-text>
      <v-textarea
        v-model="localScript"
        placeholder="Paste your long text script here... (max 50,000 characters)"
        rows="15"
        variant="outlined"
        counter
        :rules="rules"
        :maxlength="50000"
      />
      <div class="text-caption text-grey mt-2">
        <v-chip size="small" color="primary" variant="outlined">
          {{ characterCount }} / 50,000 characters
        </v-chip>
        <v-chip size="small" color="secondary" variant="outlined" class="ml-2">
          ~{{ estimatedWords }} words
        </v-chip>
      </div>
    </v-card-text>
  </v-card>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  }
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

const rules = [
  v => v.length <= 50000 || 'Script must be less than 50,000 characters'
]
</script>
