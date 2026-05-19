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
        <p>• <span class="text-gray-300">Beispiel:</span> Ein SQLi-Versuch (Score 5) triggert bei Threshold 5 einen Block</p>
      </div>
    </div>

    <div class="space-y-5">
      <!-- Inbound -->
      <div>
        <div class="flex items-center justify-between mb-2">
          <label class="text-xs text-gray-400">Inbound Threshold</label>
          <span :class="['text-xs font-medium px-2 py-0.5 rounded', inboundIndicator.cls]">{{ inboundIndicator.label }}</span>
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

      <!-- Outbound -->
      <div>
        <div class="flex items-center justify-between mb-2">
          <label class="text-xs text-gray-400">Outbound Threshold</label>
          <span :class="['text-xs font-medium px-2 py-0.5 rounded', outboundIndicator.cls]">{{ outboundIndicator.label }}</span>
        </div>
        <div class="flex items-center gap-3">
          <input
            type="range"
            min="1"
            max="20"
            :value="outbound"
            @input="$emit('update:outbound', Number(($event.target as HTMLInputElement).value))"
            class="flex-1 accent-blue-500"
          />
          <input
            type="number"
            :value="outbound"
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

const props = defineProps<{ inbound: number; outbound: number }>()
defineEmits<{
  (e: 'update:inbound', v: number): void
  (e: 'update:outbound', v: number): void
}>()

const infoOpen = ref(false)

function scoreIndicator(value: number, recommendedDefault: number) {
  if (value <= 3) return { label: 'Conservative — Low false positives', cls: 'bg-green-900/50 text-green-300' }
  if (value <= 5) return { label: 'Recommended — Good balance', cls: 'bg-blue-900/50 text-blue-300' }
  if (value <= 9) return { label: 'Strict — Higher false positive risk', cls: 'bg-yellow-900/50 text-yellow-300' }
  return { label: 'Very Strict — Expect false positives', cls: 'bg-red-900/50 text-red-300' }
}

const inboundIndicator = computed(() => scoreIndicator(props.inbound, 5))
const outboundIndicator = computed(() => scoreIndicator(props.outbound, 4))
</script>
