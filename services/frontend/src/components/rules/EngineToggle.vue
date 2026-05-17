<template>
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-5">
    <h2 class="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">WAF Engine Mode</h2>
    <div class="flex gap-3">
      <button
        v-for="mode in modes"
        :key="mode.value"
        @click="$emit('update:modelValue', mode.value)"
        :class="[
          'px-5 py-2 rounded-lg font-semibold text-sm transition-all border',
          modelValue === mode.value ? mode.activeClass : 'bg-gray-800 text-gray-400 border-gray-700 hover:bg-gray-700'
        ]"
      >
        {{ mode.label }}
      </button>
    </div>
    <p class="mt-3 text-xs text-gray-500">{{ currentDescription }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ modelValue: string }>()
defineEmits<{ (e: 'update:modelValue', v: string): void }>()

const modes = [
  { value: 'On', label: 'On', activeClass: 'bg-green-600 text-white border-green-500', desc: 'Full blocking mode — requests matching rules will be blocked.' },
  { value: 'DetectionOnly', label: 'Detection Only', activeClass: 'bg-yellow-500 text-gray-900 border-yellow-400', desc: 'Detection mode — violations are logged but not blocked.' },
  { value: 'Off', label: 'Off', activeClass: 'bg-red-600 text-white border-red-500', desc: 'WAF disabled — no inspection or blocking.' },
]

const currentDescription = computed(() => modes.find(m => m.value === props.modelValue)?.desc ?? '')
</script>
