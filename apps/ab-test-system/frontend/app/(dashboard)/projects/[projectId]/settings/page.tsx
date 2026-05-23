'use client'

import { use } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { SdkKeysList } from '@/components/settings/sdk-keys-list'
import { CreateSdkKeyDialog } from '@/components/settings/create-sdk-key-dialog'

interface Props {
  params: Promise<{ projectId: string }>
}

export default function SettingsPage({ params }: Props) {
  const { projectId } = use(params)

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-gray-900">SDK Keys</h1>
        <p className="mt-1 text-sm text-gray-500">
          Authenticate your SDK with the AB Platform API using these keys.
        </p>
      </div>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle className="text-base">API Keys</CardTitle>
            <CardDescription className="mt-1">
              Keep these secret — treat them like passwords.
            </CardDescription>
          </div>
          <CreateSdkKeyDialog projectId={projectId} />
        </CardHeader>
        <CardContent>
          <SdkKeysList projectId={projectId} />
        </CardContent>
      </Card>
    </div>
  )
}
