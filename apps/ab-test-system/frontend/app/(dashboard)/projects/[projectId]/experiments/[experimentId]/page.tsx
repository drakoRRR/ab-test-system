'use client'

import { use } from 'react'
import Link from 'next/link'
import { ArrowLeft, BarChart2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import { ExperimentStatusBadge } from '@/components/experiments/experiment-status-badge'
import { ExperimentLifecycleControls } from '@/components/experiments/experiment-lifecycle-controls'
import { useExperiment } from '@/hooks/use-experiments'
import { formatNumber } from '@/lib/utils'

interface Props {
  params: Promise<{ projectId: string; experimentId: string }>
}

export default function ExperimentDetailPage({ params }: Props) {
  const { projectId, experimentId } = use(params)
  const { data: exp, isLoading, isError } = useExperiment(projectId, experimentId)

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-56" />
        <Skeleton className="h-40 w-full" />
      </div>
    )
  }

  if (isError || !exp) return <p className="text-destructive">Failed to load experiment.</p>

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <Link href={`/projects/${projectId}/experiments`}>
          <Button variant="ghost" size="sm">
            <ArrowLeft className="h-4 w-4" />
          </Button>
        </Link>
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold">{exp.name}</h1>
            <ExperimentStatusBadge status={exp.status} />
          </div>
          <p className="text-sm text-muted-foreground font-mono mt-0.5">{exp.key}</p>
        </div>
        <Link href={`/projects/${projectId}/experiments/${experimentId}/analytics`}>
          <Button variant="outline" size="sm">
            <BarChart2 className="h-4 w-4" />
            Analytics
          </Button>
        </Link>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm font-medium text-muted-foreground">Details</CardTitle>
        </CardHeader>
        <CardContent className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-xs text-muted-foreground">Traffic</p>
            <p className="text-sm font-medium">{exp.trafficPercent}%</p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Status</p>
            <ExperimentStatusBadge status={exp.status} />
          </div>
          {exp.description && (
            <div className="col-span-2">
              <p className="text-xs text-muted-foreground">Description</p>
              <p className="text-sm">{exp.description}</p>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm font-medium text-muted-foreground">
            Variants ({exp.variants.length})
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          {exp.variants.map((v, i) => (
            <div key={v.id} className="flex items-center justify-between rounded-md border px-3 py-2">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium">{v.name}</span>
                <span className="text-xs text-muted-foreground font-mono">{v.key}</span>
                {i === 0 && <Badge variant="outline" className="text-xs">Control</Badge>}
              </div>
              <span className="text-sm text-muted-foreground">{v.weight}%</span>
            </div>
          ))}
        </CardContent>
      </Card>

      <Separator />
      <ExperimentLifecycleControls experiment={exp} projectId={projectId} />
    </div>
  )
}
