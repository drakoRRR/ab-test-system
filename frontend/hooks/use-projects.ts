import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  listProjects,
  getProject,
  createProject,
  updateProject,
  deleteProject,
} from '@/lib/api/projects'
import { queryKeys } from '@/lib/query-keys'
import type { CreateProjectRequest, UpdateProjectRequest } from '@/types'

export function useProjects() {
  return useQuery({
    queryKey: queryKeys.projects.list(),
    queryFn: listProjects,
  })
}

export function useProject(id: string) {
  return useQuery({
    queryKey: queryKeys.projects.detail(id),
    queryFn: () => getProject(id),
    enabled: Boolean(id),
  })
}

export function useCreateProject() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: CreateProjectRequest) => createProject(body),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.projects.list() }),
  })
}

export function useUpdateProject(id: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: UpdateProjectRequest) => updateProject(id, body),
    onSuccess: (updated) => {
      qc.setQueryData(queryKeys.projects.detail(id), updated)
      qc.invalidateQueries({ queryKey: queryKeys.projects.list() })
    },
  })
}

export function useDeleteProject() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteProject(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.projects.all() }),
  })
}
