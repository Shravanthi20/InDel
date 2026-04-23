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
    <div className="space-y-12">
      <header className="border-b border-gray-100 dark:border-gray-800 pb-12">
        <p className="text-[10px] font-black uppercase tracking-[0.5em] text-brand-primary mb-4">{eyebrow}</p>
        <h1 className="text-5xl font-black tracking-tighter text-gray-900 dark:text-white mb-6 leading-tight">{title}</h1>
        <p className="max-w-2xl text-lg leading-relaxed text-gray-500 dark:text-gray-400">{description}</p>
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
    <section className={`enterprise-panel p-10 ${className}`}>
      <div className="mb-10">
        <h2 className="text-2xl font-black tracking-tight text-gray-900 dark:text-white">{title}</h2>
        {subtitle ? <p className="mt-2 text-sm text-gray-500 dark:text-gray-400 font-medium">{subtitle}</p> : null}
      </div>
      {children}
    </section>
  )
}

export function StatCard({
  label,
  value,
  tone = 'default',
  onClick,
}: {
  label: string
  value: string
  tone?: 'default' | 'warm' | 'alert' | 'brand'
  onClick?: () => void
}) {
  const tones = {
    default: 'border-gray-100 bg-white dark:border-gray-800 dark:bg-gray-950 text-gray-900 dark:text-white',
    warm: 'border-orange-100 bg-orange-50 dark:border-orange-900/20 dark:bg-orange-900/10 text-orange-600 dark:text-orange-400',
    alert: 'border-pink-100 bg-pink-50 dark:border-pink-900/20 dark:bg-pink-900/10 text-pink-600 dark:text-pink-400',
    brand: 'border-brand-primary/20 bg-brand-soft dark:border-brand-primary/30 dark:bg-brand-primary/5 text-brand-primary',
  }

  return (
    <div 
      className={`rounded-2xl border p-8 shadow-soft-xl hover:shadow-2xl transition-all cursor-pointer ${tones[tone]}`}
      onClick={onClick}
    >
      <p className="text-[10px] font-black uppercase tracking-[0.3em] text-gray-400 dark:text-gray-500 mb-2">{label}</p>
      <p className="text-4xl font-black tracking-tight leading-none">{value}</p>
    </div>
  )
}

export function ImpactCard({ loss, payout }: { loss: number; payout: number }) {
  const percentage = loss > 0 ? Math.round((payout / loss) * 100) : 0;
  return (
    <div className="enterprise-panel bg-brand-soft/30 border-brand-primary/10 p-10 flex flex-col items-center justify-center text-center space-y-6">
      <p className="text-[11px] font-black uppercase tracking-[0.5em] text-brand-primary opacity-80 mb-2">The Economic Proof</p>
      <div className="flex items-center gap-12">
        <div className="text-left">
          <p className="text-[10px] font-black uppercase text-gray-400 tracking-widest mb-1">Economic Loss</p>
          <p className="text-4xl font-black text-gray-900 dark:text-white leading-none">₹{Math.round(loss).toLocaleString()}</p>
        </div>
        <div className="h-10 w-px bg-brand-primary/20"></div>
        <div className="text-left">
          <p className="text-[10px] font-black uppercase text-brand-primary tracking-widest mb-1">Verified Payout</p>
          <p className="text-4xl font-black text-brand-primary leading-none">₹{Math.round(payout).toLocaleString()}</p>
        </div>
      </div>
      <div className="pt-2">
        <span className="px-5 py-2 rounded-full bg-brand-primary text-white text-[11px] font-black uppercase tracking-[0.2em] shadow-lg shadow-brand-primary/20">
          {percentage}% Risk Protected
        </span>
      </div>
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
