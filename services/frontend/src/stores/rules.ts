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

  // restart status: null | 'pending' | 'success' | 'timeout'
  const restartStatus = ref<null | 'pending' | 'success' | 'timeout'>(null)
  const logsOpen = ref(false)

  async function save() {
    saving.value = true
    error.value = null
    successMessage.value = null
    restartStatus.value = 'pending'
    try {
      const result = await saveRulesConfig(config.value)
      savedConfig.value = JSON.parse(JSON.stringify(config.value))

      const initiatedAt = result.restart_initiated_at
      if (initiatedAt) {
        let attempts = 0
        const poll = setInterval(async () => {
          attempts++
          try {
            const status = await fetchSPOAStatus()
            if (status.started_at && new Date(status.started_at) > new Date(initiatedAt)) {
              clearInterval(poll)
              restartStatus.value = 'success'
              const t = new Date(status.started_at).toLocaleTimeString()
              successMessage.value = `Config saved — WAF restarted at ${t}`
              setTimeout(() => { successMessage.value = null; restartStatus.value = null }, 6000)
            } else if (attempts >= 15) {
              clearInterval(poll)
              restartStatus.value = 'timeout'
              logsOpen.value = true
            }
          } catch {
            if (attempts >= 15) {
              clearInterval(poll)
              restartStatus.value = 'timeout'
              logsOpen.value = true
            }
          }
        }, 1000)
      } else {
        // fallback: old behaviour
        restartStatus.value = 'success'
        successMessage.value = 'Config saved & WAF reloaded'
        setTimeout(() => { successMessage.value = null; restartStatus.value = null }, 4000)
      }
    } catch (e: any) {
      error.value = e?.response?.data?.error ?? e?.message ?? 'Failed to save config'
      restartStatus.value = null
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

  async function disableRuleOnly(id: string) {
    // Add id to disabled_rule_ids without reloading config from server
    addRuleId(id)
    await save()
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
    restartStatus,
    logsOpen,
    hasUnsavedChanges,
    loadConfig,
    save,
    disableRuleOnly,
    isCategoryDisabled,
    toggleCategory,
    addRuleId,
    removeRuleId,
    toggleRuleId,
  }
})
