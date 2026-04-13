import { useLocation } from 'react-router-dom'
import { useTheme } from '../../context/ThemeContext'

export default function Navbar() {
  const { theme, toggleTheme } = useTheme()
  const location = useLocation()
  
  const pathSegments = location.pathname.split('/').filter(Boolean)
  const breadcrumbs = pathSegments.map((segment, index) => ({
    label: segment.charAt(0).toUpperCase() + segment.slice(1).replace(/-/g, ' '),
    path: '/' + pathSegments.slice(0, index + 1).join('/')
  }))

  return (
    <nav className="flex h-16 w-full items-center justify-between border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 px-12">
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider">
          <span className="text-slate-400">Insurer</span>
          {breadcrumbs.length > 0 && <span className="text-slate-300">/</span>}
          {breadcrumbs.map((bc, i) => (
            <div key={bc.path} className="flex items-center gap-1.5">
              <span className={i === breadcrumbs.length - 1 ? 'text-[var(--brand-primary)] dark:text-[var(--brand-soft)]' : 'text-slate-500'}>
                {bc.label}
              </span>
              {i < breadcrumbs.length - 1 && <span className="text-slate-300">/</span>}
            </div>
          ))}
          {breadcrumbs.length === 0 && <span className="text-[var(--brand-primary)] dark:text-[var(--brand-soft)]">Overview</span>}
        </div>
      </div>

      <div className="flex items-center gap-6">
        <div className="relative group">
          <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-slate-400">
            <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
          <input
            type="text"
            placeholder="Search console..."
            className="w-48 rounded-md border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 py-1.5 pl-9 pr-3 text-[11px] text-slate-900 dark:text-white outline-none focus:border-[var(--brand-primary)] transition-none"
          />
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={toggleTheme}
            className="flex h-8 w-8 items-center justify-center rounded border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-slate-500 hover:text-slate-900 dark:hover:text-white transition-none"
          >
            {theme === 'light' ? (
              <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
              </svg>
            ) : (
              <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 3v1m0 18v1m9-9h1M1 12H0m17.364-7.364l.707-.707M6.343 17.657l-.707.707M17.364 17.657l-.707-.707M6.343 6.343l.707-.707M14.5 12a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
              </svg>
            )}
          </button>

          <div className="h-8 w-8 flex items-center justify-center rounded border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800 text-slate-500 font-bold text-[10px] cursor-pointer">
            OP
          </div>
        </div>
      </div>
    </nav>
  )
}
