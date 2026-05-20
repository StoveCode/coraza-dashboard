<template>
  <div class="p-6">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-xl font-bold text-gray-100">Rules Management</h1>
      <div class="flex items-center gap-3">
        <!-- Unsaved changes badge -->
        <span v-if="store.hasUnsavedChanges" class="text-xs text-yellow-400 bg-yellow-400/10 px-3 py-1 rounded-full border border-yellow-400/30">
          Unsaved changes
        </span>
        <!-- Success badge -->
        <transition name="fade">
          <span v-if="store.successMessage" class="text-xs text-green-400 bg-green-400/10 px-3 py-1 rounded-full border border-green-400/30">
            &#10003; {{ store.successMessage }}
          </span>
        </transition>
        <!-- Error -->
        <span v-if="store.error" class="text-xs text-red-400 bg-red-400/10 px-3 py-1 rounded-full border border-red-400/30">
          &#10007; {{ store.error }}
        </span>
        <a
          href="https://playground.coraza.io/"
          target="_blank"
          rel="noopener noreferrer"
          class="flex items-center gap-1.5 px-4 py-2 rounded-lg bg-gray-800 hover:bg-gray-700 text-gray-300 text-sm transition-colors border border-gray-700"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/>
          </svg>
          Coraza Playground
        </a>
        <button
          @click="store.save()"
          :disabled="store.saving || !store.hasUnsavedChanges"
          class="px-5 py-2 rounded-lg bg-blue-600 hover:bg-blue-700 disabled:opacity-40 text-white text-sm font-semibold transition-all flex items-center gap-2"
        >
          <svg v-if="store.saving" class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
          </svg>
          {{ store.saving ? 'Saving...' : 'Save Changes' }}
        </button>
      </div>
    </div>

    <div v-if="store.loading" class="text-center text-gray-500 py-12">Loading configuration...</div>

    <template v-else>
      <!-- Engine Mode -->
      <EngineToggle v-model="store.config.engine_mode" />

      <div class="border-t border-gray-800/60 pt-5 mt-5"></div>

      <!-- Paranoia + Request Inspection -->
      <div class="grid grid-cols-5 gap-5">
        <div class="col-span-2">
          <ParanoiaSlider
            v-model="store.config.paranoia_level"
            v-model:enabled="store.config.paranoia_level_enabled"
          />
        </div>
        <div class="col-span-3">
          <RequestInspectionPanel
            v-model:inbound="store.config.inbound_threshold"
          />
        </div>
      </div>

      <div class="border-t border-gray-800/60 pt-5 mt-5"></div>

      <!-- Data Leakage Prevention -->
      <DataLeakagePanel
        v-model:outbound="store.config.outbound_threshold"
        v-model:responseCheck="store.config.response_check"
      />

      <div class="border-t border-gray-800/60 pt-5 mt-5"></div>

      <!-- CRS Categories -->
      <CategoryToggles
        :categories="store.categories"
        :disabled-tags="store.config.disabled_tags"
        :disabled-rule-ids="store.config.disabled_rule_ids"
        :paranoia-level="store.config.paranoia_level"
        :paranoia-level-enabled="store.config.paranoia_level_enabled"
        :catalog="catalog"
        @toggle-tag="store.toggleCategory"
        @toggle-rule="store.toggleRuleId"
      />

      <div class="border-t border-gray-800/60 pt-5 mt-5"></div>

      <!-- Disabled Rule IDs -->
      <DisabledRuleIds
        :rule-ids="store.config.disabled_rule_ids"
        @add="store.addRuleId"
        @remove="store.removeRuleId"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRulesStore } from '../stores/rules'
import { fetchRuleCatalog, type CRSRule } from '../api/rules'
import EngineToggle from '../components/rules/EngineToggle.vue'
import ParanoiaSlider from '../components/rules/ParanoiaSlider.vue'
import RequestInspectionPanel from '../components/rules/RequestInspectionPanel.vue'
import DataLeakagePanel from '../components/rules/DataLeakagePanel.vue'
import CategoryToggles from '../components/rules/CategoryToggles.vue'
import DisabledRuleIds from '../components/rules/DisabledRuleIds.vue'

const store = useRulesStore()
const catalog = ref<CRSRule[]>([])

async function loadCatalog() {
  try {
    catalog.value = await fetchRuleCatalog()
  } catch {
    // catalog stays empty — UI shows "No rules in catalog" per category
  }
}

onMounted(() => {
  store.loadConfig()
  loadCatalog()
})
</script>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.4s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
