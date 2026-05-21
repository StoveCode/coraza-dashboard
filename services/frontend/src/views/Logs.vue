<template>
  <div class="p-6 text-gray-100">
    <h1 class="text-2xl font-bold mb-6 text-white">Logs</h1>

    <!-- Controls -->
    <div class="flex flex-wrap gap-3 mb-4 items-end">
      <!-- Service selector -->
      <div class="flex flex-col gap-1">
        <label class="text-xs text-gray-400 uppercase tracking-wide">Service</label>
        <select
          v-model="selectedService"
          @change="onServiceChange"
          class="bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm text-gray-100 focus:outline-none focus:border-red-400"
        >
          <option v-for="s in services" :key="s.value" :value="s.value">{{ s.label }}</option>
        </select>
      </div>

      <!-- Tail count -->
      <div class="flex flex-col gap-1">
        <label class="text-xs text-gray-400 uppercase tracking-wide">Tail</label>
        <select
          v-model="tailCount"
          @change="onTailChange"
          class="bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm text-gray-100 focus:outline-none focus:border-red-400"
        >
          <option :value="50">50</option>
          <option :value="100">100</option>
          <option :value="200">200</option>
          <option :value="500">500</option>
        </select>
      </div>

      <!-- Text filter -->
      <div class="flex flex-col gap-1 flex-1 min-w-48">
        <label class="text-xs text-gray-400 uppercase tracking-wide">Filter</label>
        <input
          v-model="filterText"
          type="text"
          placeholder="Search logs..."
          class="bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-red-400"
        />
      </div>

      <!-- Level filter -->
      <div class="flex flex-col gap-1">
        <label class="text-xs text-gray-400 uppercase tracking-wide">Level</label>
        <select
          v-model="levelFilter"
          class="bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm text-gray-100 focus:outline-none focus:border-red-400"
        >
          <option value="ALL">ALL</option>
          <option value="ERROR">ERROR</option>
          <option value="WARN">WARN</option>
          <option value="INFO">INFO</option>
        </select>
      </div>

      <!-- Actions -->
      <div class="flex flex-col gap-1">
        <label class="text-xs text-gray-400 uppercase tracking-wide invisible">Actions</label>
        <div class="flex gap-2">
          <button
            @click="refresh"
            :disabled="loading"
            class="bg-gray-700 hover:bg-gray-600 disabled:opacity-50 text-white px-3 py-2 rounded text-sm font-medium transition-colors"
            title="Refresh"
          >
            <span :class="loading ? 'animate-spin inline-block' : ''">↻</span>
          </button>
          <button
            @click="toggleAutoRefresh"
            :class="autoRefresh ? 'bg-red-600 hover:bg-red-700' : 'bg-gray-700 hover:bg-gray-600'"
            class="text-white px-3 py-2 rounded text-sm font-medium transition-colors"
            title="Toggle auto-refresh (10s)"
          >
            ⏱ {{ autoRefresh ? 'On' : 'Off' }}
          </button>
          <button
            @click="copyToClipboard"
            class="bg-gray-700 hover:bg-gray-600 text-white px-3 py-2 rounded text-sm font-medium transition-colors"
            title="Copy to clipboard"
          >
            📋
          </button>
        </div>
      </div>
    </div>

    <!-- Status bar -->
    <div class="flex items-center gap-3 mb-2 text-xs text-gray-500">
      <span>
        <span
          :class="containerRunning === true ? 'text-green-400' : containerRunning === false ? 'text-red-400' : 'text-gray-500'"
        >●</span>
        {{ statusLabel }}
      </span>
      <span v-if="fetchError" class="text-red-400">{{ fetchError }}</span>
      <span v-else>{{ filteredLines.length }} lines shown</span>
      <span v-if="lastRefresh" class="ml-auto">Last refresh: {{ lastRefresh }}</span>
    </div>

    <!-- Log output -->
    <div
      ref="logContainer"
      class="bg-gray-900 border border-gray-800 rounded-lg overflow-auto font-mono text-xs leading-5"
      style="max-height: 65vh; min-height: 200px;"
    >
      <div v-if="loading && filteredLines.length === 0" class="p-4 text-gray-500 text-center">
        Loading…
      </div>
      <div v-else-if="filteredLines.length === 0" class="p-4 text-gray-500 text-center">
        No log lines match the current filter.
      </div>
      <div v-else>
        <div
          v-for="(line, i) in filteredLines"
          :key="i"
          :class="lineClass(line)"
          class="px-4 py-0.5 whitespace-pre-wrap break-all hover:bg-gray-800/50"
        >
          <span class="text-gray-600 select-none mr-2">{{ String(i + 1).padStart(4, ' ') }}</span>
          <span v-html="formatLine(line)"></span>
        </div>
      </div>
    </div>

    <!-- Bottom actions -->
    <div class="flex justify-between items-center mt-2 text-xs text-gray-500">
      <span>Showing {{ filteredLines.length }} / {{ rawLines.length }} lines</span>
      <button
        @click="scrollToBottom"
        class="bg-gray-700 hover:bg-gray-600 text-white px-3 py-1.5 rounded text-xs font-medium transition-colors"
        title="Scroll to bottom"
      >
        ↓ Scroll to bottom
      </button>
    </div>

    <!-- Copy toast -->
    <div
      v-if="copyToast"
      class="fixed bottom-6 right-6 bg-green-700 text-white px-4 py-2 rounded shadow-lg text-sm transition-opacity"
    >
      Copied to clipboard!
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { fetchServiceLogs } from '../api/system'

