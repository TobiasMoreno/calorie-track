import { Bar, BarChart, CartesianGrid, ResponsiveContainer, Tooltip, YAxis } from 'recharts'
import { formatShortDay } from '../lib/date'
import type { DaySummary } from '../types'

export function SevenDayChart({ days }: { days: DaySummary[] }) {
  const data = days.map((day) => ({ ...day, label: formatShortDay(day.date), shownCalories: day.hasData ? day.calories : null }))
  const recorded = days.filter((day) => day.hasData)
  const average = recorded.length ? Math.round(recorded.reduce((sum, day) => sum + day.calories, 0) / recorded.length) : null
  return (
    <section className="chart-panel" aria-labelledby="week-title">
      <div className="panel-header">
        <div><h2 id="week-title">Últimos siete días</h2><p>{average === null ? 'Todavía no hay días registrados' : `Promedio de ${average} kcal en ${recorded.length} ${recorded.length === 1 ? 'día registrado' : 'días registrados'}`}</p></div>
      </div>
      <div className="chart-wrap">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={data} margin={{ top: 8, right: 8, bottom: 0, left: -10 }}>
            <CartesianGrid vertical={false} stroke="var(--grid)" strokeDasharray="3 5" />
            <YAxis axisLine={false} tickLine={false} allowDecimals={false} tick={{ fill: 'var(--faint)', fontSize: 11 }} width={46} />
            <Tooltip cursor={{ fill: 'var(--raised)', opacity: 0.5 }} content={({ active, payload, label }) => active && payload?.length ? <div className="chart-tooltip"><span>{label}</span><strong>{payload[0].value} kcal</strong></div> : null} />
            <Bar dataKey="shownCalories" fill="var(--accent)" radius={[3, 3, 0, 0]} maxBarSize={34} />
          </BarChart>
        </ResponsiveContainer>
      </div>
      <div className="chart-days" aria-hidden="true">{data.map((day) => <span key={day.date}>{day.label}</span>)}</div>
      {recorded.length < 7 && <p className="chart-footnote">Los días sin registros se muestran vacíos, no como 0 kcal.</p>}
    </section>
  )
}

