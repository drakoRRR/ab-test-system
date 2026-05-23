'use client'

import { useState } from 'react'
import { useForm, useFieldArray } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Plus, Trash2, FlaskConical } from 'lucide-react'
import { toast } from 'sonner'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useCreateExperiment } from '@/hooks/use-experiments'
import { ApiError } from '@/lib/api/client'

const variantSchema = z.object({
  key: z.string().min(1).max(50),
  name: z.string().min(1).max(100),
  weight: z.number().int().min(1).max(100),
})

const schema = z.object({
  key: z.string().regex(/^[a-z0-9][a-z0-9-]*[a-z0-9]$/, 'Lowercase, numbers, hyphens (e.g. btn-color)'),
  name: z.string().min(1, 'Name is required').max(100),
  description: z.string().max(500).optional(),
  trafficPercent: z.number().min(1).max(100),
  variants: z.array(variantSchema).min(2, 'At least 2 variants required'),
})

type FormValues = z.infer<typeof schema>

interface Props {
  projectId: string
}

const VARIANT_COLORS = ['bg-indigo-500', 'bg-violet-500', 'bg-sky-500', 'bg-teal-500']

export function CreateExperimentDialog({ projectId }: Props) {
  const [open, setOpen] = useState(false)
  const { mutate, isPending } = useCreateExperiment(projectId)

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      trafficPercent: 100,
      variants: [
        { key: 'control', name: 'Control', weight: 50 },
        { key: 'treatment', name: 'Treatment', weight: 50 },
      ],
    },
  })

  const { fields, append, remove } = useFieldArray({ control: form.control, name: 'variants' })
  const variants = form.watch('variants')
  const totalWeight = variants.reduce((s, v) => s + (v.weight || 0), 0)
  const trafficPercent = form.watch('trafficPercent')

  function onSubmit(values: FormValues) {
    mutate(values, {
      onSuccess: () => {
        toast.success('Experiment created')
        setOpen(false)
        form.reset()
      },
      onError: (err) =>
        toast.error(err instanceof ApiError ? err.message : 'Failed to create experiment'),
    })
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button size="sm" />}>
        <Plus className="h-4 w-4" />
        New experiment
      </DialogTrigger>

      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-indigo-50">
              <FlaskConical className="h-5 w-5 text-indigo-600" />
            </div>
            <div>
              <DialogTitle className="text-lg">Create experiment</DialogTitle>
              <p className="text-sm text-muted-foreground">Configure your A/B test</p>
            </div>
          </div>
        </DialogHeader>

        <form onSubmit={form.handleSubmit(onSubmit)} className="mt-2 space-y-6">
          {/* Basic info */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label>Key <span className="text-muted-foreground font-normal">(immutable)</span></Label>
              <Input
                placeholder="btn-color-test"
                className="font-mono"
                {...form.register('key')}
              />
              {form.formState.errors.key && (
                <p className="text-xs text-destructive">{form.formState.errors.key.message}</p>
              )}
            </div>
            <div className="space-y-1.5">
              <Label>Name</Label>
              <Input placeholder="Button color test" {...form.register('name')} />
              {form.formState.errors.name && (
                <p className="text-xs text-destructive">{form.formState.errors.name.message}</p>
              )}
            </div>
          </div>

          <div className="space-y-1.5">
            <Label>Description <span className="text-muted-foreground font-normal">(optional)</span></Label>
            <Input placeholder="What are you testing and why?" {...form.register('description')} />
          </div>

          {/* Traffic */}
          <div className="rounded-lg border bg-muted/30 p-4 space-y-3">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium">Traffic allocation</p>
                <p className="text-xs text-muted-foreground">
                  Percentage of users entering this experiment
                </p>
              </div>
              <div className="flex items-center gap-2">
                <Input
                  type="number"
                  min={1}
                  max={100}
                  className="w-20 text-center font-medium"
                  {...form.register('trafficPercent', { valueAsNumber: true })}
                />
                <span className="text-sm text-muted-foreground">%</span>
              </div>
            </div>
            <input
              type="range"
              min={1}
              max={100}
              value={trafficPercent || 100}
              onChange={(e) => form.setValue('trafficPercent', Number(e.target.value))}
              className="w-full accent-indigo-600"
            />
          </div>

          {/* Variants */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium">Variants</p>
                <p className="text-xs text-muted-foreground">
                  Total weight:{' '}
                  <span className={totalWeight === 100 ? 'text-green-600 font-medium' : 'text-amber-600 font-medium'}>
                    {totalWeight}
                  </span>
                  {' '}/ 100
                </p>
              </div>
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={() => append({ key: '', name: '', weight: 50 })}
                disabled={fields.length >= 4}
              >
                <Plus className="h-3.5 w-3.5" />
                Add variant
              </Button>
            </div>

            {/* Weight distribution bar */}
            {totalWeight > 0 && (
              <div className="flex h-2 w-full overflow-hidden rounded-full bg-muted">
                {variants.map((v, i) => {
                  const pct = totalWeight > 0 ? ((v.weight || 0) / totalWeight) * 100 : 0
                  return (
                    <div
                      key={i}
                      className={`transition-all ${VARIANT_COLORS[i % VARIANT_COLORS.length]}`}
                      style={{ width: `${pct}%` }}
                    />
                  )
                })}
              </div>
            )}

            {/* Header row */}
            <div className="grid grid-cols-[1fr_1fr_80px_36px] gap-2 px-1">
              <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Key</span>
              <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Name</span>
              <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Weight %</span>
              <span />
            </div>

            {fields.map((field, i) => (
              <div
                key={field.id}
                className="grid grid-cols-[1fr_1fr_80px_36px] gap-2 items-center rounded-lg border bg-card p-2"
              >
                <div className="flex items-center gap-2">
                  <div className={`h-2 w-2 flex-none rounded-full ${VARIANT_COLORS[i % VARIANT_COLORS.length]}`} />
                  <Input
                    placeholder="control"
                    className="font-mono text-sm h-8"
                    {...form.register(`variants.${i}.key`)}
                  />
                </div>
                <div className="flex items-center gap-1.5">
                  <Input
                    placeholder="Control"
                    className="text-sm h-8"
                    {...form.register(`variants.${i}.name`)}
                  />
                  {i === 0 && (
                    <span className="flex-none rounded-full bg-indigo-100 px-1.5 py-0.5 text-[10px] font-medium text-indigo-700">
                      control
                    </span>
                  )}
                </div>
                <Input
                  type="number"
                  min={1}
                  max={100}
                  className="text-center text-sm h-8"
                  {...form.register(`variants.${i}.weight`, { valueAsNumber: true })}
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => remove(i)}
                  disabled={fields.length <= 2}
                  className="h-8 w-8 p-0 text-muted-foreground hover:text-destructive"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </Button>
              </div>
            ))}

            {form.formState.errors.variants?.root && (
              <p className="text-xs text-destructive">
                {form.formState.errors.variants.root.message}
              </p>
            )}
          </div>

          <div className="flex justify-end gap-2 border-t pt-4">
            <Button type="button" variant="outline" onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending || totalWeight !== 100}>
              {isPending ? 'Creating…' : 'Create experiment'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
