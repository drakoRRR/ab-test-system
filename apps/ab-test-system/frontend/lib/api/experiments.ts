import { apiClient } from './client'
import type { Experiment, CreateExperimentRequest, UpdateExperimentRequest } from '@/types'

const base = (projectId: string) => `/projects/${projectId}/experiments`

export function listExperiments(projectId: string): Promise<Experiment[]> {
  return apiClient.get<Experiment[]>(base(projectId))
}

export function getExperiment(projectId: string, experimentId: string): Promise<Experiment> {
  return apiClient.get<Experiment>(`${base(projectId)}/${experimentId}`)
}

export function createExperiment(
  projectId: string,
  body: CreateExperimentRequest,
): Promise<Experiment> {
  return apiClient.post<Experiment>(base(projectId), body)
}

export function updateExperiment(
  projectId: string,
  experimentId: string,
  body: UpdateExperimentRequest,
): Promise<Experiment> {
  return apiClient.patch<Experiment>(`${base(projectId)}/${experimentId}`, body)
}

export function deleteExperiment(projectId: string, experimentId: string): Promise<void> {
  return apiClient.delete<void>(`${base(projectId)}/${experimentId}`)
}

export function startExperiment(projectId: string, experimentId: string): Promise<Experiment> {
  return apiClient.post<Experiment>(`${base(projectId)}/${experimentId}/start`)
}

export function pauseExperiment(projectId: string, experimentId: string): Promise<Experiment> {
  return apiClient.post<Experiment>(`${base(projectId)}/${experimentId}/pause`)
}

export function resumeExperiment(projectId: string, experimentId: string): Promise<Experiment> {
  return apiClient.post<Experiment>(`${base(projectId)}/${experimentId}/resume`)
}

export function completeExperiment(projectId: string, experimentId: string): Promise<Experiment> {
  return apiClient.post<Experiment>(`${base(projectId)}/${experimentId}/complete`)
}
