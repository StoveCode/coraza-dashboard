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
  data: Array<{ rule_id: string; rule_msg: string; count: number }>
}>()

const chartData = computed(() => ({
  labels: props.data.map(d => `${d.rule_id}: ${d.rule_msg.slice(0, 30)}...`),
  datasets: [
    {
      label: 'Hits',
      data: props.data.map(d => d.count),
      backgroundColor: 'rgba(234, 179, 8, 0.7)',
      borderColor: '#eab308',
      borderWidth: 1,
      borderRadius: 3,
    },
  ],
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: true,
  indexAxis: 'y' as const,
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: '#1c1c28',
      borderColor: '#3a3a50',
      borderWidth: 1,
      titleColor: '#c0c0d8',
      bodyColor: '#a0a0c0',
      callbacks: {
        title: (items: { dataIndex: number }[]) => {
          const d = props.data[items[0].dataIndex]
          return `Rule ${d.rule_id}`
        },
        label: (item: { raw: unknown }) => ` Hits: ${item.raw}`,
        afterLabel: (item: { dataIndex: number }) => {
          const d = props.data[item.dataIndex]
          return d.rule_msg
        },
      },
    },
  },
  scales: {
    x: {
      grid: { color: '#22222f' },
      ticks: { color: '#5a5a78', precision: 0 },
      beginAtZero: true,
    },
    y: {
      grid: { color: '#22222f' },
      ticks: { color: '#5a5a78', font: { size: 11 } },
    },
  },
}
</script>
