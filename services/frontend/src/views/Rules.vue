<template>
  <div class="p-6 space-y-5">
    <!-- Header -->
    <div class="flex items-center justify-between">
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

      <!-- Paranoia + Thresholds -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
        <ParanoiaSlider
          v-model="store.config.paranoia_level"
          v-model:enabled="store.config.paranoia_level_enabled"
        />
        <ThresholdInputs
          v-model:inbound="store.config.inbound_threshold"
          v-model:outbound="store.config.outbound_threshold"
        />
      </div>

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
import ThresholdInputs from '../components/rules/ThresholdInputs.vue'
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
