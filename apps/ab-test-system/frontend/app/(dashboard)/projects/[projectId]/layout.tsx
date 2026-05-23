import { Sidebar } from '@/components/layout/sidebar'

interface Props {
  children: React.ReactNode
  params: Promise<{ projectId: string }>
}

export default async function ProjectLayout({ children, params }: Props) {
  const { projectId } = await params

  return (
    <>
      <Sidebar projectId={projectId} />
      <main className="flex-1 overflow-y-auto p-6">{children}</main>
    </>
  )
}
