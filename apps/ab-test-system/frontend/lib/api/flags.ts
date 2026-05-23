import { apiClient } from './client'
import type { Flag, CreateFlagRequest, UpdateFlagRequest } from '@/types'

export function listFlags(projectId: string): Promise<Flag[]> {
  return apiClient.get<Flag[]>(`/projects/${projectId}/flags`)
}

export function getFlag(projectId: string, flagKey: string): Promise<Flag> {
  return apiClient.get<Flag>(`/projects/${projectId}/flags/${flagKey}`)
}

export function createFlag(projectId: string, body: CreateFlagRequest): Promise<Flag> {
  return apiClient.post<Flag>(`/projects/${projectId}/flags`, body)
}

export function updateFlag(
  projectId: string,
  flagKey: string,
  body: UpdateFlagRequest,
): Promise<Flag> {
  return apiClient.patch<Flag>(`/projects/${projectId}/flags/${flagKey}`, body)
}

export function deleteFlag(projectId: string, flagKey: string): Promise<void> {
  return apiClient.delete<void>(`/projects/${projectId}/flags/${flagKey}`)
}
