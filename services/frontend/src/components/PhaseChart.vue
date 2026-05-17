<template>
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-4">
    <h3 class="text-sm font-semibold text-gray-400 mb-3">Phase Distribution</h3>
    <Doughnut v-if="chartData" :data="chartData" :options="options" />
    <div v-else class="h-48 flex items-center justify-center text-gray-600">No data</div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Doughnut } from 'vue-chartjs'
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import type { TopEntry } from '../api/stats'

ChartJS.register(ArcElement, Tooltip, Legend)

const props = defineProps<{ data: TopEntry[] | null | undefined }>()

const COLORS = ['#ef4444', '#f97316', '#eab308', '#22c55e', '#3b82f6', '#8b5cf6']

const chartData = computed(() => {
  if (!props.data?.length) return null
  return {
    labels: props.data.map(e => e.label),
    datasets: [{
      data: props.data.map(e => e.count),
      backgroundColor: props.data.map((_, i) => COLORS[i % COLORS.length]),
      borderWidth: 0,
    }]
  }
})

const options = {
  responsive: true,
  plugins: {
    legend: {
      display: true,
      labels: { color: '#9ca3af', boxWidth: 12, font: { size: 11 } }
    }
  }
}
</script>
