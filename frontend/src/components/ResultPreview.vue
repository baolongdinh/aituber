<template>
  <v-card v-if="videoUrl">
    <v-card-title class="text-h5">
      <v-icon left color="success">mdi-check-circle</v-icon>
      Video Ready!
    </v-card-title>
    <v-card-text>
      <!-- Video Player -->
      <div class="video-container mb-4">
        <video
          :src="videoUrl"
          controls
          class="video-player"
        />
      </div>

      <!-- Action Buttons -->
      <div class="d-flex gap-2">
        <v-btn
          :href="videoUrl"
          download
          color="primary"
          size="large"
          prepend-icon="mdi-download"
        >
          Download Video
        </v-btn>

        <v-btn
          color="secondary"
          size="large"
          prepend-icon="mdi-refresh"
          @click="$emit('reset')"
        >
          Generate Another
        </v-btn>

        <v-btn
          color="info"
          size="large"
          prepend-icon="mdi-share-variant"
          @click="copyLink"
        >
          Copy Link
        </v-btn>
      </div>

      <!-- Success Message -->
      <v-alert
        v-if="linkCopied"
        type="success"
        variant="tonal"
        closable
        class="mt-4"
      >
        Link copied to clipboard!
      </v-alert>
    </v-card-text>
  </v-card>
</template>

<script setup>
import { ref } from 'vue'

defineProps({
  videoUrl: {
    type: String,
    default: null
  }
})

defineEmits(['reset'])

const linkCopied = ref(false)

const copyLink = async () => {
  try {
    await navigator.clipboard.writeText(window.location.origin + props.videoUrl)
    linkCopied.value = true
    setTimeout(() => {
      linkCopied.value = false
    }, 3000)
  } catch (err) {
    console.error('Failed to copy link:', err)
  }
}
</script>

<script>
export default {
  name: 'ResultPreview'
}
</script>

<style scoped>
.video-container {
  width: 100%;
  background: #000;
  border-radius: 8px;
  overflow: hidden;
}

.video-player {
  width: 100%;
  height: auto;
  display: block;
}
</style>
