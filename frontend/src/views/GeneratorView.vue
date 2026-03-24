<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import { videoApi } from '@/api/video.api'
import { useUIStore } from '@/stores/ui.store'
import { useVideo } from '@/composables/useVideo'
import { useAuthStore } from '@/stores/auth.store'
import { useAuth } from '@/composables/useAuth'
import { storeToRefs } from 'pinia'

const uiStore = useUIStore()
const { platform } = storeToRefs(uiStore)
const { isGenerating, progress, currentStep, seriesParts, error, generate, checkActiveTask } = useVideo()
const { isAuthenticated } = storeToRefs(useAuthStore())
const { login } = useAuth()

const selectedTask = ref(null)
const isTaskDetailOpen = ref(false)

const tasks = ref([])
const isLoadingHistory = ref(true)
const isVoiceDropdownOpen = ref(false)
const voiceOptions = [
  { value: 'banmai', label: 'Ban Mai (Nữ)', icon: 'face_3', gender: 'female', providers: ['fpt', 'hub'] },
  { value: 'leminh', label: 'Lê Minh (Nữ)', icon: 'face_4', gender: 'female', providers: ['fpt'] },
  { value: 'minhquang', label: 'Minh Quang (Nam)', icon: 'face_6', gender: 'male', providers: ['fpt', 'hub'] },
  { value: 'giahuy', label: 'Gia Huy (Nam)', icon: 'face_2', gender: 'male', providers: ['fpt'] }
]

const form = reactive({
  topic: '',
  content_name: '',
  is_series: false,
  num_parts: 5,
  voice: 'banmai',
  tts_provider: 'fpt',
  t2v_model: 'flux-1-dev',
  stock_keywords: ''
})

// Filter voices based on selected TTS provider
const availableVoices = computed(() => {
  return voiceOptions.filter(voice => voice.providers.includes(form.tts_provider))
})

const selectedVoiceLabel = computed(() => {
  const voice = availableVoices.value.find(o => o.value === form.voice)
  return voice ? voice.label : 'Chọn giọng đọc'
})

let historyPollingTimer = null

const groupedTasks = computed(() => {
  const groups = []
  const seriesMap = new Map()

  tasks.value.forEach(task => {
    if (task.series_id) {
      // Extract base content name without the part number
      const baseContentName = task.content_name?.split(' - Tập ')[0] || task.topic
      const partTitle = `${baseContentName} - Tập ${task.part_index + 1}`
      groups.push({
        id: task.id,
        type: 'series_part',
        platform: task.platform,
        topic: task.topic,
        content_name: partTitle,
        created_at: task.created_at,
        status: task.status,
        progress: task.progress || 0,
        current_step: task.current_step,
        job: task,
        series_id: task.series_id,
        part_index: task.part_index
      })
    } else {
      groups.push({ id: task.id, type: 'single', job: task, created_at: task.created_at })
    }
  })

  // Sort by creation time (newest first)
  groups.sort((a, b) => new Date(b.created_at) - new Date(a.created_at))

  return groups
})

async function fetchTasks() {
  if (!isAuthenticated.value) return
  try {
    const res = await videoApi.getTasks({ limit: 50, platform: platform.value })
    tasks.value = res.data.data
  } catch (e) {
    console.error('Failed to fetch history', e)
  } finally {
    isLoadingHistory.value = false
  }
}

async function resumeTask(jobId) {
  try {
    await videoApi.resumeTask(jobId)
    await fetchTasks()
    checkActiveTask()
  } catch (e) {
    alert('Không thể tiếp tục tiến trình này.')
  }
}

async function cancelTask(jobId) {
  if (!confirm('Hủy tiến trình này?')) return
  try {
    await videoApi.cancelTask(jobId)
    await fetchTasks()
  } catch (e) {
    alert('Lỗi khi hủy tiến trình.')
  }
}

function startHistoryPolling() {
  stopHistoryPolling()
  historyPollingTimer = setInterval(fetchTasks, 5000)
}

function stopHistoryPolling() {
  if (historyPollingTimer) {
    clearInterval(historyPollingTimer)
    historyPollingTimer = null
  }
}

watch(isAuthenticated, (val) => {
  if (val) {
    fetchTasks()
    startHistoryPolling()
  } else {
    stopHistoryPolling()
    tasks.value = []
  }
})

watch(isGenerating, (val) => {
  if (!val) fetchTasks() // Refresh history once generation finishes
})

// Watch for provider changes and auto-select appropriate voice
watch(() => form.tts_provider, (newProvider, oldProvider) => {
  if (oldProvider && newProvider !== oldProvider) {
    // Check if current voice supports the new provider
    const currentVoice = voiceOptions.find(v => v.value === form.voice)
    if (!currentVoice || !currentVoice.providers.includes(newProvider)) {
      // Auto-select first available voice for new provider
      const firstAvailableVoice = availableVoices.value[0]
      if (firstAvailableVoice) {
        form.voice = firstAvailableVoice.value
      }
    }
  }
})

async function handleGenerate() {
  if (!form.topic) return
  
  if (!isAuthenticated.value) {
    await login()
    return
  }

  try {
    await generate({
      topic: form.topic,
      platform: platform.value,
      // Automatically generate a content name if not provided
      content_name: form.topic.slice(0, 20).replace(/\s+/g, '-').toLowerCase() + '-' + Date.now().toString().slice(-4),
      is_series: form.is_series,
      num_parts: form.is_series ? form.num_parts : 1,
      voice: form.voice,
      tts_provider: form.tts_provider,
      t2v_model: form.t2v_model,
      stock_keywords: form.stock_keywords
    })
  } catch (e) {
    console.error('Generation failed')
  }
}

function scrollToHistory() {
  const el = document.querySelector('.history-section')
  if (el) el.scrollIntoView({ behavior: 'smooth' })
}

watch(platform, () => {
  fetchTasks()
})

function showTaskDetails(task) {
  selectedTask.value = task
  isTaskDetailOpen.value = true
}

function togglePartDetails(partIndex) {
  if (selectedTask.value?.expandedPart === partIndex) {
    selectedTask.value.expandedPart = null
  } else {
    selectedTask.value.expandedPart = partIndex
  }
}

// Helper functions to determine step status
function isStepCompleted(task, step) {
  if (task.status === 'completed') return true
  if (task.status === 'failed') return false
  
  const progress = task.progress || 0
  const currentStep = task.current_step || ''
  
  switch (step) {
    case 'script':
      return progress >= 25 || currentStep.toLowerCase().includes('script')
    case 'audio':
      return progress >= 50 || currentStep.toLowerCase().includes('audio') || currentStep.toLowerCase().includes('tts')
    case 'video':
      return progress >= 75 || currentStep.toLowerCase().includes('video') || currentStep.toLowerCase().includes('generating')
    case 'final':
      return task.status === 'completed'
    default:
      return false
  }
}

