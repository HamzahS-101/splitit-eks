import { useState } from 'react'
import { useNavigate } from 'react-router'
import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import { MemberInput } from '@/components/member-input'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { api } from '@/lib/api'
import { rememberGroup } from '@/lib/recent'

export function HomePage() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [members, setMembers] = useState<string[]>([])

  const create = useMutation({
    mutationFn: () => api.createGroup(name.trim(), members),
    onSuccess: (group) => {
      rememberGroup(group.slug, group.name)
      navigate(`/g/${group.slug}`)
    },
    onError: (error) => toast.error(error.message),
  })

  const canSubmit = name.trim().length > 0 && members.length >= 2 && !create.isPending

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-semibold tracking-tight">Split shared costs with one link</h1>
        <p className="max-w-md leading-relaxed text-muted-foreground">
          Make a group, add who's in it and share the link. splitit keeps the running total and works out
          the fewest payments to settle up. No accounts needed.
        </p>
      </div>

      <form
        className="flex flex-col gap-5 rounded-xl border bg-card p-5 md:p-6"
        onSubmit={(e) => {
          e.preventDefault()
          if (canSubmit) create.mutate()
        }}
      >
        <div className="flex flex-col gap-2">
          <Label htmlFor="group-name">Group name</Label>
          <Input
            id="group-name"
            value={name}
            maxLength={80}
            onChange={(e) => setName(e.target.value)}
            placeholder="Lisbon, March"
            autoComplete="off"
          />
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="member-input">People</Label>
          <MemberInput id="member-input" value={members} onChange={setMembers} />
          <p className="text-xs text-muted-foreground">Press Enter after each name. You need at least two.</p>
        </div>

        <div className="flex justify-end">
          <Button type="submit" disabled={!canSubmit}>
            {create.isPending ? 'Creating group' : 'Create group'}
          </Button>
        </div>
      </form>
    </div>
  )
}
