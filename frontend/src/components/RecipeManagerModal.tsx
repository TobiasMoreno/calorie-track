import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { AlertCircle, ArrowLeft, Pencil, Plus, Trash2, X } from 'lucide-react'
import { api } from '../lib/api'
import type { CreateItemInput, Food, Recipe, RecipeInput } from '../types'
import { QuickPortionCreator } from './QuickPortionCreator'

interface RecipeDraftItem {
  key: number
  foodId: number
  quantity: number
  portionChoice: 'grams' | number
  cookingMethod: string
  notes: string
}

let recipeItemKey = 10_000
const emptyIngredient = (): RecipeDraftItem => ({ key: recipeItemKey++, foodId: 0, quantity: 1, portionChoice: 'grams', cookingMethod: '', notes: '' })

export function RecipeManagerModal({ open, onClose, onChanged }: { open: boolean; onClose: () => void; onChanged: () => void }) {
  const queryClient = useQueryClient()
  const recipes = useQuery({ queryKey: ['recipes'], queryFn: () => api.recipes(), enabled: open })
  const foods = useQuery({ queryKey: ['foods'], queryFn: () => api.foods(), enabled: open })
  const foodMap = useMemo(() => new Map((foods.data ?? []).map((food) => [food.id, food])), [foods.data])
  const [editingId, setEditingId] = useState<number | null | 'new'>(null)
  const [name, setName] = useState('')
  const [servings, setServings] = useState(2)
  const [notes, setNotes] = useState('')
  const [items, setItems] = useState<RecipeDraftItem[]>([emptyIngredient()])
  const [foodSearch, setFoodSearch] = useState('')
  const [recipeSearch, setRecipeSearch] = useState('')

  const saveMutation = useMutation({
    mutationFn: (input: RecipeInput) => editingId === 'new' ? api.createRecipe(input) : api.updateRecipe(Number(editingId), input),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['recipes'] })
      setEditingId(null)
      onChanged()
    },
  })
  const deleteMutation = useMutation({
    mutationFn: api.deleteRecipe,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['recipes'] })
      onChanged()
    },
  })

  useEffect(() => {
    if (!open) return
    setEditingId(null)
    setRecipeSearch('')
    saveMutation.reset()
    deleteMutation.reset()
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  function startNew() {
    setName('')
    setServings(2)
    setNotes('')
    setItems([emptyIngredient()])
    setFoodSearch('')
    setEditingId('new')
  }

  function startEdit(recipe: Recipe) {
    setName(recipe.name)
    setServings(recipe.servings)
    setNotes(recipe.notes ?? '')
    setItems(recipe.ingredients.map((ingredient) => {
      const food = foodMap.get(ingredient.foodId)
      const portionExists = ingredient.portionId && food?.portions.some((portion) => portion.id === ingredient.portionId)
      return {
        key: recipeItemKey++,
        foodId: ingredient.foodId,
        quantity: portionExists ? ingredient.quantity : ingredient.estimatedGrams,
        portionChoice: portionExists ? ingredient.portionId! : 'grams',
        cookingMethod: ingredient.cookingMethod ?? '',
        notes: ingredient.notes ?? '',
      }
    }))
    setFoodSearch('')
    setEditingId(recipe.id)
  }

  function updateItem(key: number, patch: Partial<RecipeDraftItem>) {
    setItems((current) => current.map((item) => item.key === key ? { ...item, ...patch } : item))
  }

  function chooseFood(item: RecipeDraftItem, foodId: number) {
    const food = foodMap.get(foodId)
    const portion = food?.portions.find((value) => value.isDefault) ?? food?.portions[0]
    updateItem(item.key, { foodId, portionChoice: portion?.id ?? 'grams', quantity: portion ? 1 : 100 })
  }

  function submit(event: React.FormEvent) {
    event.preventDefault()
    if (!name.trim() || servings <= 0 || items.some((item) => !item.foodId || item.quantity <= 0)) return
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
    saveMutation.mutate({ name: name.trim(), servings, notes: notes.trim(), items: payload })
  }

  if (!open) return null
  const visibleRecipes = (recipes.data ?? []).filter((recipe) => recipe.name.toLowerCase().includes(recipeSearch.toLowerCase()))
  const visibleFoods = (foods.data ?? []).filter((food) => food.name.toLowerCase().includes(foodSearch.toLowerCase()))
  const total = items.reduce((sum, item) => sum + estimateDraft(item, foodMap.get(item.foodId)), 0)

  return (
    <div className="modal-backdrop" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <section className="modal recipe-modal" role="dialog" aria-modal="true" aria-labelledby="recipes-title">
        <div className="modal-header">
          <div className="modal-title-with-back">
            {editingId !== null && <button className="icon-button" onClick={() => setEditingId(null)} aria-label="Volver a recetas"><ArrowLeft size={17} /></button>}
            <div><h2 id="recipes-title">{editingId === null ? 'Comidas frecuentes' : editingId === 'new' ? 'Nueva receta' : 'Editar receta'}</h2><p>{editingId === null ? 'Guardá una preparación una vez y registrala en segundos.' : 'Las modificaciones sólo afectarán registros futuros.'}</p></div>
          </div>
          <button className="icon-button" onClick={onClose} aria-label="Cerrar"><X size={18} /></button>
        </div>

        {editingId === null ? (
          <div className="recipe-library">
            <div className="recipe-library-tools">
              <label className="field"><span>Buscar preparación</span><input value={recipeSearch} onChange={(event) => setRecipeSearch(event.target.value)} placeholder="Milanesas caseras, desayuno…" /></label>
              <button className="primary-button" onClick={startNew}><Plus size={16} /> Nueva receta</button>
            </div>
            {recipes.isLoading && <div className="recipe-empty">Cargando preparaciones…</div>}
            {recipes.isError && <div className="inline-error"><AlertCircle size={15} /> No se pudieron cargar las recetas.</div>}
            {deleteMutation.isError && <div className="inline-error"><AlertCircle size={15} /> {deleteMutation.error.message}</div>}
            {!recipes.isLoading && visibleRecipes.length === 0 && <div className="recipe-empty"><strong>Todavía no guardaste preparaciones.</strong><p>Creá la primera con los alimentos y cantidades que usás habitualmente.</p><button className="secondary-button" onClick={startNew}>Crear primera receta</button></div>}
            <div className="recipe-list">
              {visibleRecipes.map((recipe) => (
                <article className="recipe-row" key={recipe.id}>
                  <div><h3>{recipe.name}</h3><p>{recipe.ingredients.map((ingredient) => ingredient.foodName).join(' · ')}</p><span>{recipe.servings} {recipe.servings === 1 ? 'porción' : 'porciones'} · {Math.round(recipe.caloriesPerServing)} kcal por porción</span></div>
                  <strong className="num">{Math.round(recipe.totalCalories)} kcal</strong>
                  <div className="recipe-actions"><button onClick={() => startEdit(recipe)}><Pencil size={14} /> Editar</button><button className="danger-action" onClick={() => window.confirm(`¿Eliminar “${recipe.name}”? Los registros anteriores se conservarán.`) && deleteMutation.mutate(recipe.id)}><Trash2 size={14} /> Eliminar</button></div>
                </article>
              ))}
            </div>
          </div>
        ) : (
          <form onSubmit={submit}>
            <div className="recipe-basics">
              <label className="field"><span>Nombre de la preparación</span><input autoFocus maxLength={100} value={name} onChange={(event) => setName(event.target.value)} placeholder="Milanesas caseras con puré" required /></label>
              <label className="field"><span>Porciones que rinde</span><input type="number" min="0.1" max="100" step="0.1" value={servings} onChange={(event) => setServings(Number(event.target.value))} required /></label>
              <label className="field full-field"><span>Notas</span><input maxLength={500} value={notes} onChange={(event) => setNotes(event.target.value)} placeholder="Detalles que quieras recordar" /></label>
            </div>
            <div className="item-section-heading"><h3>Ingredientes</h3><span>{items.length} {items.length === 1 ? 'ingrediente' : 'ingredientes'}</span></div>
            <label className="field search-field"><span>Filtrar catálogo</span><input value={foodSearch} onChange={(event) => setFoodSearch(event.target.value)} placeholder="Buscar alimento" /></label>
            <div className="food-items">
              {items.map((item, index) => {
                const food = foodMap.get(item.foodId)
                const portion = item.portionChoice === 'grams' ? undefined : food?.portions.find((value) => value.id === item.portionChoice)
                const grams = portion ? item.quantity * portion.grams : item.quantity
                return (
                  <fieldset className="food-item" key={item.key}>
                    <legend>Ingrediente {index + 1}</legend>
                    <button type="button" className="remove-item" disabled={items.length === 1} onClick={() => setItems((current) => current.filter((value) => value.key !== item.key))} aria-label={`Quitar ingrediente ${index + 1}`}><Trash2 size={15} /></button>
                    <label className="field full-field"><span>Alimento</span><select value={item.foodId || ''} onChange={(event) => chooseFood(item, Number(event.target.value))} required><option value="" disabled>Seleccionar alimento</option>{visibleFoods.map((option) => <option value={option.id} key={option.id}>{option.name} · {option.caloriesPer100g} kcal/100 g</option>)}</select></label>
                    <label className="field"><span>Cantidad</span><input type="number" min="0.1" max="10000" step="0.1" value={item.quantity} onChange={(event) => updateItem(item.key, { quantity: Number(event.target.value) })} required /></label>
                    <label className="field"><span>Unidad o tamaño</span><select value={item.portionChoice} onChange={(event) => updateItem(item.key, { portionChoice: event.target.value === 'grams' ? 'grams' : Number(event.target.value) })}><option value="grams">Gramos</option>{food?.portions.map((value) => <option value={value.id} key={value.id}>{value.name} · {value.grams} g</option>)}</select></label>
                    {food && <QuickPortionCreator foodId={food.id} onCreated={(created) => updateItem(item.key, { portionChoice: created.id, quantity: 1 })} />}
                    <label className="field"><span>Cocción</span><select value={item.cookingMethod} onChange={(event) => updateItem(item.key, { cookingMethod: event.target.value })}><option value="">Sin especificar</option><option>Al horno</option><option>Frito</option><option>A la plancha</option><option>Hervido</option><option>Crudo</option></select></label>
                    <label className="field"><span>Nota</span><input maxLength={160} value={item.notes} onChange={(event) => updateItem(item.key, { notes: event.target.value })} placeholder="Marca, salsa, detalle…" /></label>
                    {food && <div className="food-reference"><strong>{Math.round(grams)} g · {Math.round(estimateDraft(item, food))} kcal</strong><p>{food.referenceSource}</p></div>}
                  </fieldset>
                )
              })}
            </div>
            <button type="button" className="add-item-button" onClick={() => setItems((current) => [...current, emptyIngredient()])}><Plus size={15} /> Agregar ingrediente</button>
            {saveMutation.isError && <div className="inline-error"><AlertCircle size={15} /> {saveMutation.error.message}</div>}
            <div className="modal-footer"><div><span>Total de la receta</span><strong className="num">{Math.round(total)} kcal · {servings > 0 ? Math.round(total / servings) : 0} por porción</strong></div><button type="button" className="text-button" onClick={() => setEditingId(null)}>Cancelar</button><button className="primary-button" disabled={saveMutation.isPending || !name.trim() || items.some((item) => !item.foodId)}>{saveMutation.isPending ? 'Guardando…' : 'Guardar receta'}</button></div>
          </form>
        )}
      </section>
    </div>
  )
}

function estimateDraft(item: RecipeDraftItem, food?: Food) {
  if (!food) return 0
  const portion = item.portionChoice === 'grams' ? undefined : food.portions.find((value) => value.id === item.portionChoice)
  const grams = portion ? item.quantity * portion.grams : item.quantity
  return grams * food.caloriesPer100g / 100
}

