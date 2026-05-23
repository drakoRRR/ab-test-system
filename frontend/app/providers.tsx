'use client'

import { useState, useEffect } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { NuqsAdapter } from 'nuqs/adapters/next/app'
import { onAuthStateChanged } from 'firebase/auth'
import { auth, startTokenSync } from '@/lib/auth'
import { useAuthStore } from '@/store/auth-store'

function AuthSync() {
  const { setUser, setLoading } = useAuthStore()

  useEffect(() => {
    const unsubscribeToken = startTokenSync()
    const unsubscribeAuth = onAuthStateChanged(auth, (user) => {
      setUser(user)
      setLoading(false)
    })
    return () => {
      unsubscribeToken()
      unsubscribeAuth()
    }
  }, [setUser, setLoading])

  return null
}

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: { staleTime: 30_000, retry: 1 },
        },
      }),
  )

  return (
    <QueryClientProvider client={queryClient}>
      <NuqsAdapter>
        <AuthSync />
        {children}
      </NuqsAdapter>
    </QueryClientProvider>
  )
}
