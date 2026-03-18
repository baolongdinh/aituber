# AITuber / ViralCraft Design System

## 1. Visual Archetype
- **Atmosphere**: Futuristic, Premium, Social Media Focused (TikTok/YouTube)
- **Concept**: Frosted Dark (Bề mặt tối, mờ ảo với điểm nhấn ánh sáng neon)
- **Brand**: ViralCraft AI
- **Theme**: Dark Mode only. Background: #0a0a0c với hiệu ứng Radial Glow màu Tím/Hồng (#a14bff / #ff3f6c).

## 2. Core Tokens
- **Palette (Dynamic Theme)**:
    - **TikTok Mode**:
        - `Primary`: #a14bff (Purple)
        - `Secondary`: #ff3f6c (Pink)
        - `Glow`: Radial gradient of Purple/Pink
    - **YouTube Mode**:
        - `Primary`: #ff0000 (Red)
        - `Secondary`: #b30000 (Dark Red)
        - `Glow`: Radial gradient of Red/Dark Red
    - `Surface`: #121214 (Dark Gray Card)
    - `Background`: #0a0a0c
- **Typography**: 
    - Font Family: "Inter" or "Outfit", sans-serif

## 3. UI Patterns
- **Theme Switching**: Cần có một platform switcher rõ rệt ở Header. Khi switch, toàn bộ màu `Primary`, `Glow` và `Active States` sẽ thay đổi tương ứng.
- **Split View Processing**: Giữ nguyên layout 2 cột cho Generator.
- **Library/Explore**: Dạng Grid thẻ (Glass cards) với preview video rực rỡ.

## 4. Design System Notes for Stitch Generation
[IMPORTANT: Copy this section to every Stitch prompt]

**DESIGN SYSTEM (REQUIRED for ViralCraft):**
- **Dual-Theme Logic**: 
    - If platform is TikTok: Use **TikTok Purple (#a14bff)** and **Pink (#ff3f6c)**.
    - If platform is YouTube: Use **YouTube Red (#ff0000)** and **Dark Red (#b30000)**.
- **Theme**: Ultra Dark Mode with platform-specific Radial Glow.
- **Style**: Frosted Dark cards (#121214) with subtle 1px borders.
- **Aesthetics**: Professional Split Screen or Grid layout. High-contrast labels, sleek "Gemini AI" branding.
