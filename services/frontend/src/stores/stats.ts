import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchStats, type StatsResponse } from '@/api/stats'

export const useStatsStore = defineStore('stats', () => {
  const stats = ref<StatsResponse | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function load() {
    loading.value = true
    error.value = null
    try {
      stats.value = await fetchStats()
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch stats'
    } finally {
      loading.value = false
    }
  }

  return { stats, loading, error, load }
})
