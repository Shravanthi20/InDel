import type { ReactNode } from 'react'
import { NavLink } from 'react-router-dom'

const navItems = [
  { to: '/god-mode/temperature', label: 'Temperature' },
  { to: '/god-mode/rain', label: 'Rain' },
  { to: '/god-mode/aqi', label: 'AQI' },
  { to: '/god-mode/traffic', label: 'Traffic' },
  { to: '/god-mode/results', label: 'Results' },
  { to: '/god-mode/batches', label: 'Batches' },
]

export default function Sidebar({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-screen bg-slate-50 text-slate-900">
      <aside className="sticky top-0 flex h-screen w-72 flex-col border-r border-slate-200 bg-white px-5 py-6 shadow-sm">
        <div className="rounded-[1.5rem] border border-slate-200 bg-slate-50 p-5 shadow-sm">
          <div className="text-[11px] uppercase tracking-[0.4em] text-sky-700">InDel</div>
          <h1 className="mt-3 text-2xl font-black tracking-tight text-slate-900">Platform Dashboard</h1>
          <p className="mt-2 text-sm leading-6 text-slate-600">
            Simulation workspace for disruption, risk, and payout testing.
          </p>
        </div>

        <nav className="mt-6 space-y-2 text-sm font-medium">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) => [
                'block rounded-2xl border px-4 py-3 transition-all duration-200',
                isActive
                  ? 'border-sky-300 bg-sky-50 text-sky-800 shadow-sm'
                  : 'border-slate-200 bg-white text-slate-700 hover:border-slate-300 hover:bg-slate-50 hover:text-slate-900',
              ].join(' ')}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>

        <div className="mt-auto rounded-[1.5rem] border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">
          <div className="text-[11px] uppercase tracking-[0.35em] text-slate-500">Judge tools</div>
          <p className="mt-2 leading-6 text-slate-600">
            Use God Mode to test environmental inputs, risk scoring, payouts, and batch behavior.
          </p>
        </div>
      </aside>

      <main className="min-w-0 flex-1">{children}</main>
    </div>
  )
}
