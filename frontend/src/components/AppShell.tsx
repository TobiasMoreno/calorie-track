import { useEffect, useState, type ReactNode } from 'react'
import { BookOpen, CalendarDays, Moon, Settings, Sun } from 'lucide-react'

export function AppShell({ children, activeView, onOpenCalendar, onOpenRecipes, onOpenSettings, onGoToday }: { children: ReactNode; activeView: 'day' | 'calendar'; onOpenCalendar: () => void; onOpenRecipes: () => void; onOpenSettings: () => void; onGoToday: () => void }) {
  const [theme, setTheme] = useState<'dark' | 'light'>(() => (localStorage.getItem('calorietrack-theme') === 'light' ? 'light' : 'dark'))

  useEffect(() => {
    document.documentElement.dataset.theme = theme
    localStorage.setItem('calorietrack-theme', theme)
  }, [theme])

  return (
    <>
      <header className="topbar">
        <div className="topbar-inner">
          <button className="brand" onClick={onGoToday} aria-label="Ir a hoy">
            <span className="brand-mark" aria-hidden="true"><span /></span>
            <span>CalorieTrack</span>
          </button>
          <nav aria-label="Navegación principal">
            <button className={activeView === 'day' ? 'active' : ''} onClick={onGoToday}><span>Hoy</span></button>
            <button className={activeView === 'calendar' ? 'active' : ''} onClick={onOpenCalendar}><CalendarDays size={15} /> <span>Historial</span></button>
            <button onClick={onOpenRecipes}><BookOpen size={15} /> <span>Recetas</span></button>
            <button onClick={onOpenSettings}><Settings size={15} /> <span>Configuración</span></button>
            <button className="icon-button" onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')} aria-label={theme === 'dark' ? 'Usar tema claro' : 'Usar tema oscuro'}>
              {theme === 'dark' ? <Sun size={17} /> : <Moon size={17} />}
            </button>
          </nav>
        </div>
      </header>
      <main className="app-canvas">{children}</main>
    </>
  )
}

