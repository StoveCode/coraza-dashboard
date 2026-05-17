<template>
  <div class="p-6 space-y-6">
    <!-- Stat Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <StatCard label="Total Blocks" :value="stats?.total_blocks ?? 0" value-class="text-red-400" />
      <StatCard label="Total Detections" :value="stats?.total_detections ?? 0" value-class="text-yellow-400" />
      <StatCard label="Unique IPs" :value="uniqueIPs" value-class="text-blue-400" />
      <StatCard label="Top Rule" :value="topRule" value-class="text-orange-400" />
    </div>

    <!-- Timeline (full width) -->
    <TimelineChart :data="stats?.events_per_hour" />

    <!-- 2-column charts -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <TopIPsChart :data="stats?.top_ips" />
      <TopRulesChart :data="stats?.top_rules" />
      <TopTagsChart :data="stats?.top_tags" />
      <PhaseChart :data="stats?.top_phases" />
    </div>

    <div v-if="loading" class="text-center text-gray-500 py-4">Loading...</div>
    <div v-if="error" class="text-center text-red-400 py-4">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useStatsStore } from '../stores/stats'
import StatCard from '../components/StatCard.vue'
import TimelineChart from '../components/TimelineChart.vue'
import TopIPsChart from '../components/TopIPsChart.vue'
import TopRulesChart from '../components/TopRulesChart.vue'
import TopTagsChart from '../components/TopTagsChart.vue'
import PhaseChart from '../components/PhaseChart.vue'

const store = useStatsStore()
const { stats, loading, error } = store

const uniqueIPs = computed(() => store.stats?.top_ips?.length ?? 0)
const topRule = computed(() => {
  const r = store.stats?.top_rules?.[0]
  return r ? String(r.rule_id) : '-'
})

onMounted(() => store.load())
</script>
