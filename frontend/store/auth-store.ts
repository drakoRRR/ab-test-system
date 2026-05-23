'use client'

import { create } from 'zustand'
import { useShallow } from 'zustand/shallow'
import type { User } from 'firebase/auth'

interface AuthState {
  user: User | null
  isLoading: boolean
  setUser: (user: User | null) => void
  setLoading: (v: boolean) => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isLoading: true,
  setUser: (user) => set({ user }),
  setLoading: (isLoading) => set({ isLoading }),
}))

export const useUser = () => useAuthStore((s) => s.user)
export const useIsAuthenticated = () => useAuthStore((s) => s.user !== null)
export const useAuthStatus = () =>
  useAuthStore(useShallow((s) => ({ user: s.user, isLoading: s.isLoading })))
