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
    },

    /**
     * Generate a multi-part series
     * @param {Object} data - { topic, content_name, platform, num_parts, voice, speaking_speed }
     * @returns {Promise<Object>} { series_id, status, num_parts }
     */
    async generateSeries(data) {
        const response = await axios.post(`${API_BASE}/generate-series`, data)
        return response.data
    },

    /**
     * Get series status
     * @param {string} seriesId - Series ID
     * @returns {Promise<Object>} { status, overall_progress, parts }
     */
    async getSeriesStatus(seriesId) {
        const response = await axios.get(`${API_BASE}/series-status/${seriesId}`)
        return response.data
    },

    /**
     * Retry a specific part of a series
     * @param {string} seriesId - Series ID
     * @param {number} partIndex - Part index (0-based)
     * @returns {Promise<Object>} { status, part_index }
     */
    async retrySeriesPart(seriesId, partIndex) {
        const response = await axios.post(`${API_BASE}/retry-series-part/${seriesId}/${partIndex}`)
        return response.data
    }
}

