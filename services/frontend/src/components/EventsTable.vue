<template>
  <div class="card overflow-hidden">
    <!-- Table -->
    <div class="overflow-x-auto">
      <table class="w-full text-xs">
        <thead>
          <tr class="bg-gray-900 border-b border-gray-700">
            <th class="px-3 py-2 text-left text-gray-400 font-semibold uppercase tracking-wider">Timestamp</th>
            <th class="px-3 py-2 text-left text-gray-400 font-semibold uppercase tracking-wider">IP</th>
            <th class="px-3 py-2 text-left text-gray-400 font-semibold uppercase tracking-wider">Method</th>
            <th class="px-3 py-2 text-left text-gray-400 font-semibold uppercase tracking-wider">URI</th>
            <th class="px-3 py-2 text-left text-gray-400 font-semibold uppercase tracking-wider">Rule ID</th>
            <th class="px-3 py-2 text-left text-gray-400 font-semibold uppercase tracking-wider">Rule Msg</th>
            <th class="px-3 py-2 text-left text-gray-400 font-semibold uppercase tracking-wider">Severity</th>
            <th class="px-3 py-2 text-left text-gray-400 font-semibold uppercase tracking-wider">Action</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="event in events" :key="event.id">
            <tr
              class="border-b border-gray-800 hover:bg-gray-750 cursor-pointer transition-colors"
              @click="toggleRow(event.id)"
            >
              <td class="px-3 py-2 text-gray-400 whitespace-nowrap">{{ formatTs(event.timestamp) }}</td>
              <td class="px-3 py-2 text-gray-200 font-mono whitespace-nowrap">{{ event.client_ip }}</td>
              <td class="px-3 py-2">
                <span class="font-mono font-semibold" :class="methodColor(event.method)">{{ event.method }}</span>
              </td>
              <td class="px-3 py-2 text-gray-300 max-w-xs truncate font-mono" :title="event.uri">{{ event.uri }}</td>
              <td class="px-3 py-2 text-indigo-300 font-mono whitespace-nowrap">{{ event.rule_id }}</td>
              <td class="px-3 py-2 text-gray-400 max-w-xs truncate" :title="event.rule_msg">{{ event.rule_msg }}</td>
              <td class="px-3 py-2 text-gray-400 whitespace-nowrap">{{ event.severity }}</td>
              <td class="px-3 py-2 whitespace-nowrap">
                <span :class="event.action === 'block' ? 'badge-block' : 'badge-detect'">
                  {{ event.action }}
                </span>
              </td>
            </tr>
            <!-- Raw Log Accordion -->
            <tr v-if="expandedRows.has(event.id)" class="border-b border-gray-800 bg-gray-950">
              <td colspan="8" class="px-4 py-3">
                <div class="text-xs text-gray-400 mb-1">Raw Log</div>
                <pre class="text-xs text-green-300 bg-gray-900 rounded p-3 overflow-x-auto border border-gray-700">{{ JSON.stringify(event.raw_log, null, 2) }}</pre>
              </td>
            </tr>
          </template>
          <tr v-if="!events.length">
            <td colspan="8" class="px-4 py-8 text-center text-gray-500">Keine Events gefunden.</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div class="flex items-center justify-between px-4 py-3 border-t border-gray-700 bg-gray-900">
      <span class="text-xs text-gray-400">
        {{ offset + 1 }}–{{ Math.min(offset + limit, total) }} von {{ total }} Events
      </span>
      <div class="flex gap-2">
        <button
          class="btn-ghost text-xs"
          :disabled="offset === 0"
          :class="{ 'opacity-40 cursor-not-allowed': offset === 0 }"
          @click="$emit('page', offset - limit)"
        >← Zurück</button>
        <button
          class="btn-ghost text-xs"
          :disabled="offset + limit >= total"
          :class="{ 'opacity-40 cursor-not-allowed': offset + limit >= total }"
          @click="$emit('page', offset + limit)"
        >Weiter →</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { WAFEvent } from '@/api/events'

const props = defineProps<{
  events: WAFEvent[]
  total: number
  limit: number
  offset: number
}>()

defineEmits<{ page: [offset: number] }>()

const expandedRows = ref(new Set<string>())

function toggleRow(id: string) {
  if (expandedRows.value.has(id)) {
    expandedRows.value.delete(id)
  } else {
    expandedRows.value.add(id)
  }
}

function formatTs(ts: string): string {
  return new Date(ts).toLocaleString('de-DE', {
    year: '2-digit',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

function methodColor(method: string): string {
  switch (method.toUpperCase()) {
    case 'GET': return 'text-green-400'
    case 'POST': return 'text-blue-400'
    case 'PUT': return 'text-yellow-400'
    case 'DELETE': return 'text-red-400'
    default: return 'text-gray-300'
  }
}
</script>
