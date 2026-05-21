<template>
  <div class="overflow-x-auto">
    <!-- Toast -->
    <div
      v-if="toastVisible"
      class="fixed bottom-6 right-6 z-50 bg-green-800 border border-green-600 text-green-200 text-sm px-4 py-2 rounded-lg shadow-lg transition-opacity"
    >
      Rule disabled
    </div>

    <table class="w-full text-sm text-left">
      <thead class="text-xs text-gray-400 uppercase border-b border-gray-800">
        <tr>
          <th class="px-3 py-2">Time</th>
          <th class="px-3 py-2">Action</th>
          <th class="px-3 py-2">Client IP</th>
          <th class="px-3 py-2">URI</th>
          <th class="px-3 py-2">Rule</th>
          <th class="px-3 py-2">Score</th>
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
              {{ e.disruptive ? 'Block' : 'Detect' }}
            </span>
          </td>
          <td class="px-3 py-2 font-mono text-blue-300">{{ e.client_ip }}</td>
          <td class="px-3 py-2 max-w-xs truncate text-gray-200" :title="e.uri">{{ e.uri }}</td>
          <td class="px-3 py-2">
            <div class="flex items-center gap-1">
              <span class="text-orange-400 font-mono text-xs">{{ e.rule_id }}</span>
              <button
                v-if="!PROTECTED_RULE_IDS.includes(e.rule_id)"
                @click="disableRule(e.rule_id)"
                class="text-gray-600 hover:text-red-400 transition-colors"
                title="Disable this rule"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="12" cy="12" r="10"/>
                  <line x1="4.93" y1="4.93" x2="19.07" y2="19.07"/>
                </svg>
              </button>
            </div>
            <span class="text-gray-300 text-xs">{{ e.rule_msg }}</span>
          </td>
          <td class="px-3 py-2">
            <span :class="['text-xs font-bold px-1.5 py-0.5 rounded', scoreBadgeClass(e.rule_id, e.severity)]">
              {{ scoreBadgeText(e.rule_id, e.severity) }}
            </span>
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
          <td colspan="10" class="px-3 py-6 text-center text-gray-500">No events found</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { WAFEvent } from '../api/events'
import { useRulesStore } from '../stores/rules'

defineProps<{ events: WAFEvent[] }>()

const rulesStore = useRulesStore()

const PROTECTED_RULE_IDS = [949110, 949111]
const toastVisible = ref(false)

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

function scoreBadgeText(ruleId: number, severity: string): string {
  if (PROTECTED_RULE_IDS.includes(ruleId)) return 'BLOCK'
  switch (severity?.toUpperCase()) {
    case 'CRITICAL': return '+5'
    case 'ERROR': return '+4'
    case 'WARNING': return '+3'
    case 'NOTICE': return '+2'
    default: return '?'
  }
}

function scoreBadgeClass(ruleId: number, severity: string): string {
  if (PROTECTED_RULE_IDS.includes(ruleId)) return 'bg-purple-900/60 text-purple-300'
  switch (severity?.toUpperCase()) {
    case 'CRITICAL': return 'bg-red-900/60 text-red-300'
    case 'ERROR': return 'bg-orange-900/60 text-orange-300'
    case 'WARNING': return 'bg-yellow-900/60 text-yellow-300'
    case 'NOTICE': return 'bg-gray-700 text-gray-300'
    default: return 'bg-gray-700 text-gray-400'
  }
}

async function disableRule(ruleId: number) {
  await rulesStore.disableRuleOnly(String(ruleId))
  toastVisible.value = true
  setTimeout(() => { toastVisible.value = false }, 3000)
}
</script>
