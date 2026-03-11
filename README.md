# ⚡ ViralCraft AI: The Ultimate TikTok & YouTube Creator

**ViralCraft AI** is a professional-grade, automated video generation engine powered by **Google Gemini**. It transforms simple ideas or complex topics into high-engagement, viral video series for TikTok and YouTube, handling everything from scriptwriting and voice synthesis to visual orchestration and final composition.

![ViralCraft UI Overview](https://raw.githubusercontent.com/baolongdinh/aituber/main/frontend/src/assets/preview.png) *(Note: Replace with actual asset path or keep as placeholder)*

## 🚀 Key Capabilities

### 🧠 Intelligent Brain (Gemini 2.0/3.1)
- Generates viral-ready scripts in Vietnamese with a high-pacing "Gen Z" style.
- Automatically handles **Series Extraction**: Turn one topic into a 10-20 part series with logical flow and separate hook/CTA for each part.

### 🎥 Multi-Tier Visual Fallback System
Never worry about API failures again. ViralCraft uses a 5-tier orchestration logic to guarantee high-quality visuals for EVERY segment:
1.  **Tier 1:** Pexels Pro Stock Video (High-quality matches).
2.  **Tier 2:** AI Video Generation (HuggingFace T2V - Tencent Hunyuan/Mochi).
3.  **Tier 3:** AI Image Generation + Animation (HuggingFace FLUX.1 / FLUX Schnell).
4.  **Tier 4:** Google Gemini T2I (High-res cinematic images).
5.  **Tier 5:** Ultra Fallback (Automatic "Natural 4K" search or cinematic placeholders).

### 🎙️ Professional Voice Synthesis
- **ElevenLabs (Pro):** Deep, expressive, and human-like voices for premium content.
- **FPT.AI (Standard):** High-speed, reliable Vietnamese voice synthesis.
- Perfect auto-synchronization between audio duration and video segment timing.

### � Platform-Specific Optimization
- **TikTok/Shorts:** Vertical format (9:16), no intro/outro, fast-paced transitions.
- **YouTube:** Landscape format (16:9) with automated Intro & Outro concatenation for branding.

### 💬 Dynamic "Burned-in" Subtitles
- Automatically generates and burns high-quality, perfectly synced SRT subtitles into the video.
- Designed to be readable, clean, and visually engaging.

---

## 🛠 Tech Stack

-   **Backend:** Go (Golang) - Highly parallelized processing engine.
-   **Frontend:** Vue 3 + Vite + Custom CSS (Premium Modern Dark UI).
-   **Processing:** FFmpeg (Complex filter-graphs for transitions, scaling, and subtitle burning).
-   **AI Engines:** Google Gemini, HuggingFace Inference API, Pexels API, ElevenLabs API.

---

## 📋 Getting Started

### Prerequisites
- **Go 1.21+**
- **Node.js 18+**
- **FFmpeg** (Ensure it's in your system PATH)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/baolongdinh/aituber.git
   cd aituber
   ```

2. **Backend Setup**
   ```bash
   cd backend
   cp .env.example .env
   # Fill in your API Keys (GEMINI_API_KEY, PEXELS_API_KEY, HF_TOKEN, etc.)
   go run .
   ```

3. **Frontend Setup**
   ```bash
   cd ../frontend
   npm install
   npm run dev
   ```

---

## 🏗 Project Architecture

```
aituber/
├── backend/
│   ├── handlers/       # HTTP Logic & Failure Propagation
│   ├── services/       # StockVideo, Gemini, Audio orchestration
│   ├── utils/          # FFmpeg wrappers & File management
│   └── main.go         # Service entry point
├── frontend/
│   ├── src/            # Vue 3 modern UI components
│   └── vite.config.js  # Build configuration
└── README.md
```

## 🔧 Internal Failure Resiliency
ViralCraft is built for durability. It features:
- **Mocking Infrastructure:** Unit tests simulate API failures (402, 404, 500) to ensure fallback logic works offline.
- **Strict Error Handling:** Jobs fail explicitly if a segment cannot be generated, preventing "glitched" final videos.
- **Timeline Precision:** Subtitle offsets are calculated from zero-base to prevent desync on multi-segment renders.

---

## 📝 License
Built with ❤️ for AI Content Creators. Licensed under MIT.
