'use client'

import { use } from 'react'
import { PageHeader } from '@/components/shared/page-header'
import { EmptyState } from '@/components/shared/empty-state'
import { FlagCard } from '@/components/flags/flag-card'
import { CreateFlagDialog } from '@/components/flags/create-flag-dialog'
import { useFlags } from '@/hooks/use-flags'

interface Props {
  params: Promise<{ projectId: string }>
}

export default function FlagsPage({ params }: Props) {
  const { projectId } = use(params)
  const { data, isLoading, isError } = useFlags(projectId)

  if (isLoading) return null
  if (isError) return <p className="text-destructive">Failed to load flags.</p>

  return (
    <div className="space-y-6">
      <PageHeader
        title="Feature Flags"
        description="Control feature rollout and experiments"
        action={<CreateFlagDialog projectId={projectId} />}
      />

      {!data?.length ? (
        <EmptyState
          title="No feature flags"
          description="Create a flag to control feature rollout."
          action={<CreateFlagDialog projectId={projectId} />}
        />
      ) : (
        <div className="flex flex-col gap-3">
          {data.map((flag) => (
            <FlagCard key={flag.id} flag={flag} projectId={projectId} />
          ))}
        </div>
      )}
    </div>
  )
}
