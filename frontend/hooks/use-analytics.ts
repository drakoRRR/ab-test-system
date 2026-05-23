import { useQuery } from '@tanstack/react-query'
import { getExperimentAnalytics } from '@/lib/api/analytics'
import { queryKeys } from '@/lib/query-keys'

export function useExperimentAnalytics(projectId: string, experimentId: string) {
  return useQuery({
    queryKey: queryKeys.experiments.analytics(experimentId),
    queryFn: () => getExperimentAnalytics(projectId, experimentId),
    enabled: Boolean(projectId) && Boolean(experimentId),
    refetchInterval: 30_000,
  })
}
