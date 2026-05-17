<template>
  <div class="flex flex-wrap gap-3 items-end">
    <div>
      <label class="block text-xs text-gray-400 mb-1">Action</label>
      <select v-model="localDisruptive" class="bg-gray-800 border border-gray-700 rounded px-2 py-1 text-sm text-gray-100">
        <option value="">All</option>
        <option value="true">Blocked</option>
        <option value="false">Detected</option>
      </select>
    </div>
    <div>
      <label class="block text-xs text-gray-400 mb-1">Client IP</label>
      <input v-model="localIP" type="text" placeholder="1.2.3.4"
        class="bg-gray-800 border border-gray-700 rounded px-2 py-1 text-sm text-gray-100 w-36" />
    </div>
    <div>
      <label class="block text-xs text-gray-400 mb-1">From</label>
      <input v-model="localFrom" type="datetime-local"
        class="bg-gray-800 border border-gray-700 rounded px-2 py-1 text-sm text-gray-100" />
    </div>
    <div>
      <label class="block text-xs text-gray-400 mb-1">To</label>
      <input v-model="localTo" type="datetime-local"
        class="bg-gray-800 border border-gray-700 rounded px-2 py-1 text-sm text-gray-100" />
    </div>
    <button @click="emit('apply', buildFilter())"
      class="bg-red-600 hover:bg-red-700 text-white px-4 py-1.5 rounded text-sm transition-colors">
      Apply
    </button>
    <button @click="reset" class="bg-gray-700 hover:bg-gray-600 text-gray-300 px-4 py-1.5 rounded text-sm transition-colors">
      Reset
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { EventsFilter } from '../api/events'

const emit = defineEmits<{ apply: [filter: EventsFilter] }>()

const localDisruptive = ref('')
const localIP = ref('')
const localFrom = ref('')
const localTo = ref('')

function buildFilter(): EventsFilter {
  const f: EventsFilter = {}
  if (localDisruptive.value !== '') f.disruptive = localDisruptive.value === 'true'
  if (localIP.value) f.client_ip = localIP.value
  if (localFrom.value) f.from = new Date(localFrom.value).toISOString()
  if (localTo.value) f.to = new Date(localTo.value).toISOString()
  return f
}

function reset() {
  localDisruptive.value = ''
  localIP.value = ''
  localFrom.value = ''
  localTo.value = ''
  emit('apply', {})
}
</script>
