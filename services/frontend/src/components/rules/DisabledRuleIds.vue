<template>
  <div class="bg-gray-900/80 rounded-2xl p-6">
    <h2 class="text-xs font-semibold text-gray-500 uppercase tracking-widest mb-4">Disable Individual Rules</h2>
    <div class="flex gap-2 mb-3">
      <input
        v-model="newId"
        @keyup.enter="add"
        type="text"
        placeholder="Rule ID e.g. 920350"
        class="bg-gray-800 border border-gray-700 rounded-lg px-3 py-1.5 text-gray-100 text-sm focus:outline-none focus:border-blue-500 w-52"
        :class="{ 'border-red-500': validationError }"
      />
      <button
        @click="add"
        class="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded-lg transition-colors"
      >+ Add</button>
    </div>
    <p v-if="validationError" class="text-xs text-red-400 mb-2">{{ validationError }}</p>
    <div v-if="ruleIds.length > 0" class="flex flex-wrap gap-1.5">
      <span
        v-for="id in ruleIds"
        :key="id"
        class="flex items-center gap-1 bg-gray-700 text-gray-200 text-xs px-2 py-0.5 rounded-full"
      >
        {{ id }}
        <button @click="$emit('remove', id)" class="text-gray-400 hover:text-red-400 transition-colors text-xs leading-none">×</button>
      </span>
    </div>
    <p v-else class="text-xs text-gray-500">No individual rules disabled.</p>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

defineProps<{ ruleIds: string[] }>()
const emit = defineEmits<{
  (e: 'add', id: string): void
  (e: 'remove', id: string): void
}>()

const newId = ref('')
const validationError = ref('')

function add() {
  const trimmed = newId.value.trim()
  if (!trimmed) return
  if (!/^\d+$/.test(trimmed)) {
    validationError.value = 'Rule ID must be numeric'
    return
  }
  validationError.value = ''
  emit('add', trimmed)
  newId.value = ''
}
</script>
