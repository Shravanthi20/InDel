import { useLocation } from 'react-router-dom'
import { useTheme } from '../../context/ThemeContext'
import { Search, Moon, Sun } from 'lucide-react'
import { useLocalization } from '../../context/LocalizationContext'

export default function Navbar() {
  const { theme, toggleTheme } = useTheme()
  const { t } = useLocalization()
  const location = useLocation()

  const segmentLabelMap: Record<string, string> = {
    workers: t('sidebar.workers'),
    zones: t('sidebar.zones'),
    analytics: t('sidebar.analytics'),
    batches: t('sidebar.viewBatches'),
    disruptions: t('sidebar.chaosEngine'),
    reconciliation: t('sidebar.reconciliation'),
  }
  
  const pathSegments = location.pathname.split('/').filter(Boolean)
  const breadcrumbs = pathSegments.map((segment, index) => ({
    label: segmentLabelMap[segment] || segment.charAt(0).toUpperCase() + segment.slice(1).replace(/-/g, ' '),
    path: '/' + pathSegments.slice(0, index + 1).join('/')
  }))

  return (
    <nav className="flex h-20 w-full items-center justify-between border-b border-gray-100 dark:border-gray-800 bg-white/80 dark:bg-gray-950/80 backdrop-blur-md sticky top-0 z-40 px-16">
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2 text-[10px] font-black uppercase tracking-[0.2em]">
          <span className="text-gray-400">{t('navbar.platform')}</span>
          {breadcrumbs.length > 0 && <span className="text-gray-300">/</span>}
          {breadcrumbs.map((bc, i) => (
            <div key={bc.path} className="flex items-center gap-2">
              <span className={i === breadcrumbs.length - 1 ? 'text-brand-primary' : 'text-gray-500'}>
                {bc.label}
              </span>
              {i < breadcrumbs.length - 1 && <span className="text-gray-300">/</span>}
            </div>
          ))}
          {breadcrumbs.length === 0 && <span className="text-brand-primary">{t('sidebar.overview')}</span>}
        </div>
      </div>

      <div className="flex items-center gap-8">
        <div className="relative group">
          <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4 text-gray-400">
            <Search className="h-4 w-4" />
          </div>
          <input
            type="text"
            placeholder={t('navbar.searchPlaceholder')}
            className="w-64 rounded-xl border border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-900 py-2.5 pl-11 pr-4 text-xs text-gray-900 dark:text-white outline-none focus:border-brand-primary/50 focus:ring-4 focus:ring-brand-primary/5 transition-all"
          />
        </div>

        <div className="flex items-center gap-4">
          <button
            onClick={toggleTheme}
            className="flex h-10 w-10 items-center justify-center rounded-xl border border-gray-100 dark:border-gray-800 bg-white dark:bg-gray-950 text-gray-500 hover:text-brand-primary transition-all shadow-sm"
          >
            {theme === 'light' ? <Moon className="h-4.5 w-4.5" /> : <Sun className="h-4.5 w-4.5" />}
          </button>

          <div className="h-10 w-10 flex items-center justify-center rounded-xl bg-gray-900 text-white font-black text-[10px] cursor-pointer shadow-lg shadow-gray-900/10">
            AD
          </div>
        </div>
      </div>
    </nav>
  )
}
