import { useSyncExternalStore } from 'react'

export type RecentGroup = { slug: string; name: string; visitedAt: number }

const KEY = 'splitit:groups'
const listeners = new Set<() => void>()
let cache: RecentGroup[] | null = null

function read(): RecentGroup[] {
  if (cache) return cache
  try {
    cache = JSON.parse(localStorage.getItem(KEY) ?? '[]')
  } catch {
    cache = []
  }
  return cache!
}

function persist(groups: RecentGroup[]) {
  try {
    localStorage.setItem(KEY, JSON.stringify(groups))
    return true
  } catch {
    return false
  }
}

function write(groups: RecentGroup[]) {
  cache = groups
  persist(groups)
  listeners.forEach((l) => l())
}

export function rememberGroup(slug: string, name: string) {
  const others = read().filter((g) => g.slug !== slug)
  write([{ slug, name, visitedAt: Date.now() }, ...others].slice(0, 12))
}

export function forgetGroup(slug: string) {
  write(read().filter((g) => g.slug !== slug))
}

export function useRecentGroups() {
  return useSyncExternalStore(
    (l) => {
      listeners.add(l)
      return () => listeners.delete(l)
    },
    read,
  )
}
