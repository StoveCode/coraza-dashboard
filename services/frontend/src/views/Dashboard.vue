<template>
  <div>
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-lg font-bold text-white">WAF Overview</h1>
      <div class="flex items-center gap-3 text-xs text-gray-400">
        <span v-if="statsStore.loading" class="flex items-center gap-1">
          <svg class="w-3 h-3 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
          </svg>
          Lade...
        </span>
        <span v-else>Aktualisiert: {{ lastRefresh }}</span>
        <span class="text-gray-600">Auto-Refresh 30s</span>
      </div>
    </div>

    <!-- Error -->
    <div v-if="statsStore.error" class="mb-4 p-3 bg-red-900/40 border border-red-700/50 rounded text-red-300 text-sm">
      {{ statsStore.error }}
    </div>

    <!-- Stat Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
      <StatCard
        label="Total Blocks"
        :value="stats?.total_blocks ?? 0"
        color="red"
      />
      <StatCard
        label="Total Detections"
        :value="stats?.total_detections ?? 0"
        color="yellow"
      />
      <StatCard
        label="Unique IPs"
        :value="uniqueIPs"
        color="blue"
      />
      <StatCard
        label="Most Hit Rule"
        :value="topRule?.rule_id ?? '—'"
        :sub="topRule?.rule_msg?.slice(0, 40) ?? ''"
        color="default"
      />
    </div>

    <!-- Charts grid -->
    <div class="grid grid-cols-1 xl:grid-cols-2 gap-4 mb-6">
      <TimelineChart
        title="Events per Hour (last 24h)"
        :data="stats?.events_per_hour ?? []"
      />
      <TopIPsChart
        title="Top Blocked IPs"
        :data="stats?.top_ips ?? []"
      />
    </div>

    <div class="grid grid-cols-1">
      <TopRulesChart
        title="Top WAF Rules"
        :data="stats?.top_rules ?? []"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useStatsStore } from '@/stores/stats'
import StatCard from '@/components/StatCard.vue'
import TimelineChart from '@/components/TimelineChart.vue'
import TopIPsChart from '@/components/TopIPsChart.vue'
import TopRulesChart from '@/components/TopRulesChart.vue'

const statsStore = useStatsStore()
const lastRefresh = ref('—')
let interval: ReturnType<typeof setInterval>

const stats = computed(() => statsStore.stats)
const uniqueIPs = computed(() => stats.value?.top_ips?.length ?? 0)
const topRule = computed(() => stats.value?.top_rules?.[0] ?? null)

async function refresh() {
  await statsStore.load()
  lastRefresh.value = new Date().toLocaleTimeString('de-DE')
}

onMounted(() => {
  refresh()
  interval = setInterval(refresh, 30_000)
})

onUnmounted(() => {
  clearInterval(interval)
})
</script>
