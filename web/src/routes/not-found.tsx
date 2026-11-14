import { Link } from 'react-router'
import { Button } from '@/components/ui/button'

export function NotFound({ group = false }: { group?: boolean }) {
  return (
    <div className="flex flex-col items-start gap-4">
      <div className="flex flex-col gap-1.5">
        <h1 className="text-xl font-semibold tracking-tight">
          {group ? "This group doesn't exist" : 'Page not found'}
        </h1>
        <p className="max-w-md leading-relaxed text-muted-foreground">
          {group
            ? 'The link may be wrong, or the group was removed after 30 days without any activity.'
            : "There's nothing at this address."}
        </p>
      </div>
      <Button asChild variant="outline" size="sm">
        <Link to="/">Start a new group</Link>
      </Button>
    </div>
  )
}
