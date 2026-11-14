import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import type { Expense, Group } from '@/lib/api'
import { formatMoney } from '@/lib/money'

type Props = {
  group: Group
  expenses: Expense[] | undefined
  onAdd: () => void
  onEdit: (expense: Expense) => void
}

function dayLabel(isoDate: string) {
  const [y, m, d] = isoDate.split('-').map(Number)
  const date = new Date(y, m - 1, d)
  const now = new Date()
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const diffDays = Math.round((startOfToday.getTime() - date.getTime()) / 86_400_000)
  if (diffDays === 0) return 'Today'
  if (diffDays === 1) return 'Yesterday'
  return date.toLocaleDateString('en-GB', {
    weekday: 'short',
    day: 'numeric',
    month: 'short',
    year: date.getFullYear() === now.getFullYear() ? undefined : 'numeric',
  })
}

export function ExpenseList({ group, expenses, onAdd, onEdit }: Props) {
  if (!expenses) {
    return (
      <div className="flex flex-col gap-2">
        {[0, 1, 2].map((i) => (
          <Skeleton key={i} className="h-14 w-full" />
        ))}
      </div>
    )
  }

  if (expenses.length === 0) {
    return (
      <div className="flex flex-col items-start gap-3 rounded-xl border border-dashed p-8">
        <div className="flex flex-col gap-1">
          <p className="font-medium">No expenses yet</p>
          <p className="text-muted-foreground">Add the first one and balances will start to build up.</p>
        </div>
        <Button size="sm" variant="outline" onClick={onAdd}>
          <Plus />
          Add expense
        </Button>
      </div>
    )
  }

  const names = Object.fromEntries(group.members.map((m) => [m.id, m.name]))
  const days: { day: string; items: Expense[] }[] = []
  for (const e of expenses) {
    const last = days[days.length - 1]
    if (last?.day === e.spent_on) last.items.push(e)
    else days.push({ day: e.spent_on, items: [e] })
  }

  return (
    <div className="flex flex-col gap-6">
      {days.map(({ day, items }) => (
        <section key={day} className="flex flex-col gap-1">
          <h3 className="px-3 pb-1 text-xs font-medium text-muted-foreground">{dayLabel(day)}</h3>
          {items.map((e) => (
            <button
              key={e.id}
              onClick={() => onEdit(e)}
              className="flex w-full items-center gap-4 rounded-lg px-3 py-2.5 text-left transition-colors hover:bg-accent focus-visible:bg-accent focus-visible:outline-none"
            >
              <div className="min-w-0 flex-1">
                <p className="truncate font-medium">{e.description}</p>
                <p className="truncate text-xs text-muted-foreground">
                  {names[e.paid_by] ?? 'Someone'} paid,{' '}
                  {e.split_type === 'equal' ? `split ${e.shares.length} ways` : 'custom split'}
                </p>
              </div>
              <span className="font-mono tabular">{formatMoney(e.amount_pence, group.currency)}</span>
            </button>
          ))}
        </section>
      ))}
    </div>
  )
}
