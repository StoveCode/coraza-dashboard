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
