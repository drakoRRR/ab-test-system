'use client'

import { useRouter } from 'next/navigation'
import { LogOut, User } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { signOut } from '@/lib/auth'
import { useUser } from '@/store/auth-store'
import { toast } from 'sonner'

export function Header() {
  const user = useUser()
  const router = useRouter()

  async function handleSignOut() {
    await signOut()
    router.push('/login')
    toast.success('Signed out')
  }

  return (
    <header className="flex h-14 items-center justify-between border-b bg-card px-6">
      <span className="text-sm font-semibold tracking-tight">SplitLab</span>
      <div className="flex items-center gap-3">
        {user && (
          <span className="flex items-center gap-1.5 text-sm text-muted-foreground">
            <User className="h-4 w-4" />
            {user.displayName ?? user.email}
          </span>
        )}
        <Button variant="ghost" size="sm" onClick={handleSignOut}>
          <LogOut className="h-4 w-4" />
          Sign out
        </Button>
      </div>
    </header>
  )
}
