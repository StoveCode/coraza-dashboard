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
