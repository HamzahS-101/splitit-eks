const formatters = new Map<string, Intl.NumberFormat>()

export function formatMoney(pence: number, currency = 'GBP') {
  let f = formatters.get(currency)
  if (!f) {
    f = new Intl.NumberFormat('en-GB', { style: 'currency', currency })
    formatters.set(currency, f)
  }
  return f.format(pence / 100)
}

export function parsePounds(input: string): number | null {
  const v = input.trim().replace(/^£/, '').replace(/,/g, '')
  if (!/^\d+(\.\d{1,2})?$/.test(v)) return null
  const [whole, frac = ''] = v.split('.')
  return Number(whole) * 100 + Number(frac.padEnd(2, '0'))
}

export function penceToInput(pence: number) {
  return (pence / 100).toFixed(2)
}

export function splitEqually(pence: number, count: number) {
  if (count === 0) return []
  const base = Math.floor(pence / count)
  const remainder = pence % count
  return Array.from({ length: count }, (_, i) => base + (i < remainder ? 1 : 0))
}
