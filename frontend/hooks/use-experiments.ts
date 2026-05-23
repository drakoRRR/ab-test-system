import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  listExperiments,
  getExperiment,
  createExperiment,
  updateExperiment,
  deleteExperiment,
  startExperiment,
  pauseExperiment,
  resumeExperiment,
  completeExperiment,
} from '@/lib/api/experiments'
import { queryKeys } from '@/lib/query-keys'
import type { CreateExperimentRequest, UpdateExperimentRequest } from '@/types'

export function useExperiments(projectId: string) {
  return useQuery({
    queryKey: queryKeys.experiments.byProject(projectId),
    queryFn: () => listExperiments(projectId),
    enabled: Boolean(projectId),
  })
}

export function useExperiment(projectId: string, experimentId: string) {
  return useQuery({
    queryKey: queryKeys.experiments.detail(experimentId),
    queryFn: () => getExperiment(projectId, experimentId),
    enabled: Boolean(projectId) && Boolean(experimentId),
  })
}

export function useCreateExperiment(projectId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: CreateExperimentRequest) => createExperiment(projectId, body),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: queryKeys.experiments.byProject(projectId) }),
  })
}

export function useUpdateExperiment(projectId: string, experimentId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: UpdateExperimentRequest) =>
      updateExperiment(projectId, experimentId, body),
    onSuccess: (updated) => {
      qc.setQueryData(queryKeys.experiments.detail(experimentId), updated)
      qc.invalidateQueries({ queryKey: queryKeys.experiments.byProject(projectId) })
    },
  })
}

export function useDeleteExperiment(projectId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (experimentId: string) => deleteExperiment(projectId, experimentId),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: queryKeys.experiments.byProject(projectId) }),
  })
}

function useLifecycleMutation(
  projectId: string,
  action: (pid: string, eid: string) => Promise<import('@/types').Experiment>,
) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (experimentId: string) => action(projectId, experimentId),
    onSuccess: (updated) => {
      qc.setQueryData(queryKeys.experiments.detail(updated.id), updated)
      qc.invalidateQueries({ queryKey: queryKeys.experiments.byProject(projectId) })
    },
  })
}

export function useStartExperiment(projectId: string) {
  return useLifecycleMutation(projectId, startExperiment)
}

export function usePauseExperiment(projectId: string) {
  return useLifecycleMutation(projectId, pauseExperiment)
}

export function useResumeExperiment(projectId: string) {
  return useLifecycleMutation(projectId, resumeExperiment)
}

export function useCompleteExperiment(projectId: string) {
  return useLifecycleMutation(projectId, completeExperiment)
}
