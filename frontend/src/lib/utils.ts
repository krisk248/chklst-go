import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Robustly parse a timestamp into a Date.
 *
 * The backend stores timestamps as `YYYY-MM-DD HH:MM:SS.ffffff` (space separator,
 * 6-digit microseconds). Chromium parses this leniently (but as local time), while
 * Firefox returns `Invalid Date` for it. Either way the raw string is unsafe to feed
 * straight into `new Date()`. This normalizes to ISO-8601 (T separator, millisecond
 * precision) and falls back to manual component parsing.
 *
 * Returns `null` for empty/zero/unparseable values instead of an Invalid Date, so
 * callers can skip bad rows explicitly rather than silently producing NaN.
 */
export function parseTimestamp(value: string | Date | null | undefined): Date | null {
  if (!value) return null
  if (value instanceof Date) return isNaN(value.getTime()) ? null : value

  let s = String(value).trim()
  if (!s || s.startsWith('0001-')) return null // Go zero-value time

  // "YYYY-MM-DD HH:MM:SS.ffffff" -> "YYYY-MM-DDTHH:MM:SS.fff"
  s = s.replace(' ', 'T').replace(/(\.\d{3})\d+/, '$1')

  let d = new Date(s)
  if (isNaN(d.getTime())) {
    const m = s.match(/^(\d{4})-(\d{2})-(\d{2})[T ](\d{2}):(\d{2})(?::(\d{2}))?/)
    if (m) {
      d = new Date(+m[1], +m[2] - 1, +m[3], +m[4], +m[5], +(m[6] || 0))
    }
  }
  return isNaN(d.getTime()) ? null : d
}

export function formatDate(date: string | Date): string {
  const d = parseTimestamp(date)
  if (!d) return '—'
  return d.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

export function formatDateTime(date: string | Date): string {
  const d = parseTimestamp(date)
  if (!d) return '—'
  return d.toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function truncate(str: string, length: number): string {
  return str.length > length ? str.substring(0, length) + '...' : str
}

export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: NodeJS.Timeout

  return function executedFunction(...args: Parameters<T>) {
    const later = () => {
      clearTimeout(timeout)
      func(...args)
    }

    clearTimeout(timeout)
    timeout = setTimeout(later, wait)
  }
}
