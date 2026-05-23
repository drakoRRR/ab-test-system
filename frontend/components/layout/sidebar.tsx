'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { FlaskConical, Flag, Settings, ChevronLeft, LayoutDashboard } from 'lucide-react'
import { cn } from '@/lib/utils'

interface NavItemProps {
  href: string
  icon: React.ReactNode
  label: string
  exact?: boolean
}

function NavItem({ href, icon, label, exact = false }: NavItemProps) {
  const pathname = usePathname()
  const active = exact ? pathname === href : pathname.startsWith(href)

  return (
    <Link
      href={href}
      className={cn(
        'flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors',
        active
          ? 'bg-indigo-50 text-indigo-700 font-medium'
          : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900',
      )}
    >
      {icon}
      {label}
    </Link>
  )
}

interface SidebarProps {
  projectId: string
}

export function Sidebar({ projectId }: SidebarProps) {
  const base = `/projects/${projectId}`

  return (
    <aside className="flex h-full w-56 flex-col border-r border-gray-200 bg-white">
      <div className="border-b border-gray-200 px-4 py-3">
        <Link
          href="/projects"
          className="flex items-center gap-1 text-xs text-gray-500 hover:text-gray-900"
        >
          <ChevronLeft className="h-3 w-3" />
          All Projects
        </Link>
      </div>
      <nav className="flex flex-col gap-1 px-2 py-3">
        <NavItem
          href={base}
          icon={<LayoutDashboard className="h-4 w-4" />}
          label="Overview"
          exact
        />
        <NavItem
          href={`${base}/flags`}
          icon={<Flag className="h-4 w-4" />}
          label="Feature Flags"
        />
        <NavItem
          href={`${base}/experiments`}
          icon={<FlaskConical className="h-4 w-4" />}
          label="Experiments"
        />
        <NavItem
          href={`${base}/settings`}
          icon={<Settings className="h-4 w-4" />}
          label="SDK Keys"
        />
      </nav>
    </aside>
  )
}

export function RootSidebar() {
  return null
}
