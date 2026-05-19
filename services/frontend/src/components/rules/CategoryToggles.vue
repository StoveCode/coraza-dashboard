<template>
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-5">
    <h2 class="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">CRS Rule Categories</h2>
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
      <div
        v-for="cat in categories"
        :key="cat.tag"
        class="flex items-center justify-between bg-gray-800 rounded-lg px-4 py-3"
        :class="{ 'opacity-60': isCovered(cat.tag) }"
      >
        <div>
          <div class="flex items-center gap-2">
            <p class="text-sm text-gray-200 font-medium">{{ cat.label }}</p>
            <span
              v-if="isCovered(cat.tag)"
              class="text-xs bg-blue-900/50 text-blue-300 px-1.5 py-0.5 rounded"
            >PL {{ coveredByPL(cat.tag) }}</span>
          </div>
          <p class="text-xs text-gray-500 mt-0.5">{{ cat.description }}</p>
        </div>
        <button
          @click="!isCovered(cat.tag) && $emit('toggle', cat.tag)"
          :disabled="isCovered(cat.tag)"
          :class="[
            'relative inline-flex h-6 w-11 flex-shrink-0 rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ml-4',
            isCovered(cat.tag)
              ? 'bg-blue-700/50 cursor-not-allowed opacity-50'
              : isDisabled(cat.tag) ? 'bg-gray-600' : 'bg-green-500'
          ]"
          :title="isCovered(cat.tag)
            ? `Active via Paranoia Level ${coveredByPL(cat.tag)} — disable PL to control individually`
            : isDisabled(cat.tag) ? 'Enable' : 'Disable'"
        >
          <span
            :class="[
              'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
              (isCovered(cat.tag) || !isDisabled(cat.tag)) ? 'translate-x-5' : 'translate-x-0'
            ]"
          ></span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { RuleCategory } from '../../api/rules'

const props = defineProps<{
  categories: RuleCategory[]
  disabledTags: string[]
  paranoiaLevel?: number
}>()

defineEmits<{ (e: 'toggle', tag: string): void }>()

const PL_COVERED_TAGS: Record<number, string[]> = {
  1: ['attack-sqli', 'attack-xss', 'attack-lfi', 'attack-scanner', 'attack-protocol'],
  2: ['attack-rce', 'attack-rfi', 'attack-injection', 'attack-protocol'],
  3: ['attack-ssrf', 'attack-reputation'],
  4: [],
}

function coveredByPL(tag: string): number | null {
  const pl = props.paranoiaLevel ?? 0
  for (let level = 1; level <= pl; level++) {
    if (PL_COVERED_TAGS[level]?.includes(tag)) {
      return level
    }
  }
  return null
}

function isCovered(tag: string): boolean {
  return coveredByPL(tag) !== null
}

function isDisabled(tag: string): boolean {
  return props.disabledTags.includes(tag)
}
</script>
