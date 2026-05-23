'use client'

import { use } from 'react'
import { PageHeader } from '@/components/shared/page-header'
import { EmptyState } from '@/components/shared/empty-state'
import { ExperimentCard } from '@/components/experiments/experiment-card'
import { CreateExperimentDialog } from '@/components/experiments/create-experiment-dialog'
import { useExperiments } from '@/hooks/use-experiments'

interface Props {
  params: Promise<{ projectId: string }>
}

export default function ExperimentsPage({ params }: Props) {
  const { projectId } = use(params)
  const { data, isLoading, isError } = useExperiments(projectId)

  if (isLoading) return null
  if (isError) return <p className="text-destructive">Failed to load experiments.</p>

  return (
    <div className="space-y-6">
      <PageHeader
        title="Experiments"
        description="Run A/B tests and analyze results"
        action={<CreateExperimentDialog projectId={projectId} />}
      />

      {!data?.length ? (
        <EmptyState
          title="No experiments"
          description="Create an experiment to start A/B testing."
          action={<CreateExperimentDialog projectId={projectId} />}
        />
      ) : (
        <div className="flex flex-col gap-3">
          {data.map((exp) => (
            <ExperimentCard key={exp.id} experiment={exp} projectId={projectId} />
          ))}
        </div>
      )}
    </div>
  )
}
