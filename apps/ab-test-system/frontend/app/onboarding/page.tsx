'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { createOrganization, getCurrentUser } from '@/lib/api/users'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { toast } from 'sonner'

export default function OnboardingPage() {
  const [name, setName] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [isChecking, setIsChecking] = useState(true)
  const router = useRouter()

  useEffect(() => {
    getCurrentUser()
      .then((user) => {
        if (user.organizationId) {
          router.replace('/projects')
        } else {
          setIsChecking(false)
        }
      })
      .catch(() => {
        setIsChecking(false)
      })
  }, [router])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return

    setIsLoading(true)
    try {
      await createOrganization(name.trim())
      router.push('/projects')
    } catch {
      toast.error('Failed to create organization. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  if (isChecking) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <Skeleton className="h-64 w-80 rounded-xl" />
      </div>
    )
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Create your organization</CardTitle>
          <CardDescription>This is your workspace for managing experiments and feature flags</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">Organization name</Label>
              <Input
                id="name"
                placeholder="Acme Corp"
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={isLoading}
                autoFocus
              />
            </div>
            <Button type="submit" className="w-full" disabled={isLoading || !name.trim()}>
              {isLoading ? 'Creating…' : 'Create organization'}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
