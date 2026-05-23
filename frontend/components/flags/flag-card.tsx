'use client'

import { useState } from 'react'
import { Trash2, Pencil } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { FlagToggle } from './flag-toggle'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { useDeleteFlag, useUpdateFlag } from '@/hooks/use-flags'
import { toast } from 'sonner'
import { ApiError } from '@/lib/api/client'
import type { Flag } from '@/types'

interface Props {
  flag: Flag
  projectId: string
}

function EditRolloutDialog({
  flag,
  projectId,
  open,
  onOpenChange,
}: {
  flag: Flag
  projectId: string
  open: boolean
  onOpenChange: (v: boolean) => void
}) {
  const currentPercent = flag.rules?.find((r) => r.type === 'percentage')
    ? Math.round((flag.rules.find((r) => r.type === 'percentage')!.value) * 100)
    : 100
  const [percent, setPercent] = useState(currentPercent)
  const { mutate: update, isPending } = useUpdateFlag(projectId, flag.key)

  function handleSave() {
    const rules = percent < 100 ? [{ type: 'percentage' as const, value: percent / 100 }] : []
    update(
      { rules },
      {
        onSuccess: () => {
          toast.success('Rollout updated')
          onOpenChange(false)
        },
        onError: (err) =>
          toast.error(err instanceof ApiError ? err.message : 'Failed to update flag'),
      },
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-xs">
        <DialogHeader>
          <DialogTitle>Edit rollout — {flag.name}</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 pt-1">
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label>Rollout percentage</Label>
              <span className="text-sm font-medium tabular-nums">{percent}%</span>
            </div>
            <input
              type="range"
              min={0}
              max={100}
              value={percent}
              onChange={(e) => setPercent(Number(e.target.value))}
              className="w-full accent-indigo-600"
            />
            <p className="text-xs text-muted-foreground">
              {percent === 100
                ? 'Enabled for all users'
                : percent === 0
                  ? 'Disabled for all users'
                  : `Enabled for ${percent}% of users (hash-based, stable)`}
            </p>
          </div>
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button onClick={handleSave} disabled={isPending}>
              {isPending ? 'Saving…' : 'Save'}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

export function FlagCard({ flag, projectId }: Props) {
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const { mutate: del, isPending } = useDeleteFlag(projectId)

  const rolloutPercent = flag.rules?.find((r) => r.type === 'percentage')
    ? Math.round(flag.rules.find((r) => r.type === 'percentage')!.value * 100)
    : 100

  function handleDelete() {
    del(flag.key, {
      onSuccess: () => {
        toast.success('Flag deleted')
        setConfirmOpen(false)
      },
      onError: (err) =>
        toast.error(err instanceof ApiError ? err.message : 'Failed to delete flag'),
    })
  }

  return (
    <>
      <Card>
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="text-base">{flag.name}</CardTitle>
              <p className="mt-0.5 text-xs text-muted-foreground font-mono">{flag.key}</p>
            </div>
            <div className="flex items-center gap-2">
              <FlagToggle projectId={projectId} flagKey={flag.key} enabled={flag.enabled} />
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setEditOpen(true)}
                className="text-muted-foreground hover:text-foreground"
                title="Edit rollout"
              >
                <Pencil className="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setConfirmOpen(true)}
                className="text-muted-foreground hover:text-destructive"
                title="Delete flag"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent className="pt-0">
          <p className="text-xs text-muted-foreground">
            {rolloutPercent === 100
              ? 'All users'
              : `${rolloutPercent}% rollout`}
          </p>
        </CardContent>
      </Card>

      <EditRolloutDialog
        flag={flag}
        projectId={projectId}
        open={editOpen}
        onOpenChange={setEditOpen}
      />

      <ConfirmDialog
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        title="Delete flag"
        description={`This will permanently delete "${flag.name}". This action cannot be undone.`}
        confirmLabel="Delete"
        onConfirm={handleDelete}
        isPending={isPending}
      />
    </>
  )
}
