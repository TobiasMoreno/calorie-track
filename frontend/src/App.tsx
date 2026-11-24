import { lazy, Suspense, useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { AlertCircle, Plus, RefreshCw } from 'lucide-react'
import { api } from './lib/api'
import { todayInArgentina } from './lib/date'
import { AppShell } from './components/AppShell'
import { DateNavigator } from './components/DateNavigator'
import { CalorieSummary } from './components/CalorieSummary'
import { MealChecklist } from './components/MealChecklist'
import { MealEntryModal } from './components/MealEntryModal'
import { RecipeManagerModal } from './components/RecipeManagerModal'
import { SettingsModal } from './components/SettingsModal'
import { Toast } from './components/Toast'
import { CalendarScreen } from './components/CalendarScreen'
import { EditMealEntryModal } from './components/EditMealEntryModal'
import type { MealEntry } from './types'

const SevenDayChart = lazy(() => import('./components/SevenDayChart').then((module) => ({ default: module.SevenDayChart })))

export default function App() {
  const queryClient = useQueryClient()
  const [date, setDate] = useState(todayInArgentina)
  const [view, setView] = useState<'day' | 'calendar'>('day')
  const [entryTaskId, setEntryTaskId] = useState<number | null>(null)
  const [entryModalOpen, setEntryModalOpen] = useState(false)
  const [editingEntry, setEditingEntry] = useState<MealEntry | null>(null)
  const [recipesOpen, setRecipesOpen] = useState(false)
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [toast, setToast] = useState<string | null>(null)

  const dashboard = useQuery({ queryKey: ['dashboard', date], queryFn: () => api.dashboard(date), enabled: view === 'day' })
  const settings = useQuery({ queryKey: ['settings'], queryFn: api.settings })

  useEffect(() => {
    if (toast) {
      const timer = window.setTimeout(() => setToast(null), 3200)
      return () => window.clearTimeout(timer)
    }
  }, [toast])

  const refreshData = () => {
    queryClient.invalidateQueries({ queryKey: ['dashboard'] })
    queryClient.invalidateQueries({ queryKey: ['calendar'] })
  }

  const statusMutation = useMutation({
    mutationFn: ({ id, status }: { id: number; status: 'pending' | 'skipped' }) => api.updateTask(id, status),
    onSuccess: refreshData,
  })
  const deleteMutation = useMutation({
    mutationFn: api.deleteEntry,
    onSuccess: () => { refreshData(); setToast('Registro eliminado') },
  })
  const extraTaskMutation = useMutation({
    mutationFn: (name: string) => api.createTask(date, name),
    onSuccess: () => { refreshData(); setToast('Comida adicional creada') },
  })

  const openEntry = (taskId?: number) => {
    setEntryTaskId(taskId ?? null)
    setEntryModalOpen(true)
  }

  return (
    <AppShell
      activeView={view}
      onOpenCalendar={() => setView('calendar')}
      onOpenRecipes={() => setRecipesOpen(true)}
      onOpenSettings={() => setSettingsOpen(true)}
      onGoToday={() => { setDate(todayInArgentina()); setView('day') }}
    >
      {view === 'day' && dashboard.isLoading && <DashboardSkeleton />}

      {view === 'day' && dashboard.isError && (
        <div className="state-panel" role="alert">
          <AlertCircle size={18} />
          <div><strong>No pudimos cargar tu día.</strong><p>{dashboard.error.message}</p></div>
          <button className="text-button" onClick={() => dashboard.refetch()}><RefreshCw size={15} /> Reintentar</button>
        </div>
      )}

      {view === 'day' && dashboard.data && (
        <>
          <section className="title-row" aria-labelledby="dashboard-title">
            <div>
              <h1 id="dashboard-title">{date === todayInArgentina() ? 'Tu día' : 'Detalle del día'}</h1>
              <p>{date === todayInArgentina() ? 'Registrá lo que comés; lo pendiente nunca cuenta como cero.' : 'Revisá y corregí este día sin modificar el resto del historial.'}</p>
            </div>
            <button className="primary-button" onClick={() => openEntry()}>
              <Plus size={17} /> <span>Registrar comida</span>
            </button>
          </section>

          <DateNavigator date={date} onChange={setDate} />

          <div className="dashboard-grid">
            <div className="main-column">
              <CalorieSummary
                dashboard={dashboard.data}
                onConfigure={() => setSettingsOpen(true)}
              />
              <Suspense fallback={<div className="skeleton skeleton-chart" aria-label="Cargando resumen semanal" />}>
                <SevenDayChart days={dashboard.data.lastSevenDays} />
              </Suspense>
            </div>

            <aside className="side-column" aria-labelledby="meals-title">
              <div className="section-heading">
                <div><h2 id="meals-title">Comidas del día</h2><p>{dashboard.data.tasks.filter((task) => task.status === 'pending').length} pendientes</p></div>
              </div>
              <MealChecklist
                tasks={dashboard.data.tasks}
                entries={dashboard.data.entries}
                onRegister={openEntry}
                onEdit={setEditingEntry}
                onStatus={(id, status) => statusMutation.mutate({ id, status })}
                onDelete={(id) => deleteMutation.mutate(id)}
                onAddTask={(name) => extraTaskMutation.mutateAsync(name)}
                busy={statusMutation.isPending || deleteMutation.isPending || extraTaskMutation.isPending}
              />
            </aside>
          </div>
        </>
      )}

      {view === 'calendar' && <CalendarScreen focusDate={date} onSelectDate={(selectedDate) => { setDate(selectedDate); setView('day') }} />}

      <MealEntryModal
        open={entryModalOpen}
        date={date}
        tasks={dashboard.data?.tasks ?? []}
        initialTaskId={entryTaskId}
        onClose={() => setEntryModalOpen(false)}
        onSaved={() => { setEntryModalOpen(false); refreshData(); setToast('Comida registrada') }}
        onManageRecipes={() => { setEntryModalOpen(false); setRecipesOpen(true) }}
      />

      <EditMealEntryModal
        entry={editingEntry}
        onClose={() => setEditingEntry(null)}
        onSaved={() => { setEditingEntry(null); refreshData(); setToast('Registro actualizado') }}
      />

      <RecipeManagerModal
        open={recipesOpen}
        onClose={() => setRecipesOpen(false)}
        onChanged={() => setToast('Recetas actualizadas')}
      />

      <SettingsModal
        open={settingsOpen}
        settings={settings.data}
        onClose={() => setSettingsOpen(false)}
        onSaved={() => {
          setSettingsOpen(false)
          queryClient.invalidateQueries({ queryKey: ['settings'] })
          refreshData()
          setToast('Objetivo actualizado')
        }}
      />

      {toast && <Toast message={toast} />}
    </AppShell>
  )
}

function DashboardSkeleton() {
  return (
    <div className="loading-layout" aria-label="Cargando dashboard">
      <div className="skeleton skeleton-title" />
      <div className="skeleton skeleton-controls" />
      <div className="dashboard-grid">
        <div className="main-column"><div className="skeleton skeleton-panel" /><div className="skeleton skeleton-chart" /></div>
        <div className="skeleton skeleton-side" />
      </div>
    </div>
  )
}

