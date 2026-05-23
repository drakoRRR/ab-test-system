'use client'

import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from 'recharts'
import { formatPercent } from '@/lib/utils'
import type { VariantAnalytics } from '@/types'

interface Props {
  variants: VariantAnalytics[]
}

const COLORS = ['hsl(217, 91%, 60%)', 'hsl(142, 71%, 45%)', 'hsl(38, 92%, 50%)', 'hsl(0, 84%, 60%)']

export function ConversionRateChart({ variants }: Props) {
  const data = variants.map((v) => ({
    name: v.variantKey,
    rate: v.conversionRate,
  }))

  return (
    <ResponsiveContainer width="100%" height={240}>
      <BarChart data={data} margin={{ top: 4, right: 8, left: 0, bottom: 0 }}>
        <CartesianGrid strokeDasharray="3 3" vertical={false} />
        <XAxis dataKey="name" tick={{ fontSize: 12 }} />
        <YAxis tickFormatter={(v) => formatPercent(v)} tick={{ fontSize: 12 }} width={52} />
        <Tooltip
          formatter={(value) => [formatPercent(Number(value)), 'Conv. rate']}
          labelStyle={{ fontWeight: 600 }}
        />
        <Bar dataKey="rate" radius={[4, 4, 0, 0]}>
          {data.map((_, i) => (
            <Cell key={i} fill={COLORS[i % COLORS.length]} />
          ))}
        </Bar>
      </BarChart>
    </ResponsiveContainer>
  )
}
