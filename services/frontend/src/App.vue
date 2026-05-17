<template>
  <div class="min-h-screen bg-gray-950">
    <!-- Navbar -->
    <nav class="bg-gray-900 border-b border-gray-700 sticky top-0 z-50">
      <div class="max-w-screen-2xl mx-auto px-4 sm:px-6">
        <div class="flex items-center justify-between h-14">
          <!-- Logo -->
          <div class="flex items-center gap-3">
            <svg class="w-6 h-6 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
            </svg>
            <span class="font-bold text-white tracking-wide text-sm">CORAZA WAF</span>
          </div>
          <!-- Nav Links -->
          <div class="flex items-center gap-1">
            <RouterLink
              to="/"
              class="px-4 py-2 rounded text-sm font-medium transition-colors"
              :class="$route.path === '/' ? 'bg-gray-700 text-white' : 'text-gray-400 hover:text-white hover:bg-gray-800'"
            >
              Dashboard
            </RouterLink>
            <RouterLink
              to="/events"
              class="px-4 py-2 rounded text-sm font-medium transition-colors"
              :class="$route.path === '/events' ? 'bg-gray-700 text-white' : 'text-gray-400 hover:text-white hover:bg-gray-800'"
            >
              Events
            </RouterLink>
          </div>
          <!-- Health indicator -->
          <div class="flex items-center gap-2 text-xs text-gray-500">
            <span
              class="w-2 h-2 rounded-full"
              :class="healthOk ? 'bg-green-500' : 'bg-red-500'"
            ></span>
            <span>{{ healthOk ? 'API Online' : 'API Offline' }}</span>
          </div>
        </div>
      </div>
    </nav>

    <!-- Main content -->
    <main class="max-w-screen-2xl mx-auto px-4 sm:px-6 py-6">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
import { checkHealth } from '@/api/events'

const healthOk = ref(false)

onMounted(async () => {
  try {
    await checkHealth()
    healthOk.value = true
  } catch {
    healthOk.value = false
  }
})
</script>