const services = [
  { value: 'coraza-spoa', label: 'coraza-spoa (WAF)' },
  { value: 'backend', label: 'backend' },
  { value: 'haproxy', label: 'haproxy' },
  { value: 'fluentbit', label: 'fluentbit' },
  { value: 'postgres', label: 'postgres' },
]

const selectedService = ref('coraza-spoa')
const tailCount = ref(100)
const filterText = ref('')
const levelFilter = ref('ALL')
const loading = ref(false)
const autoRefresh = ref(false)
const rawLines = ref<string[]>([])
const containerRunning = ref<boolean | null>(null)
const fetchError = ref<string | null>(null)
const lastRefresh = ref('')
const copyToast = ref(false)
const logContainer = ref<HTMLElement | null>(null)

function scrollToBottom() {
  if (logContainer.value) {
    logContainer.value.scrollTop = logContainer.value.scrollHeight
  }
}

let autoRefreshTimer: ReturnType<typeof setInterval> | null = null

const statusLabel = computed(() => {
  if (containerRunning.value === true) return `${selectedService.value}: running`
  if (containerRunning.value === false) return `${selectedService.value}: stopped`
  return `${selectedService.value}: unknown`
})

const filteredLines = computed(() => {
  let lines = rawLines.value
  if (levelFilter.value !== 'ALL') {
    const lv = levelFilter.value
    lines = lines.filter(l => l.toUpperCase().includes(lv))
  }
  if (filterText.value.trim()) {
    const needle = filterText.value.toLowerCase()
    lines = lines.filter(l => l.toLowerCase().includes(needle))
  }
  return lines
})

function lineClass(line: string): string {
  const upper = line.toUpperCase()
  if (upper.includes('ERROR') || upper.includes('FATAL') || upper.includes('CRIT')) return 'text-red-400'
  if (upper.includes('WARN')) return 'text-yellow-300'
  if (upper.includes('INFO')) return 'text-blue-300'
  return 'text-gray-400'
}

// Highlight timestamps (ISO-like) with dimmer color
function formatLine(line: string): string {
  // Escape HTML first
  const escaped = line
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
  // Dim timestamps like 2026-05-21T11:00:00Z or 2026-05-21T11:00:00.000Z
  return escaped.replace(
    /(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z?)/g,
    '<span class="text-gray-500">$1</span>'
  )
}

async function refresh() {
  loading.value = true
  fetchError.value = null
  try {
    const data = await fetchServiceLogs(selectedService.value, tailCount.value)
    rawLines.value = data.lines ?? []
    containerRunning.value = data.running ?? null
    if (data.error) fetchError.value = data.error
    lastRefresh.value = new Date().toLocaleTimeString()
  } catch (e: unknown) {
    fetchError.value = e instanceof Error ? e.message : 'fetch failed'
  } finally {
    loading.value = false
  }
}

function onServiceChange() {
  rawLines.value = []
  refresh()
}

function onTailChange() {
  refresh()
}

function toggleAutoRefresh() {
  autoRefresh.value = !autoRefresh.value
  if (autoRefresh.value) {
    autoRefreshTimer = setInterval(refresh, 10000)
  } else {
    if (autoRefreshTimer) clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }
}

async function copyToClipboard() {
  const text = filteredLines.value.join('\n')
  try {
    await navigator.clipboard.writeText(text)
    copyToast.value = true
    setTimeout(() => { copyToast.value = false }, 2000)
  } catch {
    // fallback
  }
}

// Stop auto-refresh when component unmounts
onUnmounted(() => {
  if (autoRefreshTimer) clearInterval(autoRefreshTimer)
})

onMounted(() => {
  refresh()
})

// Reset filter when service changes
watch(selectedService, () => {
  filterText.value = ''
  levelFilter.value = 'ALL'
})
</script>
