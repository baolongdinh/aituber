<template>
  <v-card class="mt-4">
    <v-card-title class="text-h5">
      <v-icon left>mdi-cog</v-icon>
      Configuration
    </v-card-title>
    <v-card-text>
      <v-select
        v-model="localConfig.voice"
        :items="voiceOptions"
        label="Voice"
        variant="outlined"
        prepend-inner-icon="mdi-account-voice"
      />

      <div class="mt-4">
        <label class="text-subtitle-2 mb-2 d-block">Speaking Speed: {{ localConfig.speaking_speed }}x</label>
        <v-slider
          v-model="localConfig.speaking_speed"
          min="0.5"
          max="2.0"
          step="0.1"
          thumb-label
          color="primary"
        />
      </div>

      <v-select
        v-model="localConfig.video_style"
        :items="styleOptions"
        label="Video Style"
        variant="outlined"
        prepend-inner-icon="mdi-palette"
        class="mt-4"
        v-if="localConfig.video_source === 'ai'"
      />

      <v-divider class="my-4"></v-divider>

      <label class="text-subtitle-2 mb-2 d-block">Video Source</label>
      <v-radio-group v-model="localConfig.video_source" inline>
        <v-radio label="AI Generated" value="ai"></v-radio>
        <v-radio label="Stock Video (Pexels)" value="stock"></v-radio>
      </v-radio-group>

      <v-expand-transition>
        <div v-if="localConfig.video_source === 'stock'">
          <v-text-field
            v-model="localConfig.stock_keywords"
            label="Search Keywords (e.g., nature, city)"
            placeholder="Enter keywords to search for video"
            variant="outlined"
            prepend-inner-icon="mdi-magnify"
            hint="We will find a video matching these keywords and loop it to match audio duration"
            persistent-hint
          />
        </div>
      </v-expand-transition>
    </v-card-text>
  </v-card>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({
      voice: 'banmai',
      speaking_speed: 1.0,
      video_style: 'modern',
      video_source: 'ai',
      stock_keywords: ''
    })
  }
})

const emit = defineEmits(['update:modelValue'])

const localConfig = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const voiceOptions = [
  { title: 'Female North (Ban Mai)', value: 'banmai' },
  { title: 'Female South (Lê Minh)', value: 'leminh' },
  { title: 'Male North (Minh Quang)', value: 'minhquang' },
  { title: 'Male South (Gia Huy)', value: 'giahuy' }
]

const styleOptions = [
  { title: 'Modern', value: 'modern' },
  { title: 'Cinematic', value: 'cinematic' },
  { title: 'Minimal', value: 'minimal' },
  { title: 'Abstract', value: 'abstract' }
]
</script>
