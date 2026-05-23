import { Badge } from '@/components/ui/badge'

interface Props {
  significant: boolean | null | undefined
  pValue: number | null | undefined
}

export function SignificanceBadge({ significant, pValue }: Props) {
  if (significant === null || significant === undefined) {
    return <Badge variant="outline">No data</Badge>
  }

  return (
    <div className="flex items-center gap-1.5">
      <Badge variant={significant ? 'default' : 'secondary'}>
        {significant ? 'Significant' : 'Not significant'}
      </Badge>
      {pValue !== null && pValue !== undefined && (
        <span className="text-xs text-muted-foreground">p={pValue.toFixed(4)}</span>
      )}
    </div>
  )
}
