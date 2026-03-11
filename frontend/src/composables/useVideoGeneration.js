import { ref } from "vue";
import { videoApi } from "../api/videoApi";

export function useVideoGeneration() {
  const generating = ref(false);
  const progress = ref(0);
  const currentStep = ref("");
  const jobStatus = ref("idle");
  const videoUrl = ref(null);
  const savedPath = ref(null);
  const error = ref(null);
  const jobId = ref(null);

  // Series-specific state
  const isSeries = ref(false);
  const seriesParts = ref([]);
  const seriesId = ref(null);

  let pollInterval = null;

  const pollStatus = async (id, checkingSeries) => {
    // Clear existing if any
    if (pollInterval) clearInterval(pollInterval);

    pollInterval = setInterval(async () => {
      try {
        if (checkingSeries) {
          const status = await videoApi.getSeriesStatus(id);
          progress.value = status.overall_progress;
          jobStatus.value = status.status;
          seriesParts.value = status.parts || [];

          if (status.status === "completed" || status.status === "failed" || status.status === "partial_failed") {
            clearInterval(pollInterval);
            pollInterval = null;
            generating.value = false;
            if (status.status === "failed") {
              error.value = "Quá trình tạo Series thất bại";
            }
          }
        } else {
          // Single video checking
          const status = await videoApi.getStatus(id);
          progress.value = status.progress;
          currentStep.value = status.current_step;
          jobStatus.value = status.status;

          if (status.status === "completed") {
            clearInterval(pollInterval);
            pollInterval = null;
            videoUrl.value = status.video_url;
            savedPath.value = status.saved_path || null;
            generating.value = false;
          } else if (status.status === "failed") {
            clearInterval(pollInterval);
            pollInterval = null;
            error.value = status.error || "Video generation failed";
            generating.value = false;
            jobStatus.value = "failed";
          }
        }
      } catch (err) {
        console.error("Failed to poll status:", err);
        error.value = err.message;
        // Don't necessarily stop polling on network error, but maybe after too many?
      }
    }, 2000); // Poll every 2 seconds
  };

  const generateVideo = async (topic, contentName, platform, config, isSeriesMode = false, numParts = 2) => {
    generating.value = true;
    error.value = null;
    progress.value = 0;
    currentStep.value = "Initializing...";
    jobStatus.value = "processing";
    videoUrl.value = null;
    savedPath.value = null;

    isSeries.value = isSeriesMode;
    seriesParts.value = [];
    seriesId.value = null;
    jobId.value = null;

    try {
      if (isSeriesMode) {
        const result = await videoApi.generateSeries({
          topic,
          content_name: contentName,
          platform,
          num_parts: numParts,
          voice: config.voice,
          speaking_speed: config.speaking_speed,
          tts_provider: config.tts_provider,
        });

        seriesId.value = result.series_id;
        pollStatus(result.series_id, true);
      } else {
        const result = await videoApi.generateVideo({
          topic,
          content_name: contentName,
          platform,
          voice: config.voice,
          speaking_speed: config.speaking_speed,
          tts_provider: config.tts_provider,
        });

        jobId.value = result.job_id;
        pollStatus(result.job_id, false);
      }
    } catch (err) {
      console.error("Failed to start generation:", err);
      error.value = err.response?.data?.error || err.message;
      generating.value = false;
      jobStatus.value = "failed";
    }
  };

  const reset = () => {
    if (pollInterval) {
      clearInterval(pollInterval);
    }
    generating.value = false;
    progress.value = 0;
    currentStep.value = "";
    jobStatus.value = "idle";
    videoUrl.value = null;
    savedPath.value = null;
    error.value = null;
    jobId.value = null;
    isSeries.value = false;
    seriesParts.value = [];
    seriesId.value = null;
  };

  const retryPart = async (idx) => {
    if (!seriesId.value) return;

    try {
      error.value = null;
      jobStatus.value = "processing";
      generating.value = true;

      await videoApi.retrySeriesPart(seriesId.value, idx);

      // If polling isn't active, restart it
      if (!pollInterval) {
        pollStatus(seriesId.value, true);
      }
    } catch (err) {
      console.error("Failed to retry part:", err);
      error.value = `Retry failed: ${err.response?.data?.error || err.message}`;
    }
  };

  return {
    generating,
    progress,
    currentStep,
    jobStatus,
    videoUrl,
    savedPath,
    error,
    jobId,
    isSeries,
    seriesParts,
    seriesId,
    generateVideo,
    retryPart,
    reset,
  };
}

