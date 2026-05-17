import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchStats, type Stats } from '../api/stats'

export const useStatsStore = defineStore('stats', () => {
  const stats = ref<Stats | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function load() {
    loading.value = true
    error.value = null
    try {
      stats.value = await fetchStats()
    } catch (e: any) {
      error.value = e.message ?? 'Failed to load stats'
    } finally {
      loading.value = false
    }
  }

  return { stats, loading, error, load }
})
