import { api } from './client'

export interface RulesConfig {
  engine_mode: 'On' | 'DetectionOnly' | 'Off'
  paranoia_level: number
  paranoia_level_enabled: boolean
  inbound_threshold: number
  outbound_threshold: number
  disabled_rule_ids: string[]
  disabled_tags: string[]
}

export interface RuleCategory {
  tag: string
  label: string
  description: string
}

export async function fetchRulesConfig(): Promise<RulesConfig> {
  const resp = await api.get<RulesConfig>('/api/rules/config')
  return resp.data
}

export async function saveRulesConfig(cfg: RulesConfig): Promise<{ status: string; message: string }> {
  const resp = await api.put<{ status: string; message: string }>('/api/rules/config', cfg)
  return resp.data
}

export async function fetchRuleCategories(): Promise<RuleCategory[]> {
  const resp = await api.get<RuleCategory[]>('/api/rules/categories')
  return resp.data
}

export interface CRSRule {
  id: number
  msg: string
  paranoia_level: number
  tag: string
  severity: string
  directive?: string
}

export async function fetchRuleCatalog(): Promise<CRSRule[]> {
  const resp = await api.get<CRSRule[]>('/api/rules/catalog')
  return resp.data
}
