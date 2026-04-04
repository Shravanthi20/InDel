import { useEffect, useMemo, useState } from 'react'
import { getOverview } from '../api/insurer'

export default function Overview() {
  const [stats, setStats] = useState<any>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    getOverview()
      .then((res) => setStats(res.data?.data ?? res.data))
      .catch((err) => setError(err?.response?.data?.error?.message || 'Failed to load overview.'))
  }, [])

  const cards = useMemo(() => {
    if (!stats) return []
    return [
      { title: 'Pool Health', value: String(stats.pool_health ?? 'healthy'), note: 'Current weekly reserve state' },
      { title: 'Loss Ratio', value: `${Math.round((stats.loss_ratio ?? 0) * 100)}%`, note: 'Claims versus premiums' },
      { title: 'Active Workers', value: String(stats.active_workers ?? 0), note: 'Workers on active policies' },
      { title: 'Pending Claims', value: String(stats.pending_claims ?? 0), note: 'Awaiting review or processing' },
      { title: 'Approved Claims', value: String(stats.approved_claims ?? 0), note: 'Already approved or paid' },
      { title: 'Reserve', value: `₹${Number(stats.reserve ?? 0).toLocaleString('en-IN')}`, note: 'Current reserve after payouts' },
    ]
  }, [stats])

  if (error) {
    return <div className="p-6 text-rose-600">{error}</div>
  }

  if (!stats) return <div className="p-6">Loading...</div>

  return (
    <div className="grid gap-6 p-6 md:grid-cols-2 xl:grid-cols-3">
      {cards.map((card) => (
        <div key={card.title} className="rounded-2xl bg-white p-6 shadow">
          <h3 className="text-sm font-semibold uppercase tracking-[0.2em] text-slate-500">{card.title}</h3>
          <p className="mt-3 text-3xl font-black text-slate-950">{card.value}</p>
          <p className="mt-2 text-sm text-slate-500">{card.note}</p>
        </div>
      ))}

      <div className="rounded-2xl bg-slate-950 p-6 text-white md:col-span-2 xl:col-span-3">
        <h3 className="text-sm font-semibold uppercase tracking-[0.2em] text-orange-300">Portfolio summary</h3>
        <div className="mt-4 grid gap-4 md:grid-cols-3">
          <div className="rounded-xl border border-white/10 bg-white/5 p-4">
            <p className="text-xs uppercase tracking-[0.2em] text-slate-400">Reserve utilization</p>
            <p className="mt-2 text-2xl font-bold">{Math.round((stats.reserve_utilization ?? 0) * 100)}%</p>
          </div>
          <div className="rounded-xl border border-white/10 bg-white/5 p-4">
            <p className="text-xs uppercase tracking-[0.2em] text-slate-400">Pending claims</p>
            <p className="mt-2 text-2xl font-bold">{stats.pending_claims ?? 0}</p>
          </div>
          <div className="rounded-xl border border-white/10 bg-white/5 p-4">
            <p className="text-xs uppercase tracking-[0.2em] text-slate-400">Pool health</p>
            <p className="mt-2 text-2xl font-bold capitalize">{String(stats.pool_health ?? 'healthy')}</p>
          </div>
        </div>
      </div>
    </div>
  )
}
