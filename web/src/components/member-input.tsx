import { useState } from 'react'
import { X } from 'lucide-react'
import { toast } from 'sonner'

type Props = {
  id: string
  value: string[]
  onChange: (names: string[]) => void
  max?: number
}

export function MemberInput({ id, value, onChange, max = 20 }: Props) {
  const [draft, setDraft] = useState('')

  const add = () => {
    const name = draft.trim()
    if (!name) return
    if (value.some((v) => v.toLowerCase() === name.toLowerCase())) {
      toast.error(`${name} is already in the list`)
      return
    }
    if (value.length >= max) {
      toast.error(`A group can have up to ${max} people`)
      return
    }
    onChange([...value, name])
    setDraft('')
  }

  return (
    <div className="flex min-h-9 flex-wrap items-center gap-1.5 rounded-md border border-input px-2 py-1.5 transition-[color,box-shadow] focus-within:border-ring focus-within:ring-[3px] focus-within:ring-ring/50">
      {value.map((name) => (
        <span key={name} className="inline-flex items-center gap-1 rounded-md bg-secondary py-0.5 pr-1 pl-2">
          {name}
          <button
            type="button"
            aria-label={`Remove ${name}`}
            onClick={() => onChange(value.filter((v) => v !== name))}
            className="rounded p-0.5 text-muted-foreground transition-colors hover:text-foreground"
          >
            <X className="size-3" />
          </button>
        </span>
      ))}
      <input
        id={id}
        value={draft}
        maxLength={40}
        onChange={(e) => setDraft(e.target.value.replace(',', ''))}
        onBlur={add}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ',') {
            e.preventDefault()
            add()
          } else if (e.key === 'Backspace' && !draft && value.length > 0) {
            onChange(value.slice(0, -1))
          }
        }}
        placeholder={value.length ? 'Add another' : 'Amara, Tomás, Priya'}
        className="min-w-28 flex-1 bg-transparent px-1 py-0.5 outline-none placeholder:text-muted-foreground"
      />
    </div>
  )
}
