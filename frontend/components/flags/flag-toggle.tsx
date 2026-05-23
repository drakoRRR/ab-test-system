'use client'

import { useToggleFlag } from '@/hooks/use-flags'
import { toast } from 'sonner'
import { ApiError } from '@/lib/api/client'

interface Props {
  projectId: string
  flagKey: string
  enabled: boolean
}

export function FlagToggle({ projectId, flagKey, enabled }: Props) {
  const { mutate, isPending } = useToggleFlag(projectId)

  function handleToggle() {
    mutate(
      { flagKey, enabled: !enabled },
      {
        onSuccess: () => toast.success(`Flag ${!enabled ? 'enabled' : 'disabled'}`),
        onError: (err) =>
          toast.error(err instanceof ApiError ? err.message : 'Failed to update flag'),
      },
    )
  }

  return (
    <button
      role="switch"
      aria-checked={enabled}
      onClick={handleToggle}
      disabled={isPending}
      className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50 ${
        enabled ? 'bg-primary' : 'bg-input'
      }`}
    >
      <span
        className={`inline-block h-3.5 w-3.5 rounded-full bg-background transition-transform ${
          enabled ? 'translate-x-4' : 'translate-x-1'
        }`}
      />
    </button>
  )
}
