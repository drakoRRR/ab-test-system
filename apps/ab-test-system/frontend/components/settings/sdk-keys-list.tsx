'use client'

import { toast } from 'sonner'
import { Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { useSdkKeys, useRevokeSdkKey } from '@/hooks/use-sdk-keys'
import type { ApiKey } from '@/types'

interface Props {
  projectId: string
}

function KeyRow({ apiKey, projectId }: { apiKey: ApiKey; projectId: string }) {
  const { mutate: revoke, isPending } = useRevokeSdkKey(projectId)

  function handleRevoke() {
    revoke(apiKey.id, {
      onSuccess: () => toast.success('SDK key revoked'),
      onError: () => toast.error('Failed to revoke key'),
    })
  }

  const display = apiKey.prefix ? `${apiKey.prefix}...` : 'sdk_••••••••'

  return (
    <div className="flex items-center justify-between py-3">
      <div>
        <p className="text-sm font-medium text-foreground">{apiKey.name}</p>
        <p className="mt-0.5 font-mono text-xs text-muted-foreground">{display}</p>
        <p className="mt-0.5 text-xs text-muted-foreground">
          Created {new Date(apiKey.createdAt).toLocaleDateString()}
          {apiKey.revokedAt && (
            <span className="ml-2 text-red-500">
              · Revoked {new Date(apiKey.revokedAt).toLocaleDateString()}
            </span>
          )}
        </p>
      </div>
      {!apiKey.revokedAt && (
        <AlertDialog>
          <AlertDialogTrigger
            render={
              <Button variant="ghost" size="sm" className="text-red-500 hover:text-red-700">
                <Trash2 className="h-4 w-4" />
              </Button>
            }
          />
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Revoke &ldquo;{apiKey.name}&rdquo;?</AlertDialogTitle>
              <AlertDialogDescription>
                Any SDK using this key will immediately lose access. This cannot be undone.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction
                onClick={handleRevoke}
                disabled={isPending}
                className="bg-red-600 hover:bg-red-700"
              >
                Revoke
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}
    </div>
  )
}

export function SdkKeysList({ projectId }: Props) {
  const { data: keys, isLoading, isError } = useSdkKeys(projectId)

  if (isLoading) return <p className="py-4 text-sm text-muted-foreground">Loading…</p>
  if (isError) return <p className="py-4 text-sm text-red-500">Failed to load SDK keys.</p>

  if (!keys?.length) {
    return (
      <p className="py-8 text-center text-sm text-muted-foreground">
        No SDK keys yet. Create one to authenticate your SDK.
      </p>
    )
  }

  return (
    <div className="divide-y divide-border">
      {keys.map((key) => (
        <KeyRow key={key.id} apiKey={key} projectId={projectId} />
      ))}
    </div>
  )
}
