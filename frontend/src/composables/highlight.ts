/**
 * Server search headlines arrive with neutral [[hl]] / [[/hl]] markers around
 * matches. Escape the whole fragment first (it is raw DB text), then swap the
 * markers for <mark> tags. Safe to feed the result to v-html.
 */
export function renderHeadline(raw: string | undefined | null): string {
  if (!raw) return ''
  const escaped = raw
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
  return escaped.replaceAll('[[hl]]', '<mark>').replaceAll('[[/hl]]', '</mark>')
}
