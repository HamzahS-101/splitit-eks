import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { ArrowRight } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { api, type Group, type Transfer } from '@/lib/api'
import { formatMoney } from '@/lib/money'
import { useBalances, useRefreshGroup, useSettlements } from '@/lib/queries'
import { cn } from '@/lib/utils'

function SectionTitle({ children }: { children: React.ReactNode }) {
  return <h2 className="text-xs font-medium text-muted-foreground">{children}</h2>
}

export function BalancesPanel({ group }: { group: Group }) {
  const balances = useBalances(group.slug)
  const settlements = useSettlements(group.slug)
  const refresh = useRefreshGroup(group.slug)
  const [newMember, setNewMember] = useState('')

  const names = Object.fromEntries(group.members.map((m) => [m.id, m.name]))
  const money = (pence: number) => formatMoney(pence, group.currency)

  const markPaid = useMutation({
    mutationFn: (t: Transfer) => api.createSettlement(group.slug, t),
    onSuccess: (s) => {
      refresh()
      toast.success(`${names[s.from_member]} paid ${names[s.to_member]} ${money(s.amount_pence)}`)
    },
    onError: (error) => toast.error(error.message),
  })

  const undo = useMutation({
    mutationFn: (id: string) => api.deleteSettlement(group.slug, id),
    onSuccess: () => refresh(),
    onError: (error) => toast.error(error.message),
  })

  const addMember = useMutation({
    mutationFn: () => api.addMember(group.slug, newMember.trim()),
    onSuccess: (m) => {
      setNewMember('')
      refresh()
      toast.success(`${m.name} added`)
    },
    onError: (error) => toast.error(error.message),
  })

  if (!balances.data) {
    return (
      <div className="flex flex-col gap-3">
        {group.members.map((m) => (
          <Skeleton key={m.id} className="h-10 w-full" />
        ))}
      </div>
    )
  }

  const { balances: rows, transfers } = balances.data
  const max = Math.max(1, ...rows.map((b) => Math.abs(b.net_pence)))

  return (
    <div className="flex flex-col gap-10">
      <section className="flex flex-col gap-3">
        <SectionTitle>Balances</SectionTitle>
        <div className="flex flex-col gap-3">
          {rows.map((b) => {
            const width = `${(Math.abs(b.net_pence) / max) * 50}%`
            return (
              <div key={b.member_id} className="grid grid-cols-[minmax(0,6.5rem)_1fr_6.5rem] items-center gap-4">
                <span className="truncate">{b.name}</span>
                <div className="relative h-1.5 rounded-full bg-muted">
                  <div className="absolute inset-y-[-3px] left-1/2 w-px bg-border" />
                  {b.net_pence > 0 && (
                    <div
                      className="absolute inset-y-0 left-1/2 rounded-r-full bg-positive transition-[width] duration-500 ease-out"
                      style={{ width }}
                    />
                  )}
                  {b.net_pence < 0 && (
                    <div
                      className="absolute inset-y-0 right-1/2 rounded-l-full bg-negative transition-[width] duration-500 ease-out"
                      style={{ width }}
                    />
                  )}
                </div>
                <div className="text-right">
                  <p
                    className={cn(
                      'font-mono tabular',
                      b.net_pence > 0 && 'text-positive',
                      b.net_pence < 0 && 'text-negative',
                    )}
                  >
                    {b.net_pence > 0 ? '+' : b.net_pence < 0 ? '−' : ''}
                    {money(Math.abs(b.net_pence))}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {b.net_pence > 0 ? 'gets back' : b.net_pence < 0 ? 'owes' : 'settled up'}
                  </p>
                </div>
              </div>
            )
          })}
        </div>
      </section>

      <section className="flex flex-col gap-3">
        <SectionTitle>Settle up</SectionTitle>
        {transfers.length === 0 ? (
          <p className="text-muted-foreground">Everyone is square. Nothing to pay.</p>
        ) : (
          <div className="flex flex-col gap-2">
            {transfers.map((t) => (
              <div
                key={`${t.from_member}-${t.to_member}`}
                className="flex flex-wrap items-center gap-x-3 gap-y-2 rounded-lg border bg-card px-4 py-3"
              >
                <span className="font-medium">{names[t.from_member]}</span>
                <ArrowRight className="size-4 text-muted-foreground" />
                <span className="font-medium">{names[t.to_member]}</span>
                <span className="ml-auto font-mono tabular">{money(t.amount_pence)}</span>
                <Button
                  size="sm"
                  variant="secondary"
                  disabled={markPaid.isPending}
                  onClick={() => markPaid.mutate(t)}
                >
                  Mark paid
                </Button>
              </div>
            ))}
          </div>
        )}
        <p className="text-xs text-muted-foreground">
          The fewest payments that clear every balance. Pay each other however you like, then mark it here.
        </p>
      </section>

      {settlements.data && settlements.data.length > 0 && (
        <section className="flex flex-col gap-1">
          <SectionTitle>Payments made</SectionTitle>
          <div className="flex flex-col">
            {settlements.data.map((s) => (
              <div key={s.id} className="flex items-center gap-3 py-2">
                <span className="min-w-0 flex-1 truncate">
                  {names[s.from_member]} paid {names[s.to_member]}
                  <span className="text-muted-foreground">
                    {' '}
                    on {new Date(s.created_at).toLocaleDateString('en-GB', { day: 'numeric', month: 'short' })}
                  </span>
                </span>
                <span className="font-mono tabular">{money(s.amount_pence)}</span>
                <Button size="xs" variant="ghost" disabled={undo.isPending} onClick={() => undo.mutate(s.id)}>
                  Undo
                </Button>
              </div>
            ))}
          </div>
        </section>
      )}

      <section className="flex flex-col gap-3">
        <SectionTitle>People</SectionTitle>
        <div className="overflow-hidden rounded-lg border">
          <table className="w-full">
            <thead className="bg-muted text-xs text-muted-foreground">
              <tr>
                <th className="px-4 py-2 text-left font-medium">Name</th>
                <th className="px-4 py-2 text-right font-medium">Paid</th>
                <th className="px-4 py-2 text-right font-medium">Their share</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {rows.map((b) => (
                <tr key={b.member_id}>
                  <td className="truncate px-4 py-2.5">{b.name}</td>
                  <td className="px-4 py-2.5 text-right font-mono tabular">{money(b.paid_pence)}</td>
                  <td className="px-4 py-2.5 text-right font-mono text-muted-foreground tabular">
                    {money(b.share_pence)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <form
          className="flex gap-2"
          onSubmit={(e) => {
            e.preventDefault()
            if (newMember.trim() && !addMember.isPending) addMember.mutate()
          }}
        >
          <Input
            aria-label="New person's name"
            value={newMember}
            maxLength={40}
            onChange={(e) => setNewMember(e.target.value)}
            placeholder="Add someone"
            autoComplete="off"
          />
          <Button type="submit" variant="outline" disabled={!newMember.trim() || addMember.isPending}>
            Add
          </Button>
        </form>
      </section>
    </div>
  )
}
