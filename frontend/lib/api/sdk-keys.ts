import { apiClient } from './client'
import type { ApiKey, ApiKeyCreated, CreateApiKeyRequest } from '@/types'

export function listSdkKeys(projectId: string): Promise<ApiKey[]> {
  return apiClient.get<ApiKey[]>(`/projects/${projectId}/api-keys`)
}

export function createSdkKey(projectId: string, body: CreateApiKeyRequest): Promise<ApiKeyCreated> {
  return apiClient.post<ApiKeyCreated>(`/projects/${projectId}/api-keys`, body)
}

export function revokeSdkKey(projectId: string, keyId: string): Promise<void> {
  return apiClient.delete<void>(`/projects/${projectId}/api-keys/${keyId}`)
}
