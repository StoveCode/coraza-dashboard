<template>
  <div class="card p-4">
    <h3 class="text-sm font-semibold text-gray-300 mb-3">{{ title }}</h3>
    <Line :data="chartData" :options="chartOptions" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler)

const props = defineProps<{
  title: string
  data: Array<{ hour: string; count: number }>
}>()

const chartData = computed(() => ({
  labels: props.data.map(d => {
    const date = new Date(d.hour)
    return date.toLocaleTimeString('de-DE', { hour: '2-digit', minute: '2-digit' })
  }),
  datasets: [
    {
      label: 'Events',
      data: props.data.map(d => d.count),
      borderColor: '#818cf8',
      backgroundColor: 'rgba(129, 140, 248, 0.1)',
      borderWidth: 2,
      pointRadius: 3,
      pointBackgroundColor: '#818cf8',
      fill: true,
      tension: 0.3,
    },
  ],
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: true,
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: '#1c1c28',
      borderColor: '#3a3a50',
      borderWidth: 1,
      titleColor: '#c0c0d8',
      bodyColor: '#a0a0c0',
    },
  },
  scales: {
    x: {
      grid: { color: '#22222f' },
      ticks: { color: '#5a5a78', maxTicksLimit: 12 },
    },
    y: {
      grid: { color: '#22222f' },
      ticks: { color: '#5a5a78', precision: 0 },
      beginAtZero: true,
    },
  },
}
</script>
