# Text-to-Video Generator Service

A simple web tool that converts long text scripts into complete videos with AI-generated visuals and text-to-speech narration.

## 🎯 Features

- Convert text scripts (up to 50,000 characters) into complete videos
- AI-powered text-to-speech with multiple voices
- AI-generated video visuals
- Seamless audio transitions with crossfade
- Smooth video transitions
- Perfect audio-video synchronization
- API key rotation to avoid rate limits

## 🛠 Tech Stack

- **Backend**: Golang 1.21+ with Gin framework
- **Frontend**: Vue.js 3 with Vuetify (coming soon)
- **Video Processing**: FFmpeg
- **External APIs**: 
  - FPT.AI / Zalo AI (Text-to-Speech)
  - Pika Labs / Leonardo.AI (Video Generation)

## 📋 Prerequisites

1. **Go 1.21+** - [Install Go](https://golang.org/doc/install)
2. **FFmpeg** - Install via package manager:
   ```bash
   # Ubuntu/Debian
   sudo apt install ffmpeg
   
   # macOS
   brew install ffmpeg
   ```
3. **Node.js 18+** - [Install Node.js](https://nodejs.org/)
4. **API Keys**:
   - FPT.AI TTS API key(s)
   - Pika Labs / Leonardo.AI API key(s)

## 🚀 Quick Start

### 1. Clone and Setup

```bash
cd /home/aiozlong/DATA/CODE/PROD/aituber
```

### 2. Configure Environment

```bash
# Copy example env file
cp .env.example .env

# Edit .env and add your API keys
nano .env
```

Required environment variables:
```bash
# Add multiple API keys separated by commas for load balancing
TTS_API_KEYS=your_fpt_key_1,your_fpt_key_2,your_fpt_key_3
VIDEO_API_KEYS=your_pika_key_1,your_pika_key_2
```

### 3. Run Backend

```bash
cd backend
go run main.go
```

Server will start on `http://localhost:8080`

### 4. Test Backend

```bash
# Health check
curl http://localhost:8080/health
```

## 📁 Project Structure

```
aituber/
├── backend/
│   ├── main.go              # Entry point
│   ├── config/              # Configuration management
│   │   └── config.go
│   ├── models/              # Data structures
│   │   └── types.go
│   ├── utils/               # Utilities
│   │   ├── api_key_pool.go  # API key rotation
│   │   ├── file_manager.go  # File operations
│   │   └── ffmpeg.go        # FFmpeg wrapper
│   ├── services/            # Business logic (TODO)
│   ├── handlers/            # HTTP handlers (TODO)
│   └── temp/                # Temporary files
├── frontend/                # Vue.js app (TODO)
└── .env.example             # Environment template
```

## 🔧 Development Status

### ✅ Completed (Phase 1 & 2)
- [x] Project structure
- [x] Go module initialization
- [x] Configuration management with validation
- [x] API key pool with smart rotation
- [x] Rate limiting and blacklisting
- [x] File manager utilities
- [x] FFmpeg wrapper (merge, transitions, encoding)
- [x] Gin server with CORS
- [x] Basic API endpoints structure

### 🚧 In Progress (Phase 3)
- [ ] Text processing service
- [ ] Audio service
- [ ] Video service
- [ ] Composer service
- [ ] HTTP handlers
- [ ] Frontend components

## 🎬 How It Works

1. **Text Input** → User pastes long text script
2. **Text Chunking** → Split into audio-friendly chunks (4500 chars)
3. **TTS Generation** → Convert each chunk to audio with API rotation
4. **Audio Merging** → Combine audio with crossfade transitions
5. **Video Segmentation** → Split text into 5-6s segments
6. **Video Generation** → Create AI videos for each segment
7. **Video Merging** → Combine videos with smooth transitions
8. **Final Composition** → Sync audio with video
9. **Output** → Download complete video

## 🔑 API Key Rotation Strategy

The system uses a smart API key pool that:
- Randomly selects keys, preferring less-used ones
- Automatically blacklists keys that hit rate limits
- Retries failed requests with different keys
- Tracks usage statistics

## 🎨 Video Quality Settings

Default settings ensure high quality:
- **Resolution**: 1920x1080 (Full HD)
- **Frame Rate**: 30 FPS
- **Video Bitrate**: 5 Mbps
- **Audio Bitrate**: 192 kbps
- **Audio Sample Rate**: 44.1 kHz
- **Crossfade Duration**: 0.3s
- **Video Transition**: Fade, 0.5s

## 📝 License

MIT License

## 🤝 Contributing

This is a work in progress. More features coming soon!

---

**Built with ❤️ for seamless video generation**
