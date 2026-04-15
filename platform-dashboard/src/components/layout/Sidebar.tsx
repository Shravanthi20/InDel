import { useEffect, useState, type ReactNode } from 'react'
import { NavLink } from 'react-router-dom'
import Navbar from './Navbar'
import { LayoutDashboard, Users, Map, BarChart3, Zap, ShieldCheck, Sparkles, CalendarClock, CircleDollarSign, FileSearch, Database } from 'lucide-react'
import { getZones } from '../../api/platform'

const platformNav = [
  { to: '/', label: 'Overview', icon: LayoutDashboard },
  { to: '/workers', label: 'Workers', icon: Users },
  { to: '/zones', label: 'Zones', icon: Map },
  { to: '/analytics', label: 'Analytics', icon: BarChart3 },
]

const opsNav = [
  { to: '/batches', label: 'View Batches', icon: Sparkles },
  { to: '/god-mode/batch-simulation', label: 'Batch Simulation', icon: ShieldCheck },
  { to: '/synthetic-data', label: 'Synthetic Data', icon: Database },
  { to: '/disruptions', label: 'Chaos Engine', icon: Zap },
  { to: '/weekly-cycle', label: 'Weekly Cycle', icon: CalendarClock },
  { to: '/payout-ops', label: 'Payout Ops', icon: CircleDollarSign },
  { to: '/reconciliation', label: 'Reconciliation', icon: FileSearch },
]

function navClass(isActive: boolean) {
  return [
    'group flex items-center gap-3 px-4 py-2.5 text-sm font-medium border-l-2 transition-none',
    isActive
      ? 'bg-[var(--brand-soft)] dark:bg-slate-800 border-[var(--brand-primary)] text-slate-900 dark:text-white'
      : 'border-transparent text-slate-500 hover:text-slate-900 dark:hover:text-white hover:bg-slate-50 dark:hover:bg-slate-800/50',
  ].join(' ')
}

export default function Sidebar({ children }: { children: ReactNode }) {
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
      <aside className="fixed left-0 top-0 z-50 h-screen w-64 border-r border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900">
        <div className="flex h-full flex-col py-8">
          <div className="mb-10 flex items-center gap-3 px-6">
            <div className="flex h-8 w-8 items-center justify-center rounded bg-[var(--brand-primary)]">
              <ShieldCheck className="h-4 w-4 text-white" />
            </div>
            <h1 className="text-lg font-black tracking-tight text-slate-900 dark:text-white font-['Outfit']">InDel <span className="font-light text-slate-500">Platform</span></h1>
          </div>

          <div className="flex-1 space-y-8 overflow-y-auto no-scrollbar">
            <section>
              <p className="mb-2 px-6 text-[9px] font-black uppercase tracking-[0.3em] text-slate-400">Inventory</p>
              <nav className="flex flex-col">
                {platformNav.map((item) => (
                  <NavLink key={item.to} to={item.to} end={item.to === '/'} className={({ isActive }) => navClass(isActive)}>
                    <item.icon className="h-4 w-4" />
                    {item.label}
                  </NavLink>
                ))}
              </nav>
            </section>

            <section>
              <p className="mb-2 px-6 text-[9px] font-black uppercase tracking-[0.3em] text-slate-400">Operations</p>
              <nav className="flex flex-col">
                {opsNav.map((item) => (
                  <NavLink key={item.to} to={item.to} className={({ isActive }) => navClass(isActive)}>
                    <item.icon className="h-4 w-4" />
                    {item.label}
                  </NavLink>
                ))}
              </nav>
            </section>
          </div>

          <div className="mt-auto px-4 pt-4 border-t border-slate-100 dark:border-slate-800">
             <div className="rounded-lg bg-slate-50 dark:bg-slate-800/50 p-4 border border-slate-100 dark:border-slate-800">
                <div className="flex items-center gap-2 mb-1">
                   <div className={`h-1.5 w-1.5 rounded-full ${backendStatus === 'ready' ? 'bg-emerald-500' : backendStatus === 'error' ? 'bg-rose-500' : 'bg-amber-500'}`}></div>
                   <p className="text-[9px] font-black uppercase tracking-widest text-emerald-600 dark:text-emerald-400">
                     {backendStatus === 'ready' ? 'Backend Connected' : backendStatus === 'error' ? 'Backend Offline' : 'Connecting'}
                   </p>
                </div>
                <p className="text-[10px] text-slate-500 leading-tight">
                  {backendZoneCount === null ? 'Loading zone inventory...' : `${backendZoneCount} zones loaded from backend`}
                </p>
             </div>
          </div>
        </div>
      </aside>

      {/* Main Surface */}
      <div className="flex-1 pl-64">
        <Navbar />
        <main className="min-h-screen px-12 py-12 max-w-[1400px] mx-auto">
          {children}
        </main>
      </div>
    </div>
  )
}
