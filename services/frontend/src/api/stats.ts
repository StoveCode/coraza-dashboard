import client from './client'

export interface StatsResponse {
  total_blocks: number
  total_detections: number
  top_ips: Array<{ ip: string; count: number }>
  top_rules: Array<{ rule_id: string; rule_msg: string; count: number }>
  events_per_hour: Array<{ hour: string; count: number }>
}

export async function fetchStats(): Promise<StatsResponse> {
  const { data } = await client.get<StatsResponse>('/api/stats')
  return data
}
