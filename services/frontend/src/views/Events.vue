<template>
  <div>
    <!-- Header -->
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-lg font-bold text-white">WAF Events</h1>
      <span v-if="eventsStore.loading" class="text-xs text-gray-400 flex items-center gap-1">
        <svg class="w-3 h-3 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
        </svg>
        Lade...
      </span>
    </div>

    <!-- Error -->
    <div v-if="eventsStore.error" class="mb-4 p-3 bg-red-900/40 border border-red-700/50 rounded text-red-300 text-sm">
      {{ eventsStore.error }}
    </div>

    <!-- Filter -->
    <div class="mb-4">
      <FilterBar v-model="filters" @apply="applyFilters" />
    </div>

    <!-- Table -->
    <EventsTable
      :events="eventsStore.events"
      :total="eventsStore.total"
      :limit="limit"
      :offset="offset"
      @page="onPage"
    />
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { useEventsStore } from '@/stores/events'
import FilterBar from '@/components/FilterBar.vue'
import EventsTable from '@/components/EventsTable.vue'

const eventsStore = useEventsStore()
const limit = ref(50)
const offset = ref(0)

const filters = reactive({
  from: '',
  to: '',
  action: '',
  client_ip: '',
})

function buildParams() {
  const params: Record<string, string | number> = {
    limit: limit.value,
    offset: offset.value,
  }
  if (filters.from) params.from = new Date(filters.from).toISOString()
  if (filters.to) params.to = new Date(filters.to).toISOString()
  if (filters.action) params.action = filters.action
  if (filters.client_ip) params.client_ip = filters.client_ip
  return params
}

function applyFilters() {
  offset.value = 0
  eventsStore.load(buildParams())
}

function onPage(newOffset: number) {
  offset.value = newOffset
  eventsStore.load(buildParams())
}

onMounted(() => {
  eventsStore.load(buildParams())
})
</script>
