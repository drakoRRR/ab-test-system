'use client'

import { use } from 'react'
import Link from 'next/link'
import { FlaskConical, Flag, ArrowRight, BarChart2 } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { CreateExperimentDialog } from '@/components/experiments/create-experiment-dialog'
import { CreateFlagDialog } from '@/components/flags/create-flag-dialog'
import { useExperiments } from '@/hooks/use-experiments'
import { useFlags } from '@/hooks/use-flags'
import { useProject } from '@/hooks/use-projects'
import type { ExperimentStatus } from '@/types'

const statusColors: Record<ExperimentStatus, string> = {
  draft: 'bg-gray-100 text-gray-600',
  running: 'bg-green-100 text-green-700',
  paused: 'bg-yellow-100 text-yellow-700',
  completed: 'bg-blue-100 text-blue-700',
}

interface Props {
  params: Promise<{ projectId: string }>
}

export default function ProjectOverviewPage({ params }: Props) {
  const { projectId } = use(params)
  const { data: project } = useProject(projectId)
  const { data: experiments } = useExperiments(projectId)
  const { data: flags } = useFlags(projectId)

  const runningCount = experiments?.filter((e) => e.status === 'running').length ?? 0
  const activeExperiments = experiments?.filter((e) => e.status === 'running').slice(0, 5) ?? []
  const recentFlags = flags?.slice(0, 5) ?? []

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-gray-900">{project?.name ?? 'Project'}</h1>
        <p className="mt-1 text-sm text-gray-500">Overview of experiments and feature flags</p>
      </div>

      {/* KPI cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Running Experiments</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold text-gray-900">{runningCount}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Feature Flags</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold text-gray-900">{flags?.length ?? 0}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">Total Experiments</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold text-gray-900">{experiments?.length ?? 0}</p>
          </CardContent>
        </Card>
      </div>

      {/* Active experiments */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Active Experiments</CardTitle>
          <CreateExperimentDialog projectId={projectId} />
        </CardHeader>
        <CardContent>
          {activeExperiments.length === 0 ? (
            <p className="py-4 text-center text-sm text-gray-500">No running experiments yet.</p>
          ) : (
            <div className="divide-y divide-gray-100">
              {activeExperiments.map((exp) => (
                <Link
                  key={exp.id}
                  href={`/projects/${projectId}/experiments/${exp.id}`}
                  className="flex items-center justify-between py-3 hover:bg-gray-50 -mx-2 px-2 rounded"
                >
                  <div className="flex items-center gap-3">
                    <FlaskConical className="h-4 w-4 text-gray-400" />
                    <div>
                      <p className="text-sm font-medium text-gray-900">{exp.name}</p>
                      <p className="text-xs text-gray-500">{exp.key} · {exp.trafficPercent}% traffic · {exp.variants.length} variants</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${statusColors[exp.status]}`}>
                      {exp.status}
                    </span>
                    <Link
                      href={`/projects/${projectId}/experiments/${exp.id}/analytics`}
                      onClick={(e) => e.stopPropagation()}
                      className="flex items-center gap-1 rounded px-2 py-0.5 text-xs text-indigo-600 hover:bg-indigo-50"
                    >
                      <BarChart2 className="h-3 w-3" />
                      Analytics
                    </Link>
                    <ArrowRight className="h-4 w-4 text-gray-400" />
                  </div>
                </Link>
              ))}
            </div>
          )}
          {(experiments?.length ?? 0) > 0 && (
            <div className="mt-3 border-t pt-3">
              <Link href={`/projects/${projectId}/experiments`} className="text-sm text-indigo-600 hover:text-indigo-800">
                View all experiments →
              </Link>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Feature flags */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Feature Flags</CardTitle>
          <CreateFlagDialog projectId={projectId} />
        </CardHeader>
        <CardContent>
          {recentFlags.length === 0 ? (
            <p className="py-4 text-center text-sm text-gray-500">No feature flags yet.</p>
          ) : (
            <div className="divide-y divide-gray-100">
              {recentFlags.map((flag) => (
                <div key={flag.id} className="flex items-center justify-between py-3">
                  <div className="flex items-center gap-3">
                    <Flag className="h-4 w-4 text-gray-400" />
                    <div>
                      <p className="text-sm font-medium text-gray-900">{flag.name}</p>
                      <p className="text-xs text-gray-500">{flag.key}</p>
                    </div>
                  </div>
                  <Badge variant={flag.enabled ? 'default' : 'secondary'}>
                    {flag.enabled ? 'ON' : 'OFF'}
                  </Badge>
                </div>
              ))}
            </div>
          )}
          {(flags?.length ?? 0) > 0 && (
            <div className="mt-3 border-t pt-3">
              <Link href={`/projects/${projectId}/flags`} className="text-sm text-indigo-600 hover:text-indigo-800">
                View all flags →
              </Link>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
