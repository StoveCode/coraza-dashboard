<template>
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-4">
    <!-- Title -->
    <div class="flex items-center gap-2 mb-3">
      <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2a1 1 0 01-.293.707L13 13.414V19a1 1 0 01-.553.894l-4 2A1 1 0 017 21v-7.586L3.293 6.707A1 1 0 013 6V4z"/>
      </svg>
      <span class="text-sm font-semibold text-gray-400 uppercase tracking-wider">Filter</span>
    </div>

    <!-- Row 1: filters -->
    <div class="flex flex-wrap gap-3 items-end">
      <div>
        <label class="block text-xs text-gray-400 mb-1">Action</label>
        <select v-model="localDisruptive" class="bg-gray-800 border border-gray-700 rounded-lg px-2 py-1.5 text-sm text-gray-100 focus:outline-none focus:border-blue-500">
          <option value="">All</option>
          <option value="true">Blocked</option>
          <option value="false">Detected</option>
        </select>
      </div>
      <div>
        <label class="block text-xs text-gray-400 mb-1">Client IP</label>
        <input v-model="localIP" type="text" placeholder="1.2.3.4"
          class="bg-gray-800 border border-gray-700 rounded-lg px-2 py-1.5 text-sm text-gray-100 w-36 focus:outline-none focus:border-blue-500" />
      </div>
      <div>
        <label class="block text-xs text-gray-400 mb-1">Rule ID</label>
        <input v-model.number="localRuleId" type="number" placeholder="949110"
          class="bg-gray-800 border border-gray-700 rounded-lg px-2 py-1.5 text-sm text-gray-100 w-28 focus:outline-none focus:border-blue-500" />
      </div>
      <div>
        <label class="block text-xs text-gray-400 mb-1">Tag</label>
        <select v-model="localTag" class="bg-gray-800 border border-gray-700 rounded-lg px-2 py-1.5 text-sm text-gray-100 focus:outline-none focus:border-blue-500">
          <option value="">All</option>
          <option value="attack-sqli">attack-sqli</option>
          <option value="attack-xss">attack-xss</option>
          <option value="attack-rce">attack-rce</option>
          <option value="attack-lfi">attack-lfi</option>
          <option value="attack-rfi">attack-rfi</option>
          <option value="attack-ssrf">attack-ssrf</option>
          <option value="attack-injection">attack-injection</option>
          <option value="attack-scanner">attack-scanner</option>
          <option value="attack-protocol">attack-protocol</option>
          <option value="attack-reputation">attack-reputation</option>
        </select>
      </div>
      <div>
        <label class="block text-xs text-gray-400 mb-1">From</label>
        <input v-model="localFrom" type="datetime-local"
          class="bg-gray-800 border border-gray-700 rounded-lg px-2 py-1.5 text-sm text-gray-100 focus:outline-none focus:border-blue-500" />
      </div>
      <div>
        <label class="block text-xs text-gray-400 mb-1">To</label>
        <input v-model="localTo" type="datetime-local"
          class="bg-gray-800 border border-gray-700 rounded-lg px-2 py-1.5 text-sm text-gray-100 focus:outline-none focus:border-blue-500" />
      </div>

      <!-- Row 2: actions (right-aligned) -->
      <div class="flex gap-2 ml-auto mt-auto">
        <button @click="reset" class="bg-gray-700 hover:bg-gray-600 text-gray-300 px-4 py-1.5 rounded-lg text-sm transition-colors">
          Reset
        </button>
        <button @click="emit('apply', buildFilter())"
          class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-1.5 rounded-lg text-sm transition-colors">
          Apply
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { EventsFilter } from '../api/events'

const emit = defineEmits<{ apply: [filter: EventsFilter] }>()

const localDisruptive = ref('')
const localIP = ref('')
const localRuleId = ref<number | null>(null)
const localTag = ref('')
const localFrom = ref('')
const localTo = ref('')

function buildFilter(): EventsFilter {
  const f: EventsFilter = {}
  if (localDisruptive.value !== '') f.disruptive = localDisruptive.value === 'true'
  if (localIP.value) f.client_ip = localIP.value
  if (localRuleId.value) f.rule_id = localRuleId.value
  if (localTag.value) f.tag = localTag.value
  if (localFrom.value) f.from = new Date(localFrom.value).toISOString()
  if (localTo.value) f.to = new Date(localTo.value).toISOString()
  return f
}

function reset() {
  localDisruptive.value = ''
  localIP.value = ''
  localRuleId.value = null
  localTag.value = ''
  localFrom.value = ''
  localTo.value = ''
  emit('apply', {})
}
</script>
