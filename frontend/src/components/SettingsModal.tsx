import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { AlertCircle, Plus, Trash2, X } from 'lucide-react'
import { api } from '../lib/api'
import type { UserSettings } from '../types'

export function SettingsModal({ open, settings, onClose, onSaved }: { open: boolean; settings?: UserSettings; onClose: () => void; onSaved: () => void }) {
  const queryClient = useQueryClient()
  const [target, setTarget] = useState('')
  const [meals, setMeals] = useState<string[]>(['Desayuno', 'Almuerzo', 'Merienda', 'Cena'])
  const [selectedFoodId, setSelectedFoodId] = useState(0)
  const [foodCalories, setFoodCalories] = useState('')
  const [foodDetail, setFoodDetail] = useState('')
  const foods = useQuery({ queryKey: ['foods'], queryFn: () => api.foods(), enabled: open })
  const mutation = useMutation({ mutationFn: api.saveSettings, onSuccess: onSaved })
  const foodMutation = useMutation({
    mutationFn: () => api.updateFood(selectedFoodId, { caloriesPer100g: Number(foodCalories), referenceDetail: foodDetail }),
    onSuccess: async () => { await queryClient.invalidateQueries({ queryKey: ['foods'] }) },
  })

  useEffect(() => {
    if (!open) return
    setTarget(settings?.dailyCalorieTarget?.toString() ?? '')
    setMeals(settings?.defaultMeals ?? ['Desayuno', 'Almuerzo', 'Merienda', 'Cena'])
    mutation.reset()
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, settings])

  useEffect(() => {
    if (!open || !foods.data?.length) return
    const selected = foods.data.find((food) => food.id === selectedFoodId) ?? foods.data[0]
    setSelectedFoodId(selected.id)
    setFoodCalories(String(selected.caloriesPer100g))
    setFoodDetail(selected.referenceDetail)
  // Keep the active edit stable while the query refreshes after saving.
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, foods.data])

  function selectFood(id: number) {
    const food = foods.data?.find((value) => value.id === id)
    if (!food) return
    setSelectedFoodId(food.id)
    setFoodCalories(String(food.caloriesPer100g))
    setFoodDetail(food.referenceDetail)
    foodMutation.reset()
  }

  function submit(event: React.FormEvent) {
    event.preventDefault()
    const normalizedMeals = meals.map((meal) => meal.trim()).filter(Boolean)
    mutation.mutate({
      dailyCalorieTarget: target ? Number(target) : null,
      timezone: settings?.timezone ?? 'America/Argentina/Cordoba',
      defaultMeals: normalizedMeals,
    })
  }

  if (!open) return null
  return (
    <div className="modal-backdrop" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <section className="modal settings-modal" role="dialog" aria-modal="true" aria-labelledby="settings-title">
        <div className="modal-header"><div><h2 id="settings-title">Configuración</h2><p>Definí el objetivo con el que querés comparar tus registros.</p></div><button className="icon-button" onClick={onClose} aria-label="Cerrar"><X size={18} /></button></div>
        <form onSubmit={submit}>
          <div className="settings-row">
            <div><h3>Objetivo diario</h3><p>Es un valor manual de referencia. La aplicación no lo interpreta como tu gasto energético real.</p></div>
            <div className="settings-well">
              <label className="field"><span>Calorías por día</span><div className="joined-input"><input type="number" min="500" max="10000" step="10" value={target} onChange={(event) => setTarget(event.target.value)} placeholder="Sin configurar" /><span>kcal</span></div></label>
              <p>Podés dejarlo vacío. No generamos recomendaciones automáticas ni restricciones.</p>
            </div>
          </div>
          <div className="settings-divider" />
          <div className="settings-row">
            <div><h3>Checklist habitual</h3><p>Estas tareas se crean cuando abrís una fecha nueva. Los días existentes no se modifican.</p></div>
            <div className="settings-well">
              <div className="meal-settings-list">
                {meals.map((meal, index) => <div className="joined-row" key={index}><input aria-label={`Comida ${index + 1}`} value={meal} maxLength={60} onChange={(event) => setMeals((current) => current.map((value, position) => position === index ? event.target.value : value))} /><button type="button" className="icon-button" disabled={meals.length === 1} onClick={() => setMeals((current) => current.filter((_, position) => position !== index))} aria-label={`Eliminar ${meal}`}><Trash2 size={15} /></button></div>)}
              </div>
              <button type="button" className="text-button" onClick={() => setMeals((current) => [...current, ''])}><Plus size={14} /> Agregar comida habitual</button>
            </div>
          </div>
          <div className="settings-divider" />
          <div className="settings-row">
            <div><h3>Referencias nutricionales</h3><p>Corregí el catálogo según tu receta o la etiqueta del producto. Los registros anteriores conservan el valor que usaron.</p></div>
            <div className="settings-well">
              <label className="field"><span>Alimento</span><select value={selectedFoodId} onChange={(event) => selectFood(Number(event.target.value))}>{foods.data?.map((food) => <option value={food.id} key={food.id}>{food.name}</option>)}</select></label>
              <label className="field catalog-field"><span>Calorías cada 100 g</span><input type="number" min="0" max="2000" step="0.1" value={foodCalories} onChange={(event) => setFoodCalories(event.target.value)} /></label>
              <label className="field catalog-field"><span>Fuente o motivo del ajuste</span><textarea rows={3} maxLength={500} value={foodDetail} onChange={(event) => setFoodDetail(event.target.value)} /></label>
              {foodMutation.isError && <div className="inline-error"><AlertCircle size={15} /> {foodMutation.error.message}</div>}
              {foodMutation.isSuccess && <p className="saved-message">Referencia guardada. El historial no fue modificado.</p>}
              <button type="button" className="secondary-button catalog-save" disabled={!selectedFoodId || foodMutation.isPending || foodDetail.trim().length < 5} onClick={() => foodMutation.mutate()}>{foodMutation.isPending ? 'Guardando…' : 'Guardar referencia'}</button>
            </div>
          </div>
          {mutation.isError && <div className="inline-error"><AlertCircle size={15} /> {mutation.error.message}</div>}
          <div className="modal-footer simple-footer"><button type="button" className="text-button" onClick={onClose}>Cancelar</button><button className="primary-button" disabled={mutation.isPending || meals.every((meal) => !meal.trim())}>{mutation.isPending ? 'Guardando…' : 'Guardar cambios'}</button></div>
        </form>
      </section>
    </div>
  )
}

