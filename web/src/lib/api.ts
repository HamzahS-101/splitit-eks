export type Member = { id: string; name: string }

export type Group = {
  slug: string
  name: string
  currency: string
  created_at: string
  members: Member[]
}

export type Share = { member_id: string; amount_pence: number }

export type SplitType = 'equal' | 'custom'

export type Expense = {
  id: string
  paid_by: string
  description: string
  amount_pence: number
  split_type: SplitType
  spent_on: string
  created_at: string
  shares: Share[]
}

export type ExpenseInput = {
  description: string
  amount_pence: number
  paid_by: string
  spent_on?: string
  split_type: SplitType
  member_ids?: string[]
  shares?: Share[]
}

export type Balance = {
  member_id: string
  name: string
  paid_pence: number
  share_pence: number
  net_pence: number
}

export type Transfer = { from_member: string; to_member: string; amount_pence: number }

export type Settlement = Transfer & { id: string; created_at: string }

export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`/api${path}`, {
    ...init,
    headers: init?.body ? { 'Content-Type': 'application/json' } : undefined,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => null)
    throw new ApiError(res.status, body?.error ?? 'Something went wrong. Try again.')
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

const json = (method: string, body: unknown): RequestInit => ({ method, body: JSON.stringify(body) })

export const api = {
  createGroup: (name: string, members: string[]) =>
    request<Group>('/groups', json('POST', { name, members })),
  getGroup: (slug: string) => request<Group>(`/groups/${slug}`),
  addMember: (slug: string, name: string) =>
    request<Member>(`/groups/${slug}/members`, json('POST', { name })),

  listExpenses: (slug: string) =>
    request<{ expenses: Expense[] }>(`/groups/${slug}/expenses`).then((r) => r.expenses),
  createExpense: (slug: string, input: ExpenseInput) =>
    request<Expense>(`/groups/${slug}/expenses`, json('POST', input)),
  updateExpense: (slug: string, id: string, input: ExpenseInput) =>
    request<Expense>(`/groups/${slug}/expenses/${id}`, json('PUT', input)),
  deleteExpense: (slug: string, id: string) =>
    request<void>(`/groups/${slug}/expenses/${id}`, { method: 'DELETE' }),

  getBalances: (slug: string) =>
    request<{ balances: Balance[]; transfers: Transfer[] }>(`/groups/${slug}/balances`),

  listSettlements: (slug: string) =>
    request<{ settlements: Settlement[] }>(`/groups/${slug}/settlements`).then((r) => r.settlements),
  createSettlement: (slug: string, t: Transfer) =>
    request<Settlement>(`/groups/${slug}/settlements`, json('POST', t)),
  deleteSettlement: (slug: string, id: string) =>
    request<void>(`/groups/${slug}/settlements/${id}`, { method: 'DELETE' }),
}
