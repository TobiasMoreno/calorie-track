import { useEffect, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { AlertCircle, Plus, X } from 'lucide-react'
import { api } from '../lib/api'
import type { FoodPortion, FoodPortionInput } from '../types'

export function QuickPortionCreator({ foodId, onCreated }: { foodId: number; onCreated: (portion: FoodPortion) => void }) {
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const [name, setName] = useState('')
  const [grams, setGrams] = useState('')
  const [unit, setUnit] = useState<FoodPortionInput['unit']>('porción')
  const [isDefault, setIsDefault] = useState(false)
  const mutation = useMutation({
    mutationFn: () => api.createPortion(foodId, { name: name.trim(), grams: Number(grams), unit, size: 'personal', isDefault }),
    onSuccess: async (portion) => {
      await queryClient.invalidateQueries({ queryKey: ['foods'] })
      onCreated(portion)
      setOpen(false)
      setName('')
      setGrams('')
      setUnit('porción')
      setIsDefault(false)
    },
  })

  useEffect(() => {
    setOpen(false)
    mutation.reset()
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [foodId])

  if (!open) {
    return <button type="button" className="quick-portion-toggle" onClick={() => setOpen(true)}><Plus size={13} /> Nueva porción personal</button>
  }

  const valid = name.trim().length >= 2 && Number(grams) > 0
  return (
    <div className="quick-portion-editor">
      <div className="quick-portion-heading"><strong>Nueva porción personal</strong><button type="button" onClick={() => setOpen(false)} aria-label="Cerrar editor de porción"><X size={14} /></button></div>
      <label className="field"><span>Nombre</span><input value={name} maxLength={80} onChange={(event) => setName(event.target.value)} placeholder="Mi plato normal" /></label>
      <label className="field"><span>Equivale a</span><div className="portion-grams-input"><input type="number" min="0.1" max="5000" step="0.1" value={grams} onChange={(event) => setGrams(event.target.value)} placeholder="250" /><span>g</span></div></label>
      <label className="field"><span>Tipo</span><select value={unit} onChange={(event) => setUnit(event.target.value as FoodPortionInput['unit'])}><option value="porción">Porción</option><option value="unidad">Unidad</option><option value="cucharada">Cucharada</option></select></label>
      <label className="check-field"><input type="checkbox" checked={isDefault} onChange={(event) => setIsDefault(event.target.checked)} /><span>Usar automáticamente para este alimento</span></label>
      {mutation.isError && <div className="inline-error"><AlertCircle size={14} /> {mutation.error.message}</div>}
      <button type="button" className="secondary-button" disabled={!valid || mutation.isPending} onClick={() => mutation.mutate()}>{mutation.isPending ? 'Guardando…' : 'Guardar y usar'}</button>
    </div>
  )
}
