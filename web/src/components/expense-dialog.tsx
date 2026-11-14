import { useMemo, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { api, type Expense, type ExpenseInput, type Group, type SplitType } from '@/lib/api'
import { formatMoney, parsePounds, penceToInput, splitEqually } from '@/lib/money'
import { useRefreshGroup } from '@/lib/queries'
import { cn } from '@/lib/utils'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  group: Group
  expense?: Expense
}

function today() {
  const d = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

function MoneyInput(props: React.ComponentProps<typeof Input>) {
  return (
    <div className="relative">
      <span className="pointer-events-none absolute inset-y-0 left-2.5 flex items-center text-muted-foreground">£</span>
      <Input inputMode="decimal" autoComplete="off" {...props} className={cn('pl-6 font-mono tabular', props.className)} />
    </div>
  )
}

export function ExpenseDialog({ open, onOpenChange, group, expense }: Props) {
  const members = group.members
  const refresh = useRefreshGroup(group.slug)

  const [description, setDescription] = useState(expense?.description ?? '')
  const [amount, setAmount] = useState(expense ? penceToInput(expense.amount_pence) : '')
  const [paidBy, setPaidBy] = useState(expense?.paid_by ?? members[0]?.id ?? '')
  const [spentOn, setSpentOn] = useState(expense?.spent_on ?? today())
  const [splitType, setSplitType] = useState<SplitType>(expense?.split_type ?? 'equal')
  const [included, setIncluded] = useState<Set<string>>(
    () => new Set(expense?.split_type === 'equal' ? expense.shares.map((s) => s.member_id) : members.map((m) => m.id)),
  )
  const [custom, setCustom] = useState<Record<string, string>>(() =>
    expense?.split_type === 'custom'
      ? Object.fromEntries(expense.shares.map((s) => [s.member_id, penceToInput(s.amount_pence)]))
      : {},
  )
  const [confirmDelete, setConfirmDelete] = useState(false)

  const amountPence = parsePounds(amount)

  const equalShares = useMemo(() => {
    const ids = members.filter((m) => included.has(m.id)).map((m) => m.id)
    const amounts = splitEqually(amountPence ?? 0, ids.length)
    return Object.fromEntries(ids.map((id, i) => [id, amounts[i]]))
  }, [members, included, amountPence])

  const customParsed = members.map((m) => {
    const raw = custom[m.id]?.trim() ?? ''
    return { id: m.id, pence: raw === '' ? 0 : parsePounds(raw) }
  })
  const customValid = customParsed.every((c) => c.pence !== null)
  const customTotal = customParsed.reduce((sum, c) => sum + (c.pence ?? 0), 0)
  const remaining = (amountPence ?? 0) - customTotal

  const splitValid = splitType === 'equal' ? included.size > 0 : customValid && customTotal > 0 && remaining === 0
  const valid = description.trim() !== '' && amountPence !== null && amountPence > 0 && paidBy !== '' && splitValid

  const done = (message: string) => {
    refresh()
    onOpenChange(false)
    toast.success(message)
  }

  const save = useMutation({
    mutationFn: () => {
      const input: ExpenseInput = {
        description: description.trim(),
        amount_pence: amountPence!,
        paid_by: paidBy,
        spent_on: spentOn || undefined,
        split_type: splitType,
      }
      if (splitType === 'equal') {
        input.member_ids = members.filter((m) => included.has(m.id)).map((m) => m.id)
      } else {
        input.shares = customParsed
          .filter((c) => (c.pence ?? 0) > 0)
          .map((c) => ({ member_id: c.id, amount_pence: c.pence! }))
      }
      return expense ? api.updateExpense(group.slug, expense.id, input) : api.createExpense(group.slug, input)
    },
    onSuccess: () => done(expense ? 'Expense updated' : 'Expense added'),
    onError: (error) => toast.error(error.message),
  })

  const remove = useMutation({
    mutationFn: () => api.deleteExpense(group.slug, expense!.id),
    onSuccess: () => done('Expense deleted'),
    onError: (error) => toast.error(error.message),
  })

  const toggle = (id: string, on: boolean) =>
    setIncluded((prev) => {
      const next = new Set(prev)
      if (on) next.add(id)
      else next.delete(id)
      return next
    })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90dvh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{expense ? 'Edit expense' : 'Add expense'}</DialogTitle>
        </DialogHeader>

        <form
          id="expense-form"
          className="flex flex-col gap-4"
          onSubmit={(e) => {
            e.preventDefault()
            if (valid && !save.isPending) save.mutate()
          }}
        >
          <div className="flex flex-col gap-2">
            <Label htmlFor="description">What was it for?</Label>
            <Input
              id="description"
              value={description}
              maxLength={120}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Dinner at Taberna da Rua"
              autoComplete="off"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="flex flex-col gap-2">
              <Label htmlFor="amount">Amount</Label>
              <MoneyInput id="amount" value={amount} onChange={(e) => setAmount(e.target.value)} placeholder="0.00" />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="spent-on">Date</Label>
              <Input id="spent-on" type="date" value={spentOn} onChange={(e) => setSpentOn(e.target.value)} />
            </div>
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="paid-by">Paid by</Label>
            <Select value={paidBy} onValueChange={setPaidBy}>
              <SelectTrigger id="paid-by" className="w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {members.map((m) => (
                  <SelectItem key={m.id} value={m.id}>
                    {m.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="flex flex-col gap-2">
            <div className="flex items-center justify-between gap-3">
              <Label>Split</Label>
              <div role="radiogroup" aria-label="Split type" className="inline-flex rounded-lg bg-muted p-0.5">
                {(['equal', 'custom'] as const).map((type) => (
                  <button
                    key={type}
                    type="button"
                    role="radio"
                    aria-checked={splitType === type}
                    onClick={() => setSplitType(type)}
                    className={cn(
                      'rounded-md px-3 py-1 text-xs font-medium text-muted-foreground transition-colors',
                      splitType === type && 'bg-card text-foreground shadow-sm',
                    )}
                  >
                    {type === 'equal' ? 'Equally' : 'Custom amounts'}
                  </button>
                ))}
              </div>
            </div>

            <div className="flex flex-col divide-y rounded-lg border">
              {members.map((m) =>
                splitType === 'equal' ? (
                  <label key={m.id} className="flex cursor-pointer items-center gap-3 px-3 py-2.5">
                    <Checkbox checked={included.has(m.id)} onCheckedChange={(v) => toggle(m.id, v === true)} />
                    <span className="flex-1">{m.name}</span>
                    <span className="font-mono text-muted-foreground tabular">
                      {included.has(m.id) ? formatMoney(equalShares[m.id] ?? 0) : '-'}
                    </span>
                  </label>
                ) : (
                  <div key={m.id} className="flex items-center gap-3 px-3 py-2">
                    <Label htmlFor={`share-${m.id}`} className="flex-1 font-normal">
                      {m.name}
                    </Label>
                    <MoneyInput
                      id={`share-${m.id}`}
                      value={custom[m.id] ?? ''}
                      onChange={(e) => setCustom((prev) => ({ ...prev, [m.id]: e.target.value }))}
                      placeholder="0.00"
                      className="h-8 w-28 text-right"
                    />
                  </div>
                ),
              )}
            </div>

            {splitType === 'custom' && amountPence !== null && amountPence > 0 && (
              <p
                className={cn(
                  'text-right text-xs tabular',
                  remaining === 0 && customValid ? 'text-positive' : 'text-muted-foreground',
                  remaining < 0 && 'text-negative',
                )}
              >
                {!customValid
                  ? 'Check the amounts above'
                  : remaining === 0
                    ? 'Adds up'
                    : remaining > 0
                      ? `${formatMoney(remaining)} left to assign`
                      : `${formatMoney(-remaining)} too much`}
              </p>
            )}
          </div>
        </form>

        <DialogFooter className="sm:justify-between">
          {expense ? (
            <Button
              type="button"
              variant="ghost"
              className="text-destructive hover:text-destructive"
              disabled={remove.isPending}
              onClick={() => (confirmDelete ? remove.mutate() : setConfirmDelete(true))}
            >
              {confirmDelete ? 'Click again to delete' : 'Delete'}
            </Button>
          ) : (
            <span />
          )}
          <Button type="submit" form="expense-form" disabled={!valid || save.isPending}>
            {save.isPending ? 'Saving' : expense ? 'Save changes' : 'Add expense'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
