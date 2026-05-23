import { apiClient } from './client'

export interface CurrentUser {
  id: string
  email: string
  name: string
  photoURL?: string
  organizationId?: string
  role?: string
}

export function upsertCurrentUser(email: string, name: string): Promise<CurrentUser> {
  return apiClient.post<CurrentUser>('/users', { email, name })
}

export function getCurrentUser(): Promise<CurrentUser> {
  return apiClient.get<CurrentUser>('/users/me')
}

export function createOrganization(name: string): Promise<{ id: string; name: string; createdAt: string }> {
  return apiClient.post('/organizations', { name })
}
