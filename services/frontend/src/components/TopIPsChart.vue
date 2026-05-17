<template>
  <div class="card p-4">
    <h3 class="text-sm font-semibold text-gray-300 mb-3">{{ title }}</h3>
    <Bar :data="chartData" :options="chartOptions" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Bar } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js'

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend)

const props = defineProps<{
  title: string
  data: Array<{ ip: string; count: number }>
}>()

const chartData = computed(() => ({
  labels: props.data.map(d => d.ip),
  datasets: [
    {
      label: 'Requests',
      data: props.data.map(d => d.count),
      backgroundColor: 'rgba(239, 68, 68, 0.7)',
      borderColor: '#ef4444',
      borderWidth: 1,
      borderRadius: 3,
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
      ticks: { color: '#5a5a78' },
    },
    y: {
      grid: { color: '#22222f' },
      ticks: { color: '#5a5a78', precision: 0 },
      beginAtZero: true,
    },
  },
}
</script>
