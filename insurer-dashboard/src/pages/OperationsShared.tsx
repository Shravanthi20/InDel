import type { ReactNode } from 'react'

export function PageShell({
  eyebrow,
  title,
  description,
  children,
}: {
  eyebrow: string
  title: string
  description: string
  children: ReactNode
}) {
  return (
    <div className="space-y-6">
      <header className="rounded-[28px] border border-slate-200 bg-white/85 p-8 shadow-[0_18px_50px_rgba(15,23,42,0.08)] backdrop-blur">
        <p className="text-xs uppercase tracking-[0.35em] text-orange-500">{eyebrow}</p>
        <h1 className="mt-3 text-3xl font-black tracking-tight text-slate-950">{title}</h1>
        <p className="mt-3 max-w-3xl text-sm leading-6 text-slate-600">{description}</p>
      </header>
      {children}
    </div>
  )
}

export function Panel({
  title,
  subtitle,
  children,
}: {
  title: string
  subtitle?: string
  children: ReactNode
}) {
  return (
    <section className="rounded-[24px] border border-slate-200 bg-white/90 p-6 shadow-[0_14px_40px_rgba(15,23,42,0.06)]">
      <div className="mb-5">
        <h2 className="text-lg font-bold text-slate-950">{title}</h2>
        {subtitle ? <p className="mt-1 text-sm text-slate-500">{subtitle}</p> : null}
      </div>
      {children}
    </section>
  )
}

export function StatCard({
  label,
  value,
  tone = 'default',
}: {
  label: string
  value: string
  tone?: 'default' | 'warm' | 'alert'
}) {
  const tones = {
    default: 'border-slate-200 bg-slate-50 text-slate-900',
    warm: 'border-orange-200 bg-orange-50 text-orange-950',
    alert: 'border-rose-200 bg-rose-50 text-rose-950',
  }

  return (
    <div className={`rounded-2xl border p-4 ${tones[tone]}`}>
      <p className="text-xs uppercase tracking-[0.24em] opacity-70">{label}</p>
      <p className="mt-3 text-2xl font-black">{value}</p>
    </div>
  )
}

export function ResultBox({ children }: { children: ReactNode }) {
  return (
    <div className="rounded-2xl border border-slate-200 bg-slate-950 p-4 text-sm text-slate-100 shadow-inner">
      {children}
    </div>
  )
}
