<template>
  <div class="p-6 space-y-6">
    <!-- Stat Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <StatCard label="Total Blocks" :value="store.stats?.total_blocks ?? 0" value-class="text-red-400" />
      <StatCard label="Total Detections" :value="store.stats?.total_detections ?? 0" value-class="text-yellow-400" />
      <StatCard label="Unique IPs" :value="uniqueIPs" value-class="text-blue-400" />
      <StatCard label="Top Rule" :value="topRule" value-class="text-orange-400" />
    </div>

    <!-- Timeline (full width) -->
    <ChartCard title="Events / Hour (last 24h)">
      <template #default="{ expanded }">
        <TimelineChart :data="store.stats?.events_per_hour" :expanded="expanded" />
      </template>
    </ChartCard>

    <!-- 2-column charts -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <ChartCard title="Top Client IPs">
        <template #default="{ expanded }">
          <TopIPsChart :data="store.stats?.top_ips" :expanded="expanded" />
        </template>
      </ChartCard>
      <ChartCard title="Top Rules">
        <template #default="{ expanded }">
          <TopRulesChart :data="store.stats?.top_rules" :expanded="expanded" />
        </template>
      </ChartCard>
      <ChartCard title="Top OWASP CRS Tags">
        <template #default="{ expanded }">
          <TopTagsChart :data="store.stats?.top_tags" :expanded="expanded" />
        </template>
      </ChartCard>
      <ChartCard title="Phase Distribution">
        <template #default="{ expanded }">
          <PhaseChart :data="store.stats?.top_phases" :expanded="expanded" />
        </template>
      </ChartCard>
    </div>

    <div v-if="store.loading" class="text-center text-gray-500 py-4">Loading...</div>
    <div v-if="store.error" class="text-center text-red-400 py-4">{{ store.error }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useStatsStore } from '../stores/stats'
import StatCard from '../components/StatCard.vue'
import ChartCard from '../components/ChartCard.vue'
import TimelineChart from '../components/TimelineChart.vue'
import TopIPsChart from '../components/TopIPsChart.vue'
import TopRulesChart from '../components/TopRulesChart.vue'
import TopTagsChart from '../components/TopTagsChart.vue'
import PhaseChart from '../components/PhaseChart.vue'

const store = useStatsStore()

const uniqueIPs = computed(() => store.stats?.top_ips?.length ?? 0)
const topRule = computed(() => {
  const r = store.stats?.top_rules?.[0]
  return r ? String(r.rule_id) : '-'
})

onMounted(() => { store.load(); refreshTimer = setInterval(() => store.load(), 30000) })
let refreshTimer: ReturnType<typeof setInterval>
onUnmounted(() => clearInterval(refreshTimer))
</script>
