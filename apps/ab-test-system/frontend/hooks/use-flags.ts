import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { listFlags, createFlag, updateFlag, deleteFlag } from '@/lib/api/flags'
import { queryKeys } from '@/lib/query-keys'
import type { CreateFlagRequest, UpdateFlagRequest } from '@/types'

export function useFlags(projectId: string) {
  return useQuery({
    queryKey: queryKeys.flags.byProject(projectId),
    queryFn: () => listFlags(projectId),
    enabled: Boolean(projectId),
  })
}

export function useCreateFlag(projectId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: CreateFlagRequest) => createFlag(projectId, body),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.flags.byProject(projectId) }),
  })
}

export function useToggleFlag(projectId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ flagKey, enabled }: { flagKey: string; enabled: boolean }) =>
      updateFlag(projectId, flagKey, { enabled }),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.flags.byProject(projectId) }),
  })
}

export function useUpdateFlag(projectId: string, flagKey: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: UpdateFlagRequest) => updateFlag(projectId, flagKey, body),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.flags.byProject(projectId) }),
  })
}

export function useDeleteFlag(projectId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (flagKey: string) => deleteFlag(projectId, flagKey),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.flags.byProject(projectId) }),
  })
}
