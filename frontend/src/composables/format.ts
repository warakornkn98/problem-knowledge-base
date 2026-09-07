/** Absolute date-time, e.g. "7 Sep 2026, 14:32". */
export function formatDateTime(iso: string | null | undefined): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '—'
  return d.toLocaleString(undefined, {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

/** Coarse relative time, e.g. "3 days ago", "just now". */
export function formatRelative(iso: string | null | undefined): string {
  if (!iso) return '—'
  const d = new Date(iso)
  const ms = Date.now() - d.getTime()
  if (Number.isNaN(ms)) return '—'
  const sec = Math.round(ms / 1000)
  const min = Math.round(sec / 60)
  const hr = Math.round(min / 60)
  const day = Math.round(hr / 24)
  if (sec < 45) return 'just now'
  if (min < 60) return `${min} min ago`
  if (hr < 24) return `${hr} hr ago`
  if (day < 30) return `${day} day${day === 1 ? '' : 's'} ago`
  return formatDateTime(iso)
}
