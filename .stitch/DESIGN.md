# AITuber Design System

## 1. Visual Archetype
- **Atmosphere**: Futuristic, Premium, Clean, Tech-focused
- **Concept**: Glassmorphism (Hiệu ứng kính mờ, đổ bóng nhẹ, điểm nhấn màu sắc rực rỡ)
- **RÀNG BUỘC THIẾT KẾ**: Ưu tiên sự đơn giản, rõ ràng. Chỉ thiết kế các control mà BE đã hỗ trợ. KHÔNG "vẽ" thêm tính năng ảo.
- **Theme**: Dark Mode mặc định.

## 2. Core Tokens
- **Palette**:
    - `Primary`: #40baf7 (AIOZ Blue) - For actions, CTA, focused states
    - `Surface`: #0a0a0c (Deep Black) - Main background
    - `Glass`: rgba(255, 255, 255, 0.05) with Backdrop Filter: Blur(20px)
    - `Accent`: #6366f1 (Indigo) - For gradients and highlights
    - `Success`: #10b981 (Emerald)
    - `Error`: #ef4444 (Red)
- **Typography**: 
    - Font Family: "Inter", sans-serif (Google Fonts)
    - Heading: Bold 600-700
    - Body: 14px-16px, Regular 400
- **Spacing**: 4px grid (4, 8, 12, 16, 24, 32, 48, 64)
- **Roundness**: `16px` (Large) for cards, `50px` (Pill) for buttons/badges

## 3. Component Guidelines
- **Buttons**:
    - Primary: Gradient primary-accent, white text
    - Secondary: Glass border, blurred background
- **Cards**:
    - Blurred background, 1px white border (10% opacity)
    - Subtle 0.1 opacity hover scale
- **Inputs**:
    - Minimalist glass-style fields with focused border glow

## 4. UI Patterns
- **Navigation**: Sticky top bar, glass background, centered links or side-drawer on mobile.
- **Grids**: Responsive layout (1 col mobile, 3-4 col desktop).
- **Transitions**: Smooth 0.3s ease-in-out for all interactions.

## 5. Design System Notes for Stitch Generation
[IMPORTANT: Copy this section to every Stitch prompt]

**DESIGN SYSTEM (REQUIRED for AITuber):**
- **Theme**: Dark Mode only. Background: #0a0a0c
- **Style**: Ultra-modern glassmorphism with Frosted Glass cards (rgba(255, 255, 255, 0.05) + backdrop-blur)
- **Palette**: Vibrant Primary: #40baf7. Accent Gradients: #40baf7 to #6366f1
- **Typography**: Clean Sans-Serif (Inter), Bold headers
- **Borders**: 1px subtle white (10% opacity) for cards and containers
- **Aesthetics**: Premium shadows, high-contrast text, sleek rounded corners (16px)
