import { api } from './client'

export interface WAFEvent {
  id: string
  timestamp: string
  client_ip: string
  server: string
  uri: string
  rule_id: number
  rule_msg: string
  rule_file: string
  severity: string
  severity_id: number
  phase: string
  phase_id: number
  disruptive: boolean
  tags: string[]
  data: string
  unique_id: string
  anomaly_score: number
  block_type: string
}

export interface EventsFilter {
  limit?: number
  offset?: number
  from?: string
  to?: string
  disruptive?: boolean | null
  client_ip?: string
  rule_id?: number
  tag?: string
  block_type?: string
}

export interface EventsResponse {
  total: number
  events: WAFEvent[]
}

export async function fetchEvents(filter: EventsFilter = {}): Promise<EventsResponse> {
  const params: Record<string, string | number> = {
    limit: filter.limit ?? 50,
    offset: filter.offset ?? 0,
  }
  if (filter.from) params.from = filter.from
  if (filter.to) params.to = filter.to
  if (filter.disruptive !== null && filter.disruptive !== undefined) {
    params.disruptive = String(filter.disruptive)
  }
  if (filter.client_ip) params.client_ip = filter.client_ip
  if (filter.rule_id) params.rule_id = filter.rule_id
  if (filter.tag) params.tag = filter.tag
  if (filter.block_type) params.block_type = filter.block_type

  const resp = await api.get<EventsResponse>('/api/events', { params })
  return resp.data
}
