import type { RecordMap } from '../types/dbf'

export function rowToObject(columns: string[], row: unknown[]): RecordMap {
  const o: RecordMap = {}
  for (let i = 0; i < columns.length && i < row.length; i++) {
    o[columns[i].toUpperCase()] = row[i]
  }
  return o
}

export const firstOf = (o: RecordMap, keys: string[]): string =>
  keys.find((k) => o[k] !== undefined) ?? keys[0]

export const toNumber = (v: unknown): number =>
  typeof v === 'number' ? v : Number.parseFloat(String(v ?? 0)) || 0

export const toBool = (v: unknown): boolean =>
  typeof v === 'boolean' ? v : String(v).toLowerCase() === 'true'

export const toDate = (v: unknown): Date | null => {
  if (v instanceof Date) return v
  const s = String(v ?? '')
  if (!s) return null
  const d = new Date(s)
  return Number.isNaN(d.getTime()) ? null : d
}
