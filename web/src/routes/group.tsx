import { useEffect, useState } from 'react'
import { useParams, useSearchParams } from 'react-router'
import { Link2, Plus } from 'lucide-react'
import { toast } from 'sonner'
import { BalancesPanel } from '@/components/balances-panel'
import { ExpenseDialog } from '@/components/expense-dialog'
import { ExpenseList } from '@/components/expense-list'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import type { Expense } from '@/lib/api'
import { formatMoney } from '@/lib/money'
import { isNotFound, useExpenses, useGroup } from '@/lib/queries'
import { forgetGroup, rememberGroup } from '@/lib/recent'
import { NotFound } from '@/routes/not-found'

type DialogState = { open: boolean; expense?: Expense; key: number }

export function GroupPage() {
  const { slug = '' } = useParams()
  const [params, setParams] = useSearchParams()
  const group = useGroup(slug)
  const expenses = useExpenses(slug, group.isSuccess)
  const [dialog, setDialog] = useState<DialogState>({ open: false, key: 0 })

  const name = group.data?.name
  useEffect(() => {
    if (name) rememberGroup(slug, name)
  }, [slug, name])

  const missing = isNotFound(group.error)
  useEffect(() => {
    if (missing) forgetGroup(slug)
  }, [slug, missing])

  if (group.isPending) {
    return (
      <div className="flex flex-col gap-8">
        <div className="flex flex-col gap-2">
          <Skeleton className="h-7 w-48" />
          <Skeleton className="h-4 w-64" />
        </div>
        <Skeleton className="h-40 w-full" />
      </div>
    )
  }

  if (missing) return <NotFound group />

  if (group.isError) {
    return (
      <div className="flex flex-col items-start gap-3">
        <p className="font-medium">Couldn't load this group</p>
        <p className="text-muted-foreground">{group.error.message}</p>
        <Button size="sm" variant="outline" onClick={() => group.refetch()}>
          Try again
        </Button>
      </div>
    )
  }

  const g = group.data
  const tab = params.get('tab') === 'balances' ? 'balances' : 'expenses'
  const total = expenses.data?.reduce((sum, e) => sum + e.amount_pence, 0) ?? 0

  const openDialog = (expense?: Expense) => setDialog((d) => ({ open: true, expense, key: d.key + 1 }))

  const copyLink = async () => {
    try {
      await navigator.clipboard.writeText(window.location.origin + `/g/${g.slug}`)
      toast.success('Link copied. Anyone with it can view and edit this group.')
    } catch {
      toast.error("Couldn't copy. Copy the address from your browser instead.")
    }
  }

  return (
    <div className="flex flex-col gap-8">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex min-w-0 flex-col gap-1">
          <h1 className="truncate text-xl font-semibold tracking-tight">{g.name}</h1>
          <p className="text-muted-foreground">{g.members.map((m) => m.name).join(', ')}</p>
        </div>
        <div className="flex shrink-0 gap-2">
          <Button variant="outline" size="sm" onClick={copyLink}>
            <Link2 />
            Copy link
          </Button>
          <Button size="sm" onClick={() => openDialog()}>
            <Plus />
            Add expense
          </Button>
        </div>
      </header>

      <Tabs value={tab} onValueChange={(v) => setParams(v === 'balances' ? { tab: v } : {}, { replace: true })}>
        <div className="flex items-end justify-between gap-4 border-b">
          <TabsList variant="line">
            <TabsTrigger value="expenses">Expenses</TabsTrigger>
            <TabsTrigger value="balances">Balances</TabsTrigger>
          </TabsList>
          {expenses.data && expenses.data.length > 0 && (
            <p className="hidden pb-2 text-xs text-muted-foreground sm:block">
              <span className="font-mono text-foreground tabular">{formatMoney(total, g.currency)}</span> across{' '}
              {expenses.data.length} {expenses.data.length === 1 ? 'expense' : 'expenses'}
            </p>
          )}
        </div>
        <TabsContent value="expenses" className="pt-6">
          <ExpenseList group={g} expenses={expenses.data} onAdd={() => openDialog()} onEdit={openDialog} />
        </TabsContent>
        <TabsContent value="balances" className="pt-6">
          <BalancesPanel group={g} />
        </TabsContent>
      </Tabs>

      <ExpenseDialog
        key={dialog.key}
        open={dialog.open}
        onOpenChange={(open) => setDialog((d) => ({ ...d, open }))}
        group={g}
        expense={dialog.expense}
      />
    </div>
  )
}
