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
                <span
                  v-else-if="isCategoryFullyExcludedByPL(cat.tag)"
                  class="text-xs bg-gray-700 text-gray-500 px-1.5 py-0.5 rounded flex-shrink-0"
                >Excluded by PL{{ paranoiaLevel }}</span>
              </div>
              <p class="text-xs text-gray-500 mt-0.5">{{ cat.description }}</p>
            </div>
          </div>
          <div class="flex items-center gap-3 ml-3 flex-shrink-0">
            <span class="text-xs text-gray-500">{{ activeRuleCount(cat.tag) }} / {{ rulesForTag(cat.tag).length }} active</span>
            <!-- Category Toggle -->
            <button
              @click.stop="!isCategoryFullyCoveredByPL(cat.tag) && !isCategoryFullyExcludedByPL(cat.tag) && $emit('toggle-tag', cat.tag)"
              :disabled="isCategoryFullyCoveredByPL(cat.tag) || isCategoryFullyExcludedByPL(cat.tag)"
              :class="[
                'relative inline-flex h-6 w-11 flex-shrink-0 rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
                isCategoryFullyCoveredByPL(cat.tag)
                  ? 'bg-blue-700/50 cursor-not-allowed opacity-50'
                  : isCategoryFullyExcludedByPL(cat.tag)
                    ? 'bg-gray-600/40 cursor-not-allowed opacity-40'
                    : isTagDisabled(cat.tag) ? 'bg-gray-600' : 'bg-green-500'
              ]"
              :title="isCategoryFullyCoveredByPL(cat.tag)
                ? `Active via Paranoia Level ${paranoiaLevel} — disable PL to control individually`
                : isCategoryFullyExcludedByPL(cat.tag)
                  ? `Excluded by Paranoia Level ${paranoiaLevel} — all rules require higher PL`
                  : isTagDisabled(cat.tag) ? 'Enable category' : 'Disable category'"
            >
              <span
                :class="[
                  'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                  (isCategoryFullyCoveredByPL(cat.tag) || (!isCategoryFullyExcludedByPL(cat.tag) && !isTagDisabled(cat.tag))) ? 'translate-x-5' : 'translate-x-0'
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
            :class="{
              'opacity-70': isCoveredByPL(rule) && !isTagDisabled(cat.tag),
              'opacity-40': isExcludedByPL(rule) || isTagDisabled(cat.tag)
            }"
          >
            <!-- PL Badge -->
            <span
              class="text-xs font-mono px-1.5 py-0.5 rounded flex-shrink-0"
              :class="isCoveredByPL(rule) ? 'bg-blue-900/50 text-blue-300' : isExcludedByPL(rule) ? 'bg-gray-700/50 text-gray-500' : 'bg-gray-700 text-gray-400'"
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
            <!-- Info Button -->
            <button
              @click.stop="showDirective(rule)"
              class="text-gray-500 hover:text-gray-300 transition-colors p-1 rounded flex-shrink-0"
              title="View rule directive"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
            </button>
            <!-- Rule Toggle -->
            <button
              @click.stop="isManuallyToggleable(rule) && !isTagDisabled(cat.tag) && $emit('toggle-rule', String(rule.id))"
              :disabled="!isManuallyToggleable(rule) || isTagDisabled(cat.tag)"
              :class="[
                'relative inline-flex h-5 w-9 flex-shrink-0 rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
                isCoveredByPL(rule)
                  ? 'cursor-not-allowed bg-blue-700/40'
                  : isExcludedByPL(rule) || isTagDisabled(cat.tag)
                    ? 'cursor-not-allowed bg-gray-600/40'
                    : isRuleDisabled(rule.id) ? 'bg-gray-600' : 'bg-green-500'
              ]"
              :title="isCoveredByPL(rule)
                ? `Active via Paranoia Level ${rule.paranoia_level}`
                : isExcludedByPL(rule)
                  ? `Excluded by Paranoia Level ${paranoiaLevel} (rule requires PL${rule.paranoia_level})`
                  : isTagDisabled(cat.tag)
                    ? 'Category disabled'
                    : isRuleDisabled(rule.id) ? 'Enable rule' : 'Disable rule'"
            >
              <span
                :class="[
                  'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                  (isCoveredByPL(rule) || (isManuallyToggleable(rule) && !isTagDisabled(cat.tag) && !isRuleDisabled(rule.id))) ? 'translate-x-4' : 'translate-x-0'
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

  <!-- Rule Directive Modal -->
  <Teleport to="body">
    <div v-if="activeDirectiveRule" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <!-- Backdrop -->
      <div class="absolute inset-0 bg-black/60" @click="activeDirectiveRule = null"></div>
      <!-- Modal -->
      <div class="relative bg-gray-900 border border-gray-700 rounded-xl w-full max-w-2xl shadow-2xl">
        <!-- Header -->
        <div class="flex items-center justify-between px-5 py-4 border-b border-gray-800">
          <div>
            <span class="text-sm font-mono text-gray-400 mr-2">{{ activeDirectiveRule.id }}</span>
            <span class="text-sm font-semibold text-gray-100">{{ activeDirectiveRule.msg }}</span>
          </div>
          <button @click="activeDirectiveRule = null" class="text-gray-500 hover:text-gray-300">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
            </svg>
          </button>
        </div>
        <!-- Badges -->
        <div class="flex gap-2 px-5 py-3 border-b border-gray-800">
          <span class="text-xs px-2 py-1 rounded bg-gray-800 text-gray-400">
            PL{{ activeDirectiveRule.paranoia_level }}
          </span>
          <span class="text-xs px-2 py-1 rounded" :class="severityClass(activeDirectiveRule.severity)">
            {{ activeDirectiveRule.severity }}
          </span>
          <span class="text-xs px-2 py-1 rounded bg-gray-800 text-gray-400">
            {{ activeDirectiveRule.tag }}
          </span>
        </div>
        <!-- WAF Directive -->
        <div class="px-5 py-4">
          <div class="flex items-center justify-between mb-2">
            <span class="text-xs font-semibold text-gray-400 uppercase tracking-wider">WAF Directive</span>
            <div class="flex gap-2">
              <!-- Copy Button -->
              <button
                @click="copyDirective"
                class="flex items-center gap-1.5 text-xs text-gray-400 hover:text-gray-200 bg-gray-800 hover:bg-gray-700 px-3 py-1.5 rounded transition-colors"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                </svg>
                {{ copied ? 'Copied!' : 'Copy' }}
              </button>
              <!-- Playground Link -->
              <a
                href="https://playground.coraza.io/"
                target="_blank"
                rel="noopener noreferrer"
                class="flex items-center gap-1.5 text-xs text-blue-400 hover:text-blue-300 bg-blue-900/30 hover:bg-blue-900/50 px-3 py-1.5 rounded transition-colors border border-blue-800/50"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/>
                </svg>
                Test in Playground
              </a>
            </div>
          </div>
          <!-- Directive Code Block -->
          <pre class="bg-gray-950 border border-gray-800 rounded-lg p-4 text-xs text-gray-300 font-mono overflow-x-auto whitespace-pre-wrap break-all max-h-64 overflow-y-auto">{{ activeDirectiveRule.directive || `# Rule ${activeDirectiveRule.id}\nSecRuleRemoveById ${activeDirectiveRule.id}` }}</pre>
        </div>
        <!-- Disable Hint -->
        <div class="px-5 py-3 border-t border-gray-800 bg-gray-950/50 rounded-b-xl">
          <p class="text-xs text-gray-500">
            To disable this rule, use:
            <code class="text-gray-400 font-mono">SecRuleRemoveById {{ activeDirectiveRule.id }}</code>
          </p>
        </div>
      </div>
    </div>
  </Teleport>
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
const activeDirectiveRule = ref<CRSRule | null>(null)
const copied = ref(false)

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

