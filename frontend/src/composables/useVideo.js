import { ref, computed } from 'vue'
import { videoApi } from '@/api/video.api'
import { useRouter } from 'vue-router'
import { useUIStore } from '@/stores/ui.store'
import { storeToRefs } from 'pinia'

export function useVideo() {
    const router = useRouter()
    const uiStore = useUIStore()
    const { platform } = storeToRefs(uiStore)
    const isGenerating = ref(false)
    const progress = ref(0)
    const currentStep = ref('')
    const activeJobId = ref(null)
    const activeSeriesId = ref(null)
    const seriesParts = ref([])
    const error = ref(null)

    async function checkActiveTask() {
        try {
            const res = await videoApi.getActiveTask({ platform: platform.value })
            const data = res.data.data
            if (data.type === 'job') {
                activeJobId.value = data.job_id
                isGenerating.value = true
                startPolling()
            } else if (data.type === 'series') {
                activeSeriesId.value = data.series_id
                isGenerating.value = true
                startPolling()
            }
        } catch (e) {
            console.error('Failed to check active task:', e)
        }
    }

    async function generate(data) {
        if (isGenerating.value) {
            error.value = 'Một tiến trình khác đang chạy. Vui lòng đợi.'
            return
        }
        isGenerating.value = true
        error.value = null
        progress.value = 5
        currentStep.value = 'Đang khởi tạo...'

        try {
            let res
            if (data.is_series) {
                res = await videoApi.generateSeries(data)
                activeSeriesId.value = res.data.data.series_id
            } else {
                res = await videoApi.generate(data)
                activeJobId.value = res.data.data.job_id
            }

            startPolling()
            return res.data.data
        } catch (err) {
            error.value = err.response?.data?.message || 'Không thể bắt đầu tạo video'
            isGenerating.value = false
            throw err
        }
    }

    function startPolling() {
        const timer = setInterval(async () => {
            try {
                if (activeSeriesId.value) {
                    const res = await videoApi.getSeriesStatus(activeSeriesId.value)
                    const series = res.data.data
                    seriesParts.value = series.jobs
                    progress.value = calculateOverallProgress(series.jobs)

                    if (series.status === 'completed' || series.status === 'failed') {
                        clearInterval(timer)
                        isGenerating.value = false
                    }
                } else if (activeJobId.value) {
                    const res = await videoApi.getStatus(activeJobId.value)
                    const job = res.data.data
                    progress.value = job.progress || 0
                    currentStep.value = job.status_message || 'Đang xử lý...'

                    if (job.status === 'completed' || job.status === 'failed') {
                        clearInterval(timer)
                        isGenerating.value = false
                        if (job.status === 'completed') {
                            router.push(`/job/${activeJobId.value}`)
                        }
                    }
                }
            } catch (e) {
                console.error('Polling error:', e)
            }
        }, 3000)
    }

    function calculateOverallProgress(jobs) {
        if (!jobs || jobs.length === 0) return 0
        const totalProgress = jobs.reduce((acc, job) => acc + (job.progress || 0), 0)
        return Math.floor(totalProgress / jobs.length)
    }

    return {
        isGenerating,
        progress,
        currentStep,
        seriesParts,
        error,
        generate,
        checkActiveTask
    }
}
