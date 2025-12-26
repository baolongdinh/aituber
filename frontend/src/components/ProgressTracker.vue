<template>
  <v-card>
    <v-card-title class="text-h5">
      <v-icon left>mdi-progress-clock</v-icon>
      Progress
    </v-card-title>
    <v-card-text>
      <!-- Status Badge -->
      <v-chip
        :color="statusColor"
        :prepend-icon="statusIcon"
        size="large"
        class="mb-4"
      >
        {{ statusText }}
      </v-chip>

      <!-- Progress Bar -->
      <v-progress-linear
        v-if="status !== 'idle'"
        :model-value="progress"
        :color="progressColor"
        height="30"
        striped
        class="mb-4"
      >
        <template v-slot:default>
          <strong class="text-white">{{ progress }}%</strong>
        </template>
      </v-progress-linear>

      <!-- Current Step -->
      <v-alert
        v-if="currentStep && status === 'processing'"
        type="info"
        variant="tonal"
        class="mb-4"
      >
        <strong>Current Step:</strong> {{ currentStep }}
      </v-alert>

      <!-- Error Message -->
      <v-alert
        v-if="error"
        type="error"
        variant="tonal"
        closable
        class="mb-4"
      >
        <strong>Error:</strong> {{ error }}
      </v-alert>

      <!-- Processing Steps Timeline -->
      <v-timeline
        v-if="status !== 'idle'"
        density="compact"
        align="start"
        class="mt-4"
      >
        <v-timeline-item
          v-for="step in steps"
          :key="step.name"
          :dot-color="step.completed ? 'success' : (step.current ? 'primary' : 'grey')"
          size="small"
        >
          <div class="d-flex align-center">
            <v-icon
              :color="step.completed ? 'success' : (step.current ? 'primary' : 'grey')"
              size="small"
              class="mr-2"
            >
              {{ step.completed ? 'mdi-check-circle' : (step.current ? 'mdi-loading mdi-spin' : 'mdi-circle-outline') }}
            </v-icon>
            <span :class="{ 'font-weight-bold': step.current }">{{ step.name }}</span>
          </div>
        </v-timeline-item>
      </v-timeline>
    </v-card-text>
  </v-card>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  status: {
    type: String,
    default: 'idle'
  },
  progress: {
    type: Number,
    default: 0
  },
  currentStep: {
    type: String,
    default: ''
  },
  error: {
    type: String,
    default: null
  }
})

const statusColor = computed(() => {
  switch (props.status) {
    case 'processing': return 'primary'
    case 'completed': return 'success'
    case 'failed': return 'error'
    default: return 'grey'
  }
})

const statusIcon = computed(() => {
  switch (props.status) {
    case 'processing': return 'mdi-loading mdi-spin'
    case 'completed': return 'mdi-check-circle'
    case 'failed': return 'mdi-alert-circle'
    default: return 'mdi-help-circle'
  }
})

const statusText = computed(() => {
  switch (props.status) {
    case 'processing': return 'Processing...'
    case 'completed': return 'Completed'
    case 'failed': return 'Failed'
    default: return 'Ready'
  }
})

const progressColor = computed(() => {
  return props.status === 'completed' ? 'success' : 'primary'
})

const steps = computed(() => {
  const allSteps = [
    { name: 'Initializing', threshold: 5 },
    { name: 'Splitting text for audio', threshold: 10 },
    { name: 'Generating audio chunks', threshold: 20 },
    { name: 'Merging audio with crossfade', threshold: 40 },
    { name: 'Splitting text for video', threshold: 45 },
    { name: 'Generating video prompts', threshold: 50 },
    { name: 'Generating video segments', threshold: 55 },
    { name: 'Merging video with transitions', threshold: 80 },
    { name: 'Composing final video', threshold: 90 },
    { name: 'Complete', threshold: 100 }
  ]

  return allSteps.map(step => ({
    ...step,
    completed: props.progress > step.threshold,
    current: props.progress >= step.threshold && props.progress < (allSteps[allSteps.indexOf(step) + 1]?.threshold || 101)
  }))
})
</script>
