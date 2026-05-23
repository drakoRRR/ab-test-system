'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { signInWithGoogle } from '@/lib/auth'
import { upsertCurrentUser } from '@/lib/api/users'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { toast } from 'sonner'

export default function LoginPage() {
  const [isLoading, setIsLoading] = useState(false)
  const router = useRouter()

  async function handleGoogleSignIn() {
    setIsLoading(true)
    try {
      const user = await signInWithGoogle()
      const profile = await upsertCurrentUser(user.email ?? '', user.displayName ?? '')
      router.push(profile.organizationId ? '/projects' : '/onboarding')
    } catch {
      toast.error('Sign in failed. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/30">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">SplitLab</CardTitle>
          <CardDescription>Sign in to manage your experiments and feature flags</CardDescription>
        </CardHeader>
        <CardContent>
          <Button className="w-full" onClick={handleGoogleSignIn} disabled={isLoading}>
            {isLoading ? 'Signing in…' : 'Continue with Google'}
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}
