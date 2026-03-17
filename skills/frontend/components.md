# Frontend Component Patterns

## TRIGGER
Read this file when you need to know how to organize and write Vue components, Tailwind patterns, common UI components.

---

## COMPONENT CATEGORIES

### 1. Common Components (`src/components/common/`)
Reusable everywhere in the app:

**AppButton.vue**
```vue
<script setup>
const props = defineProps({
  variant: { type: String, default: 'primary' }, // primary | secondary | danger | ghost
  size:    { type: String, default: 'md' },       // sm | md | lg
  loading: { type: Boolean, default: false },
  disabled:{ type: Boolean, default: false },
})
const emit = defineEmits(['click'])
</script>

<template>
  <button
    :class="[
      'rounded-lg font-medium transition-all duration-200',
      size === 'sm' ? 'px-3 py-1.5 text-sm' : size === 'lg' ? 'px-6 py-3 text-lg' : 'px-4 py-2 text-base',
      variant === 'primary'   ? 'bg-indigo-600 text-white hover:bg-indigo-700' : '',
      variant === 'secondary' ? 'bg-gray-100  text-gray-800 hover:bg-gray-200' : '',
      variant === 'danger'    ? 'bg-red-600   text-white hover:bg-red-700'     : '',
      variant === 'ghost'     ? 'bg-transparent text-indigo-600 hover:bg-indigo-50' : '',
      (disabled || loading)   ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer',
    ]"
    :disabled="disabled || loading"
    @click="!disabled && !loading && emit('click', $event)"
  >
    <span v-if="loading" class="mr-2 animate-spin">⟳</span>
    <slot />
  </button>
</template>
```

**AppToast.vue** — used for notifications
**AppModal.vue** — used for dialogs/confirmations
**AppBadge.vue** — used for status badges
**AppSkeleton.vue** — used for loading skeletons

### 2. Domain Components (`src/components/<domain>/`)
Components belonging to a specific feature:
- `nft/NftCard.vue`, `nft/NftGrid.vue`, `nft/NftFilters.vue`
- `user/UserAvatar.vue`, `user/UserStats.vue`
- `wallet/ConnectWalletButton.vue`, `wallet/WalletBadge.vue`

---

## TAILWIND PATTERNS

### Glassmorphism (used for cards)
```html
<div class="bg-white/10 backdrop-blur-md border border-white/20 rounded-2xl p-6 shadow-xl">
  <!-- card content -->
</div>
```

### Dark Mode Card
```html
<div class="bg-gray-900 border border-gray-800 rounded-2xl p-6 hover:border-indigo-500 transition-colors">
  <!-- card content -->
</div>
```

### Gradient Text
```html
<h1 class="bg-gradient-to-r from-indigo-400 to-purple-400 bg-clip-text text-transparent font-bold">
  Title
</h1>
```

### Loading Skeleton
```html
<div class="animate-pulse space-y-3">
  <div class="h-4 bg-gray-700 rounded w-3/4"></div>
  <div class="h-4 bg-gray-700 rounded w-1/2"></div>
</div>
```

### Empty State
```html
<div class="flex flex-col items-center justify-center py-16 text-gray-500">
  <div class="text-5xl mb-4">🎨</div>
  <p class="text-lg font-medium">No NFTs found</p>
  <p class="text-sm">Start by minting your first NFT</p>
</div>
```

---

## FORM HANDLING PATTERN

```vue
<script setup>
import { reactive, ref } from 'vue'

const form = reactive({
  username: '',
  email: '',
})
const errors = ref({})
const isSubmitting = ref(false)

async function handleSubmit() {
  // Client-side validation
  errors.value = {}
  if (!form.username) errors.value.username = 'Username is required'
  if (Object.keys(errors.value).length) return

  isSubmitting.value = true
  try {
    await someStore.submit(form)
  } catch (err) {
    errors.value.api = err.response?.data?.error || 'Something went wrong'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <form @submit.prevent="handleSubmit">
    <div>
      <input v-model="form.username" />
      <p v-if="errors.username" class="text-red-400 text-sm">{{ errors.username }}</p>
    </div>
    <p v-if="errors.api" class="text-red-400 text-sm">{{ errors.api }}</p>
    <AppButton type="submit" :loading="isSubmitting">Submit</AppButton>
  </form>
</template>
```

---

## TAILWIND CONFIG TEMPLATE

```js
// tailwind.config.js
export default {
  content: ['./index.html', './src/**/*.{vue,js}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        brand: {
          50:  '#eef2ff',
          500: '#6366f1',
          600: '#4f46e5',
          900: '#1e1b4b',
        }
      },
      fontFamily: {
        sans: ['Inter', 'sans-serif'],
      },
    },
  },
}
```

---

## DO / DON'T

