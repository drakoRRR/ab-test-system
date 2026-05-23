import Link from 'next/link'
import { BarChart2, ArrowRight } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { ExperimentStatusBadge } from './experiment-status-badge'
import { ExperimentLifecycleControls } from './experiment-lifecycle-controls'
import type { Experiment } from '@/types'

interface Props {
  experiment: Experiment
  projectId: string
}

const HAS_ANALYTICS: Experiment['status'][] = ['running', 'paused', 'completed']

export function ExperimentCard({ experiment, projectId }: Props) {
  const base = `/projects/${projectId}/experiments/${experiment.id}`

  return (
    <Card className="transition-shadow hover:shadow-sm">
      <CardContent className="p-4">
        <div className="flex items-start justify-between gap-4">

          {/* Left: info */}
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <Link
                href={base}
                className="text-base font-semibold text-foreground hover:text-indigo-600 transition-colors"
              >
                {experiment.name}
              </Link>
              <ExperimentStatusBadge status={experiment.status} />
            </div>
            <p className="mt-0.5 font-mono text-xs text-muted-foreground">{experiment.key}</p>
            <div className="mt-2 flex gap-3 text-xs text-muted-foreground">
              <span>{experiment.trafficPercent}% traffic</span>
              <span>·</span>
              <span>{experiment.variants.length} variants</span>
            </div>
          </div>

          {/* Right: actions */}
          <div className="flex flex-col items-end gap-2">
            <div className="flex items-center gap-1.5">
              {HAS_ANALYTICS.includes(experiment.status) && (
                <Link
                  href={`${base}/analytics`}
                  className="flex items-center gap-1 rounded-md border border-indigo-200 bg-indigo-50 px-2.5 py-1 text-xs font-medium text-indigo-700 hover:bg-indigo-100 transition-colors"
                >
                  <BarChart2 className="h-3.5 w-3.5" />
                  Analytics
                </Link>
              )}
              <Link
                href={base}
                className="flex items-center gap-1 rounded-md border px-2.5 py-1 text-xs font-medium text-muted-foreground hover:bg-muted transition-colors"
              >
                Details
                <ArrowRight className="h-3 w-3" />
              </Link>
            </div>
            <ExperimentLifecycleControls experiment={experiment} projectId={projectId} />
          </div>

        </div>
      </CardContent>
    </Card>
  )
}
