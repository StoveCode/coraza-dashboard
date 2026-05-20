<template>
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-5">
    <!-- Header -->
    <div class="flex items-center justify-between mb-3">
      <h2 class="text-sm font-semibold text-gray-400 uppercase tracking-wider">Anomaly Score Thresholds</h2>
      <button @click="infoOpen = !infoOpen" class="text-gray-500 hover:text-gray-300 transition-colors" title="What is anomaly scoring?">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
      </button>
    </div>

    <!-- Collapsible info box -->
    <div v-show="infoOpen" class="mb-4 bg-gray-800/60 border border-gray-700 rounded-lg p-3 text-xs text-gray-400 space-y-1">
      <p>Das Anomaly Scoring sammelt Punkte pro getriggerter Rule. Erst wenn die Gesamtpunktzahl den Threshold überschreitet, wird blockiert.</p>
      <div class="pt-1 space-y-0.5">
        <p>• <span class="text-gray-300">Inbound:</span> Prüft eingehende Requests (SQL Injection, XSS, etc.)</p>
        <p>• <span class="text-gray-300">Outbound:</span> Prüft Server-Antworten auf Data Leakage</p>
        <p>• <span class="text-gray-300">Typische Rule-Scores:</span> Critical=5, Error=4, Warning=3, Notice=2</p>
        <p>• <span class="text-gray-300">Niedriger Threshold = aggressiver</span> (weniger Punkte nötig zum Blocken)</p>
      </div>
    </div>

    <div class="space-y-6">
      <!-- Inbound -->
      <div>
        <div class="flex items-center justify-between mb-2">
          <label class="text-xs text-gray-400">Inbound Threshold</label>
          <span :class="['text-xs font-medium px-2 py-0.5 rounded', inboundIndicator.cls]">{{ inboundIndicator.label }}</span>
        </div>

        <!-- Visual Score Bar -->
        <div class="relative h-8 mb-1">
          <div class="absolute inset-x-0 top-3 h-1.5 bg-gray-700 rounded-full"></div>
          <!-- Score badges -->
          <template v-for="score in scoreBadges" :key="score.name">
            <div
              class="absolute -translate-x-1/2 flex flex-col items-center"
              :style="{ left: scoreToPercent(score.value) + '%' }"
            >
              <div :class="['text-xs px-1 py-0.5 rounded font-mono leading-none', score.cls]" style="margin-top: 0px; font-size: 9px;">
                {{ score.name }}={{ score.value }}
              </div>
              <div class="w-px h-2 bg-gray-600 mt-0.5"></div>
            </div>
          </template>
          <!-- Threshold marker -->
          <div
            class="absolute -translate-x-1/2 flex flex-col items-center pointer-events-none"
            :style="{ left: scoreToPercent(inbound) + '%' }"
          >
            <div class="text-blue-400 font-bold text-base leading-none" style="margin-top: -2px">▼</div>
          </div>
        </div>

        <!-- Examples -->
        <div class="text-xs text-gray-500 mb-2 min-h-[1.2rem]">
          <span v-if="inboundExamples.length">blocks: {{ inboundExamples.join(' · ') }}</span>
          <span v-else class="text-red-400">blocks everything (threshold too low)</span>
        </div>

        <div class="flex items-center gap-3">
          <input
            type="range"
            min="1"
            max="20"
            :value="inbound"
            @input="$emit('update:inbound', Number(($event.target as HTMLInputElement).value))"
            class="flex-1 accent-blue-500"
          />
          <input
            type="number"
            :value="inbound"
            @input="$emit('update:inbound', Number(($event.target as HTMLInputElement).value))"
            min="1"
            max="20"
            class="w-16 bg-gray-800 border border-gray-700 rounded-lg px-2 py-1 text-gray-100 text-sm focus:outline-none focus:border-blue-500 text-center"
          />
        </div>
        <div class="flex justify-between text-xs text-gray-600 mt-1 px-0.5">
          <span>1</span><span>5</span><span>10</span><span>15</span><span>20</span>
        </div>
      </div>

      <!-- Response Check Toggle -->
      <div class="flex items-center justify-between bg-gray-800/50 rounded-lg px-3 py-2">
        <div>
          <span class="text-xs font-medium text-gray-300">Response Check</span>
          <p class="text-xs text-gray-500 mt-0.5">Enable outbound response scanning</p>
        </div>
        <button
          @click="$emit('update:responseCheck', !responseCheck)"
          :class="[
            'relative inline-flex h-5 w-9 items-center rounded-full transition-colors focus:outline-none',
            responseCheck ? 'bg-blue-600' : 'bg-gray-600'
          ]"
        >
          <span
            :class="[
              'inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform',
              responseCheck ? 'translate-x-4' : 'translate-x-1'
            ]"
          />
        </button>
      </div>

      <!-- Outbound -->
      <div :class="{ 'opacity-50': !responseCheck, 'relative': !responseCheck }">
        <div class="flex items-center justify-between mb-2">
          <label class="text-xs text-gray-400">Outbound Threshold</label>
          <span :class="['text-xs font-medium px-2 py-0.5 rounded', outboundIndicator.cls]">{{ outboundIndicator.label }}</span>
        </div>

        <!-- Disabled overlay -->
        <div v-if="!responseCheck" class="absolute inset-0 z-10 flex items-center justify-center bg-gray-900/60 rounded-lg">
          <span class="text-xs text-gray-400 bg-gray-800 px-3 py-1.5 rounded border border-gray-700">
            Response Check disabled — outbound threshold has no effect
          </span>
        </div>

        <!-- Visual Score Bar -->
        <div class="relative h-8 mb-1">
          <div class="absolute inset-x-0 top-3 h-1.5 bg-gray-700 rounded-full"></div>
          <template v-for="score in scoreBadges" :key="score.name">
            <div
              class="absolute -translate-x-1/2 flex flex-col items-center"
              :style="{ left: scoreToPercent(score.value) + '%' }"
            >
              <div :class="['text-xs px-1 py-0.5 rounded font-mono leading-none', score.cls]" style="font-size: 9px;">
                {{ score.name }}={{ score.value }}
              </div>
              <div class="w-px h-2 bg-gray-600 mt-0.5"></div>
            </div>
          </template>
          <div
            class="absolute -translate-x-1/2 flex flex-col items-center pointer-events-none"
            :style="{ left: scoreToPercent(outbound) + '%' }"
          >
            <div class="text-blue-400 font-bold text-base leading-none" style="margin-top: -2px">▼</div>
          </div>
        </div>

        <div class="text-xs text-gray-500 mb-2 min-h-[1.2rem]">
          <span v-if="outboundExamples.length">blocks: {{ outboundExamples.join(' · ') }}</span>
          <span v-else class="text-red-400">blocks everything (threshold too low)</span>
        </div>

        <div class="flex items-center gap-3">
          <input
            type="range"
            min="1"
            max="20"
            :value="outbound"
            :disabled="!responseCheck"
            @input="$emit('update:outbound', Number(($event.target as HTMLInputElement).value))"
            class="flex-1 accent-blue-500"
          />
          <input
            type="number"
            :value="outbound"
            :disabled="!responseCheck"
            @input="$emit('update:outbound', Number(($event.target as HTMLInputElement).value))"
            min="1"
            max="20"
            class="w-16 bg-gray-800 border border-gray-700 rounded-lg px-2 py-1 text-gray-100 text-sm focus:outline-none focus:border-blue-500 text-center"
          />
        </div>
        <div class="flex justify-between text-xs text-gray-600 mt-1 px-0.5">
          <span>1</span><span>5</span><span>10</span><span>15</span><span>20</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps<{
  inbound: number
  outbound: number
  responseCheck: boolean
}>()
defineEmits<{
  (e: 'update:inbound', v: number): void
  (e: 'update:outbound', v: number): void
  (e: 'update:responseCheck', v: boolean): void
}>()

