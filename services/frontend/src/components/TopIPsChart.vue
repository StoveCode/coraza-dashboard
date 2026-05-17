<template>
  <div :class="expanded ? 'h-[500px]' : 'h-40'">
    <Bar v-if="chartData" :data="chartData" :options="options" />
    <div v-else class="h-full flex items-center justify-center text-gray-600">No data</div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Bar } from 'vue-chartjs'
import { Chart as ChartJS, CategoryScale, LinearScale, BarElement, Tooltip } from 'chart.js'
import type { TopEntry } from '../api/stats'

ChartJS.register(CategoryScale, LinearScale, BarElement, Tooltip)

const props = defineProps<{
  data: TopEntry[] | null | undefined
  expanded?: boolean
}>()

const chartData = computed(() => {
  if (!props.data?.length) return null
  return {
    labels: props.data.map(e => e.label),
    datasets: [{
      label: 'Requests',
      data: props.data.map(e => e.count),
      backgroundColor: '#3b82f6',
    }]
  }
})

const options = {
  responsive: true,
  maintainAspectRatio: false,
  indexAxis: 'y' as const,
  plugins: { legend: { display: false } },
  scales: {
    x: { ticks: { color: '#6b7280' }, grid: { color: '#1f2937' }, beginAtZero: true },
    y: { ticks: { color: '#9ca3af', font: { family: 'monospace' } }, grid: { display: false } },
  }
}
</script>
