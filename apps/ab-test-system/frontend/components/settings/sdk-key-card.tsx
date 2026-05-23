'use client'

import { useState } from 'react'
import { Trash2, Key } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { useRevokeSdkKey } from '@/hooks/use-sdk-keys'
import { toast } from 'sonner'
import { ApiError } from '@/lib/api/client'
import type { ApiKey } from '@/types'

interface Props {
  sdkKey: ApiKey
  projectId: string
}

export function SdkKeyCard({ sdkKey, projectId }: Props) {
  const [confirmOpen, setConfirmOpen] = useState(false)
  const { mutate: revoke, isPending } = useRevokeSdkKey(projectId)

  function handleRevoke() {
    revoke(sdkKey.id, {
      onSuccess: () => {
        toast.success('API key revoked')
        setConfirmOpen(false)
      },
      onError: (err) =>
        toast.error(err instanceof ApiError ? err.message : 'Failed to revoke key'),
    })
  }

  const isRevoked = Boolean(sdkKey.revokedAt)

  return (
    <>
      <Card>
        <CardContent className="flex items-center justify-between py-4">
          <div className="flex items-center gap-3">
            <Key className="h-4 w-4 text-muted-foreground" />
            <div>
              <p className="text-sm font-medium">{sdkKey.name}</p>
              <p className="text-xs text-muted-foreground font-mono">{sdkKey.prefix}••••••••</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            {isRevoked ? (
              <Badge variant="destructive">Revoked</Badge>
            ) : (
              <Badge variant="secondary">Active</Badge>
            )}
            {!isRevoked && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setConfirmOpen(true)}
                className="text-muted-foreground hover:text-destructive"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      <ConfirmDialog
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        title="Revoke API key"
        description={`This will permanently revoke "${sdkKey.name}". Any SDK using this key will stop working.`}
        confirmLabel="Revoke"
        onConfirm={handleRevoke}
        isPending={isPending}
      />
    </>
  )
}