const infoOpen = ref(false)

const scoreBadges = [
  { name: 'NOTICE', value: 2, cls: 'bg-gray-700 text-gray-300' },
  { name: 'WARN', value: 3, cls: 'bg-yellow-900/60 text-yellow-300' },
  { name: 'ERROR', value: 4, cls: 'bg-orange-900/60 text-orange-300' },
  { name: 'CRIT', value: 5, cls: 'bg-red-900/60 text-red-300' },
]

const SCORE_VALUES: { name: string; value: number }[] = [
  { name: 'CRITICAL', value: 5 },
  { name: 'ERROR', value: 4 },
  { name: 'WARNING', value: 3 },
  { name: 'NOTICE', value: 2 },
]

function scoreToPercent(v: number): number {
  return Math.min(100, Math.max(0, ((v - 1) / 19) * 100))
}

function thresholdIndicator(value: number) {
  if (value <= 4) return { label: 'Very Aggressive — High false positive risk', cls: 'bg-red-900/50 text-red-300' }
  if (value === 5) return { label: 'Recommended (CRS default)', cls: 'bg-green-900/50 text-green-300' }
  if (value <= 9) return { label: 'Relaxed — Some attacks may bypass', cls: 'bg-yellow-900/50 text-yellow-300' }
  return { label: 'Permissive — Tuning mode only', cls: 'bg-orange-900/50 text-orange-300' }
}

const inboundIndicator = computed(() => thresholdIndicator(props.inbound))
const outboundIndicator = computed(() => thresholdIndicator(props.outbound))

function getBlockExamples(threshold: number): string[] {
  // Find the smallest combinations of score values that sum to >= threshold
  type Combo = { parts: string[]; total: number }
  const combos: Combo[] = []

  function generate(remaining: number, idx: number, parts: string[], total: number) {
    if (total >= threshold) {
      combos.push({ parts: [...parts], total })
      return
    }
    if (idx >= SCORE_VALUES.length) return
    for (let i = idx; i < SCORE_VALUES.length; i++) {
      const s = SCORE_VALUES[i]
      generate(remaining - s.value, i, [...parts, s.name], total + s.value)
    }
  }

  generate(threshold, 0, [], 0)

  // Sort by fewest parts, then by total score
  combos.sort((a, b) => a.parts.length - b.parts.length || a.total - b.total)

  // Deduplicate by description
  const seen = new Set<string>()
  const results: string[] = []
  for (const c of combos) {
    if (results.length >= 3) break
    const key = c.parts.sort().join('+')
    if (!seen.has(key)) {
      seen.add(key)
      const desc = c.parts.length === 1
        ? `1× ${c.parts[0]} (${c.total} pts)`
        : `${c.parts.map(p => `${p}`).join('+')} (${c.total} pts)`
      results.push(desc)
    }
  }
  return results
}

const inboundExamples = computed(() => getBlockExamples(props.inbound))
const outboundExamples = computed(() => getBlockExamples(props.outbound))
</script>
