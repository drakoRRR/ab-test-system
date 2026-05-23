import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { listSdkKeys, createSdkKey, revokeSdkKey } from '@/lib/api/sdk-keys'
import { queryKeys } from '@/lib/query-keys'
import type { CreateApiKeyRequest } from '@/types'

export function useSdkKeys(projectId: string) {
  return useQuery({
    queryKey: queryKeys.sdkKeys.byProject(projectId),
    queryFn: () => listSdkKeys(projectId),
    enabled: Boolean(projectId),
  })
}

export function useCreateSdkKey(projectId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: CreateApiKeyRequest) => createSdkKey(projectId, body),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.sdkKeys.byProject(projectId) }),
  })
}

export function useRevokeSdkKey(projectId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (keyId: string) => revokeSdkKey(projectId, keyId),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.sdkKeys.byProject(projectId) }),
  })
}
