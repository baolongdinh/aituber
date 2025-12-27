<template>
  <v-app>
    <v-app-bar color="primary" dark prominent>
      <v-toolbar-title class="text-h4 font-weight-bold">
        <v-icon size="large" class="mr-2">mdi-video-vintage</v-icon>
        AI Video Generator
      </v-toolbar-title>
      <v-spacer />
      <v-chip variant="outlined">
        <v-icon left>mdi-tools</v-icon>
        Beta
      </v-chip>
    </v-app-bar>

    <v-main>
      <v-container fluid class="pa-6">
        <v-row>
          <!-- Left Column: Input & Config -->
          <v-col cols="12" md="6">
            <ScriptInput v-model="script" />
            <ConfigPanel v-model="config" />

            <v-btn
              block
              size="x-large"
              color="primary"
              :loading="generating"
              :disabled="!canGenerate"
              @click="handleGenerate"
              class="mt-4"
            >
              <v-icon left>mdi-play-circle</v-icon>
              Generate Video
            </v-btn>
          </v-col>

          <!-- Right Column: Progress & Result -->
          <v-col cols="12" md="6">
            <ProgressTracker
              :status="jobStatus"
              :progress="progress"
              :current-step="currentStep"
              :error="error"
            />

            <div class="mt-4">
              <ResultPreview
                :video-url="videoUrl"
                :job-id="jobId"
                @reset="handleReset"
              />
            </div>
          </v-col>
        </v-row>
      </v-container>
    </v-main>

    <v-footer app color="grey-lighten-3" class="pa-4">
      <div class="text-center w-100">
        <span class="text-caption">
          Built with ❤️ for seamless video generation |
          <a href="https://github.com" target="_blank">GitHub</a>
        </span>
      </div>
    </v-footer>
  </v-app>
</template>

<script setup>
import { ref, computed } from "vue";
import ScriptInput from "./components/ScriptInput.vue";
import ConfigPanel from "./components/ConfigPanel.vue";
import ProgressTracker from "./components/ProgressTracker.vue";
import ResultPreview from "./components/ResultPreview.vue";
import { useVideoGeneration } from "./composables/useVideoGeneration";

// State
const script = ref("");
const config = ref({
  voice: "banmai",
  speaking_speed: 1.0,
  video_style: "modern",
  video_source: "ai",
  stock_keywords: "",
});

// Video generation composable
const {
  generating,
  progress,
  currentStep,
  jobStatus,
  videoUrl,
  error,
  jobId,
  generateVideo,
  reset,
} = useVideoGeneration();

// Computed
const canGenerate = computed(() => {
  return (
    script.value.length > 0 && script.value.length <= 50000 && !generating.value
  );
});

// Methods
const handleGenerate = async () => {
  await generateVideo(script.value, config.value);
};

const handleReset = () => {
  reset();
  script.value = "";
};
</script>

<style>
/* Global styles */
.v-application {
  font-family: "Roboto", sans-serif;
}
</style>
