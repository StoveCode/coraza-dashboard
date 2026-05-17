<template>
  <div class="p-2 bg-gray-900 rounded-xl border border-gray-800">
    <p class="text-xs text-gray-500 uppercase tracking-wider">{{ label }}</p>
    <p class="text-2xl font-bold" :class="valueClass">{{ value }}</p>
    <p v-if="sub" class="text-xs text-gray-500 mt-1">{{ sub }}</p>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  label: string
  value: string | number
  sub?: string
  valueClass?: string
}>()
</script>
