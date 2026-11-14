import { useState } from 'react'
import { Link, NavLink, Outlet } from 'react-router'
import { Menu, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Sheet, SheetContent, SheetTitle, SheetTrigger } from '@/components/ui/sheet'
import { useRecentGroups } from '@/lib/recent'
import { cn } from '@/lib/utils'

function Wordmark({ onClick }: { onClick?: () => void }) {
  return (
    <Link to="/" onClick={onClick} className="text-[15px] font-semibold tracking-tight">
      split<span className="text-primary">it</span>
    </Link>
  )
}

function SidebarContent({ onNavigate }: { onNavigate?: () => void }) {
  const groups = useRecentGroups()

  return (
    <div className="flex h-full flex-col gap-6 p-3">
      <div className="flex h-9 items-center px-2">
        <Wordmark onClick={onNavigate} />
      </div>

      <Button asChild variant="outline" size="sm" className="justify-start">
        <Link to="/" onClick={onNavigate}>
          <Plus />
          New group
        </Link>
      </Button>

      <nav className="flex min-h-0 flex-col gap-0.5 overflow-y-auto">
        <p className="px-2 pb-1.5 text-xs font-medium text-muted-foreground">Your groups</p>
        {groups.length === 0 ? (
          <p className="px-2 text-xs leading-relaxed text-muted-foreground">
            Groups you create or open on this device show up here.
          </p>
        ) : (
          groups.map((g) => (
            <NavLink
              key={g.slug}
              to={`/g/${g.slug}`}
              onClick={onNavigate}
              className={({ isActive }) =>
                cn(
                  'truncate rounded-md px-2 py-1.5 text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-foreground',
                  isActive && 'bg-sidebar-accent text-foreground',
                )
              }
            >
              {g.name}
            </NavLink>
          ))
        )}
      </nav>

      <p className="mt-auto px-2 text-xs leading-relaxed text-muted-foreground">
        Anyone with a group's link can view and edit it.
      </p>
    </div>
  )
}

export function AppShell() {
  const [open, setOpen] = useState(false)
  const close = () => setOpen(false)

  return (
    <div className="min-h-dvh md:grid md:grid-cols-[248px_1fr]">
      <aside className="sticky top-0 hidden h-dvh border-r border-sidebar-border bg-sidebar md:block">
        <SidebarContent />
      </aside>

      <header className="sticky top-0 z-10 flex h-14 items-center justify-between border-b bg-background/90 px-4 backdrop-blur md:hidden">
        <Wordmark />
        <Sheet open={open} onOpenChange={setOpen}>
          <SheetTrigger asChild>
            <Button variant="ghost" size="icon-sm" aria-label="Open groups">
              <Menu />
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="w-72 border-sidebar-border bg-sidebar p-0">
            <SheetTitle className="sr-only">Groups</SheetTitle>
            <SidebarContent onNavigate={close} />
          </SheetContent>
        </Sheet>
      </header>

      <main className="min-w-0">
        <div className="mx-auto w-full max-w-3xl px-4 py-8 md:px-10 md:py-12">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
