<template>
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-5">
    <h2 class="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">Paranoia Level</h2>
    <div class="flex gap-2 mb-4">
      <button
        v-for="lvl in [1,2,3,4]"
        :key="lvl"
        @click="$emit('update:modelValue', lvl)"
        :class="[
          'w-10 h-10 rounded-lg font-bold text-sm transition-all border',
          modelValue === lvl ? 'bg-blue-600 text-white border-blue-500' : 'bg-gray-800 text-gray-400 border-gray-700 hover:bg-gray-700'
        ]"
      >{{ lvl }}</button>
    </div>
    <div class="flex justify-between text-xs text-gray-500 mb-1">
      <span>Conservative</span>
      <span>Strict</span>
    </div>
    <div class="w-full h-1.5 bg-gray-700 rounded-full">
      <div class="h-1.5 bg-blue-500 rounded-full transition-all" :style="{ width: ((modelValue - 1) / 3 * 100) + '%' }"></div>
    </div>
    <p class="mt-3 text-xs text-gray-500">{{ descriptions[modelValue - 1] }}</p>
  </div>
</template>

<script setup lang="ts">
defineProps<{ modelValue: number }>()
defineEmits<{ (e: 'update:modelValue', v: number): void }>()

const descriptions = [
  'PL1 — Basic rules, low false positives. Recommended for most sites.',
  'PL2 — Adds more rules. May need tuning for some applications.',
  'PL3 — Strict rules. Expect false positives without tuning.',
  'PL4 — Maximum security. Significant tuning required.',
]
</script>