function isStepCurrent(task, step) {
  if (task.status !== 'processing') return false
  
  const progress = task.progress || 0
  const currentStep = task.current_step || ''
  
  switch (step) {
    case 'script':
      return progress < 25 && currentStep.toLowerCase().includes('script')
    case 'audio':
      return progress >= 25 && progress < 50 && (currentStep.toLowerCase().includes('audio') || currentStep.toLowerCase().includes('tts'))
    case 'video':
      return progress >= 50 && progress < 75 && (currentStep.toLowerCase().includes('video') || currentStep.toLowerCase().includes('generating'))
    case 'final':
      return progress >= 75 && task.status === 'processing'
    default:
      return false
  }
}

function getStepStatus(task, step) {
  if (isStepCompleted(task, step)) return 'Hoàn thành'
  if (isStepCurrent(task, step)) return 'Đang xử lý...'
  return 'Chờ xử lý'
}

function closeTaskDetails() {
  isTaskDetailOpen.value = false
  selectedTask.value = null
}

function handleOutsideClick(e) {
  if (isVoiceDropdownOpen.value && !e.target.closest('.custom-dropdown-container')) {
    isVoiceDropdownOpen.value = false
  }
}

onMounted(() => {
  checkActiveTask()
  document.addEventListener('click', handleOutsideClick)
  if (isAuthenticated.value) {
    fetchTasks()
    startHistoryPolling()
  }
})

onUnmounted(() => {
  stopHistoryPolling()
  document.removeEventListener('click', handleOutsideClick)
})
</script>

