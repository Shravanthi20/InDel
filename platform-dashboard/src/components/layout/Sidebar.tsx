import { useEffect, useState, type ReactNode } from 'react'
import { NavLink } from 'react-router-dom'
import Navbar from './Navbar'
import { LayoutDashboard, Users, Map, BarChart3, Zap, ShieldCheck, Sparkles, FileSearch } from 'lucide-react'
import { getZones } from '../../api/platform'
import { useLocalization, type Language, type TranslationKey } from '../../context/LocalizationContext'

type NavItem = {
  to: string
  labelKey: TranslationKey
  icon: typeof LayoutDashboard
}

const platformNav: NavItem[] = [
  { to: '/', labelKey: 'sidebar.overview', icon: LayoutDashboard },
  { to: '/workers', labelKey: 'sidebar.workers', icon: Users },
  { to: '/zones', labelKey: 'sidebar.zones', icon: Map },
  { to: '/analytics', labelKey: 'sidebar.analytics', icon: BarChart3 },
]

const opsNav: NavItem[] = [
  { to: '/batches', labelKey: 'sidebar.viewBatches', icon: Sparkles },
  { to: '/disruptions', labelKey: 'sidebar.chaosEngine', icon: Zap },
  { to: '/reconciliation', labelKey: 'sidebar.reconciliation', icon: FileSearch },
]

function navClass(isActive: boolean) {
  return [
    'group flex items-center gap-3 px-6 py-3.5 text-sm font-bold transition-all relative',
    isActive
      ? 'bg-brand-soft/50 text-brand-primary'
      : 'text-gray-500 hover:text-gray-900 dark:hover:text-white hover:bg-gray-50 dark:hover:bg-gray-800/30',
  ].join(' ')
}

export default function Sidebar({ children }: { children: ReactNode }) {
  const { language, setLanguage, t } = useLocalization()
  const [backendZoneCount, setBackendZoneCount] = useState<number | null>(null)
  const [backendStatus, setBackendStatus] = useState<'loading' | 'ready' | 'error'>('loading')

  useEffect(() => {
    let mounted = true

    getZones()
      .then((response) => {
        if (!mounted) return
        setBackendZoneCount(response.data?.zones?.length ?? 0)
        setBackendStatus('ready')
      })
      .catch(() => {
        if (!mounted) return
        setBackendStatus('error')
      })

    return () => {
      mounted = false
    }
  }, [])

  return (
    <div className="flex min-h-screen bg-[var(--bg-main)]">
      {/* Enterprise Sidebar */}
      <aside className="fixed left-0 top-0 z-50 h-screen w-72 border-r border-gray-100 dark:border-gray-800 bg-white dark:bg-gray-950 shadow-soft-xl">
        <div className="flex h-full flex-col py-10">
          <div className="mb-12 flex items-center gap-4 px-8">
            <div className="flex h-10 w-10 items-center justify-center rounded-2xl bg-brand-primary shadow-lg shadow-brand-primary/20">
              <ShieldCheck className="h-5 w-5 text-white" />
            </div>
            <h1 className="text-xl font-black tracking-tighter text-gray-900 dark:text-white font-['Outfit'] italic">In<span className="text-brand-primary">Del</span></h1>
          </div>

          <div className="flex-1 space-y-8 overflow-y-auto no-scrollbar">
            <section>
              <p className="mb-2 px-6 text-[9px] font-black uppercase tracking-[0.3em] text-slate-400">{t('sidebar.inventory')}</p>
              <nav className="flex flex-col">
                {platformNav.map((item) => (
                  <NavLink key={item.to} to={item.to} end={item.to === '/'} className={({ isActive }) => navClass(isActive)}>
                    <item.icon className="h-4 w-4" />
                    {t(item.labelKey)}
                  </NavLink>
                ))}
              </nav>
            </section>

            <section>
              <p className="mb-2 px-6 text-[9px] font-black uppercase tracking-[0.3em] text-slate-400">{t('sidebar.operations')}</p>
              <nav className="flex flex-col">
                {opsNav.map((item) => (
                  <NavLink key={item.to} to={item.to} className={({ isActive }) => navClass(isActive)}>
                    <item.icon className="h-4 w-4" />
                    {t(item.labelKey)}
                  </NavLink>
                ))}
              </nav>
            </section>
          </div>

          <div className="mt-auto px-4 pt-4 border-t border-slate-100 dark:border-slate-800">
             <div className="mb-3">
                <label className="mb-2 block text-[9px] font-black uppercase tracking-[0.3em] text-slate-400">{t('common.language')}</label>
                <select
                  value={language}
                  onChange={(e) => setLanguage(e.target.value as Language)}
                  className="w-full rounded border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 px-3 py-2 text-[11px] font-bold text-slate-900 dark:text-white outline-none"
                >
                  <option value="en">{t('lang.english')}</option>
                  <option value="ta">{t('lang.tamil')}</option>
                  <option value="hi">{t('lang.hindi')}</option>
                </select>
             </div>
             <div className="rounded-lg bg-slate-50 dark:bg-slate-800/50 p-4 border border-slate-100 dark:border-slate-800">
                <div className="flex items-center gap-2 mb-1">
                   <div className={`h-1.5 w-1.5 rounded-full ${backendStatus === 'ready' ? 'bg-emerald-500' : backendStatus === 'error' ? 'bg-rose-500' : 'bg-amber-500'}`}></div>
                   <p className="text-[9px] font-black uppercase tracking-widest text-emerald-600 dark:text-emerald-400">
                     {backendStatus === 'ready' ? t('sidebar.backendConnected') : backendStatus === 'error' ? t('sidebar.backendOffline') : t('sidebar.connecting')}
                   </p>
                </div>
                <p className="text-[10px] text-slate-500 leading-tight">
                  {backendZoneCount === null ? t('sidebar.loadingZoneInventory') : `${backendZoneCount} ${t('sidebar.zonesLoaded')}`}
                </p>
             </div>
          </div>
        </div>
      </aside>

      {/* Main Surface */}
      <div className="flex-1 pl-72">
        <Navbar />
        <main className="min-h-screen px-16 py-16 max-w-[1500px] mx-auto">
          {children}
        </main>
      </div>
    </div>
  )
}
