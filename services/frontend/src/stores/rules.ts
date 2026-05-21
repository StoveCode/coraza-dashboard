import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { fetchRulesConfig, saveRulesConfig, fetchRuleCategories, type RulesConfig, type RuleCategory, type CRSRule, type ValidatedRuleId } from '../api/rules'
import { fetchSPOAStatus } from '../api/system'

export const useRulesStore = defineStore('rules', () => {
  const config = ref<RulesConfig>({
    engine_mode: 'On',
    paranoia_level: 1,
    paranoia_level_enabled: true,
    inbound_threshold: 5,
    outbound_threshold: 4,
    disabled_rule_ids: [],
    disabled_tags: [],
    response_check: false,
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
      await saveRulesConfig(config.value)
      savedConfig.value = JSON.parse(JSON.stringify(config.value))

      // Poll SPOA status after save (max 3 attempts, 1s apart, 500ms initial delay)
      let spoaRunning = false
      await new Promise(resolve => setTimeout(resolve, 500))
      for (let i = 0; i < 3; i++) {
        try {
          const status = await fetchSPOAStatus()
          if (status.running) {
            spoaRunning = true
            break
          }
        } catch {
          // ignore
        }
        if (i < 2) await new Promise(resolve => setTimeout(resolve, 1000))
      }

      successMessage.value = spoaRunning
        ? 'Config saved & WAF reloaded'
        : 'Config saved — WAF restart may have failed. Check logs.'
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

  function addRuleId(id: string, catalog?: CRSRule[]) {
    const trimmed = id.trim()
    if (trimmed && !config.value.disabled_rule_ids.includes(trimmed)) {
      config.value.disabled_rule_ids.push(trimmed)
      if (config.value.disabled_rule_ids_validated) {
        const rule = catalog?.find(r => String(r.id) === trimmed)
        const entry: ValidatedRuleId = {
          id: trimmed,
          msg: rule?.msg,
          tag: rule?.tag,
          severity: rule?.severity,
          orphaned: !rule,
        }
        config.value.disabled_rule_ids_validated.push(entry)
      }
    }
  }

  function removeRuleId(id: string) {
    const idx = config.value.disabled_rule_ids.indexOf(id)
    if (idx >= 0) {
      config.value.disabled_rule_ids.splice(idx, 1)
      if (config.value.disabled_rule_ids_validated) {
        const vi = config.value.disabled_rule_ids_validated.findIndex(r => r.id === id)
        if (vi >= 0) config.value.disabled_rule_ids_validated.splice(vi, 1)
      }
    }
  }

  function toggleRuleId(id: string, catalog?: CRSRule[]) {
    const idx = config.value.disabled_rule_ids.indexOf(id)
    if (idx >= 0) {
      config.value.disabled_rule_ids.splice(idx, 1)
      // keep validated list in sync
      if (config.value.disabled_rule_ids_validated) {
        const vi = config.value.disabled_rule_ids_validated.findIndex(r => r.id === id)
        if (vi >= 0) config.value.disabled_rule_ids_validated.splice(vi, 1)
      }
    } else {
      config.value.disabled_rule_ids.push(id)
      if (config.value.disabled_rule_ids_validated) {
        const rule = catalog?.find(r => String(r.id) === id)
        const entry: ValidatedRuleId = {
          id,
          msg: rule?.msg,
          tag: rule?.tag,
          severity: rule?.severity,
          orphaned: !rule,
        }
        config.value.disabled_rule_ids_validated.push(entry)
      }
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
    toggleRuleId,
  }
})
