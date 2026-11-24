import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { AlertCircle, BookOpen, Plus, Settings2, Trash2, X } from 'lucide-react'
import { api } from '../lib/api'
import type { CreateItemInput, DailyMealTask, Food } from '../types'
import { QuickPortionCreator } from './QuickPortionCreator'

interface DraftItem {
  key: number
  foodId: number
  quantity: number
  portionChoice: 'grams' | number
  cookingMethod: string
  notes: string
}

let nextKey = 1
const newItem = (): DraftItem => ({ key: nextKey++, foodId: 0, quantity: 1, portionChoice: 'grams', cookingMethod: '', notes: '' })

export function MealEntryModal({ open, date, tasks, initialTaskId, onClose, onSaved, onManageRecipes }: {
  open: boolean
  date: string
  tasks: DailyMealTask[]
  initialTaskId: number | null
  onClose: () => void
  onSaved: () => void
  onManageRecipes: () => void
}) {
  const availableTasks = tasks.filter((task) => task.status !== 'registered')
  const [mode, setMode] = useState<'manual' | 'frequent'>('manual')
  const [taskId, setTaskId] = useState(0)
  const [items, setItems] = useState<DraftItem[]>([newItem()])
  const [search, setSearch] = useState('')
  const [selectedRecipeId, setSelectedRecipeId] = useState(0)
  const [consumedServings, setConsumedServings] = useState(1)
  const foods = useQuery({ queryKey: ['foods'], queryFn: () => api.foods(), enabled: open && mode === 'manual' })
  const recipes = useQuery({ queryKey: ['recipes'], queryFn: () => api.recipes(), enabled: open })
  const manualMutation = useMutation({ mutationFn: api.createEntry, onSuccess: onSaved })
  const recipeMutation = useMutation({ mutationFn: api.createRecipeEntry, onSuccess: onSaved })

  useEffect(() => {
    if (!open) return
    const validInitial = initialTaskId && availableTasks.some((task) => task.id === initialTaskId)
    setTaskId(validInitial ? initialTaskId : (availableTasks[0]?.id ?? 0))
    setMode('manual')
    setItems([newItem()])
    setSearch('')
    setSelectedRecipeId(0)
    setConsumedServings(1)
    manualMutation.reset()
    recipeMutation.reset()
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, initialTaskId, date])

  useEffect(() => {
    if (!open) return
    const close = (event: KeyboardEvent) => event.key === 'Escape' && onClose()
    window.addEventListener('keydown', close)
    return () => window.removeEventListener('keydown', close)
  }, [open, onClose])

  const foodMap = useMemo(() => new Map((foods.data ?? []).map((food) => [food.id, food])), [foods.data])
  const filteredFoods = (foods.data ?? []).filter((food) => food.name.toLowerCase().includes(search.toLowerCase()))
  const selectedRecipe = recipes.data?.find((recipe) => recipe.id === selectedRecipeId)
  const manualEstimate = items.reduce((total, item) => total + estimateItem(item, foodMap.get(item.foodId)), 0)
  const recipeEstimate = (selectedRecipe?.caloriesPerServing ?? 0) * consumedServings
  const activeMutation = mode === 'manual' ? manualMutation : recipeMutation

  function updateItem(key: number, patch: Partial<DraftItem>) {
    setItems((current) => current.map((item) => item.key === key ? { ...item, ...patch } : item))
  }

  function chooseFood(item: DraftItem, foodId: number) {
    const food = foodMap.get(foodId)
    const defaultPortion = food?.portions.find((portion) => portion.isDefault) ?? food?.portions[0]
    updateItem(item.key, { foodId, portionChoice: defaultPortion?.id ?? 'grams', quantity: defaultPortion ? 1 : 100 })
  }

  function submit(event: React.FormEvent) {
    event.preventDefault()
    if (!taskId) return

    if (mode === 'frequent') {
      if (!selectedRecipe || consumedServings <= 0) return
      recipeMutation.mutate({
        date,
        taskId,
        recipeId: selectedRecipe.id,
        multiplier: consumedServings / selectedRecipe.servings,
      })
      return
    }

    if (items.some((item) => !item.foodId || item.quantity <= 0)) return
    const payload: CreateItemInput[] = items.map((item) => {
      const food = foodMap.get(item.foodId)!
      const portion = item.portionChoice === 'grams' ? undefined : food.portions.find((value) => value.id === item.portionChoice)
      return {
        foodId: item.foodId,
        quantity: item.quantity,
        unit: portion?.unit ?? 'g',
        portionId: portion?.id,
        cookingMethod: item.cookingMethod,
        notes: item.notes,
      }
    })
    manualMutation.mutate({ date, taskId, items: payload })
  }

  if (!open) return null

  return (
    <div className="modal-backdrop" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <section className="modal meal-modal" role="dialog" aria-modal="true" aria-labelledby="entry-modal-title">
        <div className="modal-header">
          <div><h2 id="entry-modal-title">Registrar comida</h2><p>Usá una preparación frecuente o detallá los alimentos.</p></div>
          <button className="icon-button" onClick={onClose} aria-label="Cerrar"><X size={18} /></button>
        </div>

        <form onSubmit={submit}>
          <label className="field">
            <span>Momento del día</span>
            <select value={taskId} onChange={(event) => setTaskId(Number(event.target.value))} required>
              {availableTasks.map((task) => <option value={task.id} key={task.id}>{task.name}{task.status === 'skipped' ? ' · omitida' : ''}</option>)}
            </select>
          </label>

          <div className="entry-mode-tabs" role="tablist" aria-label="Tipo de registro">
            <button type="button" role="tab" aria-selected={mode === 'manual'} className={mode === 'manual' ? 'active' : ''} onClick={() => setMode('manual')}><Plus size={15} /> Manual</button>
            <button type="button" role="tab" aria-selected={mode === 'frequent'} className={mode === 'frequent' ? 'active' : ''} onClick={() => setMode('frequent')}><BookOpen size={15} /> Frecuente</button>
          </div>

          {mode === 'manual' ? (
            <>
              <div className="item-section-heading"><h3>Alimentos</h3><span>{items.length} {items.length === 1 ? 'ítem' : 'ítems'}</span></div>
              <label className="field search-field"><span>Filtrar catálogo</span><input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Ej. milanesa, arroz, banana" /></label>

              <div className="food-items">
                {items.map((item, index) => {
                  const food = foodMap.get(item.foodId)
                  const portion = item.portionChoice === 'grams' ? undefined : food?.portions.find((value) => value.id === item.portionChoice)
                  const grams = item.portionChoice === 'grams' ? item.quantity : item.quantity * (portion?.grams ?? 0)
                  return (
                    <fieldset className="food-item" key={item.key}>
                      <legend>Alimento {index + 1}</legend>
                      <button type="button" className="remove-item" onClick={() => setItems((current) => current.filter((value) => value.key !== item.key))} disabled={items.length === 1} aria-label={`Quitar alimento ${index + 1}`}><Trash2 size={15} /></button>
                      <label className="field full-field"><span>Alimento</span><select value={item.foodId || ''} onChange={(event) => chooseFood(item, Number(event.target.value))} required><option value="" disabled>Seleccionar alimento</option>{filteredFoods.map((option) => <option key={option.id} value={option.id}>{option.name} · {option.caloriesPer100g} kcal/100 g</option>)}</select></label>
                      <label className="field"><span>Cantidad</span><input type="number" min="0.1" max="10000" step="0.1" value={item.quantity} onChange={(event) => updateItem(item.key, { quantity: Number(event.target.value) })} required /></label>
                      <label className="field"><span>Unidad o tamaño</span><select value={item.portionChoice} onChange={(event) => updateItem(item.key, { portionChoice: event.target.value === 'grams' ? 'grams' : Number(event.target.value) })}><option value="grams">Gramos</option>{food?.portions.map((value) => <option value={value.id} key={value.id}>{value.name} · {value.grams} g</option>)}</select></label>
                      {food && <QuickPortionCreator foodId={food.id} onCreated={(portion) => updateItem(item.key, { portionChoice: portion.id, quantity: 1 })} />}
                      <label className="field"><span>Cocción</span><select value={item.cookingMethod} onChange={(event) => updateItem(item.key, { cookingMethod: event.target.value })}><option value="">Sin especificar</option><option>Al horno</option><option>Frito</option><option>A la plancha</option><option>Hervido</option><option>Crudo</option></select></label>
                      <label className="field"><span>Nota o agregado</span><input value={item.notes} maxLength={160} onChange={(event) => updateItem(item.key, { notes: event.target.value })} placeholder="Marca, salsa, detalle…" /></label>
                      {food && <div className="food-reference"><strong>{Math.round(grams)} g estimados · {Math.round(estimateItem(item, food))} kcal</strong><p>{food.referenceSource}: {food.referenceDetail}</p></div>}
                    </fieldset>
                  )
                })}
              </div>

              <button type="button" className="add-item-button" onClick={() => setItems((current) => [...current, newItem()])}><Plus size={15} /> Agregar otro alimento o ingrediente</button>
              {foods.isError && <div className="inline-error"><AlertCircle size={15} /> No se pudo cargar el catálogo.</div>}
            </>
          ) : (
            <div className="frequent-picker">
              <div className="frequent-heading">
                <div><h3>Preparaciones guardadas</h3><p>Elegí una receta y ajustá solamente la cantidad consumida.</p></div>
                <button type="button" className="text-button" onClick={onManageRecipes}><Settings2 size={15} /> Administrar</button>
              </div>

              {recipes.isLoading && <div className="frequent-empty">Cargando preparaciones…</div>}
              {recipes.isError && <div className="inline-error"><AlertCircle size={15} /> No se pudieron cargar las preparaciones.</div>}
              {!recipes.isLoading && (recipes.data?.length ?? 0) === 0 && (
                <div className="frequent-empty">
                  <BookOpen size={22} />
                  <strong>Todavía no hay comidas frecuentes.</strong>
                  <p>Guardá una preparación con sus ingredientes para registrarla en pocos segundos.</p>
                  <button type="button" className="secondary-button" onClick={onManageRecipes}>Crear primera receta</button>
                </div>
              )}

              <div className="frequent-list" role="radiogroup" aria-label="Preparación frecuente">
                {recipes.data?.map((recipe) => (
                  <button
                    type="button"
                    role="radio"
                    aria-checked={selectedRecipeId === recipe.id}
                    className={`frequent-option ${selectedRecipeId === recipe.id ? 'selected' : ''}`}
                    key={recipe.id}
                    onClick={() => { setSelectedRecipeId(recipe.id); setConsumedServings(1) }}
                  >
                    <span><strong>{recipe.name}</strong><small>{recipe.ingredients.map((ingredient) => ingredient.foodName).join(' · ')}</small></span>
                    <b className="num">{Math.round(recipe.caloriesPerServing)} kcal<small>por porción</small></b>
                  </button>
                ))}
              </div>

              {selectedRecipe && (
                <div className="recipe-register-controls">
                  <label className="field"><span>Porciones consumidas</span><input type="number" min="0.1" max="20" step="0.1" value={consumedServings} onChange={(event) => setConsumedServings(Number(event.target.value))} required /></label>
                  <div><span>La receta rinde</span><strong className="num">{selectedRecipe.servings} {selectedRecipe.servings === 1 ? 'porción' : 'porciones'}</strong></div>
                  <p>Se registrarán proporcionalmente todos sus ingredientes. Podés corregir la receta desde “Administrar”.</p>
                </div>
              )}
            </div>
          )}

          {activeMutation.isError && <div className="inline-error"><AlertCircle size={15} /> {activeMutation.error.message}</div>}

          <div className="modal-footer">
            <div><span>Estimación total</span><strong className="num">{Math.round(mode === 'manual' ? manualEstimate : recipeEstimate)} kcal</strong></div>
            <button type="button" className="text-button" onClick={onClose}>Cancelar</button>
            <button className="primary-button" disabled={activeMutation.isPending || !taskId || (mode === 'manual' ? foods.isLoading : !selectedRecipe || consumedServings <= 0)}>{activeMutation.isPending ? 'Guardando…' : 'Confirmar registro'}</button>
          </div>
        </form>
      </section>
    </div>
  )
}

function estimateItem(item: DraftItem, food?: Food) {
  if (!food) return 0
  const portion = item.portionChoice === 'grams' ? undefined : food.portions.find((value) => value.id === item.portionChoice)
  const grams = portion ? item.quantity * portion.grams : item.quantity
  return grams * food.caloriesPer100g / 100
}
