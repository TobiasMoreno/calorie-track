import { Target } from 'lucide-react'
import type { Dashboard } from '../types'

const kcal = new Intl.NumberFormat('es-AR', { maximumFractionDigits: 0 })

export function CalorieSummary({ dashboard, onConfigure }: { dashboard: Dashboard; onConfigure: () => void }) {
  const { consumedCalories: consumed, targetCalories: target, remainingCalories: remaining } = dashboard
  const progress = target ? Math.min((consumed / target) * 100, 100) : 0
  const exceeded = remaining !== null && remaining < 0

  return (
    <section className="summary-panel" aria-labelledby="balance-title">
      <div className="panel-header">
        <div><h2 id="balance-title">Balance del día</h2><p>{target ? 'Comparado con tu objetivo configurado' : 'Falta configurar el objetivo diario'}</p></div>
        {!target && <button className="secondary-button" onClick={onConfigure}><Target size={15} /> <span>Configurar objetivo</span></button>}
      </div>

      <div className="kpi-strip">
        <div className="lead-kpi"><span>Registradas</span><strong className="num">{kcal.format(consumed)} <small>kcal</small></strong></div>
        <div><span>Objetivo</span><strong className="num">{target ? kcal.format(target) : '—'} <small>{target ? 'kcal' : ''}</small></strong></div>
        <div><span>{exceeded ? 'Sobre el objetivo' : 'Restantes'}</span><strong className={`num ${exceeded ? 'negative-number' : ''}`}>{remaining === null ? '—' : kcal.format(Math.abs(remaining))} <small>{remaining !== null ? 'kcal' : ''}</small></strong></div>
      </div>

      <div className="progress-copy"><span>{target ? `${Math.round(progress)}% registrado` : 'El progreso aparecerá al definir un objetivo'}</span>{target && <span>{kcal.format(consumed)} de {kcal.format(target)} kcal</span>}</div>
      <div className="progress-track" role="progressbar" aria-valuemin={0} aria-valuemax={target ?? undefined} aria-valuenow={consumed}>
        <span style={{ width: `${progress}%` }} className={exceeded ? 'over' : ''} />
      </div>
      <p className="summary-note">Este balance usa sólo lo que registraste. Una comida pendiente no se interpreta como consumo cero.</p>
    </section>
  )
}

