export const queryKeys = {
  projects: {
    all: () => ['projects'] as const,
    list: () => [...queryKeys.projects.all(), 'list'] as const,
    detail: (id: string) => [...queryKeys.projects.all(), id] as const,
  },
  flags: {
    all: () => ['flags'] as const,
    byProject: (projectId: string) => [...queryKeys.flags.all(), projectId] as const,
    detail: (projectId: string, key: string) => [...queryKeys.flags.byProject(projectId), key] as const,
  },
  experiments: {
    all: () => ['experiments'] as const,
    byProject: (projectId: string) => [...queryKeys.experiments.all(), projectId] as const,
    detail: (id: string) => [...queryKeys.experiments.all(), 'detail', id] as const,
    analytics: (id: string) => [...queryKeys.experiments.detail(id), 'analytics'] as const,
  },
  sdkKeys: {
    all: () => ['sdk-keys'] as const,
    byProject: (projectId: string) => [...queryKeys.sdkKeys.all(), projectId] as const,
  },
} as const
