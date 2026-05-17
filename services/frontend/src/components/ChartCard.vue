<template>
  <!-- Compact card -->
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-4">
    <div class="flex items-center justify-between mb-3">
      <h3 class="text-sm font-semibold text-gray-400">{{ title }}</h3>
      <button
        @click="isExpanded = true"
        class="text-gray-500 hover:text-gray-300 transition-colors p-1 rounded hover:bg-gray-800"
        title="Expand"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5v-4m0 4h-4m4 0l-5-5" />
        </svg>
      </button>
    </div>
    <slot :expanded="false" />
  </div>

  <!-- Modal overlay -->
  <Transition name="fade">
    <div
      v-if="isExpanded"
      class="fixed inset-0 z-50 flex items-center justify-center"
      @click.self="isExpanded = false"
    >
      <!-- Backdrop -->
      <div class="absolute inset-0 bg-black/70 backdrop-blur-sm" @click="isExpanded = false" />

      <!-- Modal panel -->
      <div class="relative z-10 bg-gray-900 border border-gray-700 rounded-xl p-6 w-full max-w-4xl mx-4 shadow-2xl">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-base font-semibold text-gray-200">{{ title }}</h2>
          <button
            @click="isExpanded = false"
            class="text-gray-400 hover:text-gray-200 transition-colors p-1 rounded hover:bg-gray-800"
            title="Close"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <slot :expanded="true" />
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref } from 'vue'

defineProps<{ title: string }>()

const isExpanded = ref(false)
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
