import client from './client'

export interface WAFEvent {
  id: string
  timestamp: string
  client_ip: string
  method: string
  uri: string
  rule_id: string
  rule_msg: string
  severity: string
  action: 'block' | 'detect'
  raw_log: Record<string, unknown>
}

export interface EventsResponse {
  total: number
  limit: number
  offset: number
  events: WAFEvent[]
}

export interface EventsParams {
  limit?: number
  offset?: number
  from?: string
  to?: string
  action?: 'block' | 'detect' | ''
  client_ip?: string
}

export async function fetchEvents(params: EventsParams = {}): Promise<EventsResponse> {
  const { data } = await client.get<EventsResponse>('/api/events', { params })
  return data
}

export async function checkHealth(): Promise<void> {
  await client.get('/api/health')
}
