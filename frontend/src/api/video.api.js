import api from './index'

export const videoApi = {
    // Single Video
    generate: (data) => api.post('/generate', data),
    getStatus: (jobId) => api.get(`/status/${jobId}`),
    download: (jobId) => api.get(`/download/${jobId}`, { responseType: 'blob' }),

    // Series
    generateSeries: (data) => api.post('/series/generate', data),
    getSeriesStatus: (seriesId) => api.get(`/series/status/${seriesId}`),

    // Gallery, Explore & Tasks
    getGallery: (params) => api.get('/me/videos', { params }),
    getExplore: (params) => api.get('/explore', { params }),
    getTasks: (params) => api.get('/me/tasks', { params }),
    resumeTask: (jobId) => api.post(`/status/${jobId}/resume`),
    cancelTask: (jobId) => api.delete(`/status/${jobId}`),
    getActiveTask: () => api.get('/me/active-task')
}