// Rule is active via PL (covered = within level, already active through PL)
function isCoveredByPL(rule: CRSRule): boolean {
  return props.paranoiaLevelEnabled && props.paranoiaLevel > 0
    && rule.paranoia_level <= props.paranoiaLevel
}

// Rule is inactive via PL (above the level, automatically deactivated by PL)
function isExcludedByPL(rule: CRSRule): boolean {
  return props.paranoiaLevelEnabled && props.paranoiaLevel > 0
    && rule.paranoia_level > props.paranoiaLevel
}

// Whether a rule can be toggled manually (no PL active or PL=0)
function isManuallyToggleable(rule: CRSRule): boolean {
  return !props.paranoiaLevelEnabled || props.paranoiaLevel === 0
}

function isCategoryFullyCoveredByPL(tag: string): boolean {
  const rules = rulesForTag(tag)
  if (rules.length === 0) return false
  return rules.every(r => isCoveredByPL(r))
}

function isCategoryFullyExcludedByPL(tag: string): boolean {
  const rules = rulesForTag(tag)
  if (rules.length === 0) return false
  return rules.every(r => isExcludedByPL(r))
}

function activeRuleCount(tag: string): number {
  return rulesForTag(tag).filter(r =>
    !isExcludedByPL(r) && !isTagDisabled(r.tag) && !isRuleDisabled(r.id)
  ).length
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

function showDirective(rule: CRSRule) {
  activeDirectiveRule.value = rule
  copied.value = false
}

async function copyDirective() {
  const text = activeDirectiveRule.value?.directive
    || `SecRuleRemoveById ${activeDirectiveRule.value?.id}`
  await navigator.clipboard.writeText(text)
  copied.value = true
  setTimeout(() => { copied.value = false }, 2000)
}
</script>