✅ **DO**
- Dùng `v-bind` shorthand `:prop` và `v-on` shorthand `@event`
- Luôn đặt `:key` khi dùng `v-for`
- Extract logic phức tạp ra composable
- Dùng Tailwind `group`, `peer` utilities cho interactive states
- Prefix event handlers với `handle`: `handleSubmit`, `handleClose`

❌ **DON'T**
- KHÔNG viết CSS inline (`style="..."`)
- KHÔNG dùng `v-html` trừ khi đã sanitize
- KHÔNG mix `v-if` và `v-for` trên cùng element
- KHÔNG emit events có tên giống native events (dùng prefix namespace)

---

## RESPONSIVE DESIGN (BẮT BUỘC)

> ⚠️ **MANDATORY**: Mọi UI component đều phải hoạt động tốt trên **CẢ desktop lẫn mobile/tablet**. Không có ngoại lệ.
> - Viết cho **desktop trước** (đây là màn hình ưu tiên)
> - Sau đó dùng `max-md:` / `max-sm:` để đảm bảo hiển thị đúng trên mobile
> - Trước khi commit bất kỳ UI nào, phải kiểm tra ở cả 3 breakpoint: desktop / tablet / mobile

### Chiến lược: Desktop-first với Tailwind

Tailwind mặc định là **mobile-first** (không prefix = nhỏ nhất). Dự án này dùng **desktop-first** với **max-* variant**:

| Variant      | Ý nghĩa                          | Tương đương CSS                  |
|-------------|----------------------------------|----------------------------------|
| *(không prefix)* | Áp dụng từ mobile trở lên  | áp dụng mọi lúc (baseline)       |
| `md:` | Áp dụng từ `≥ 768px` trở lên         | `@media (min-width: 768px)`      |
| `lg:` | Áp dụng từ `≥ 1024px` trở lên        | `@media (min-width: 1024px)`     |
| `max-md:` | Áp dụng khi `< 768px` (tablet/mobile) | `@media (max-width: 767px)` |
| `max-sm:` | Áp dụng khi `< 640px` (chỉ mobile)   | `@media (max-width: 639px)` |

### Quy tắc viết class (Desktop-first)

```html
<!-- ✅ Desktop-first: viết desktop trước, dùng max-md: để override xuống nhỏ hơn -->
<div class="
  flex flex-row          <!-- desktop: ngang -->
  max-md:flex-col        <!-- tablet + mobile: xuống dọc -->
">

<p class="
  text-lg                <!-- desktop -->
  max-md:text-base       <!-- tablet -->
  max-sm:text-sm         <!-- mobile -->
">

<!-- Grid: desktop 4 cột → tablet 2 cột → mobile 1 cột -->
<div class="grid grid-cols-4 max-lg:grid-cols-3 max-md:grid-cols-2 max-sm:grid-cols-1 gap-6">
```

---

### Layout Patterns

#### Page Layout (DefaultLayout.vue)
```vue
<template>
  <div class="min-h-screen bg-gray-950 text-white">
    <!-- Navbar: full desktop nav, hamburger trên mobile -->
    <AppNav />

    <!-- Main content: desktop có sidebar-aware padding + max width -->
    <main class="max-w-7xl mx-auto px-8 py-10 max-md:px-4 max-md:py-6">
      <slot />
    </main>
  </div>
</template>
```

#### Desktop Navigation (Responsive xuống mobile)
```vue
<template>
  <header class="sticky top-0 z-40 bg-gray-900/80 backdrop-blur border-b border-gray-800">
    <div class="max-w-7xl mx-auto flex items-center justify-between px-8 h-16 max-md:px-4 max-md:h-14">

      <!-- Logo -->
      <RouterLink to="/" class="font-bold text-xl max-md:text-lg">DApp</RouterLink>

      <!-- Desktop nav links — ẩn trên mobile -->
      <nav class="flex items-center gap-8 text-sm text-gray-400 max-md:hidden">
        <RouterLink to="/nfts"          class="hover:text-white transition-colors">NFTs</RouterLink>
        <RouterLink to="/my-collection" class="hover:text-white transition-colors">Collection</RouterLink>
        <RouterLink to="/profile"       class="hover:text-white transition-colors">Profile</RouterLink>
      </nav>

      <!-- Right actions -->
      <div class="flex items-center gap-3">
        <appkit-button />
        <!-- Hamburger: chỉ hiện trên mobile -->
        <button class="hidden max-md:flex p-2 rounded-lg hover:bg-gray-800" @click="menuOpen = !menuOpen">
          ☰
        </button>
      </div>
    </div>

    <!-- Mobile dropdown menu -->
    <nav v-if="menuOpen" class="hidden max-md:flex flex-col px-4 pb-4 gap-3 text-sm border-t border-gray-800">
      <RouterLink to="/nfts"          @click="menuOpen = false" class="py-2 text-gray-300 hover:text-white">NFTs</RouterLink>
      <RouterLink to="/my-collection" @click="menuOpen = false" class="py-2 text-gray-300 hover:text-white">Collection</RouterLink>
      <RouterLink to="/profile"       @click="menuOpen = false" class="py-2 text-gray-300 hover:text-white">Profile</RouterLink>
    </nav>
  </header>
</template>

<script setup>
import { ref } from 'vue'
const menuOpen = ref(false)
</script>
```

