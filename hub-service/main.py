import asyncio
import os
import uuid
import time
import logging
from typing import Optional
from fastapi import FastAPI, BackgroundTasks, HTTPException
from fastapi.staticfiles import StaticFiles
from pydantic import BaseModel
import torch
from diffusers import FluxPipeline
from PIL import Image

# Setup logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

app = FastAPI(title="Local AI Hub Service")

# Constants
OUTPUT_DIR = "outputs"
os.makedirs(OUTPUT_DIR, exist_ok=True)

# Global variables for model and queue
model_pipe = None
task_queue = asyncio.Queue()
task_statuses = {} # task_id -> {"status": "pending"|"processing"|"completed"|"failed", "url": str}

# Mount static files for output images
app.mount("/outputs", StaticFiles(directory=OUTPUT_DIR), name="outputs")

class GenerateRequest(BaseModel):
    prompt: str
    width: Optional[int] = 1024
    height: Optional[int] = 1024
    num_inference_steps: Optional[int] = 4 # Optimized for FLUX.1-schnell

def load_model():
    global model_pipe
    logger.info("Loading FLUX.1-schnell model to CPU... (This might take a while)")
    try:
        # Load the model with CPU optimizations
        model_pipe = FluxPipeline.from_pretrained(
            "black-forest-labs/FLUX.1-schnell",
            torch_dtype=torch.bfloat16 # Good for RAM and CPU
        )
        model_pipe.to("cpu")
        logger.info("Model loaded successfully!")
    except Exception as e:
        logger.error(f"Failed to load model: {e}")
        raise e

async def worker():
    """Sequential worker that processes generation tasks one by one."""
    global model_pipe
    while True:
        task = await task_queue.get()
        task_id, prompt, width, height, steps = task
        
        task_statuses[task_id]["status"] = "processing"
        logger.info(f"Processing task {task_id}: {prompt}")
        
        try:
            # Generate image
            start_time = time.time()
            image = model_pipe(
                prompt,
                width=width,
                height=height,
                num_inference_steps=steps,
                max_sequence_length=256,
                guidance_scale=0.0,
            ).images[0]
            
            # Save image
            output_path = os.path.join(OUTPUT_DIR, f"{task_id}.png")
            image.save(output_path)
            
            end_time = time.time()
            logger.info(f"Task {task_id} completed in {end_time - start_time:.2f}s")
            
            task_statuses[task_id]["status"] = "completed"
            task_statuses[task_id]["url"] = f"/outputs/{task_id}.png"
            
        except Exception as e:
            logger.error(f"Error processing task {task_id}: {e}")
            task_statuses[task_id]["status"] = "failed"
            task_statuses[task_id]["error"] = str(e)
        
        finally:
            task_queue.task_done()

@app.on_event("startup")
async def startup_event():
    load_model()
    asyncio.create_task(worker())

@app.post("/generate")
async def generate_image(request: GenerateRequest):
    task_id = str(uuid.uuid4())
    
    # Pre-register status
    task_statuses[task_id] = {
        "status": "pending",
        "url": f"/outputs/{task_id}.png", # Predictive URL
        "prompt": request.prompt
    }
    
    # Put into queue
    await task_queue.put((
        task_id, 
        request.prompt, 
        request.width, 
        request.height, 
        request.num_inference_steps
    ))
    
    logger.info(f"Task {task_id} queued")
    
    return {
        "task_id": task_id,
        "status": "pending",
        "url": f"/outputs/{task_id}.png"
    }

@app.get("/status/{task_id}")
async def get_status(task_id: str):
    if task_id not in task_statuses:
        raise HTTPException(status_code=404, detail="Task not found")
    
    status = task_statuses[task_id]
    
    # Check if the file actually exists for 'completed' check from caller's perspective
    file_exists = os.path.exists(os.path.join(OUTPUT_DIR, f"{task_id}.png"))
    
    return {
        "task_id": task_id,
        "status": status["status"],
        "url": status["url"],
        "file_ready": file_exists
    }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=5000)
