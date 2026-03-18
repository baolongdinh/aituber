<script setup>
import MainLayout from '@/layouts/MainLayout.vue'
</script>

<template>
  <div id="app-root">
    <MainLayout>
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </MainLayout>
  </div>
</template>

<style>
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>

<style>
/* Global styles are in assets/main.css */
@import '@/assets/main.css';

#app-root {
  min-height: 100vh;
}
</style>