---

### NFT Grid (Desktop-first Responsive)

```html
<!-- Desktop: 4 cột → Tablet: 2-3 cột → Mobile: 1-2 cột -->
<div class="grid grid-cols-4 gap-6  max-lg:grid-cols-3  max-md:grid-cols-2  max-sm:grid-cols-1">
  <NftCard v-for="nft in nfts" :key="nft.id" :nft="nft" />
</div>
```

### NftCard
```html
<div class="bg-gray-900 border border-gray-800 rounded-2xl overflow-hidden
            hover:border-indigo-500 transition-all duration-300 flex flex-col">
  <!-- Image -->
  <div class="aspect-square w-full overflow-hidden">
    <img :src="nft.imageUrl" class="w-full h-full object-cover hover:scale-105 transition-transform duration-500" />
  </div>

  <!-- Info: padding lớn hơn trên desktop -->
  <div class="p-5 max-md:p-3 flex flex-col gap-3">
    <h3 class="font-semibold text-base max-md:text-sm truncate">{{ nft.name }}</h3>
    <div class="flex items-center justify-between">
      <span class="text-indigo-400 font-bold">{{ nft.price }} AIOZ</span>
      <button class="bg-indigo-600 hover:bg-indigo-700 text-white text-sm px-4 py-2 rounded-lg transition-colors">
        Buy
      </button>
    </div>
  </div>
</div>
```

---

### Modal (Desktop dialog, mobile bottom sheet)
```html
<div class="fixed inset-0 z-50 flex items-center justify-center max-md:items-end bg-black/60 backdrop-blur-sm">
  <div class="
    bg-gray-900 border border-gray-800 rounded-2xl p-8 w-[480px] max-h-[80vh] overflow-y-auto
    max-md:w-full max-md:rounded-b-none max-md:rounded-t-2xl max-md:p-6
  ">
    <slot />
  </div>
</div>
```

---

### Typography & Spacing Scale (Desktop-first)

```html
<!-- Headings: desktop full size, mobile giảm xuống -->
<h1 class="text-4xl font-bold max-lg:text-3xl max-md:text-2xl">Page Title</h1>
<h2 class="text-2xl font-semibold max-md:text-xl">Section Title</h2>
<p  class="text-base max-md:text-sm text-gray-400">Body text</p>

<!-- Section padding -->
<section class="py-16 max-md:py-10 max-sm:py-6">

<!-- Standard container -->
<div class="max-w-7xl mx-auto px-8 max-md:px-4">
```

---

### Touch Targets (Mobile override)

```html
<!-- Desktop button bình thường, mobile cần đủ lớn để tap -->
<button class="px-6 py-2.5 text-sm rounded-lg max-md:py-3 max-md:min-h-[44px]">
  Action
</button>

<!-- Input: desktop bình thường, mobile dùng text-base tránh iOS auto-zoom -->
<input class="px-4 py-2.5 text-sm max-md:text-base rounded-xl w-full" />
```

---

## DO / DON'T (Responsive — BẮT BUỘC)

✅ **DO**
- **LUÔN** test UI ở 3 mức: desktop (`>1024px`) / tablet (`768px`) / mobile (`375px`)
- Viết desktop layout trước, dùng `max-md:` / `max-sm:` để đảm bảo mobile hoạt động
- Dùng `max-w-7xl mx-auto` làm container chuẩn trên desktop
- Ẩn nav links trên mobile với `max-md:hidden`, show hamburger với `hidden max-md:flex`
- Button/input trên mobile phải có `min-h-[44px]`; input dùng `text-base` để tránh iOS auto-zoom
- Dùng `aspect-square` / `aspect-video` cho image để tỷ lệ nhất quán mọi màn hình
- Dùng `w-full` cho input/button trên mobile (`max-sm:w-full`)

❌ **DON'T**
- **KHÔNG** ship UI chỉ đẹp trên desktop mà vỡ trên mobile — đây là lỗi nghiêm trọng
- KHÔNG hardcode `width: 1200px` — dùng `max-w-7xl` + `w-full`
- KHÔNG dùng giá trị spacing tùy tiện như `px-[30px]` — dùng Tailwind scale
- KHÔNG dùng `overflow: hidden` trên body/html — gây lỗi scroll trên mobile
- KHÔNG bỏ qua `max-md:` override khi layout có nhiều column trên desktop
