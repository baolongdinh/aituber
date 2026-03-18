<script setup>
import { ref, reactive, onMounted } from 'vue'
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

const form = reactive({
  topic: '',
  content_name: '',
  is_series: false,
  num_parts: 5,
  voice: 'ban_mai',
  tts_provider: 'fpt',
  t2v_model: 'flux-1-dev',
  stock_keywords: ''
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

onMounted(() => {
  checkActiveTask()
})
</script>

<template>
  <div class="generator-view" :class="{ 'is-processing': isGenerating }">
    <div class="split-layout">
      <!-- Left Column: Settings -->
      <aside class="settings-column">
        <div class="column-header">
          <h1 class="text-slate-100 font-black text-2xl">Universal Generator</h1>
          <p class="text-slate-400 text-sm">Cấu hình tham số cho video Viral của bạn</p>
        </div>

        <div class="settings-card glass-card" :class="{ 'opacity-50 pointer-events-none': isGenerating }">
          <div class="form-group">
            <label>Bạn muốn tạo video về chủ đề gì?</label>
            <textarea 
              v-model="form.topic" 
              placeholder="Vd: Lịch sử tiền điện tử, 3 bí mật về thành công..."
              class="custom-textarea"
            ></textarea>
          </div>

          <div class="form-grid">
            <div class="form-group">
              <label>Giọng đọc AI</label>
              <select v-model="form.voice" class="custom-select">
                <option value="ban_mai">Ban Mai (Nữ)</option>
                <option value="minh_quang">Minh Quang (Nam)</option>
                <option value="le_minh">Lê Minh (Nam)</option>
              </select>
            </div>
            <div class="form-group">
               <label>Nền tảng mục tiêu</label>
               <div class="platform-indicator" :class="platform">
                 <span class="material-symbols-outlined">{{ platform === 'tiktok' ? 'filter_frames' : 'play_circle' }}</span>
                 {{ platform === 'tiktok' ? 'TikTok' : 'YouTube' }}
               </div>
            </div>
          </div>

          <div class="series-toggle-box" @click="form.is_series = !form.is_series">
            <div class="toggle-info">
              <span class="material-symbols-outlined icon">layers</span>
              <div>
                <h4>Chế độ Series</h4>
                <p>Tự động tạo chuỗi các tập liên kết</p>
              </div>
            </div>
            <div class="toggle-switch" :class="{ active: form.is_series }"></div>
          </div>

          <div v-if="form.is_series" class="slider-group">
            <div class="slider-header">
              <span>Số tập phim</span>
              <span class="value">{{ form.num_parts }} tập</span>
            </div>
            <input type="range" v-model.number="form.num_parts" min="2" max="20" class="custom-slider">
          </div>

          <button 
            @click="handleGenerate" 
            class="generate-btn" 
            :disabled="!form.topic || isGenerating"
          >
            <span class="material-symbols-outlined">bolt</span>
            {{ isGenerating ? 'Đang khởi tạo...' : 'Bắt đầu tạo Video' }}
          </button>
        </div>
      </aside>

      <!-- Right Column: Results / Processing -->
      <main class="results-column">
        <!-- Idle State -->
        <div v-if="!isGenerating && seriesParts.length === 0" class="idle-state glass-card">
          <div class="placeholder-icon">
            <span class="material-symbols-outlined">movie_filter</span>
          </div>
          <h3>Sẵn sàng tạo nội dung Viral?</h3>
          <p>Nhập chủ đề ở bên trái và ViralCraft sẽ lo phần còn lại.</p>
        </div>

        <!-- Processing State -->
        <div v-else class="processing-state flex flex-col gap-6">
          <!-- Overall Status -->
          <div class="status-card glass-card" :class="platform">
            <div class="status-header">
              <div class="progress-info">
                <span class="percentage">{{ progress }}%</span>
                <div>
                  <h4>Tiến độ tổng quát</h4>
                  <p class="step">{{ currentStep }}</p>
                </div>
              </div>
              <div class="badge-pulse">
                <span class="dot"></span>
                Đang xử lý...
              </div>
            </div>
            <div class="overall-progress-bar">
              <div class="fill" :style="{ width: progress + '%' }"></div>
            </div>
          </div>

          <!-- Parts List -->
          <div class="parts-list-card glass-card">
            <div class="list-header">
              <h3>Danh sách các tập ({{ form.is_series ? form.num_parts : 1 }})</h3>
            </div>
            <div class="list-container custom-scrollbar">
              <div v-for="(part, index) in seriesParts" :key="index" class="part-row" :class="part.status">
                <div class="part-info">
                  <div class="part-icon">
                    <span class="material-symbols-outlined">{{ part.status === 'completed' ? 'check_circle' : (part.status === 'processing' ? 'sync' : 'hourglass_empty') }}</span>
                  </div>
                  <div>
                    <h4>Tập {{ index + 1 }}: {{ part.title || 'Đang chuẩn bị...' }}</h4>
                    <p class="part-status-text">{{ part.status === 'completed' ? 'Đã hoàn thành' : (part.status === 'processing' ? 'Đang tạo video...' : 'Đang chờ...') }}</p>
                  </div>
                </div>
                <div class="part-badge" :class="part.status">
                  {{ part.status === 'completed' ? 'Xong' : (part.status === 'processing' ? 'Đang chạy' : 'Chờ') }}
                </div>
                <div v-if="part.status === 'processing'" class="mini-progress">
                  <div class="fill" :style="{ width: part.progress + '%' }"></div>
                </div>
              </div>
              
              <!-- Placeholder for single job -->
              <div v-if="!form.is_series && seriesParts.length === 0" class="part-row processing">
                <div class="part-info">
                  <div class="part-icon"><span class="material-symbols-outlined">sync</span></div>
                  <div>
                    <h4>Video đơn: {{ form.content_name || 'Đang khởi tạo' }}</h4>
                    <p class="part-status-text">Đang xử lý bước cuối...</p>
                  </div>
                </div>
                <div class="part-badge processing">Đang chạy</div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.generator-view { min-height: calc(100vh - 128px); }
.split-layout { display: grid; grid-template-columns: 450px 1fr; gap: 40px; }

/* Settings Column */
.settings-column { display: flex; flex-direction: column; gap: 24px; }
.column-header { margin-bottom: 8px; }

.settings-card { padding: 32px; display: flex; flex-direction: column; gap: 24px; }

.form-group { display: flex; flex-direction: column; gap: 10px; }
.form-group label { font-size: 0.75rem; font-weight: 700; color: rgba(255, 255, 255, 0.4); text-transform: uppercase; letter-spacing: 0.05em; margin-left: 4px; }

.custom-textarea, .custom-input, .custom-select {
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 12px;
  padding: 14px;
  color: #fff;
  font-size: 0.9rem;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.custom-textarea { height: 120px; resize: none; }
.custom-textarea:focus, .custom-input:focus, .custom-select:focus {
  border-color: var(--tiktok-primary);
  box-shadow: 0 0 0 4px rgba(161, 75, 255, 0.1);
  outline: none;
}

.platform-indicator {
  padding: 14px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.9rem;
  font-weight: 700;
  text-transform: capitalize;
}
.platform-indicator.tiktok { color: var(--tiktok-primary); }
.platform-indicator.youtube { color: var(--youtube-primary); }

.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

.series-toggle-box {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 16px;
  padding: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  transition: background 0.2s;
}
.series-toggle-box:hover { background: rgba(255, 255, 255, 0.05); }

.toggle-info { display: flex; align-items: center; gap: 12px; }
.toggle-info .icon { color: var(--tiktok-primary); }
.toggle-info h4 { font-size: 0.85rem; font-weight: 700; }
.toggle-info p { font-size: 0.7rem; color: rgba(255, 255, 255, 0.4); }

.toggle-switch {
  width: 40px; height: 20px;
  background: #252529;
  border-radius: 10px;
  position: relative;
  transition: background 0.3s;
}
.toggle-switch::after {
  content: '';
  position: absolute;
  top: 2px; left: 2px;
  width: 16px; height: 16px;
  background: #fff;
  border-radius: 50%;
  transition: transform 0.3s;
}
.toggle-switch.active { background: var(--tiktok-primary); }
.toggle-switch.active::after { transform: translateX(20px); }

.slider-group { display: flex; flex-direction: column; gap: 12px; }
.slider-header { display: flex; justify-content: space-between; font-size: 0.75rem; font-weight: 700; }
.slider-header .value { color: var(--tiktok-primary); }
.custom-slider { width: 100%; accent-color: var(--tiktok-primary); cursor: pointer; }

.generate-btn {
  margin-top: 12px;
  padding: 18px;
  border-radius: 16px;
  border: none;
  background: linear-gradient(135deg, #a14bff, #ff3f6c);
  color: #fff;
  font-weight: 800;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  cursor: pointer;
  box-shadow: 0 10px 20px rgba(161, 75, 255, 0.2);
  transition: transform 0.2s, box-shadow 0.2s;
}
.generate-btn:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 15px 30px rgba(161, 75, 255, 0.3); }
.generate-btn:disabled { opacity: 0.5; cursor: not-allowed; filter: grayscale(1); }

/* Results Column */
.results-column { position: relative; }

.idle-state {
  height: 600px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 40px;
  border: 2px dashed rgba(255, 255, 255, 0.05);
  background: rgba(18, 18, 20, 0.3);
}
.placeholder-icon {
  width: 100px; height: 100px;
  background: rgba(161, 75, 255, 0.05);
  border-radius: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24px;
  color: rgba(255, 255, 255, 0.1);
}
.placeholder-icon .material-symbols-outlined { font-size: 3rem; }
.idle-state h3 { font-size: 1.5rem; font-weight: 800; margin-bottom: 12px; }
.idle-state p { color: rgba(255, 255, 255, 0.4); max-width: 300px; line-height: 1.6; }

/* Status Card */
.status-card { padding: 32px; position: relative; overflow: hidden; }
.status-card.tiktok { background: linear-gradient(135deg, rgba(161, 75, 255, 0.15), rgba(10, 10, 12, 0)); border-color: rgba(161, 75, 255, 0.2); }
.status-card.youtube { background: linear-gradient(135deg, rgba(255, 0, 0, 0.1), rgba(10, 10, 12, 0)); border-color: rgba(255, 0, 0, 0.2); }

.status-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px; }
.progress-info { display: flex; align-items: center; gap: 20px; }
.percentage { font-size: 3.5rem; font-weight: 900; letter-spacing: -0.05em; color: #fff; }
.progress-info h4 { font-size: 1.1rem; font-weight: 800; }
.progress-info .step { font-size: 0.8rem; color: rgba(255, 255, 255, 0.4); font-style: italic; }

.badge-pulse {
  padding: 6px 14px;
  background: rgba(161, 75, 255, 0.1);
  color: var(--tiktok-primary);
  border-radius: 999px;
  font-size: 0.7rem;
  font-weight: 800;
  text-transform: uppercase;
  display: flex;
  align-items: center;
  gap: 8px;
  animation: pulse-border 2s infinite;
}
.theme-youtube .badge-pulse { color: var(--youtube-primary); background: rgba(255, 0, 0, 0.1); }
.badge-pulse .dot { width: 6px; height: 6px; border-radius: 50%; background: currentColor; }

.overall-progress-bar { height: 4px; background: rgba(255, 255, 255, 0.05); border-radius: 2px; overflow: hidden; }
.overall-progress-bar .fill { height: 100%; background: linear-gradient(to right, #a14bff, #ff3f6c); transition: width 0.5s ease; }
.theme-youtube .overall-progress-bar .fill { background: linear-gradient(to right, #ff0000, #b30000); }

/* List Card */
.parts-list-card { flex: 1; display: flex; flex-direction: column; }
.list-header { padding: 24px; border-bottom: 1px solid rgba(255, 255, 255, 0.05); }
.list-header h3 { font-size: 1rem; font-weight: 700; }

.list-container { padding: 12px; display: flex; flex-direction: column; gap: 8px; max-height: 480px; overflow-y: auto; }

.part-row {
  padding: 16px;
  border-radius: 12px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid transparent;
  position: relative;
  overflow: hidden;
}

.part-row.processing { background: rgba(161, 75, 255, 0.05); border-color: rgba(161, 75, 255, 0.1); }
.part-row.completed { background: rgba(16, 185, 129, 0.05); }

.part-info { display: flex; align-items: center; gap: 16px; z-index: 1; }
.part-icon { width: 32px; height: 32px; border-radius: 8px; display: flex; align-items: center; justify-content: center; background: rgba(255, 255, 255, 0.05); }

.part-row.processing .part-icon { color: var(--tiktok-primary); animation: spin 4s linear infinite; }
.part-row.completed .part-icon { color: #10b981; }

.part-info h4 { font-size: 0.85rem; font-weight: 700; color: #fff; }
.part-status-text { font-size: 0.7rem; color: rgba(255, 255, 255, 0.3); }

.part-badge {
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 0.65rem;
  font-weight: 800;
  text-transform: uppercase;
  z-index: 1;
}
.part-badge.waiting { background: rgba(255, 255, 255, 0.05); color: rgba(255, 255, 255, 0.3); }
.part-badge.processing { background: rgba(161, 75, 255, 0.2); color: var(--tiktok-primary); }
.part-badge.completed { background: rgba(16, 185, 129, 0.2); color: #10b981; }

.mini-progress { position: absolute; bottom: 0; left: 0; right: 0; height: 1px; background: rgba(255, 255, 255, 0.05); }
.mini-progress .fill { height: 100%; background: var(--tiktok-primary); transition: width 0.3s; }

@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
@keyframes pulse-border { 0% { box-shadow: 0 0 0 0 rgba(161, 75, 255, 0.2); } 70% { box-shadow: 0 0 0 6px rgba(161, 75, 255, 0); } 100% { box-shadow: 0 0 0 0 rgba(161, 75, 255, 0); } }
</style>
