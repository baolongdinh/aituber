import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export const useUIStore = defineStore('ui', () => {
    const platform = ref(localStorage.getItem('platform') || 'tiktok')

    const isTikTok = () => platform.value === 'tiktok'
    const isYouTube = () => platform.value === 'youtube'

    function setPlatform(p) {
        platform.value = p
        localStorage.setItem('platform', p)
    }

    return {
        platform,
        isTikTok,
        isYouTube,
        setPlatform
    }
})
