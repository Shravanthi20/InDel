import type { ReactNode } from 'react'
import { NavLink } from 'react-router-dom'

const insurerNav = [
  { to: '/', label: 'Overview' },
  { to: '/loss-ratio', label: 'Loss Ratio' },
  { to: '/claims', label: 'Claims' },
  { to: '/fraud-queue', label: 'Fraud Queue' },
  { to: '/forecast', label: 'Forecast' },
  { to: '/maintenance-checks', label: 'Maintenance' },
]

const opsNav = [
  { to: '/operations/synthetic', label: 'Synthetic Data' },
  { to: '/operations/weekly-cycle', label: 'Weekly Cycle' },
  { to: '/operations/payouts', label: 'Payout Ops' },
  { to: '/operations/reconciliation', label: 'Reconciliation' },
]

function navClass(isActive: boolean) {
  return [
    'block rounded-xl px-3 py-2 text-sm transition-colors',
    isActive ? 'bg-orange-400 text-slate-950 font-semibold' : 'text-slate-200 hover:bg-slate-800 hover:text-white',
  ].join(' ')
}

export default function Sidebar({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top_left,_rgba(254,215,170,0.35),_transparent_32%),linear-gradient(180deg,_#f8fafc_0%,_#eef2ff_100%)] text-slate-900">
      <div className="flex min-h-screen">
        <aside className="w-72 border-r border-slate-800 bg-slate-950/95 px-6 py-8 text-white">
          <div className="mb-10">
            <p className="text-xs uppercase tracking-[0.35em] text-orange-300">InDel</p>
            <h1 className="mt-3 text-3xl font-black tracking-tight">Insurer Ops</h1>
            <p className="mt-3 text-sm leading-6 text-slate-300">
              Portfolio intelligence, payout controls, and synthetic-data operations in one place.
            </p>
          </div>

          <div className="space-y-8">
            <section>
              <p className="mb-3 text-xs uppercase tracking-[0.25em] text-slate-500">Analytics</p>
              <nav className="space-y-2">
                {insurerNav.map((item) => (
                  <NavLink key={item.to} to={item.to} end={item.to === '/'} className={({ isActive }) => navClass(isActive)}>
                    {item.label}
                  </NavLink>
                ))}
              </nav>
            </section>

            <section>
              <p className="mb-3 text-xs uppercase tracking-[0.25em] text-slate-500">Operations</p>
              <nav className="space-y-2">
                {opsNav.map((item) => (
                  <NavLink key={item.to} to={item.to} className={({ isActive }) => navClass(isActive)}>
                    {item.label}
                  </NavLink>
                ))}
              </nav>
            </section>
          </div>
        </aside>

        <main className="flex-1 px-6 py-8 md:px-10">{children}</main>
      </div>
    </div>
  )
}
