<template>
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-5">
    <h2 class="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">CRS Rule Categories</h2>
    <div class="space-y-2">
      <div
        v-for="cat in categories"
        :key="cat.tag"
        class="bg-gray-800 rounded-lg overflow-hidden"
      >
        <!-- Category Header -->
        <div
          class="flex items-center justify-between px-4 py-3 cursor-pointer select-none hover:bg-gray-750"
          @click="toggleOpen(cat.tag)"
        >
          <div class="flex items-center gap-2 min-w-0">
            <!-- Chevron -->
            <svg
              class="w-4 h-4 text-gray-500 flex-shrink-0 transition-transform duration-200"
              :class="{ 'rotate-90': openCategories.has(cat.tag) }"
              fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
            </svg>
            <div class="min-w-0">
              <div class="flex items-center gap-2 flex-wrap">
                <span class="text-sm text-gray-200 font-medium">{{ cat.label }}</span>
                <span
                  v-if="isCategoryFullyCoveredByPL(cat.tag)"
                  class="text-xs bg-blue-900/50 text-blue-300 px-1.5 py-0.5 rounded flex-shrink-0"
                >PL{{ paranoiaLevel }}</span>
              </div>
              <p class="text-xs text-gray-500 mt-0.5">{{ cat.description }}</p>
            </div>
          </div>
          <div class="flex items-center gap-3 ml-3 flex-shrink-0">
            <span class="text-xs text-gray-500">{{ rulesForTag(cat.tag).length }} rules</span>
            <!-- Category Toggle -->
            <button
              @click.stop="!isCategoryFullyCoveredByPL(cat.tag) && $emit('toggle-tag', cat.tag)"
              :disabled="isCategoryFullyCoveredByPL(cat.tag)"
              :class="[
                'relative inline-flex h-6 w-11 flex-shrink-0 rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
                isCategoryFullyCoveredByPL(cat.tag)
                  ? 'bg-blue-700/50 cursor-not-allowed opacity-50'
                  : isTagDisabled(cat.tag) ? 'bg-gray-600' : 'bg-green-500'
              ]"
              :title="isCategoryFullyCoveredByPL(cat.tag)
                ? `Active via Paranoia Level ${paranoiaLevel} — disable PL to control individually`
                : isTagDisabled(cat.tag) ? 'Enable category' : 'Disable category'"
            >
              <span
                :class="[
                  'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                  (isCategoryFullyCoveredByPL(cat.tag) || !isTagDisabled(cat.tag)) ? 'translate-x-5' : 'translate-x-0'
                ]"
              ></span>
            </button>
          </div>
        </div>

        <!-- Expanded Rule List -->
        <div v-if="openCategories.has(cat.tag)" class="border-t border-gray-700/50">
          <div
            v-for="rule in rulesForTag(cat.tag)"
            :key="rule.id"
            class="flex items-center gap-3 px-4 py-2.5 border-b border-gray-700/30 last:border-b-0 transition-opacity"
            :class="{ 'opacity-50': isCoveredByPL(rule) || isTagDisabled(cat.tag) }"
          >
            <!-- PL Badge -->
            <span
              class="text-xs font-mono px-1.5 py-0.5 rounded flex-shrink-0"
              :class="isCoveredByPL(rule) ? 'bg-blue-900/50 text-blue-300' : 'bg-gray-700 text-gray-400'"
            >PL{{ rule.paranoia_level }}</span>
            <!-- Rule ID -->
            <span class="text-xs font-mono text-gray-500 flex-shrink-0 w-14">{{ rule.id }}</span>
            <!-- Severity Chip -->
            <span
              class="text-xs px-1.5 py-0.5 rounded flex-shrink-0"
              :class="severityClass(rule.severity)"
            >{{ rule.severity }}</span>
            <!-- Message -->
            <span
              class="text-xs text-gray-300 flex-1 truncate"
              :title="rule.msg"
            >{{ rule.msg }}</span>
            <!-- Rule Toggle -->
            <button
              @click.stop="!isCoveredByPL(rule) && !isTagDisabled(cat.tag) && $emit('toggle-rule', String(rule.id))"
              :disabled="isCoveredByPL(rule) || isTagDisabled(cat.tag)"
              :class="[
                'relative inline-flex h-5 w-9 flex-shrink-0 rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
                (isCoveredByPL(rule) || isTagDisabled(cat.tag))
                  ? 'cursor-not-allowed ' + (isCoveredByPL(rule) ? 'bg-blue-700/40' : 'bg-gray-600/40')
                  : isRuleDisabled(rule.id) ? 'bg-gray-600' : 'bg-green-500'
              ]"
              :title="isCoveredByPL(rule) ? `Active via Paranoia Level ${rule.paranoia_level}` : isTagDisabled(cat.tag) ? 'Category disabled' : isRuleDisabled(rule.id) ? 'Enable rule' : 'Disable rule'"
            >
              <span
                :class="[
                  'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                  (isCoveredByPL(rule) || (!isTagDisabled(cat.tag) && !isRuleDisabled(rule.id))) ? 'translate-x-4' : 'translate-x-0'
                ]"
              ></span>
            </button>
          </div>
          <div v-if="rulesForTag(cat.tag).length === 0" class="px-4 py-3 text-xs text-gray-600 italic">
            No rules in catalog for this category.
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { RuleCategory, CRSRule } from '../../api/rules'

const props = defineProps<{
  categories: RuleCategory[]
  disabledTags: string[]
  disabledRuleIds: string[]
  paranoiaLevel: number
  paranoiaLevelEnabled: boolean
  catalog: CRSRule[]
}>()

defineEmits<{
  (e: 'toggle-tag', tag: string): void
  (e: 'toggle-rule', id: string): void
}>()

const openCategories = ref(new Set<string>())

function toggleOpen(tag: string) {
  if (openCategories.value.has(tag)) {
    openCategories.value.delete(tag)
  } else {
    openCategories.value.add(tag)
  }
}

function rulesForTag(tag: string): CRSRule[] {
  return props.catalog.filter(r => r.tag === tag).sort((a, b) => a.id - b.id)
}

function isTagDisabled(tag: string): boolean {
  return props.disabledTags.includes(tag)
}

function isRuleDisabled(id: number): boolean {
  return props.disabledRuleIds.includes(String(id))
}

function isCoveredByPL(rule: CRSRule): boolean {
  return props.paranoiaLevelEnabled && rule.paranoia_level <= props.paranoiaLevel
}

function isCategoryFullyCoveredByPL(tag: string): boolean {
  const rules = rulesForTag(tag)
  if (rules.length === 0) return false
  return rules.every(r => isCoveredByPL(r))
}

function severityClass(severity: string): string {
  switch (severity) {
    case 'critical': return 'bg-red-900/50 text-red-300'
    case 'error': return 'bg-orange-900/50 text-orange-300'
    case 'warning': return 'bg-yellow-900/50 text-yellow-300'
    case 'notice': return 'bg-gray-700 text-gray-400'
    default: return 'bg-gray-700 text-gray-400'
  }
}
</script>
