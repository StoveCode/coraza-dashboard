import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchEvents, type WAFEvent, type EventsParams, type EventsResponse } from '@/api/events'

export const useEventsStore = defineStore('events', () => {
  const events = ref<WAFEvent[]>([])
  const total = ref(0)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function load(params: EventsParams = {}) {
    loading.value = true
    error.value = null
    try {
      const res: EventsResponse = await fetchEvents(params)
      events.value = res.events ?? []
      total.value = res.total ?? 0
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch events'
    } finally {
      loading.value = false
    }
  }

  return { events, total, loading, error, load }
})
