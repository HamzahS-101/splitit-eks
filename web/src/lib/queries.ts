import { useQuery, useQueryClient } from '@tanstack/react-query'
import { ApiError, api } from '@/lib/api'

export const keys = {
  all: (slug: string) => ['group', slug] as const,
  group: (slug: string) => ['group', slug, 'info'] as const,
  expenses: (slug: string) => ['group', slug, 'expenses'] as const,
  balances: (slug: string) => ['group', slug, 'balances'] as const,
  settlements: (slug: string) => ['group', slug, 'settlements'] as const,
}

export function isNotFound(error: unknown) {
  return error instanceof ApiError && error.status === 404
}

export function useGroup(slug: string) {
  return useQuery({
    queryKey: keys.group(slug),
    queryFn: () => api.getGroup(slug),
    retry: (count, error) => !isNotFound(error) && count < 2,
  })
}

export function useExpenses(slug: string, enabled = true) {
  return useQuery({ queryKey: keys.expenses(slug), queryFn: () => api.listExpenses(slug), enabled })
}

export function useBalances(slug: string) {
  return useQuery({ queryKey: keys.balances(slug), queryFn: () => api.getBalances(slug) })
}

export function useSettlements(slug: string) {
  return useQuery({ queryKey: keys.settlements(slug), queryFn: () => api.listSettlements(slug) })
}

export function useRefreshGroup(slug: string) {
  const qc = useQueryClient()
  return () => qc.invalidateQueries({ queryKey: keys.all(slug) })
}
