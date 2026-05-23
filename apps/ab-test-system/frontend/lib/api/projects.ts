import { apiClient } from './client'
import type { Project, CreateProjectRequest, UpdateProjectRequest } from '@/types'

export function listProjects(): Promise<Project[]> {
  return apiClient.get<Project[]>('/projects')
}

export function getProject(id: string): Promise<Project> {
  return apiClient.get<Project>(`/projects/${id}`)
}

export function createProject(body: CreateProjectRequest): Promise<Project> {
  return apiClient.post<Project>('/projects', body)
}

export function updateProject(id: string, body: UpdateProjectRequest): Promise<Project> {
  return apiClient.patch<Project>(`/projects/${id}`, body)
}

export function deleteProject(id: string): Promise<void> {
  return apiClient.delete<void>(`/projects/${id}`)
}
