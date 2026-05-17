<template>
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-4">
    <h3 class="text-sm font-semibold text-gray-400 mb-3">Events / Hour (last 24h)</h3>
    <Line v-if="chartData" :data="chartData" :options="options" />
    <div v-else class="h-48 flex items-center justify-center text-gray-600">No data</div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Line } from 'vue-chartjs'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, Filler, Tooltip } from 'chart.js'
import type { HourBucket } from '../api/stats'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Filler, Tooltip)

const props = defineProps<{ data: HourBucket[] | null | undefined }>()

const chartData = computed(() => {
  if (!props.data?.length) return null
  return {
    labels: props.data.map(b => new Date(b.hour).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })),
    datasets: [{
      label: 'Events',
      data: props.data.map(b => b.count),
      borderColor: '#f87171',
      backgroundColor: 'rgba(248,113,113,0.15)',
      fill: true,
      tension: 0.3,
    }]
  }
})

const options = {
  responsive: true,
  plugins: { legend: { display: false } },
  scales: {
    x: { ticks: { color: '#6b7280' }, grid: { color: '#1f2937' } },
    y: { ticks: { color: '#6b7280' }, grid: { color: '#1f2937' }, beginAtZero: true },
  }
}
</script>
