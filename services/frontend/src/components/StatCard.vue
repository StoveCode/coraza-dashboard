<template>
  <div class="card p-4">
    <div class="text-xs text-gray-400 uppercase tracking-wider mb-1">{{ label }}</div>
    <div class="text-3xl font-bold" :class="valueColor">{{ displayValue }}</div>
    <div v-if="sub" class="text-xs text-gray-500 mt-1">{{ sub }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  label: string
  value: number | string
  sub?: string
  color?: 'red' | 'yellow' | 'green' | 'blue' | 'default'
}>()

const valueColor = computed(() => {
  switch (props.color) {
    case 'red': return 'text-red-400'
    case 'yellow': return 'text-yellow-400'
    case 'green': return 'text-green-400'
    case 'blue': return 'text-blue-400'
    default: return 'text-white'
  }
})

const displayValue = computed(() => {
  if (typeof props.value === 'number') {
    return props.value.toLocaleString()
  }
  return props.value
})
</script>
