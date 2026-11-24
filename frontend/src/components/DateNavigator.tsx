import { CalendarDays, ChevronLeft, ChevronRight } from 'lucide-react'
import { addDays, formatLongDate, todayInArgentina } from '../lib/date'

export function DateNavigator({ date, onChange }: { date: string; onChange: (date: string) => void }) {
  const today = todayInArgentina()
  const isToday = date === today
  return (
    <div className="date-navigator" aria-label="Navegar por fecha">
      <button className="icon-button" onClick={() => onChange(addDays(date, -1))} aria-label="Día anterior"><ChevronLeft size={18} /></button>
      <label className="date-display">
        <CalendarDays size={16} />
        <span>{isToday ? 'Hoy · ' : ''}{formatLongDate(date)}</span>
        <input type="date" value={date} onChange={(event) => event.target.value && onChange(event.target.value)} aria-label="Elegir fecha" />
      </label>
      <button className="icon-button" onClick={() => onChange(addDays(date, 1))} disabled={date >= today} aria-label="Día siguiente"><ChevronRight size={18} /></button>
    </div>
  )
}

