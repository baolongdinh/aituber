# 🤖 Local AI Hub Service

This service provides a local API to generate high-quality images using the **FLUX.1-schnell** model. It is designed to run efficiently on **CPU-only** systems with high RAM (60GB+ recommended).

## 🌟 Key Features

-   **Sequential Processing:** Uses an internal `asyncio.Queue` to process generation tasks one by one, preventing CPU/RAM exhaustion from concurrent requests.
-   **Async Return-then-Process:** The API returns a task ID and a predictive URL immediately, allowing the caller (Go Backend) to continue its workflow while the image renders in the background.
-   **CPU Optimized:** Configured with `bfloat16` and specific pipeline optimizations to leverage high-thread-count CPUs (24+ threads).
-   **Static Serving:** Automatically serves generated images from the `/outputs` directory via HTTP.

## 🛠 Installation

### 1. Prerequisites
- Python 3.10 or higher
- `pip` (Python package manager)

### 2. Setup
```bash
# Navigate to the service directory
cd hub-service

# Install dependencies
pip install -r requirements.txt
```

## 🚀 How to Run

By default, model weights (15-20GB) are stored in your home directory (`~/.cache/huggingface/hub`). If you want to keep them inside the project folder:

```bash
# Unix/Linux
export MODELS_DIR="./models"
python main.py
```

```bash
python main.py
```
*Note: On the first run, the service will download approximately 15-20GB of model weights from Hugging Face. Ensure you have a stable internet connection and sufficient disk space.*

## 🔌 API Documentation

### 1. Generate Image
**Endpoint:** `POST /generate`
**Body:**
```json
{
  "prompt": "Full cinematic shot of a futuristic city with neon lights",
  "width": 1024,
  "height": 1024,
  "num_inference_steps": 4
}
```
**Response:** Returns a `task_id` and a `url` where the image will be available once finished.

### 2. Check Status
**Endpoint:** `GET /status/{task_id}`
**Response:** Returns the current state (`pending`, `processing`, `completed`, or `failed`) and whether the file is ready to be downloaded.

## 🏗 Implementation Details

The service follows a **Producer-Consumer pattern**:
1.  **FastAPI (Producer):** Receives HTTP requests, generates a UUID for the task, pre-calculates the final filename, and pushes the task metadata into an `asyncio.Queue`.
2.  **Background Worker (Consumer):** An infinite loop that waits for tasks in the queue. It uses a global `FluxPipeline` instance to render images using `torch` on the CPU.
3.  **State Management:** An in-memory dictionary `task_statuses` tracks the progress of each request.

---
Built for **ViralCraft AI** local orchestration.
