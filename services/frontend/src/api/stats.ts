import { api } from './client'

export interface TopEntry {
  label: string
  count: number
}

export interface TopRule {
  rule_id: number
  msg: string
  count: number
}

export interface HourBucket {
  hour: string
  count: number
}

export interface Stats {
  total_blocks: number
  total_detections: number
  top_ips: TopEntry[]
  top_ips_blocked: TopEntry[]
  top_rules: TopRule[]
  top_tags: TopEntry[]
  top_phases: TopEntry[]
  events_per_hour: HourBucket[]
}

export async function fetchStats(): Promise<Stats> {
  const resp = await api.get<Stats>('/api/stats')
  return resp.data
}
