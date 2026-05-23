import { apiClient } from './client'
import type { ExperimentAnalytics } from '@/types'

export function getExperimentAnalytics(
  projectId: string,
  experimentId: string,
): Promise<ExperimentAnalytics> {
  return apiClient.get<ExperimentAnalytics>(
    `/projects/${projectId}/experiments/${experimentId}/analytics`,
  )
}