<template>
  <div class="generator-view" :class="['theme-' + platform, { 'is-processing': isGenerating }]">
    <!-- Dynamic background effect with platform-specific colors -->
    <div class="aura-glow" :class="platform"></div>
    <div class="background-gradient" :class="platform"></div>

    <!-- Active Task Banner -->
    <Transition name="fade">
      <div v-if="isGenerating" class="active-task-banner glass-card" :class="platform">
        <div class="banner-content">
          <div class="banner-info">
            <div class="spinner-small"><span class="material-symbols-outlined spin">sync</span></div>
            <div>
              <h3>Đang trong quá trình tạo Video...</h3>
              <p>{{ currentStep || 'Hệ thống đang xử lý, vui lòng chờ' }} - <strong>{{ progress }}%</strong></p>
            </div>
          </div>
          <button @click="scrollToHistory" class="view-history-btn">
            Xem lịch sử chi tiết
            <span class="material-symbols-outlined">history</span>
          </button>
        </div>
        <div class="banner-progress">
          <div class="fill" :style="{ width: progress + '%' }"></div>
        </div>
      </div>
    </Transition>

    <div class="split-layout-v3">
      <!-- Left Column: Generator Workspace -->
      <section class="generator-workspace">
        <header class="workspace-header">
          <div class="platform-chip" :class="platform">{{ platform.toUpperCase() }} MODE</div>
          <h1 class="page-title">Universal Generator</h1>
          <p class="page-subtitle">Cấu hình tham số cho video Viral của bạn</p>
        </header>

        <div class="generator-form-refined" :class="{ 'is-loading': isGenerating }">
          <!-- 01 CHỦ ĐỀ NỘI DUNG -->
          <div class="form-group-lux">
            <label class="label-industrial"><span>01</span> CHỦ ĐỀ NỘI DUNG</label>
            <textarea 
              v-model="form.topic" 
              placeholder="Vd: Lịch sử tiền điện tử, 3 bí mật về thành công..."
              class="textarea-refined"
            ></textarea>
          </div>

          <!-- 02 TTS PROVIDER & GIỌNG ĐỌC AI -->
          <div class="form-row-lux">
            <div class="form-group-lux flex-1">
              <label class="label-industrial"><span>02</span> NHÀ CUNG CẤP TTS</label>
              <div class="tts-provider-selector">
                <div class="provider-toggle-group">
                  <button 
                    @click="form.tts_provider = 'fpt'"
                    class="provider-toggle-btn"
                    :class="{ active: form.tts_provider === 'fpt' }"
                  >
                    <span class="provider-icon">🌐</span>
                    <div class="provider-info">
                      <span class="provider-name">FPT.AI</span>
                      <span class="provider-count">4 giọng</span>
                    </div>
                  </button>
                  <button 
                    @click="form.tts_provider = 'hub'"
                    class="provider-toggle-btn"
                    :class="{ active: form.tts_provider === 'hub' }"
                  >
                    <span class="provider-icon">🚀</span>
                    <div class="provider-info">
                      <span class="provider-name">Hub</span>
                      <span class="provider-count">2 giọng</span>
                    </div>
                  </button>
                </div>
              </div>
            </div>
            
            <div class="form-group-lux flex-1">
              <label class="label-industrial"><span>03</span> GIỌNG ĐỌC AI</label>
              <div class="custom-dropdown-container">
                <button 
                  @click="isVoiceDropdownOpen = !isVoiceDropdownOpen"
                  class="dropdown-trigger-refined"
                  :class="{ 'is-open': isVoiceDropdownOpen }"
                >
                  <div class="flex items-center gap-3 truncate">
                    <span class="material-symbols-outlined text-xl opacity-40">mic</span>
                    <span class="font-medium">{{ selectedVoiceLabel }}</span>
                  </div>
                  <span class="material-symbols-outlined transition-transform" :class="{ 'rotate-180': isVoiceDropdownOpen }">expand_more</span>
                </button>
                
                <Transition name="onyx-fade">
                  <div v-if="isVoiceDropdownOpen" class="dropdown-menu-refined">
                    <div 
                      v-for="opt in availableVoices" 
                      :key="opt.value"
                      @click="form.voice = opt.value; isVoiceDropdownOpen = false"
                      class="dropdown-item-refined"
                      :class="{ 'active': form.voice === opt.value }"
                    >
                      <span class="material-symbols-outlined">{{ opt.icon }}</span>
                      {{ opt.label }}
                    </div>
                  </div>
                </Transition>
              </div>
            </div>
          </div>

          <!-- 04 PLATFORM -->
          <div class="form-group-lux">
            <label class="label-industrial"><span>04</span> PLATFORM</label>
            <div class="platform-indicator-refined" :class="platform">
              <span class="material-symbols-outlined">{{ platform === 'tiktok' ? 'filter_frames' : 'play_circle' }}</span>
              {{ platform === 'tiktok' ? 'TikTok' : 'YouTube' }}
            </div>
          </div>

          <!-- 05 SERIES ENGINE -->
          <div class="form-group-lux">
            <label class="label-industrial"><span>05</span> SERIES ENGINE</label>
            <div class="engine-controls-refined">
              <div class="toggle-card-refined" :class="{ active: form.is_series }" @click="form.is_series = !form.is_series">
                <div class="flex items-center gap-4">
                  <span class="material-symbols-outlined opacity-40">layers</span>
                  <div>
                    <h4 class="text-xs font-bold font-mono">SERIES MODE</h4>
                    <p class="text-[9px] opacity-30 uppercase tracking-tighter">Auto-link</p>
                  </div>
                </div>
                <div class="switch-mini" :class="{ active: form.is_series }">
                  <div class="thumb"></div>
                </div>
              </div>

              <Transition name="fade-up">
                <div v-if="form.is_series" class="slider-box-refined">
                  <div class="batch-size-header">
                    <div class="batch-size-info">
                      <span class="batch-label">Batch Size</span>
                      <span class="batch-value" :class="'text-' + (platform === 'tiktok' ? 'purple' : 'red') + '-400'">
                        {{ form.num_parts }} PARTS
                      </span>
                    </div>
                    <div class="batch-size-indicator">
                      <div class="indicator-dot" v-for="i in Math.min(form.num_parts, 5)" :key="i"></div>
                      <span v-if="form.num_parts > 5" class="indicator-more">+{{ form.num_parts - 5 }}</span>
                    </div>
                  </div>
                  <div class="range-container">
                    <input type="range" v-model.number="form.num_parts" min="2" max="20" class="range-refined">
                    <div class="range-marks">
                      <span class="mark" v-for="i in 5" :key="i">{{ i * 4 }}</span>
                    </div>
                  </div>
                </div>
              </Transition>
            </div>
          </div>

          <button 
            @click="handleGenerate" 
            class="generate-btn-refined" 
            :disabled="!form.topic || isGenerating"
          >
            <span class="material-symbols-outlined">bolt</span>
            {{ isGenerating ? 'EXECUTING...' : 'BẮT ĐẦU TẠO NỘI DUNG' }}
          </button>
        </div>
      </section>

      <!-- Right Column: Monitor & History -->
      <main class="monitor-column">
        <!-- Idle State -->
        <Transition name="fade" mode="out-in">
          <div v-if="!isGenerating && seriesParts.length === 0" class="status-placeholder-card">
            <div class="placeholder-content">
              <span class="material-symbols-outlined text-5xl opacity-10 mb-6">movie_edit</span>
              <h3 class="text-lg font-bold mb-2">Sẵn sàng tạo nội dung Viral?</h3>
              <p class="text-xs text-secondary opacity-50">Nhập chủ đề ở bên trái và ViralCraft sẽ lo phần còn lại.</p>
            </div>
          </div>

          <!-- Cinematic Monitoring (Active) -->
          <div v-else class="active-monitor-layout">
            <div class="monitor-main-card">
              <div class="flex justify-between items-start mb-10">
                <div class="flex items-center gap-5">
                  <div class="processor-ring"><span class="material-symbols-outlined spin">settings_slow_motion_video</span></div>
                  <div>
                    <div class="flex items-center gap-3 mb-2">
                      <span class="active-pill" :class="platform">PROCESSING</span>
                      <span class="text-[10px] font-mono opacity-30">GPU-V100 | NODE-ALPHA</span>
                    </div>
                    <h3 class="monitor-title">{{ form.topic }}</h3>
                  </div>
                </div>
                <div class="percentage-display">{{ progress }}%</div>
              </div>
              
              <div class="progress-details">
                <p class="step-label">{{ currentStep || 'Initializing Neural Engine...' }}</p>
                <div class="track-refined">
                  <div class="fill" :style="{ width: progress + '%' }"></div>
                </div>
              </div>
            </div>

            <!-- Mini logs for multi-part jobs -->
            <div v-if="seriesParts.length > 0" class="mini-jobs-log">
               <div v-for="(part, idx) in seriesParts" :key="idx" class="job-row-refined">
                  <div class="flex items-center gap-3">
                    <div class="dot" :class="part.status"></div>
                    <span class="text-[10px] font-bold opacity-70">PART {{ idx+1 }}</span>
                  </div>
                  <span class="text-[10px] opacity-40 uppercase">{{ part.status }}</span>
               </div>
            </div>
          </div>
        </Transition>

        <!-- LỊCH SỬ TẠO GẦN ĐÂY -->
        <div class="recent-history-lux mt-12">
          <header class="section-label-header">
            <h3 class="flex items-center gap-3">
              <span class="material-symbols-outlined text-lg opacity-30">history</span>
              LỊCH SỬ TẠO GẦN ĐÂY
            </h3>
            <button @click="fetchTasks" class="btn-icon-refined" :disabled="isLoadingHistory">
              <span class="material-symbols-outlined" :class="{ 'spin': isLoadingHistory }">refresh</span>
            </button>
          </header>

          <div v-if="isLoadingHistory" class="history-loading">
            <div class="loader-lux"></div>
            <p class="text-[10px] opacity-30 tracking-widest mt-4">FETCHING RECORDS...</p>
          </div>

          <div v-else-if="groupedTasks.length === 0" class="history-empty">
             <span class="material-symbols-outlined text-4xl opacity-10 mb-4">layers_clear</span>
             <p class="text-xs opacity-30">Chưa có lịch sử tạo Video nào.</p>
          </div>

          <div v-else class="history-stack-v3">
            <div v-for="group in groupedTasks" :key="group.id" class="history-card-v3" :class="[group.status, { 'is-series': group.type === 'series_part' }]" @click="showTaskDetails(group)">
               <div class="flex items-center gap-5">
                 <div class="status-indicator" :class="group.status"></div>
                 <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2 mb-1">
                      <span v-if="group.type === 'series_part'" class="tag-series">TẬP {{ group.part_index + 1 }}</span>
                      <h4 class="history-title-v3 truncate">{{ group.content_name || group.topic }}</h4>
                    </div>
                    <div class="history-meta-v3">
                      <span class="status-text" :class="group.status">{{ group.status.toUpperCase() }}</span>
                      <span class="dot"></span>
                      <span>{{ new Date(group.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) }}</span>
                      <span class="dot"></span>
                      <span class="uppercase">{{ group.platform }}</span>
                    </div>
                    <!-- Progress bar for processing tasks -->
                    <div v-if="group.status === 'processing' && group.progress > 0" class="task-progress-bar">
                      <div class="progress-fill" :style="{ width: group.progress + '%' }"></div>
                      <span class="progress-text">{{ group.progress }}%</span>
                    </div>
                 </div>
                 <div class="history-actions-v3">
                    <router-link v-if="group.status === 'completed'" :to="group.type === 'series_part' ? '/job/' + group.id : '/job/' + group.id" class="action-btn-v3" @click.stop>
                      <span class="material-symbols-outlined">chevron_right</span>
                    </router-link>
                 </div>
               </div>
            </div>
          </div>
        </div>
      </main>
    </div>

    <!-- Task Detail Modal -->
    <Transition name="modal-fade">
      <div v-if="isTaskDetailOpen && selectedTask" class="modal-overlay" @click="closeTaskDetails">
        <div class="modal-container" @click.stop>
          <div class="modal-header">
            <h3 class="modal-title">Chi tiết Task</h3>
            <button @click="closeTaskDetails" class="modal-close-btn">
              <span class="material-symbols-outlined">close</span>
            </button>
          </div>
          
          <div class="modal-content">
            <!-- Task Info -->
            <div class="task-info-grid">
              <div class="info-item">
                <span class="info-label">Tên task</span>
                <span class="info-value">{{ selectedTask.content_name || selectedTask.topic }}</span>
              </div>
              
              <div class="info-item">
                <span class="info-label">Trạng thái</span>
                <span class="info-value status-badge" :class="selectedTask.status">
                  {{ selectedTask.status.toUpperCase() }}
                </span>
              </div>
              
              <div class="info-item">
                <span class="info-label">Tiến độ</span>
                <div class="progress-container">
                  <div class="progress-bar">
                    <div class="progress-fill" :style="{ width: (selectedTask.progress || 0) + '%' }"></div>
                  </div>
                  <span class="progress-percentage">{{ selectedTask.progress || 0 }}%</span>
                </div>
              </div>
              
              <div class="info-item">
                <span class="info-label">Bước hiện tại</span>
                <span class="info-value">{{ selectedTask.current_step || 'Đang chờ...' }}</span>
              </div>
              
              <div class="info-item">
                <span class="info-label">Platform</span>
                <span class="info-value platform-badge" :class="selectedTask.platform">
                  {{ selectedTask.platform.toUpperCase() }}
                </span>
              </div>
              
              <div class="info-item">
                <span class="info-label">Thời gian tạo</span>
                <span class="info-value">{{ new Date(selectedTask.created_at).toLocaleString('vi-VN') }}</span>
              </div>
              
              <div v-if="selectedTask.type === 'series_part'" class="info-item">
                <span class="info-label">Phần</span>
                <span class="info-value">Tập {{ selectedTask.part_index + 1 }}</span>
              </div>
            </div>

            <!-- Task Steps -->
            <div class="task-steps-section">
              <h4 class="steps-title">Chi tiết các bước</h4>
              <div class="task-steps-list">
                <div class="task-step-item" :class="{ 'completed': isStepCompleted(selectedTask, 'script'), 'current': isStepCurrent(selectedTask, 'script') }">
                  <div class="step-indicator">
                    <span class="material-symbols-outlined">description</span>
                  </div>
                  <div class="step-content">
                    <span class="step-name">Tạo kịch bản</span>
                    <span class="step-status">{{ getStepStatus(selectedTask, 'script') }}</span>
                  </div>
                </div>

                <div class="task-step-item" :class="{ 'completed': isStepCompleted(selectedTask, 'audio'), 'current': isStepCurrent(selectedTask, 'audio') }">
                  <div class="step-indicator">
                    <span class="material-symbols-outlined">mic</span>
                  </div>
                  <div class="step-content">
                    <span class="step-name">Tạo audio</span>
                    <span class="step-status">{{ getStepStatus(selectedTask, 'audio') }}</span>
                  </div>
                </div>

                <div class="task-step-item" :class="{ 'completed': isStepCompleted(selectedTask, 'video'), 'current': isStepCurrent(selectedTask, 'video') }">
                  <div class="step-indicator">
                    <span class="material-symbols-outlined">movie</span>
                  </div>
                  <div class="step-content">
                    <span class="step-name">Tạo video</span>
                    <span class="step-status">{{ getStepStatus(selectedTask, 'video') }}</span>
                  </div>
                </div>

                <div class="task-step-item" :class="{ 'completed': isStepCompleted(selectedTask, 'final'), 'current': isStepCurrent(selectedTask, 'final') }">
                  <div class="step-indicator">
                    <span class="material-symbols-outlined">check_circle</span>
                  </div>
                  <div class="step-content">
                    <span class="step-name">Hoàn thành</span>
                    <span class="step-status">{{ getStepStatus(selectedTask, 'final') }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
          
          <div class="modal-footer">
            <button @click="closeTaskDetails" class="modal-btn modal-btn-primary">
              Đóng
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;700&display=swap');

.generator-view {
  min-height: calc(100vh - 80px); /* Account for header */
  color: #fff;
  font-family: 'Plus Jakarta Sans', sans-serif;
  position: relative;
  overflow-x: hidden;
}

/* Platform-specific background gradients and glows */
.background-gradient {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: -2;
  opacity: 0.02;
  pointer-events: none;
}

.background-gradient.tiktok {
  background: radial-gradient(circle at 20% 50%, #a14bff 0%, transparent 50%),
              radial-gradient(circle at 80% 80%, #ff0050 0%, transparent 50%),
              radial-gradient(circle at 40% 20%, #00f2ea 0%, transparent 50%);
}

.background-gradient.youtube {
  background: radial-gradient(circle at 30% 40%, #ff0000 0%, transparent 50%),
              radial-gradient(circle at 70% 60%, #282828 0%, transparent 50%),
              radial-gradient(circle at 50% 10%, #ff6b6b 0%, transparent 50%);
}

.aura-glow {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: -1;
  pointer-events: none;
  filter: blur(120px);
  opacity: 0.1;
}

.aura-glow.tiktok {
  background: radial-gradient(ellipse 1200px 600px at 25% 30%, #a14bff 0%, transparent 40%),
              radial-gradient(ellipse 800px 400px at 75% 70%, #ff0050 0%, transparent 40%),
              radial-gradient(ellipse 600px 300px at 50% 10%, #00f2ea 0%, transparent 40%);
  animation: tiktok-pulse 8s ease-in-out infinite;
}

.aura-glow.youtube {
  background: radial-gradient(ellipse 1200px 600px at 25% 30%, #ff0000 0%, transparent 40%),
              radial-gradient(ellipse 800px 400px at 75% 70%, #282828 0%, transparent 40%),
              radial-gradient(ellipse 600px 300px at 50% 10%, #ff6b6b 0%, transparent 40%);
  animation: youtube-pulse 10s ease-in-out infinite;
}

@keyframes tiktok-pulse {
  0%, 100% { transform: scale(1) rotate(0deg); }
  50% { transform: scale(1.1) rotate(2deg); }
}

@keyframes youtube-pulse {
  0%, 100% { transform: scale(1) rotate(0deg); }
  50% { transform: scale(1.05) rotate(-1deg); }
}

.split-layout-v3 {
  display: grid;
  grid-template-columns: 1fr 480px;
  gap: 80px;
  padding: 60px 80px 60px 40px; /* Reduced left padding since MainLayout handles spacing */
  max-width: 1600px;
  margin: 0 auto;
}

/* WORKSPACE HEADER */
.workspace-header { margin-bottom: 48px; }
.platform-chip {
  display: inline-block; padding: 4px 12px; border-radius: 6px; font-weight: 800; font-size: 0.65rem;
  letter-spacing: 0.1em; background: rgba(255,255,255,0.03); margin-bottom: 20px;
}
.platform-chip.tiktok { color: #a14bff; }
.platform-chip.youtube { color: #ff0000; }
.page-title { font-size: 3.5rem; font-weight: 800; letter-spacing: -0.04em; margin-bottom: 8px; }
.page-subtitle { color: rgba(255,255,255,0.6); font-size: 1.1rem; }

/* FORM REFINEMENT */
.generator-form-refined { display: flex; flex-direction: column; gap: 40px; }
.form-group-lux { display: flex; flex-direction: column; gap: 14px; }
.label-industrial {
  font-size: 0.65rem; font-weight: 800; color: rgba(255,255,255,0.35); letter-spacing: 0.15em;
  display: flex; align-items: center; gap: 10px;
}
.label-industrial span { opacity: 0.7; font-family: 'JetBrains Mono', monospace; }

.textarea-refined {
  background: #0d0d0e; border: 1px solid rgba(255,255,255,0.05); border-radius: 20px;
  padding: 32px; color: #fff; font-size: 1.1rem; line-height: 1.6; resize: none; height: 220px;
  transition: 0.3s;
}
.textarea-refined:focus { border-color: rgba(255,255,255,0.1); background: #131315; outline: none; }

.form-row-lux { display: flex; gap: 24px; }

/* TTS PROVIDER SELECTOR */
.tts-provider-selector {
  width: 100%;
}

.provider-toggle-group {
  display: flex;
  background: rgba(255,255,255,0.02);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 16px;
  padding: 4px;
  gap: 4px;
}

.provider-toggle-btn {
  flex: 1;
  padding: 16px 20px;
  border: none;
  background: transparent;
  color: rgba(255,255,255,0.7);
  font-size: 0.85rem;
  font-weight: 600;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.provider-toggle-btn:hover {
  background: rgba(255,255,255,0.05);
  color: rgba(255,255,255,0.8);
}

.provider-toggle-btn.active {
  background: rgba(255,255,255,0.1);
  color: #fff;
  box-shadow: 0 4px 20px rgba(255,255,255,0.1);
}

.provider-toggle-btn.active.tiktok {
  background: linear-gradient(135deg, #a14bff 0%, #ff0050 100%);
  box-shadow: 0 4px 20px rgba(161, 75, 255, 0.3);
}

.provider-toggle-btn.active.youtube {
  background: linear-gradient(135deg, #ff0000 0%, #282828 100%);
  box-shadow: 0 4px 20px rgba(255, 0, 0, 0.3);
}

.provider-icon {
  font-size: 1.2rem;
}

.provider-info {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
}

.provider-name {
  font-weight: 700;
  font-size: 0.9rem;
}

.provider-count {
  font-size: 0.7rem;
  opacity: 0.8;
  font-weight: 500;
}

/* DROPDOWN & INDICATORS */
.dropdown-trigger-refined {
  width: 100%; height: 64px; padding: 0 24px; background: #0d0d0e; border-radius: 16px;
  border: 1px solid rgba(255,255,255,0.05); color: #fff; display: flex; align-items: center; justify-content: space-between;
  cursor: pointer; transition: 0.3s;
}
.dropdown-trigger-refined:hover { background: #131315; border-color: rgba(255,255,255,0.1); }
.dropdown-trigger-refined.is-open { border-color: rgba(255,255,255,0.2); }

.platform-indicator-refined {
  height: 64px; display: flex; align-items: center; gap: 16px; padding: 0 24px;
  background: #0d0d0e; border-radius: 16px; border: 1px solid rgba(255,255,255,0.03);
  font-weight: 700; font-size: 0.95rem;
}
.platform-indicator-refined.tiktok { color: #a14bff; }
.platform-indicator-refined.youtube { color: #ff0000; }

/* ENGINE CONTROLS */
.engine-controls-refined { display: flex; gap: 20px; }
.toggle-card-refined {
  padding: 24px; background: #0d0d0e; border-radius: 20px; flex: 1; border: 1px solid rgba(255,255,255,0.03);
  display: flex; justify-content: space-between; align-items: center; cursor: pointer; transition: 0.3s;
}
.toggle-card-refined:hover { background: #131315; border-color: rgba(255,255,255,0.1); }

.switch-mini { width: 36px; height: 20px; background: #222; border-radius: 20px; position: relative; transition: 0.3s; }
.switch-mini.active { background: #fff; }
.switch-mini .thumb {
  position: absolute; top: 3px; left: 3px; width: 14px; height: 14px; background: #fff; border-radius: 50%; transition: 0.3s;
}
.switch-mini.active .thumb { transform: translateX(16px); background: #000; }

/* Enhanced Batch Size Display */
.slider-box-refined { 
  padding: 24px; 
  background: #0d0d0e; 
  border-radius: 20px; 
  flex: 1; 
  border: 1px solid rgba(255,255,255,0.03); 
  display: flex; 
  flex-direction: column; 
  gap: 20px;
}

.batch-size-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.batch-size-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.batch-label {
  font-size: 0.7rem;
  font-weight: 800;
  color: rgba(255,255,255,0.5);
  letter-spacing: 0.1em;
}

.batch-value {
  font-size: 1.1rem;
  font-weight: 800;
  font-family: 'JetBrains Mono', monospace;
}

.batch-size-indicator {
  display: flex;
  align-items: center;
  gap: 4px;
}

.indicator-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: rgba(255,255,255,0.2);
  transition: all 0.3s ease;
}

.indicator-dot:nth-child(1) { background: rgba(255,255,255,0.7); }
.indicator-dot:nth-child(2) { background: rgba(255,255,255,0.6); }
.indicator-dot:nth-child(3) { background: rgba(255,255,255,0.5); }
.indicator-dot:nth-child(4) { background: rgba(255,255,255,0.4); }
.indicator-dot:nth-child(5) { background: rgba(255,255,255,0.3); }

.indicator-more {
  font-size: 0.6rem;
  font-weight: 600;
  color: rgba(255,255,255,0.4);
  font-family: 'JetBrains Mono', monospace;
}

.range-container {
  position: relative;
}

.range-refined { 
  -webkit-appearance: none; 
  appearance: none; 
  width: 100%; 
  height: 6px; 
  background: rgba(255,255,255,0.1); 
  border-radius: 3px;
  outline: none;
  cursor: pointer;
}

.range-refined::-webkit-slider-thumb {
  -webkit-appearance: none; 
  width: 20px; 
  height: 20px; 
  background: #fff; 
  border-radius: 50%; 
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(0,0,0,0.3);
  transition: all 0.2s ease;
}

.range-refined::-webkit-slider-thumb:hover {
  transform: scale(1.2);
  box-shadow: 0 4px 12px rgba(255,255,255,0.3);
}

.range-refined::-moz-range-thumb {
  width: 20px; 
  height: 20px; 
  background: #fff; 
  border-radius: 50%; 
  cursor: pointer;
  border: none;
  box-shadow: 0 2px 8px rgba(0,0,0,0.3);
}

.range-marks {
  display: flex;
  justify-content: space-between;
  margin-top: 8px;
  padding: 0 4px;
}

.mark {
  font-size: 0.6rem;
  font-weight: 600;
  color: rgba(255,255,255,0.3);
  font-family: 'JetBrains Mono', monospace;
}

.generate-btn-refined {
  height: 80px; 
  border: none; 
  border-radius: 24px;
  font-size: 1.1rem; 
  font-weight: 800; 
  display: flex; 
  align-items: center; 
  justify-content: center; 
  gap: 16px;
  cursor: pointer; 
  transition: 0.4s; 
  margin-top: 20px; 
  box-shadow: 0 20px 40px rgba(0,0,0,0.4);
  color: #fff;
}

/* Platform-specific button colors */
.theme-tiktok .generate-btn-refined {
  background: linear-gradient(135deg, #a14bff 0%, #ff0050 100%);
  box-shadow: 0 20px 40px rgba(161, 75, 255, 0.4);
}

.theme-youtube .generate-btn-refined {
  background: linear-gradient(135deg, #ff0000 0%, #cc0000 100%);
  box-shadow: 0 20px 40px rgba(255, 0, 0, 0.4);
}

.generate-btn-refined:hover:not(:disabled) { 
  transform: translateY(-6px); 
  box-shadow: 0 25px 50px rgba(0,0,0,0.6); 
}

.theme-tiktok .generate-btn-refined:hover:not(:disabled) {
  box-shadow: 0 25px 50px rgba(161, 75, 255, 0.6);
}

.theme-youtube .generate-btn-refined:hover:not(:disabled) {
  box-shadow: 0 25px 50px rgba(255, 0, 0, 0.6);
}

.generate-btn-refined:disabled { 
  opacity: 0.1; 
  cursor: not-allowed; 
  filter: grayscale(1); 
}

/* MONITOR COLUMN */
.status-placeholder-card {
  height: 240px; background: #0a0a0b; border-radius: 32px; border: 1px solid rgba(255,255,255,0.03);
  display: flex; align-items: center; justify-content: center; text-align: center;
}
.monitor-main-card {
  padding: 40px; background: #0d0d0e; border-radius: 32px; border: 1px solid rgba(255,255,255,0.05);
}
.processor-ring { width: 48px; height: 48px; border-radius: 50%; background: #000; display: flex; align-items: center; justify-content: center; color: #fff; border: 1px solid rgba(255,255,255,0.1); }
.active-pill { padding: 4px 8px; border-radius: 4px; background: #fff; color: #000; font-size: 0.6rem; font-weight: 800; }
.active-pill.tiktok { background: #a14bff; color: #fff; }
.active-pill.youtube { background: #ff0000; color: #fff; }
.monitor-title { font-size: 1.5rem; font-weight: 700; }
.percentage-display { font-size: 3.5rem; font-weight: 800; opacity: 0.1; font-family: 'JetBrains Mono', monospace; }

.track-refined { height: 4px; background: rgba(255,255,255,0.05); border-radius: 10px; overflow: hidden; }
.track-refined .fill { height: 100%; background: #fff; transition: 0.4s; }

/* HISTORY STACK */
.section-label-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; padding: 0 4px; }
.section-label-header h3 { font-size: 0.75rem; font-weight: 800; opacity: 0.4; letter-spacing: 0.1em; }
.btn-icon-refined { background: transparent; border: none; color: #fff; opacity: 0.3; cursor: pointer; transition: 0.3s; }
.btn-icon-refined:hover { opacity: 1; transform: rotate(180deg); }

.history-card-v3 {
  padding: 20px 24px; background: #0d0d0e; border-radius: 20px; border: 1px solid rgba(255,255,255,0.03);
  margin-bottom: 12px; transition: 0.3s; cursor: pointer;
}
.history-card-v3:hover { border-color: rgba(255,255,255,0.1); background: #131315; transform: translateY(-2px); }
.status-indicator { width: 4px; height: 24px; border-radius: 10px; background: rgba(255,255,255,0.1); }
.status-indicator.completed { background: #10b981; }
.status-indicator.processing { background: #3b82f6; animation: soft-pulse 2s infinite; }

/* Task progress bar */
.task-progress-bar {
  position: relative; margin-top: 8px; height: 4px; background: rgba(255,255,255,0.1); border-radius: 2px;
  overflow: hidden;
}
.task-progress-bar .progress-fill {
  height: 100%; background: linear-gradient(90deg, #3b82f6, #8b5cf6); transition: width 0.3s ease;
}
.task-progress-bar .progress-text {
  position: absolute; right: 0; top: -20px; font-size: 0.6rem; color: rgba(255,255,255,0.6);
  font-weight: 600;
}

.history-title-v3 { font-size: 1rem; font-weight: 700; margin-bottom: 4px; }
.history-meta-v3 { display: flex; align-items: center; gap: 8px; font-size: 0.7rem; opacity: 0.4; font-weight: 600; }
.history-meta-v3 .dot { width: 2px; height: 2px; background: currentColor; border-radius: 50%; opacity: 0.5; }

.action-btn-v3 {
  width: 36px; height: 36px; border-radius: 10px; background: #000; display: flex; align-items: center; justify-content: center;
  color: #fff; text-decoration: none; border: 1px solid rgba(255,255,255,0.05); transition: 0.2s;
}
.action-btn-v3:hover { background: #fff; color: #000; border-color: #fff; }

.tag-series { font-size: 0.6rem; font-weight: 800; background: rgba(255,255,255,0.05); padding: 2px 6px; border-radius: 4px; opacity: 0.5; }

@keyframes soft-pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.4; } }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }

/* Modal Styles */
.modal-overlay {
  position: fixed; top: 0; left: 0; right: 0; bottom: 0; 
  background: rgba(0, 0, 0, 0.8); backdrop-filter: blur(8px);
  display: flex; align-items: center; justify-content: center;
  z-index: 1000; padding: 20px;
}

.modal-container {
  background: #1a1a1b; border-radius: 24px; border: 1px solid rgba(255,255,255,0.1);
  max-width: 600px; width: 100%; max-height: 90vh; overflow-y: auto;
  box-shadow: 0 20px 60px rgba(0,0,0,0.5);
}

.modal-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 24px 24px 20px; border-bottom: 1px solid rgba(255,255,255,0.05);
}

.modal-title {
  font-size: 1.5rem; font-weight: 700; margin: 0;
}

.modal-close-btn {
  background: transparent; border: none; color: rgba(255,255,255,0.6);
  cursor: pointer; padding: 8px; border-radius: 8px; transition: 0.2s;
}
.modal-close-btn:hover { background: rgba(255,255,255,0.1); color: #fff; }

.modal-content {
  padding: 24px;
}

.task-info-grid {
  display: grid; gap: 20px;
}

.info-item {
  display: flex; flex-direction: column; gap: 8px;
}

.info-label {
  font-size: 0.85rem; font-weight: 600; color: rgba(255,255,255,0.6);
  text-transform: uppercase; letter-spacing: 0.05em;
}

.info-value {
  font-size: 1rem; font-weight: 500; color: #fff;
}

/* Task Steps Section */
.task-steps-section {
  margin-top: 32px; padding-top: 24px; border-top: 1px solid rgba(255,255,255,0.1);
}

.steps-title {
  font-size: 1.1rem; font-weight: 700; margin: 0 0 20px 0; color: #fff;
}

.task-steps-list {
  display: flex; flex-direction: column; gap: 12px;
}

.task-step-item {
  display: flex; align-items: center; gap: 12px; padding: 12px;
  background: rgba(255,255,255,0.02); border-radius: 8px; border: 1px solid rgba(255,255,255,0.05);
  transition: 0.2s;
}

.task-step-item.completed {
  background: rgba(16, 185, 129, 0.05); border-color: rgba(16, 185, 129, 0.2);
}

.task-step-item.current {
  background: rgba(59, 130, 246, 0.1); border-color: rgba(59, 130, 246, 0.3);
  box-shadow: 0 0 15px rgba(59, 130, 246, 0.2);
}

.step-indicator {
  width: 40px; height: 40px; border-radius: 50%; background: rgba(255,255,255,0.1);
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}

.task-step-item.completed .step-indicator {
  background: rgba(16, 185, 129, 0.2); color: #10b981;
}

.task-step-item.current .step-indicator {
  background: rgba(59, 130, 246, 0.2); color: #3b82f6;
  animation: pulse-step 2s infinite;
}

.step-content {
  flex: 1; display: flex; justify-content: space-between; align-items: center;
}

.step-name {
  font-weight: 600; color: #fff; font-size: 0.9rem;
}

.step-status {
  font-size: 0.75rem; color: rgba(255,255,255,0.6); font-weight: 500;
}

.task-step-item.completed .step-status {
  color: #10b981;
}

.task-step-item.current .step-status {
  color: #3b82f6; font-weight: 600;
}

@keyframes pulse-step {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}

.status-badge {
  display: inline-block; padding: 6px 12px; border-radius: 8px; font-size: 0.85rem;
  font-weight: 700; text-transform: uppercase;
}
.status-badge.processing { background: rgba(59, 130, 246, 0.2); color: #3b82f6; }
.status-badge.completed { background: rgba(16, 185, 129, 0.2); color: #10b981; }
.status-badge.failed { background: rgba(239, 68, 68, 0.2); color: #ef4444; }
.status-badge.queued { background: rgba(255,255,255,0.1); color: rgba(255,255,255,0.8); }

.platform-badge {
  display: inline-block; padding: 6px 12px; border-radius: 8px; font-size: 0.85rem;
  font-weight: 700; text-transform: uppercase;
}
.platform-badge.tiktok { background: rgba(161, 75, 255, 0.2); color: #a14bff; }
.platform-badge.youtube { background: rgba(255, 0, 0, 0.2); color: #ff0000; }

.progress-container {
  display: flex; align-items: center; gap: 12px;
}

.progress-bar {
  flex: 1; height: 8px; background: rgba(255,255,255,0.1); border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%; background: linear-gradient(90deg, #3b82f6, #8b5cf6);
  transition: width 0.3s ease;
}

.progress-percentage {
  font-size: 0.9rem; font-weight: 600; color: #fff; min-width: 45px;
  text-align: right;
}

/* Series Tasks Styles */
.series-tasks-container {
  display: flex; flex-direction: column; gap: 24px;
}

.series-header {
  display: flex; justify-content: space-between; align-items: center;
  padding-bottom: 16px; border-bottom: 1px solid rgba(255,255,255,0.05);
}

.series-title {
  font-size: 1.2rem; font-weight: 700; margin: 0; color: #fff;
}

.series-info {
  font-size: 0.85rem; color: rgba(255,255,255,0.6); font-weight: 600;
}

.tasks-list {
  display: flex; flex-direction: column; gap: 12px;
  max-height: 400px; overflow-y: auto;
  padding-right: 8px;
}

.tasks-list::-webkit-scrollbar {
  width: 6px;
}

.tasks-list::-webkit-scrollbar-track {
  background: rgba(255,255,255,0.05); border-radius: 3px;
}

.tasks-list::-webkit-scrollbar-thumb {
  background: rgba(255,255,255,0.2); border-radius: 3px;
}

.task-item {
  padding: 16px; background: rgba(255,255,255,0.03); border-radius: 12px;
  border: 1px solid rgba(255,255,255,0.05); transition: 0.2s;
}

.task-item:hover {
  background: rgba(255,255,255,0.05); border-color: rgba(255,255,255,0.1);
}

.task-item.is-current {
  background: rgba(59, 130, 246, 0.1); border-color: rgba(59, 130, 246, 0.3);
  box-shadow: 0 0 20px rgba(59, 130, 246, 0.2);
}

.task-item.is-completed {
  background: rgba(16, 185, 129, 0.05); border-color: rgba(16, 185, 129, 0.2);
}

.task-item.is-failed {
  background: rgba(239, 68, 68, 0.05); border-color: rgba(239, 68, 68, 0.2);
}

.task-item-header {
  display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;
}

.task-item-header.clickable {
  cursor: pointer; padding: 4px; margin: -4px; border-radius: 8px;
  transition: 0.2s;
}

.task-item-header.clickable:hover {
  background: rgba(255,255,255,0.05);
}

.task-part-info {
  display: flex; align-items: center; gap: 12px;
}

.part-number {
  font-weight: 700; font-size: 0.95rem; color: #fff;
}

.task-status {
  font-size: 0.7rem; font-weight: 700; padding: 4px 8px; border-radius: 6px;
  text-transform: uppercase;
}

.task-status.processing { background: rgba(59, 130, 246, 0.2); color: #3b82f6; }
.task-status.completed { background: rgba(16, 185, 129, 0.2); color: #10b981; }
.task-status.failed { background: rgba(239, 68, 68, 0.2); color: #ef4444; }
.task-status.queued { background: rgba(255,255,255,0.1); color: rgba(255,255,255,0.8); }

.task-progress-mini {
  display: flex; align-items: center; gap: 8px;
}

.progress-percent {
  font-size: 0.85rem; font-weight: 600; color: rgba(255,255,255,0.8);
}

.expand-icon {
  font-size: 1.2rem; color: rgba(255,255,255,0.6); transition: transform 0.2s;
}

.expand-icon.expanded {
  transform: rotate(180deg);
}

.task-step-info {
  margin-bottom: 8px;
}

.step-text {
  font-size: 0.8rem; color: rgba(255,255,255,0.7); font-style: italic;
}

.task-item .task-progress-bar {
  height: 4px; background: rgba(255,255,255,0.1); border-radius: 2px;
  overflow: hidden; margin: 0;
}

.task-item .progress-fill {
  height: 100%; background: linear-gradient(90deg, #3b82f6, #8b5cf6);
  transition: width 0.3s ease;
}

/* Task Details Expanded */
.task-details-expanded {
  margin-top: 16px; padding-top: 16px; border-top: 1px solid rgba(255,255,255,0.1);
}

.task-steps-list {
  display: flex; flex-direction: column; gap: 12px;
}

.task-step-item {
  display: flex; align-items: center; gap: 12px; padding: 12px;
  background: rgba(255,255,255,0.02); border-radius: 8px; border: 1px solid rgba(255,255,255,0.05);
  transition: 0.2s;
}

.task-step-item.completed {
  background: rgba(16, 185, 129, 0.05); border-color: rgba(16, 185, 129, 0.2);
}

.task-step-item.current {
  background: rgba(59, 130, 246, 0.1); border-color: rgba(59, 130, 246, 0.3);
  box-shadow: 0 0 15px rgba(59, 130, 246, 0.2);
}

.step-indicator {
  width: 40px; height: 40px; border-radius: 50%; background: rgba(255,255,255,0.1);
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}

.task-step-item.completed .step-indicator {
  background: rgba(16, 185, 129, 0.2); color: #10b981;
}

.task-step-item.current .step-indicator {
  background: rgba(59, 130, 246, 0.2); color: #3b82f6;
  animation: pulse-step 2s infinite;
}

.step-content {
  flex: 1; display: flex; justify-content: space-between; align-items: center;
}

.step-name {
  font-weight: 600; color: #fff; font-size: 0.9rem;
}

.step-status {
  font-size: 0.75rem; color: rgba(255,255,255,0.6); font-weight: 500;
}

.task-step-item.completed .step-status {
  color: #10b981;
}

.task-step-item.current .step-status {
  color: #3b82f6; font-weight: 600;
}

@keyframes pulse-step {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}

/* Task details transition */
.task-details-enter-active, .task-details-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}

.task-details-enter-from, .task-details-leave-to {
  max-height: 0;
  opacity: 0;
  margin-top: 0;
  padding-top: 0;
}

.task-details-enter-to, .task-details-leave-from {
  max-height: 500px;
  opacity: 1;
  margin-top: 16px;
  padding-top: 16px;
}

.series-summary {
  display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px;
  padding: 16px; background: rgba(255,255,255,0.03); border-radius: 12px;
  border: 1px solid rgba(255,255,255,0.05);
}

.summary-item {
  display: flex; flex-direction: column; align-items: center; gap: 4px;
  text-align: center;
}

.summary-item span:first-child {
  font-size: 0.75rem; color: rgba(255,255,255,0.6); font-weight: 600;
  text-transform: uppercase;
}

.summary-item span:last-child {
  font-size: 1.1rem; color: #fff; font-weight: 700;
}

.modal-footer {
  padding: 20px 24px 24px; border-top: 1px solid rgba(255,255,255,0.05);
  display: flex; justify-content: flex-end;
}

.modal-btn {
  padding: 12px 24px; border: none; border-radius: 12px; font-weight: 600;
  cursor: pointer; transition: 0.2s; font-size: 0.95rem;
}

.modal-btn-primary {
  background: linear-gradient(135deg, #3b82f6, #8b5cf6); color: #fff;
}
.modal-btn-primary:hover { transform: translateY(-2px); box-shadow: 0 8px 20px rgba(59, 130, 246, 0.4); }

/* Modal transitions */
.modal-fade-enter-active, .modal-fade-leave-active {
  transition: all 0.3s ease;
}
.modal-fade-enter-from, .modal-fade-leave-to {
  opacity: 0;
  transform: scale(0.9);
}
.modal-fade-enter-to, .modal-fade-leave-from {
  opacity: 1;
  transform: scale(1);
}

/* Transitions */
.onyx-fade-enter-active, .onyx-fade-leave-active { transition: 0.2s; }
.onyx-fade-enter-from, .onyx-fade-leave-to { opacity: 0; transform: translateY(-10px); }

.fade-up-enter-active, .fade-up-leave-active { transition: 0.3s; }
.fade-up-enter-from, .fade-up-leave-to { opacity: 0; transform: translateY(10px); }

.fade-enter-active, .fade-leave-active { transition: 0.2s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

/* Responsive Design */
@media (max-width: 1400px) {
  .split-layout-v3 {
    padding: 60px 60px 60px 80px;
    gap: 60px;
  }
}

@media (max-width: 1200px) {
  .split-layout-v3 {
    grid-template-columns: 1fr;
    padding: 40px;
  }
  
  .page-title {
    font-size: 2.5rem;
  }
}

@media (max-width: 768px) {
  .split-layout-v3 {
    padding: 20px;
  }
  
  .page-title {
    font-size: 2rem;
  }
  
  .form-row-lux {
    flex-direction: column;
    gap: 16px;
  }
  
  .engine-controls-refined {
    flex-direction: column;
  }
}
</style>
