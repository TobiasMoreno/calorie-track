import { useState } from 'react'
import { BookOpen, Check, Circle, Pencil, Plus, RotateCcw, SkipForward, Trash2 } from 'lucide-react'
import type { DailyMealTask, MealEntry, TaskStatus } from '../types'

const kcal = new Intl.NumberFormat('es-AR', { maximumFractionDigits: 0 })

export function MealChecklist({ tasks, entries, onRegister, onEdit, onStatus, onDelete, onAddTask, busy }: {
  tasks: DailyMealTask[]
  entries: MealEntry[]
  onRegister: (taskId: number) => void
  onEdit: (entry: MealEntry) => void
  onStatus: (taskId: number, status: Exclude<TaskStatus, 'registered'>) => void
  onDelete: (entryId: number) => void
  onAddTask: (name: string) => Promise<unknown>
  busy: boolean
}) {
  const [adding, setAdding] = useState(false)
  const [name, setName] = useState('Snack')
  const entryByTask = new Map(entries.map((entry) => [entry.taskId, entry]))

  async function addTask(event: React.FormEvent) {
    event.preventDefault()
    if (!name.trim()) return
    await onAddTask(name.trim())
    setName('Snack')
    setAdding(false)
  }

  return (
    <div className="checklist-panel">
      <div className="meal-list">
        {tasks.map((task) => {
          const entry = entryByTask.get(task.id)
          return (
            <article className={`meal-row status-${task.status}`} key={task.id}>
              <div className="meal-state" aria-hidden="true">
                {task.status === 'registered' ? <Check size={16} /> : task.status === 'skipped' ? <SkipForward size={15} /> : <Circle size={14} />}
              </div>
              <div className="meal-content">
                <div className="meal-title-line">
                  <div><h3>{task.name}</h3><span>{task.status === 'registered' ? 'Registrada' : task.status === 'skipped' ? 'Omitida por decisión' : 'Pendiente'}</span></div>
                  {entry && <strong className="num">{kcal.format(entry.totalCalories)} kcal</strong>}
                </div>
                {entry && (
                  <div className="entry-details">
                    {entry.recipeName && <p className="recipe-entry-label"><BookOpen size={12} /> {entry.recipeName}</p>}
                    {entry.items.map((item) => <p key={item.id}>{item.foodName} · {item.estimatedGrams} g <span>{Math.round(item.calories)} kcal</span></p>)}
                  </div>
                )}
                <div className="row-actions">
                  {task.status === 'pending' && <><button onClick={() => onRegister(task.id)} disabled={busy}>Registrar</button><button onClick={() => onStatus(task.id, 'skipped')} disabled={busy}>Omitir</button></>}
                  {task.status === 'skipped' && <><button onClick={() => onRegister(task.id)} disabled={busy}>Registrar</button><button onClick={() => onStatus(task.id, 'pending')} disabled={busy}><RotateCcw size={13} /> Volver a pendiente</button></>}
                  {entry && <><button onClick={() => onEdit(entry)} disabled={busy}><Pencil size={13} /> Editar</button><button className="danger-action" onClick={() => window.confirm(`¿Eliminar el registro de ${task.name}?`) && onDelete(entry.id)} disabled={busy}><Trash2 size={13} /> Eliminar</button></>}
                </div>
              </div>
            </article>
          )
        })}
      </div>

      {adding ? (
        <form className="inline-form" onSubmit={addTask}>
          <label><span className="sr-only">Nombre de la comida adicional</span><input autoFocus value={name} maxLength={60} onChange={(event) => setName(event.target.value)} /></label>
          <button className="secondary-button" disabled={busy}>Agregar</button>
          <button type="button" className="text-button" onClick={() => setAdding(false)}>Cancelar</button>
        </form>
      ) : (
        <button className="add-meal-button" onClick={() => setAdding(true)}><Plus size={15} /> Agregar comida o snack</button>
      )}
    </div>
  )
}

