import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { AlertCircle, ChevronLeft, ChevronRight, RefreshCw } from 'lucide-react'
import { api } from '../lib/api'
import { addMonths, formatMonth, todayInArgentina } from '../lib/date'
import type { DaySummary } from '../types'

const weekdayNames = ['Lun', 'Mar', 'Mié', 'Jue', 'Vie', 'Sáb', 'Dom']
const kcal = new Intl.NumberFormat('es-AR', { maximumFractionDigits: 0 })

export function CalendarScreen({ focusDate, onSelectDate }: { focusDate: string; onSelectDate: (date: string) => void }) {
  const [month, setMonth] = useState(focusDate.slice(0, 7))
  const calendar = useQuery({ queryKey: ['calendar', month], queryFn: () => api.calendar(month) })
  const today = todayInArgentina()
  const days = calendar.data?.days ?? []
  const leadingDays = days.length ? (new Date(`${days[0].date}T12:00:00Z`).getUTCDay() + 6) % 7 : 0
  const recorded = days.filter((day) => day.hasData)
  const incomplete = days.filter((day) => day.taskCount > 0 && day.pendingTasks > 0)
  const average = recorded.length ? recorded.reduce((sum, day) => sum + day.calories, 0) / recorded.length : null

  return (
    <section className="history-screen" aria-labelledby="history-title">
      <div className="title-row history-title-row">
        <div><h1 id="history-title">Historial</h1><p>Los días sin registros permanecen separados de un consumo real de cero.</p></div>
        <button className="secondary-button" onClick={() => { const value = todayInArgentina(); setMonth(value.slice(0, 7)) }}><span>Mes actual</span></button>
      </div>

      <div className="history-summary" aria-label="Resumen del mes">
        <div><span>Días con registros</span><strong className="num">{recorded.length}</strong></div>
        <div><span>Días incompletos</span><strong className="num">{incomplete.length}</strong></div>
        <div><span>Promedio registrado</span><strong className="num">{average === null ? '—' : `${kcal.format(average)} kcal`}</strong></div>
      </div>

      <div className="calendar-panel">
        <div className="calendar-toolbar">
          <button className="icon-button" onClick={() => setMonth(addMonths(month, -1))} aria-label="Mes anterior"><ChevronLeft size={17} /></button>
          <h2>{formatMonth(month)}</h2>
          <button className="icon-button" onClick={() => setMonth(addMonths(month, 1))} aria-label="Mes siguiente"><ChevronRight size={17} /></button>
        </div>

        {calendar.isLoading && <div className="calendar-loading"><div className="skeleton" /><div className="skeleton" /><div className="skeleton" /></div>}
        {calendar.isError && <div className="calendar-error"><AlertCircle size={17} /><span>No se pudo cargar el mes.</span><button className="text-button" onClick={() => calendar.refetch()}><RefreshCw size={14} /> Reintentar</button></div>}
        {calendar.data && (
          <>
            <div className="calendar-weekdays" aria-hidden="true">{weekdayNames.map((name) => <span key={name}>{name}</span>)}</div>
            <div className="calendar-grid">
              {Array.from({ length: leadingDays }, (_, index) => <span className="calendar-spacer" key={`spacer-${index}`} />)}
              {days.map((day) => <CalendarDay key={day.date} day={day} today={today} onSelect={() => onSelectDate(day.date)} />)}
            </div>
            <div className="calendar-legend">
              <span><i className="legend-complete" /> Completo</span>
              <span><i className="legend-incomplete" /> Pendiente</span>
              <span><i className="legend-empty" /> Sin consultar</span>
            </div>
          </>
        )}
      </div>
    </section>
  )
}

function CalendarDay({ day, today, onSelect }: { day: DaySummary; today: string; onSelect: () => void }) {
  const untouched = day.taskCount === 0
  const incomplete = !untouched && day.pendingTasks > 0
  const status = untouched ? 'empty' : incomplete ? 'incomplete' : 'complete'
  const overTarget = day.hasData && day.target !== null && day.calories > day.target
  const dayNumber = Number(day.date.slice(-2))
  return (
    <button className={`calendar-day calendar-${status} ${day.date === today ? 'is-today' : ''}`} onClick={onSelect} aria-label={`${day.date}, ${calendarDescription(day)}`}>
      <span className="calendar-day-number">{dayNumber}</span>
      {day.hasData ? <strong className={`num ${overTarget && !incomplete ? 'calendar-over' : ''}`}>{kcal.format(day.calories)}<small>kcal</small></strong> : <strong>—</strong>}
      <small>{untouched ? 'Sin consultar' : incomplete ? `${day.pendingTasks} pendiente${day.pendingTasks === 1 ? '' : 's'}` : day.hasData ? 'Completo' : 'Omitido'}</small>
    </button>
  )
}

function calendarDescription(day: DaySummary) {
  if (day.taskCount === 0) return 'sin información'
  if (day.pendingTasks > 0) return `${day.pendingTasks} comidas pendientes`
  if (day.hasData) return `${Math.round(day.calories)} kilocalorías registradas`
  return 'comidas omitidas voluntariamente'
}
