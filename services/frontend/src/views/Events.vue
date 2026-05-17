<template>
  <div class="p-6 space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-bold text-gray-100">WAF Events</h1>
      <span class="text-sm text-gray-400">{{ store.total }} total</span>
    </div>

    <FilterBar @apply="onFilter" />

    <div class="bg-gray-900 border border-gray-800 rounded-xl">
      <EventsTable :events="store.events" />
    </div>

    <!-- Pagination -->
    <div class="flex gap-2 items-center justify-between">
      <button @click="prevPage" :disabled="page === 0"
        class="px-4 py-1.5 rounded bg-gray-800 text-gray-300 disabled:opacity-40 hover:bg-gray-700 transition-colors text-sm">
        ← Prev
      </button>
      <span class="text-sm text-gray-400">Page {{ page + 1 }} of {{ totalPages }}</span>
      <button @click="nextPage" :disabled="page >= totalPages - 1"
        class="px-4 py-1.5 rounded bg-gray-800 text-gray-300 disabled:opacity-40 hover:bg-gray-700 transition-colors text-sm">
        Next →
      </button>
    </div>

    <div v-if="store.loading" class="text-center text-gray-500 py-4">Loading...</div>
    <div v-if="store.error" class="text-center text-red-400 py-4">{{ store.error }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useEventsStore } from '../stores/events'
import type { EventsFilter } from '../api/events'
import FilterBar from '../components/FilterBar.vue'
import EventsTable from '../components/EventsTable.vue'

const store = useEventsStore()

const page = ref(0)
const perPage = 50
const activeFilter = ref<EventsFilter>({})

const totalPages = computed(() => Math.max(1, Math.ceil(store.total / perPage)))

function load() {
  store.load({ ...activeFilter.value, limit: perPage, offset: page.value * perPage })
}

function onFilter(f: EventsFilter) {
  activeFilter.value = f
  page.value = 0
  load()
}

function prevPage() {
  if (page.value > 0) { page.value--; load() }
}

function nextPage() {
  if (page.value < totalPages.value - 1) { page.value++; load() }
}

onMounted(load)
</script>
