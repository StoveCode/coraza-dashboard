<template>
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-5">
    <h2 class="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">Anomaly Score Thresholds</h2>
    <div class="space-y-4">
      <div>
        <label class="block text-xs text-gray-400 mb-1">Inbound Threshold</label>
        <input
          type="number"
          :value="inbound"
          @input="$emit('update:inbound', Number(($event.target as HTMLInputElement).value))"
          min="1"
          class="w-24 bg-gray-800 border border-gray-700 rounded-lg px-3 py-1.5 text-gray-100 text-sm focus:outline-none focus:border-blue-500"
        />
        <span class="ml-2 text-xs text-gray-500">default: 5</span>
      </div>
      <div>
        <label class="block text-xs text-gray-400 mb-1">Outbound Threshold</label>
        <input
          type="number"
          :value="outbound"
          @input="$emit('update:outbound', Number(($event.target as HTMLInputElement).value))"
          min="1"
          class="w-24 bg-gray-800 border border-gray-700 rounded-lg px-3 py-1.5 text-gray-100 text-sm focus:outline-none focus:border-blue-500"
        />
        <span class="ml-2 text-xs text-gray-500">default: 4</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ inbound: number; outbound: number }>()
defineEmits<{
  (e: 'update:inbound', v: number): void
  (e: 'update:outbound', v: number): void
}>()
</script>
