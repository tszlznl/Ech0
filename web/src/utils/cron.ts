import type { ComposerTranslation } from 'vue-i18n'

export type CronFrequency = 'daily' | 'weekly' | 'monthly' | 'hourly' | 'custom'

export interface ParsedCron {
  frequency: CronFrequency
  minute: number
  hour: number
  weekday: number
  monthday: number
  hourlyInterval: number
  custom: string
}

const pad = (n: number): string => String(n).padStart(2, '0')

const clamp = (n: number, min: number, max: number): number => {
  if (Number.isNaN(n)) return min
  return Math.min(Math.max(n, min), max)
}

export function parseCron(expr: string): ParsedCron {
  const src = (expr || '').trim()
  const base: ParsedCron = {
    frequency: 'custom',
    minute: 0,
    hour: 2,
    weekday: 0,
    monthday: 1,
    hourlyInterval: 6,
    custom: '',
  }

  const daily = /^(\d+)\s+(\d+)\s+\*\s+\*\s+\*$/.exec(src)
  if (daily) {
    return {
      ...base,
      frequency: 'daily',
      minute: clamp(parseInt(daily[1], 10), 0, 59),
      hour: clamp(parseInt(daily[2], 10), 0, 23),
    }
  }

  const weekly = /^(\d+)\s+(\d+)\s+\*\s+\*\s+([0-6])$/.exec(src)
  if (weekly) {
    return {
      ...base,
      frequency: 'weekly',
      minute: clamp(parseInt(weekly[1], 10), 0, 59),
      hour: clamp(parseInt(weekly[2], 10), 0, 23),
      weekday: clamp(parseInt(weekly[3], 10), 0, 6),
    }
  }

  const monthly = /^(\d+)\s+(\d+)\s+(\d+)\s+\*\s+\*$/.exec(src)
  if (monthly) {
    return {
      ...base,
      frequency: 'monthly',
      minute: clamp(parseInt(monthly[1], 10), 0, 59),
      hour: clamp(parseInt(monthly[2], 10), 0, 23),
      monthday: clamp(parseInt(monthly[3], 10), 1, 28),
    }
  }

  const hourly = /^0\s+\*\/(\d+)\s+\*\s+\*\s+\*$/.exec(src)
  if (hourly) {
    return {
      ...base,
      frequency: 'hourly',
      hourlyInterval: clamp(parseInt(hourly[1], 10), 1, 23),
    }
  }

  return { ...base, frequency: 'custom', custom: src }
}

export function buildCron(p: ParsedCron): string {
  switch (p.frequency) {
    case 'daily':
      return `${p.minute} ${p.hour} * * *`
    case 'weekly':
      return `${p.minute} ${p.hour} * * ${p.weekday}`
    case 'monthly':
      return `${p.minute} ${p.hour} ${p.monthday} * *`
    case 'hourly':
      return `0 */${p.hourlyInterval} * * *`
    case 'custom':
      return (p.custom || '').trim()
    default:
      return ''
  }
}

const WEEKDAY_KEYS = [
  'cronEditor.weekdaySun',
  'cronEditor.weekdayMon',
  'cronEditor.weekdayTue',
  'cronEditor.weekdayWed',
  'cronEditor.weekdayThu',
  'cronEditor.weekdayFri',
  'cronEditor.weekdaySat',
] as const

export function humanizeCron(expr: string, t: ComposerTranslation): string {
  const p = parseCron(expr)
  const time = `${pad(p.hour)}:${pad(p.minute)}`
  switch (p.frequency) {
    case 'daily':
      return t('cronEditor.humanDaily', { time })
    case 'weekly':
      return t('cronEditor.humanWeekly', {
        weekday: t(WEEKDAY_KEYS[p.weekday]),
        time,
      })
    case 'monthly':
      return t('cronEditor.humanMonthly', { day: p.monthday, time })
    case 'hourly':
      return t('cronEditor.humanHourly', { n: p.hourlyInterval })
    case 'custom':
      return t('cronEditor.humanCustom')
    default:
      return ''
  }
}
