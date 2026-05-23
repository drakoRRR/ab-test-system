import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { SignificanceBadge } from './significance-badge'
import { Badge } from '@/components/ui/badge'
import { formatPercent, formatNumber } from '@/lib/utils'
import type { VariantAnalytics } from '@/types'

interface Props {
  variants: VariantAnalytics[]
}

export function VariantStatsTable({ variants }: Props) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Variant</TableHead>
          <TableHead className="text-right">Exposures</TableHead>
          <TableHead className="text-right">Conversions</TableHead>
          <TableHead className="text-right">Conv. rate</TableHead>
          <TableHead className="text-right">Uplift</TableHead>
          <TableHead>Significance</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {variants.map((v) => (
          <TableRow key={v.variantId}>
            <TableCell>
              <div className="flex items-center gap-2">
                <span className="font-medium">{v.variantName}</span>
                <span className="text-xs text-muted-foreground font-mono">{v.variantKey}</span>
                {v.isControl && (
                  <Badge variant="outline" className="text-xs">
                    Control
                  </Badge>
                )}
              </div>
            </TableCell>
            <TableCell className="text-right">{formatNumber(v.exposures)}</TableCell>
            <TableCell className="text-right">{formatNumber(v.conversions)}</TableCell>
            <TableCell className="text-right font-medium">
              {formatPercent(v.conversionRate)}
            </TableCell>
            <TableCell className="text-right">
              {v.uplift != null ? (
                <span className={v.uplift >= 0 ? 'text-green-600' : 'text-red-600'}>
                  {v.uplift >= 0 ? '+' : ''}
                  {v.uplift.toFixed(1)}%
                </span>
              ) : (
                <span className="text-muted-foreground">—</span>
              )}
            </TableCell>
            <TableCell>
              {v.isControl ? (
                <span className="text-xs text-muted-foreground">Baseline</span>
              ) : (
                <SignificanceBadge significant={v.significant} pValue={v.pValue} />
              )}
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
