import { api } from './client'

export interface SystemVersions {
  backend_crs_version: string
  spoa_crs_version: string
  spoa_image: string
}

export async function fetchSystemVersions(): Promise<SystemVersions> {
  const resp = await api.get<SystemVersions>('/api/system/versions')
  return resp.data
}

export interface SPOAStatus {
  running: boolean
  started_at?: string
  status: string
}

export async function fetchSPOAStatus(): Promise<SPOAStatus> {
  const resp = await api.get<SPOAStatus>('/api/system/spoa-status')
  return resp.data
}

export interface SPOALogs {
  lines: string[]
  count: number
}

export async function fetchSPOALogs(tail = 100): Promise<SPOALogs> {
  const resp = await api.get<SPOALogs>(`/api/system/spoa-logs?tail=${tail}`)
  return resp.data
}
