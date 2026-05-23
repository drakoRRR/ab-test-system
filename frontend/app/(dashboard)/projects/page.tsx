'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Plus, ArrowRight } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { PageHeader } from '@/components/shared/page-header'
import { EmptyState } from '@/components/shared/empty-state'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { useProjects, useCreateProject } from '@/hooks/use-projects'
import { ApiError } from '@/lib/api/client'

const schema = z.object({
  name: z.string().min(1).max(100),
  description: z.string().max(500).optional(),
})
type FormValues = z.infer<typeof schema>

function CreateProjectDialog() {
  const [open, setOpen] = useState(false)
  const { mutate, isPending } = useCreateProject()
  const form = useForm<FormValues>({ resolver: zodResolver(schema) })

  function onSubmit(values: FormValues) {
    mutate(values, {
      onSuccess: () => {
        toast.success('Project created')
        setOpen(false)
        form.reset()
      },
      onError: (err) =>
        toast.error(err instanceof ApiError ? err.message : 'Failed to create project'),
    })
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button size="sm" />}>
        <Plus className="h-4 w-4" />
        New project
      </DialogTrigger>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>Create project</DialogTitle>
        </DialogHeader>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-1.5">
            <Label>Name</Label>
            <Input placeholder="My app" {...form.register('name')} />
          </div>
          <div className="space-y-1.5">
            <Label>Description</Label>
            <Input placeholder="Optional" {...form.register('description')} />
          </div>
          <div className="flex justify-end gap-2 pt-1">
            <Button type="button" variant="outline" onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? 'Creating…' : 'Create'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export default function ProjectsPage() {
  const { data, isLoading, isError } = useProjects()
  const router = useRouter()

  if (isLoading) {
    return (
      <main className="flex-1 overflow-y-auto p-6">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <Skeleton className="h-8 w-32" />
            <Skeleton className="h-9 w-28" />
          </div>
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-24 w-full rounded-lg" />
          ))}
        </div>
      </main>
    )
  }

  if (isError) {
    return (
      <main className="flex-1 overflow-y-auto p-6">
        <p className="text-destructive">Failed to load projects.</p>
      </main>
    )
  }

  return (
    <main className="flex-1 overflow-y-auto p-6">
    <div className="space-y-6">
      <PageHeader
        title="Projects"
        description="Manage your A/B testing projects"
        action={<CreateProjectDialog />}
      />

      {!data?.length ? (
        <EmptyState
          title="No projects yet"
          description="Create a project to start managing feature flags and experiments."
          action={<CreateProjectDialog />}
        />
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {data.map((project) => (
            <Card
              key={project.id}
              className="cursor-pointer transition-shadow hover:shadow-md"
              onClick={() => router.push(`/projects/${project.id}`)}
            >
              <CardHeader className="pb-2">
                <div className="flex items-start justify-between">
                  <CardTitle className="text-base">{project.name}</CardTitle>
                  <ArrowRight className="h-4 w-4 text-muted-foreground" />
                </div>
              </CardHeader>
              {project.description && (
                <CardContent className="pt-0">
                  <p className="text-sm text-muted-foreground">{project.description}</p>
                </CardContent>
              )}
            </Card>
          ))}
        </div>
      )}
    </div>
    </main>
  )
}
