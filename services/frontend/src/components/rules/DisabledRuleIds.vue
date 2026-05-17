<template>
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-5">
    <h2 class="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">Disable Individual Rules</h2>
    <div class="flex gap-2 mb-4">
      <input
        v-model="newId"
        @keyup.enter="add"
        type="text"
        placeholder="Rule ID e.g. 920350"
        class="bg-gray-800 border border-gray-700 rounded-lg px-3 py-1.5 text-gray-100 text-sm focus:outline-none focus:border-blue-500 w-52"
      />
      <button
        @click="add"
        class="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white text-sm rounded-lg transition-colors"
      >+ Add</button>
    </div>
    <div v-if="ruleIds.length > 0" class="flex flex-wrap gap-2">
      <span
        v-for="id in ruleIds"
        :key="id"
        class="flex items-center gap-1.5 bg-gray-700 text-gray-200 text-sm px-3 py-1 rounded-full"
      >
        {{ id }}
        <button @click="$emit('remove', id)" class="text-gray-400 hover:text-red-400 transition-colors text-xs">×</button>
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

function add() {
  if (newId.value.trim()) {
    emit('add', newId.value.trim())
    newId.value = ''
  }
}
</script>
