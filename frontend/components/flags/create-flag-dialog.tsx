'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Plus } from 'lucide-react'
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
import { useCreateFlag } from '@/hooks/use-flags'
import { ApiError } from '@/lib/api/client'
import type { FlagRule } from '@/types'

const schema = z.object({
  key: z
    .string()
    .min(1, 'Key is required')
    .regex(/^[a-z0-9-]+$/, 'Only lowercase letters, numbers, and hyphens'),
  name: z.string().min(1, 'Name is required'),
  rolloutType: z.enum(['all', 'percentage']),
  rolloutPercent: z.number().min(0).max(100),
})

type FormValues = z.infer<typeof schema>

interface Props {
  projectId: string
}

export function CreateFlagDialog({ projectId }: Props) {
  const [open, setOpen] = useState(false)
  const { mutate, isPending } = useCreateFlag(projectId)
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { key: '', name: '', rolloutType: 'all', rolloutPercent: 100 },
  })

  function onSubmit(values: FormValues) {
    const rules: FlagRule[] =
      values.rolloutType === 'percentage'
        ? [{ type: 'percentage', value: values.rolloutPercent / 100 }]
        : []

    mutate(
      { key: values.key, name: values.name, rules },
      {
        onSuccess: () => {
          toast.success('Flag created')
          setOpen(false)
          form.reset({ key: '', name: '', rolloutType: 'all', rolloutPercent: 100 })
        },
        onError: () => toast.error('Failed to create flag'),
      },
    )
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button size="sm" />}>
        <Plus className="h-4 w-4" />
        New flag
      </DialogTrigger>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>Create feature flag</DialogTitle>
        </DialogHeader>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-1.5">
            <Label>Key</Label>
            <Input placeholder="dark-mode" {...form.register('key')} />
            {form.formState.errors.key && (
              <p className="text-xs text-destructive">{form.formState.errors.key.message}</p>
            )}
          </div>
          <div className="space-y-1.5">
            <Label>Name</Label>
            <Input placeholder="Dark mode" {...form.register('name')} />
          </div>
          {/* Rollout */}
          <div className="space-y-1.5">
            <Label>Rollout</Label>
            <div className="flex flex-col gap-2">
              <label className="flex items-center gap-2 text-sm cursor-pointer">
                <input
                  type="radio"
                  value="all"
                  checked={form.watch('rolloutType') === 'all'}
                  onChange={() => form.setValue('rolloutType', 'all')}
                  className="accent-indigo-600"
                />
                All users (100%)
              </label>
              <label className="flex items-center gap-2 text-sm cursor-pointer">
                <input
                  type="radio"
                  value="percentage"
                  checked={form.watch('rolloutType') === 'percentage'}
                  onChange={() => form.setValue('rolloutType', 'percentage')}
                  className="accent-indigo-600"
                />
                Percentage rollout
              </label>
            </div>
          </div>
          {form.watch('rolloutType') === 'percentage' && (
            <div className="space-y-1.5">
              <Label>Rollout Percentage: {form.watch('rolloutPercent')}%</Label>
              <input
                type="range"
                min={0}
                max={100}
                step={1}
                value={form.watch('rolloutPercent')}
                onChange={(e) => form.setValue('rolloutPercent', Number(e.target.value))}
                className="w-full accent-indigo-600"
              />
            </div>
          )}
          <div className="flex justify-end gap-2 pt-1">
            <Button type="button" variant="outline" onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? 'Creating…' : 'Create'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
