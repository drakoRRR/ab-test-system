'use client'

import { Button } from '@/components/ui/button'
import {
  useStartExperiment,
  usePauseExperiment,
  useResumeExperiment,
  useCompleteExperiment,
} from '@/hooks/use-experiments'
import type { Experiment } from '@/types'
import { toast } from 'sonner'
import { ApiError } from '@/lib/api/client'

interface Props {
  experiment: Experiment
  projectId: string
}

export function ExperimentLifecycleControls({ experiment, projectId }: Props) {
  const start = useStartExperiment(projectId)
  const pause = usePauseExperiment(projectId)
  const resume = useResumeExperiment(projectId)
  const complete = useCompleteExperiment(projectId)

  function act(
    mutation: { mutate: (id: string, opts: object) => void; isPending: boolean },
    label: string,
  ) {
    mutation.mutate(experiment.id, {
      onSuccess: () => toast.success(`Experiment ${label}`),
      onError: (err: unknown) =>
        toast.error(err instanceof ApiError ? err.message : `Failed to ${label}`),
    })
  }

  const { status } = experiment

  return (
    <div className="flex gap-2">
      {status === 'draft' && (
        <Button size="sm" onClick={() => act(start, 'started')} disabled={start.isPending}>
          Start
        </Button>
      )}
      {status === 'running' && (
        <>
          <Button
            size="sm"
            variant="outline"
            onClick={() => act(pause, 'paused')}
            disabled={pause.isPending}
          >
            Pause
          </Button>
          <Button
            size="sm"
            variant="secondary"
            onClick={() => act(complete, 'completed')}
            disabled={complete.isPending}
          >
            Complete
          </Button>
        </>
      )}
      {status === 'paused' && (
        <>
          <Button size="sm" onClick={() => act(resume, 'resumed')} disabled={resume.isPending}>
            Resume
          </Button>
          <Button
            size="sm"
            variant="secondary"
            onClick={() => act(complete, 'completed')}
            disabled={complete.isPending}
          >
            Complete
          </Button>
        </>
      )}
    </div>
  )
}
