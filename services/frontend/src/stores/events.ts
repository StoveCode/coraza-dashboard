import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchEvents, type WAFEvent, type EventsFilter } from '../api/events'

export const useEventsStore = defineStore('events', () => {
  const events = ref<WAFEvent[]>([])
  const total = ref(0)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function load(filter: EventsFilter = {}) {
    loading.value = true
    error.value = null
    try {
      const result = await fetchEvents(filter)
      events.value = result.events
      total.value = result.total
    } catch (e: any) {
      error.value = e.message ?? 'Failed to load events'
    } finally {
      loading.value = false
    }
  }

  return { events, total, loading, error, load }
})
