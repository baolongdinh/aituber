import { ref } from 'vue'
import { videoApi } from '../api/videoApi'

export function useVideoGeneration() {
    const generating = ref(false)
    const progress = ref(0)
    const currentStep = ref('')
    const jobStatus = ref('idle')
    const videoUrl = ref(null)
    const error = ref(null)
    const jobId = ref(null)

    let pollInterval = null

    const pollStatus = async (id) => {
        pollInterval = setInterval(async () => {
            try {
                const status = await videoApi.getStatus(id)

                progress.value = status.progress
                currentStep.value = status.current_step
                jobStatus.value = status.status

                if (status.status === 'completed') {
                    clearInterval(pollInterval)
                    videoUrl.value = status.video_url
                    generating.value = false
                } else if (status.status === 'failed') {
                    clearInterval(pollInterval)
                    error.value = status.error || 'Video generation failed'
                    generating.value = false
                    jobStatus.value = 'failed'
                }
            } catch (err) {
                console.error('Failed to poll status:', err)
                error.value = err.message
            }
        }, 2000) // Poll every 2 seconds
    }

    const generateVideo = async (script, config) => {
        generating.value = true
        error.value = null
        progress.value = 0
        currentStep.value = 'Initializing...'
        jobStatus.value = 'processing'
        videoUrl.value = null

        try {
            const result = await videoApi.generateVideo({
                script,
                ...config
            })

            jobId.value = result.job_id
            pollStatus(result.job_id)
        } catch (err) {
            console.error('Failed to start video generation:', err)
            error.value = err.response?.data?.error || err.message
            generating.value = false
            jobStatus.value = 'failed'
        }
    }

    const reset = () => {
        if (pollInterval) {
            clearInterval(pollInterval)
        }
        generating.value = false
        progress.value = 0
        currentStep.value = ''
        jobStatus.value = 'idle'
        videoUrl.value = null
        error.value = null
        jobId.value = null
    }

    return {
        generating,
        progress,
        currentStep,
        jobStatus,
        videoUrl,
        error,
        jobId,
        generateVideo,
        reset
    }
}
