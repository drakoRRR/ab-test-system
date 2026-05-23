import { Badge } from '@/components/ui/badge'
import type { ExperimentStatus } from '@/types'

const STATUS_CONFIG: Record<
  ExperimentStatus,
  { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }
> = {
  draft: { label: 'Draft', variant: 'secondary' },
  running: { label: 'Running', variant: 'default' },
  paused: { label: 'Paused', variant: 'outline' },
  completed: { label: 'Completed', variant: 'secondary' },
}

export function ExperimentStatusBadge({ status }: { status: ExperimentStatus }) {
  const config = STATUS_CONFIG[status]
  return <Badge variant={config.variant}>{config.label}</Badge>
}
