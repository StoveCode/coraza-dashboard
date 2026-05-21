<template>
  <div class="p-6 space-y-6">
    <!-- Stat Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-5 gap-4">
      <StatCard label="Inbound Blocked" :value="store.stats?.total_inbound_blocks ?? 0" value-class="text-red-400" />
      <StatCard label="Outbound Blocked" :value="store.stats?.total_outbound_blocks ?? 0" value-class="text-orange-400" />
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
      <!-- Anomaly Score Widget -->
      <div v-if="(store.stats?.max_anomaly_score ?? 0) > 0" class="bg-gray-900 border border-gray-800 rounded-xl p-4">
        <h3 class="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-3">Anomaly Score</h3>
        <div class="flex gap-6 mb-3">
          <div>
            <div class="text-xs text-gray-500 mb-0.5">Avg Score</div>
            <div class="text-2xl font-bold text-yellow-400">{{ store.stats?.avg_anomaly_score != null ? store.stats.avg_anomaly_score.toFixed(1) : '-' }}</div>
          </div>
          <div>
            <div class="text-xs text-gray-500 mb-0.5">Max Score</div>
            <div class="text-2xl font-bold text-red-400">{{ store.stats?.max_anomaly_score ?? '-' }}</div>
          </div>
        </div>
        <div class="space-y-1" v-if="store.stats?.score_distribution?.length">
          <div v-for="bucket in store.stats.score_distribution" :key="bucket.range" class="flex items-center gap-2">
            <span class="text-xs text-gray-400 w-10 shrink-0">{{ bucket.range }}</span>
            <div class="flex-1 bg-gray-800 rounded-full h-2 overflow-hidden">
              <div
                class="h-2 rounded-full bg-orange-500"
                :style="{ width: maxScoreBucketCount > 0 ? (bucket.count / maxScoreBucketCount * 100) + '%' : '0%' }"
              ></div>
            </div>
            <span class="text-xs text-gray-500 w-8 text-right">{{ bucket.count }}</span>
          </div>
        </div>
        <div v-else class="text-xs text-gray-600 py-2">No anomaly score data</div>
      </div>
      <!-- Top Attacking IPs — tabbed widget -->
      <div class="bg-gray-900 border border-gray-800 rounded-xl p-4">
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-sm font-semibold text-gray-400 uppercase tracking-wider">Top Attacking IPs</h3>
          <div class="flex gap-1">
            <button
              @click="ipTab = 'all'"
              :class="['px-2 py-1 text-xs rounded transition-colors', ipTab === 'all' ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700']"
            >All Events</button>
            <button
              @click="ipTab = 'blocked'"
              :class="['px-2 py-1 text-xs rounded transition-colors', ipTab === 'blocked' ? 'bg-red-700 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700']"
            >Blocked (Threshold)</button>
          </div>
        </div>
        <div class="space-y-1.5">
          <template v-if="activeIPList.length">
            <div
              v-for="entry in activeIPList.slice(0, 8)"
              :key="entry.label"
              class="flex items-center justify-between gap-2 group"
            >
              <span class="font-mono text-xs text-blue-300 truncate flex-1">{{ entry.label }}</span>
              <span class="text-xs text-gray-400 shrink-0">{{ entry.count }}</span>
              <button
                @click="viewEventsForIP(entry.label)"
                class="text-xs text-gray-500 hover:text-blue-400 transition-colors shrink-0 opacity-0 group-hover:opacity-100"
                :title="ipTab === 'blocked' ? 'View blocked events for this IP' : 'View events for this IP'"
              >→ Events</button>
            </div>
          </template>
          <div v-else class="text-xs text-gray-600 py-4 text-center">No data</div>
        </div>
      </div>

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
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useStatsStore } from '../stores/stats'
import StatCard from '../components/StatCard.vue'
import ChartCard from '../components/ChartCard.vue'
import TimelineChart from '../components/TimelineChart.vue'
import TopRulesChart from '../components/TopRulesChart.vue'
import TopTagsChart from '../components/TopTagsChart.vue'
import PhaseChart from '../components/PhaseChart.vue'

const store = useStatsStore()
const router = useRouter()

const ipTab = ref<'all' | 'blocked'>('all')

const activeIPList = computed(() => {
  if (ipTab.value === 'blocked') return store.stats?.top_ips_blocked ?? []
  return store.stats?.top_ips ?? []
})

const uniqueIPs = computed(() => store.stats?.top_ips?.length ?? 0)
const topRule = computed(() => {
  const r = store.stats?.top_rules?.[0]
  return r ? String(r.rule_id) : '-'
})

const maxScoreBucketCount = computed(() => {
  const dist = store.stats?.score_distribution
  if (!dist?.length) return 1
  return Math.max(...dist.map(b => b.count), 1)
})

function viewEventsForIP(ip: string) {
  router.push({ path: '/events', query: { client_ip: ip } })
}

let refreshTimer: ReturnType<typeof setInterval>
onMounted(() => { store.load(); refreshTimer = setInterval(() => store.load(), 30000) })
onUnmounted(() => clearInterval(refreshTimer))
</script>
