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
    <div className="space-y-10">
      <header className="border-b border-slate-200 dark:border-slate-800 pb-10">
        <p className="text-[10px] font-black uppercase tracking-[0.4em] text-slate-500 mb-3">{eyebrow}</p>
        <h1 className="text-4xl font-black tracking-tight text-slate-900 dark:text-white mb-4">{title}</h1>
        <p className="max-w-3xl text-base leading-relaxed text-slate-500 dark:text-slate-400">{description}</p>
      </header>
      <div className="pt-2">
        {children}
      </div>
    </div>
  )
}

export function Panel({
  title,
  subtitle,
  children,
  className = '',
}: {
  title: string
  subtitle?: string
  children: ReactNode
  className?: string
}) {
  return (
    <section className={`enterprise-panel p-8 ${className}`}>
      <div className="mb-8">
        <h2 className="text-xl font-bold tracking-tight text-slate-900 dark:text-white">{title}</h2>
        {subtitle ? <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">{subtitle}</p> : null}
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
    default: 'border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900 text-slate-900 dark:text-white',
    warm: 'border-orange-200 bg-orange-50 dark:border-orange-950 dark:bg-orange-900/10 text-orange-700 dark:text-orange-400',
    alert: 'border-rose-200 bg-rose-50 dark:border-rose-950 dark:bg-rose-900/10 text-rose-700 dark:text-rose-400',
  }

  return (
    <div className={`rounded-xl border p-6 shadow-sm ${tones[tone]}`}>
      <p className="text-[10px] font-black uppercase tracking-[0.2em] text-slate-500 mb-1">{label}</p>
      <p className="text-3xl font-black tracking-tight">{value}</p>
    </div>
  )
}

export function ResultBox({ children }: { children: ReactNode }) {
  return (
    <div className="rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 p-6 text-sm font-mono text-slate-700 dark:text-slate-300">
      {children}
    </div>
  )
}
