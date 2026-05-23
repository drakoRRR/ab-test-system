'use client'

import { use } from 'react'
import Link from 'next/link'
import { ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { ConversionRateChart } from '@/components/analytics/conversion-rate-chart'
import { VariantStatsTable } from '@/components/analytics/variant-stats-table'
import { EmptyState } from '@/components/shared/empty-state'
import { useExperimentAnalytics } from '@/hooks/use-analytics'
import { formatNumber } from '@/lib/utils'

interface Props {
  params: Promise<{ projectId: string; experimentId: string }>
}

export default function AnalyticsPage({ params }: Props) {
  const { projectId, experimentId } = use(params)
  const { data, isLoading, isError } = useExperimentAnalytics(projectId, experimentId)

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-48" />
        <div className="grid grid-cols-3 gap-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-24 rounded-lg" />
          ))}
        </div>
        <Skeleton className="h-64 rounded-lg" />
      </div>
    )
  }

  if (isError) return <p className="text-destructive">Failed to load analytics.</p>

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <Link href={`/projects/${projectId}/experiments/${experimentId}`}>
          <Button variant="ghost" size="sm">
            <ArrowLeft className="h-4 w-4" />
          </Button>
        </Link>
        <h1 className="text-2xl font-bold">Analytics</h1>
      </div>

      {!data || data.totalExposures === 0 ? (
        <EmptyState
          title="No data yet"
          description="Once the experiment is running, exposure and conversion events will appear here."
        />
      ) : (
        <>
          <div className="grid grid-cols-3 gap-4">
            <Card>
              <CardHeader className="pb-1">
                <CardTitle className="text-xs text-muted-foreground font-medium uppercase tracking-wide">
                  Total exposures
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold">{formatNumber(data.totalExposures)}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-1">
                <CardTitle className="text-xs text-muted-foreground font-medium uppercase tracking-wide">
                  Total conversions
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold">{formatNumber(data.totalConversions)}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-1">
                <CardTitle className="text-xs text-muted-foreground font-medium uppercase tracking-wide">
                  Overall conv. rate
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-2xl font-bold">
                  {data.totalExposures > 0
                    ? `${((data.totalConversions / data.totalExposures) * 100).toFixed(1)}%`
                    : '—'}
                </p>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Conversion rates by variant</CardTitle>
            </CardHeader>
            <CardContent>
              <ConversionRateChart variants={data.variants} />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Variant breakdown</CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              <VariantStatsTable variants={data.variants} />
            </CardContent>
          </Card>
        </>
      )}
    </div>
  )
}
