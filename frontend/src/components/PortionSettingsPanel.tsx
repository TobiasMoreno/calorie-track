import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { AlertCircle, ArrowLeft, Pencil, Plus, Star, Trash2 } from 'lucide-react'
import { api } from '../lib/api'
import type { Food, FoodPortion, FoodPortionInput } from '../types'

export function PortionSettingsPanel({ foods, selectedFoodId, onSelectFood }: { foods: Food[]; selectedFoodId: number; onSelectFood: (id: number) => void }) {
  const queryClient = useQueryClient()
  const food = foods.find((value) => value.id === selectedFoodId)
  const [editing, setEditing] = useState<FoodPortion | 'new' | null>(null)
  const [name, setName] = useState('')
  const [grams, setGrams] = useState('')
  const [unit, setUnit] = useState<FoodPortionInput['unit']>('porción')
  const [isDefault, setIsDefault] = useState(false)

  const refresh = async () => {
    await queryClient.invalidateQueries({ queryKey: ['foods'] })
  }
  const saveMutation = useMutation({
    mutationFn: () => {
      const input: FoodPortionInput = { name: name.trim(), grams: Number(grams), unit, size: 'personal', isDefault }
      return editing === 'new' ? api.createPortion(selectedFoodId, input) : api.updatePortion(editing!.id, input)
    },
    onSuccess: async () => { await refresh(); setEditing(null) },
  })
  const defaultMutation = useMutation({ mutationFn: api.setDefaultPortion, onSuccess: refresh })
  const deleteMutation = useMutation({ mutationFn: api.deletePortion, onSuccess: async () => { await refresh(); setEditing(null) } })

  function startNew() {
    setName('')
    setGrams('')
    setUnit('porción')
    setIsDefault(false)
    saveMutation.reset()
    setEditing('new')
  }

  function startEdit(portion: FoodPortion) {
    setName(portion.name)
    setGrams(String(portion.grams))
    setUnit(portion.unit)
    setIsDefault(portion.isDefault)
    saveMutation.reset()
    setEditing(portion)
  }

  if (!food) return <div className="portion-settings-empty">Seleccioná un alimento para administrar sus porciones.</div>

  if (editing) {
    const valid = name.trim().length >= 2 && Number(grams) > 0
    return (
      <div className="portion-settings-editor">
        <button type="button" className="text-button portion-back" onClick={() => setEditing(null)}><ArrowLeft size={14} /> Volver a porciones</button>
        <div className="portion-editor-grid">
          <label className="field"><span>Nombre personal</span><input autoFocus value={name} maxLength={80} onChange={(event) => setName(event.target.value)} placeholder="Mi plato normal" /></label>
          <label className="field"><span>Gramaje estimado</span><div className="portion-grams-input"><input type="number" min="0.1" max="5000" step="0.1" value={grams} onChange={(event) => setGrams(event.target.value)} /><span>g</span></div></label>
          <label className="field"><span>Tipo</span><select value={unit} onChange={(event) => setUnit(event.target.value as FoodPortionInput['unit'])}><option value="porción">Porción</option><option value="unidad">Unidad</option><option value="cucharada">Cucharada</option></select></label>
          <label className="check-field portion-default-check"><input type="checkbox" checked={isDefault} onChange={(event) => setIsDefault(event.target.checked)} /><span>Usar automáticamente</span></label>
        </div>
        {saveMutation.isError && <div className="inline-error"><AlertCircle size={14} /> {saveMutation.error.message}</div>}
        <button type="button" className="secondary-button catalog-save" disabled={!valid || saveMutation.isPending} onClick={() => saveMutation.mutate()}>{saveMutation.isPending ? 'Guardando…' : editing === 'new' ? 'Crear porción' : 'Guardar porción'}</button>
      </div>
    )
  }

  return (
    <div>
      <label className="field"><span>Alimento</span><select value={selectedFoodId} onChange={(event) => { onSelectFood(Number(event.target.value)); setEditing(null) }}>{foods.map((value) => <option value={value.id} key={value.id}>{value.name}</option>)}</select></label>
      <div className="portion-settings-list">
        {food.portions.map((portion) => (
          <div className="portion-settings-row" key={portion.id}>
            <div><strong>{portion.name}</strong><span>{portion.grams} g · {portion.unit}{portion.isCustom ? ' · Personal' : ' · Catálogo'}</span></div>
            {portion.isDefault ? <span className="default-portion"><Star size={12} /> Predeterminada</span> : <button type="button" className="text-button" disabled={defaultMutation.isPending} onClick={() => defaultMutation.mutate(portion.id)}><Star size={13} /> Usar por defecto</button>}
            {portion.isCustom && <div className="portion-row-actions"><button type="button" onClick={() => startEdit(portion)}><Pencil size={13} /> Editar</button><button type="button" className="danger-action" onClick={() => window.confirm(`¿Eliminar “${portion.name}”? Los registros anteriores conservarán sus datos.`) && deleteMutation.mutate(portion.id)}><Trash2 size={13} /> Eliminar</button></div>}
          </div>
        ))}
      </div>
      {(defaultMutation.isError || deleteMutation.isError) && <div className="inline-error"><AlertCircle size={14} /> {(defaultMutation.error ?? deleteMutation.error)?.message}</div>}
      <button type="button" className="add-item-button portion-add-button" onClick={startNew}><Plus size={14} /> Crear porción personal para {food.name}</button>
    </div>
  )
}
