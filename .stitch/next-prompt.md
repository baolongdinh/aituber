---
page: job-detail
---
# AITuber Job Progress & Result View - Simple & Aligned

Design a high-fidelity "Job Detail" or "Processing" page for AITuber. This page is shown after a user starts a generation or clicks on a recent job. It must clearly communicate the state of the backend task.

**DESIGN SYSTEM (REQUIRED for AITuber):**
- **Theme**: Dark Mode only. Background: #0a0a0c
- **Style**: Ultra-modern glassmorphism with Frosted Glass cards (rgba(255, 255, 255, 0.05) + backdrop-blur)
- **Palette**: Vibrant Primary: #40baf7. Accent Gradients: #40baf7 to #6366f1
- **Typography**: Clean Sans-Serif (Inter), Bold headers
- **Borders**: 1px subtle white (10% opacity) for cards and containers
- **Aesthetics**: Premium shadows, high-contrast text, sleek rounded corners (16px)

**PAGE STRUCTURE:**
1. **Header**:
   - Title: "[Content Name]" (e.g., "The Future of AI - Part 1")
   - Breadcrumbs: Dashboard / Jobs / [JobID]

2. **Main Status Card (Large Glass Card)**:
   - **Status Badge**: Clear indicators for "Queued", "Processing", "Completed", or "Failed" (using Success/Error/Primary colors).
   - **Progress Section**:
     - Large, sleek circular or horizontal progress bar (0-100%).
     - Percentage text: "65% complete".
     - **Current Activity**: Text label showing the active backend step (e.g., "Step: Merging audio and video tracks...").

3. **Output Content (Visible only when Status is "Completed")**:
   - **Video Player**: A centered, premium video player component.
   - **Actions**:
     - "Download Video" (Primary Action button).
     - "Back to Dashboard" (Secondary button).

4. **Error Info (Visible only when Status is "Failed")**:
   - A red-tinted glass card showing the error message details.
   - "Retry Job" button.

5. **Configuration Summary (Sidebar or Bottom Grid)**:
   - Topic: [Topic text]
   - Platform: [Icon] YouTube
   - Voice: Bella
   - Model: FLUX-1.dev

**UNIMPLEMENTED FEATURES (DO NOT INCLUDE):**
- No social sharing, comments, or likes.
- No editing tools on the result.
- No analytics or performance charts.
