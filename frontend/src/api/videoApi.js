import axios from 'axios'

const API_BASE = '/api'

export const videoApi = {
    /**
     * Generate video from script
     * @param {Object} data - { script, voice, speaking_speed, video_style }
     * @returns {Promise<Object>} { job_id, status }
     */
    async generateVideo(data) {
        const response = await axios.post(`${API_BASE}/generate`, data)
        return response.data
    },

    /**
     * Get job status
     * @param {string} jobId - Job ID
     * @returns {Promise<Object>} { status, progress, current_step, video_url?, error? }
     */
    async getStatus(jobId) {
        const response = await axios.get(`${API_BASE}/status/${jobId}`)
        return response.data
    },

    /**
     * Get download URL for video
     * @param {string} jobId - Job ID
     * @returns {string} Download URL
     */
    getDownloadUrl(jobId) {
        return `${API_BASE}/download/${jobId}`
    }
}
