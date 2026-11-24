import { useEffect, useMemo, useRef, useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { AlertCircle, BookOpen, Plus, Trash2, X } from 'lucide-react'
import { api } from '../lib/api'
import type { CreateItemInput, Food, MealEntry } from '../types'
import { QuickPortionCreator } from './QuickPortionCreator'

interface EditDraftItem {
  key: number
  foodId: number
  quantity: number
  portionChoice: 'grams' | number
  cookingMethod: string
  notes: string
}

let editItemKey = 50_000
const emptyItem = (): EditDraftItem => ({ key: editItemKey++, foodId: 0, quantity: 1, portionChoice: 'grams', cookingMethod: '', notes: '' })

export function EditMealEntryModal({ entry, onClose, onSaved }: { entry: MealEntry | null; onClose: () => void; onSaved: () => void }) {
  const open = entry !== null
  const initializedEntry = useRef<number | null>(null)
  const [items, setItems] = useState<EditDraftItem[]>([])
  const [search, setSearch] = useState('')
  const foods = useQuery({ queryKey: ['foods'], queryFn: () => api.foods(), enabled: open })
  const mutation = useMutation({ mutationFn: ({ id, payload }: { id: number; payload: { items: CreateItemInput[] } }) => api.updateEntry(id, payload), onSuccess: onSaved })
  const foodMap = useMemo(() => new Map((foods.data ?? []).map((food) => [food.id, food])), [foods.data])

  useEffect(() => {
    initializedEntry.current = null
    mutation.reset()
    setSearch('')
    if (!open) setItems([])
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, entry?.id])

  useEffect(() => {
    if (!entry || !foods.data || initializedEntry.current === entry.id) return
    setItems(entry.items.map((item) => {
      const food = foods.data.find((value) => value.id === item.foodId)
      const portionExists = item.portionId && food?.portions.some((portion) => portion.id === item.portionId)
      return {
        key: editItemKey++,
        foodId: item.foodId,
        quantity: portionExists ? item.quantity : item.estimatedGrams,
        portionChoice: portionExists ? item.portionId! : 'grams',
        cookingMethod: item.cookingMethod ?? '',
        notes: item.notes ?? '',
      }
    }))
    initializedEntry.current = entry.id
  }, [entry, foods.data])

  useEffect(() => {
    if (!open) return
    const close = (event: KeyboardEvent) => event.key === 'Escape' && onClose()
    window.addEventListener('keydown', close)
    return () => window.removeEventListener('keydown', close)
  }, [open, onClose])

  if (!entry) return null

  const filteredFoods = (foods.data ?? []).filter((food) => food.name.toLowerCase().includes(search.toLowerCase()))
  const estimate = items.reduce((sum, item) => sum + estimateItem(item, foodMap.get(item.foodId)), 0)

  function updateItem(key: number, patch: Partial<EditDraftItem>) {
    setItems((current) => current.map((item) => item.key === key ? { ...item, ...patch } : item))
  }

  function chooseFood(item: EditDraftItem, foodId: number) {
    const food = foodMap.get(foodId)
    const portion = food?.portions.find((value) => value.isDefault) ?? food?.portions[0]
    updateItem(item.key, { foodId, portionChoice: portion?.id ?? 'grams', quantity: portion ? 1 : 100 })
  }

  function submit(event: React.FormEvent) {
    event.preventDefault()
    if (!items.length || items.some((item) => !item.foodId || item.quantity <= 0)) return
    const payload = items.map<CreateItemInput>((item) => {
      const food = foodMap.get(item.foodId)!
      const portion = item.portionChoice === 'grams' ? undefined : food.portions.find((value) => value.id === item.portionChoice)
      return { foodId: item.foodId, quantity: item.quantity, unit: portion?.unit ?? 'g', portionId: portion?.id, cookingMethod: item.cookingMethod, notes: item.notes }
    })
    mutation.mutate({ id: entry.id, payload: { items: payload } })
  }

  return (
    <div className="modal-backdrop" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <section className="modal meal-modal" role="dialog" aria-modal="true" aria-labelledby="edit-entry-title">
        <div className="modal-header">
          <div><h2 id="edit-entry-title">Editar {entry.mealName.toLowerCase()}</h2><p>Se recalculará solamente este registro del {formatEntryDate(entry.date)}.</p></div>
          <button className="icon-button" onClick={onClose} aria-label="Cerrar"><X size={18} /></button>
        </div>
        <form onSubmit={submit}>
          {entry.recipeName && <div className="recipe-origin-note"><BookOpen size={15} /><div><strong>Creada desde “{entry.recipeName}”</strong><p>Estos cambios afectan únicamente este día; la receta frecuente queda intacta.</p></div></div>}
          <div className="item-section-heading"><h3>Alimentos registrados</h3><span>{items.length} {items.length === 1 ? 'ítem' : 'ítems'}</span></div>
          <label className="field search-field"><span>Filtrar catálogo</span><input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Buscar alimento" /></label>
          <div className="food-items">
            {items.map((item, index) => {
              const food = foodMap.get(item.foodId)
              const portion = item.portionChoice === 'grams' ? undefined : food?.portions.find((value) => value.id === item.portionChoice)
              const grams = portion ? item.quantity * portion.grams : item.quantity
              return (
                <fieldset className="food-item" key={item.key}>
                  <legend>Alimento {index + 1}</legend>
                  <button type="button" className="remove-item" disabled={items.length === 1} onClick={() => setItems((current) => current.filter((value) => value.key !== item.key))} aria-label={`Quitar alimento ${index + 1}`}><Trash2 size={15} /></button>
                  <label className="field full-field"><span>Alimento</span><select value={item.foodId || ''} onChange={(event) => chooseFood(item, Number(event.target.value))} required><option value="" disabled>Seleccionar alimento</option>{filteredFoods.map((option) => <option value={option.id} key={option.id}>{option.name} · {option.caloriesPer100g} kcal/100 g</option>)}</select></label>
                  <label className="field"><span>Cantidad</span><input type="number" min="0.1" max="10000" step="0.1" value={item.quantity} onChange={(event) => updateItem(item.key, { quantity: Number(event.target.value) })} required /></label>
                  <label className="field"><span>Unidad o tamaño</span><select value={item.portionChoice} onChange={(event) => updateItem(item.key, { portionChoice: event.target.value === 'grams' ? 'grams' : Number(event.target.value) })}><option value="grams">Gramos</option>{food?.portions.map((value) => <option value={value.id} key={value.id}>{value.name} · {value.grams} g</option>)}</select></label>
                  {food && <QuickPortionCreator foodId={food.id} onCreated={(created) => updateItem(item.key, { portionChoice: created.id, quantity: 1 })} />}
                  <label className="field"><span>Cocción</span><select value={item.cookingMethod} onChange={(event) => updateItem(item.key, { cookingMethod: event.target.value })}><option value="">Sin especificar</option><option>Al horno</option><option>Frito</option><option>A la plancha</option><option>Hervido</option><option>Crudo</option></select></label>
                  <label className="field"><span>Nota o agregado</span><input value={item.notes} maxLength={160} onChange={(event) => updateItem(item.key, { notes: event.target.value })} placeholder="Marca, salsa, detalle…" /></label>
                  {food && <div className="food-reference"><strong>{Math.round(grams)} g estimados · {Math.round(estimateItem(item, food))} kcal</strong><p>{food.referenceSource}: {food.referenceDetail}</p></div>}
                </fieldset>
              )
            })}
          </div>
          <button type="button" className="add-item-button" onClick={() => setItems((current) => [...current, emptyItem()])}><Plus size={15} /> Agregar alimento</button>
          {foods.isError && <div className="inline-error"><AlertCircle size={15} /> No se pudo cargar el catálogo.</div>}
          {mutation.isError && <div className="inline-error"><AlertCircle size={15} /> {mutation.error.message}</div>}
          <div className="modal-footer"><div><span>Nuevo total estimado</span><strong className="num">{Math.round(estimate)} kcal</strong></div><button type="button" className="text-button" onClick={onClose}>Cancelar</button><button className="primary-button" disabled={mutation.isPending || foods.isLoading || !items.length}>{mutation.isPending ? 'Guardando…' : 'Guardar edición'}</button></div>
        </form>
      </section>
    </div>
  )
}

function estimateItem(item: EditDraftItem, food?: Food) {
  if (!food) return 0
  const portion = item.portionChoice === 'grams' ? undefined : food.portions.find((value) => value.id === item.portionChoice)
  const grams = portion ? item.quantity * portion.grams : item.quantity
  return grams * food.caloriesPer100g / 100
}

function formatEntryDate(date: string) {
  return new Intl.DateTimeFormat('es-AR', { day: 'numeric', month: 'long', year: 'numeric', timeZone: 'UTC' }).format(new Date(`${date}T12:00:00Z`))
}
