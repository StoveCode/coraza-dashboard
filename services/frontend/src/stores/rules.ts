import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { fetchRulesConfig, saveRulesConfig, fetchRuleCategories, type RulesConfig, type RuleCategory } from '../api/rules'

export const useRulesStore = defineStore('rules', () => {
  const config = ref<RulesConfig>({
    engine_mode: 'On',
    paranoia_level: 1,
    inbound_threshold: 5,
    outbound_threshold: 4,
    disabled_rule_ids: [],
    disabled_tags: [],
  })

  const savedConfig = ref<RulesConfig | null>(null)
  const categories = ref<RuleCategory[]>([])
  const loading = ref(false)
  const saving = ref(false)
  const error = ref<string | null>(null)
  const successMessage = ref<string | null>(null)

  const hasUnsavedChanges = computed(() => {
    if (!savedConfig.value) return false
    return JSON.stringify(config.value) !== JSON.stringify(savedConfig.value)
  })

  async function loadConfig() {
    loading.value = true
    error.value = null
    try {
      const [cfg, cats] = await Promise.all([fetchRulesConfig(), fetchRuleCategories()])
      config.value = cfg
      savedConfig.value = JSON.parse(JSON.stringify(cfg))
      categories.value = cats
    } catch (e: any) {
      error.value = e?.message ?? 'Failed to load config'
    } finally {
      loading.value = false
    }
  }

  async function save() {
    saving.value = true
    error.value = null
    successMessage.value = null
    try {
      const result = await saveRulesConfig(config.value)
      savedConfig.value = JSON.parse(JSON.stringify(config.value))
      successMessage.value = result.message
      setTimeout(() => { successMessage.value = null }, 4000)
    } catch (e: any) {
      error.value = e?.response?.data?.error ?? e?.message ?? 'Failed to save config'
    } finally {
      saving.value = false
    }
  }

  function isCategoryDisabled(tag: string): boolean {
    return config.value.disabled_tags.includes(tag)
  }

  function toggleCategory(tag: string) {
    const idx = config.value.disabled_tags.indexOf(tag)
    if (idx >= 0) {
      config.value.disabled_tags.splice(idx, 1)
    } else {
      config.value.disabled_tags.push(tag)
    }
  }

  function addRuleId(id: string) {
    const trimmed = id.trim()
    if (trimmed && !config.value.disabled_rule_ids.includes(trimmed)) {
      config.value.disabled_rule_ids.push(trimmed)
    }
  }

  function removeRuleId(id: string) {
    const idx = config.value.disabled_rule_ids.indexOf(id)
    if (idx >= 0) {
      config.value.disabled_rule_ids.splice(idx, 1)
    }
  }

  return {
    config,
    categories,
    loading,
    saving,
    error,
    successMessage,
    hasUnsavedChanges,
    loadConfig,
    save,
    isCategoryDisabled,
    toggleCategory,
    addRuleId,
    removeRuleId,
  }
})
