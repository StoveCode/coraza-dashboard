<template>
  <div class="overflow-x-auto">
    <table class="w-full text-sm text-left">
      <thead class="text-xs text-gray-400 uppercase border-b border-gray-800">
        <tr>
          <th class="px-3 py-2">Time</th>
          <th class="px-3 py-2">Action</th>
          <th class="px-3 py-2">Client IP</th>
          <th class="px-3 py-2">URI</th>
          <th class="px-3 py-2">Rule</th>
          <th class="px-3 py-2">Severity</th>
          <th class="px-3 py-2">Phase</th>
          <th class="px-3 py-2">Tags</th>
          <th class="px-3 py-2">Data</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="e in events" :key="e.id" class="border-b border-gray-800 hover:bg-gray-900 transition-colors">
          <td class="px-3 py-2 text-gray-400 whitespace-nowrap">{{ formatTime(e.timestamp) }}</td>
          <td class="px-3 py-2">
            <span :class="e.disruptive ? 'text-red-400 font-semibold' : 'text-yellow-400'">
              {{ e.disruptive ? '🛑 Block' : '⚠️ Detect' }}
            </span>
          </td>
          <td class="px-3 py-2 font-mono text-blue-300">{{ e.client_ip }}</td>
          <td class="px-3 py-2 max-w-xs truncate text-gray-200" :title="e.uri">{{ e.uri }}</td>
          <td class="px-3 py-2">
            <span class="text-orange-400 font-mono text-xs">{{ e.rule_id }}</span>
            <span class="text-gray-300 ml-1 text-xs">{{ e.rule_msg }}</span>
          </td>
          <td class="px-3 py-2 text-xs font-semibold" :class="severityClass(e.severity)">{{ e.severity }}</td>
          <td class="px-3 py-2 text-xs text-gray-400">{{ e.phase }}</td>
          <td class="px-3 py-2">
            <span v-for="tag in (e.tags ?? []).slice(0, 3)" :key="tag"
              class="inline-block bg-gray-800 text-gray-300 rounded text-xs px-1 mr-1 mb-0.5">{{ tag }}</span>
          </td>
          <td class="px-3 py-2 max-w-xs truncate text-xs text-gray-400" :title="e.data">{{ e.data }}</td>
        </tr>
        <tr v-if="events.length === 0">
          <td colspan="9" class="px-3 py-6 text-center text-gray-500">No events found</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import type { WAFEvent } from '../api/events'

defineProps<{ events: WAFEvent[] }>()

function formatTime(ts: string): string {
  return new Date(ts).toLocaleString()
}

function severityClass(s: string): string {
  switch (s?.toUpperCase()) {
    case 'CRITICAL': return 'text-red-400'
    case 'ERROR': return 'text-orange-400'
    case 'WARNING': return 'text-yellow-400'
    default: return 'text-gray-400'
  }
}
</script>
