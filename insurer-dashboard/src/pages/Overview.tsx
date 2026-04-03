import { useEffect, useState } from 'react'
import { getOverview, getPoolHealth } from '../api/insurer'
import { PageShell, Panel, StatCard } from './OperationsShared'

type OverviewData = {
  active_workers: number
  pending_claims: number
  approved_claims: number
  loss_ratio: number
  reserve_utilization: number
  reserve: number
  pool_health: string
}

type PoolHealth = {
  week_premiums: number
  week_payouts: number
  net_pool: number
  pending_payouts: number
}

export default function Overview() {
  const [overview, setOverview] = useState<OverviewData | null>(null)
  const [pool, setPool] = useState<PoolHealth | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    Promise.all([getOverview(), getPoolHealth()])
      .then(([overviewData, poolHealth]) => {
        setOverview(overviewData)
        setPool(poolHealth)
      })
      .catch((err) => setError(err?.message ?? 'Failed to load overview'))
  }, [])

  return (
    <PageShell
      eyebrow="Insurer"
      title="Portfolio Overview"
      description="Track live worker coverage, claims pressure, and reserve posture from one operational surface."
    >
      {error ? <Panel title="Load Error">{error}</Panel> : null}
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <StatCard label="Active Workers" value={String(overview?.active_workers ?? 0)} />
        <StatCard label="Pending Claims" value={String(overview?.pending_claims ?? 0)} tone="alert" />
        <StatCard label="Approved Claims" value={String(overview?.approved_claims ?? 0)} tone="warm" />
        <StatCard
          label="Loss Ratio"
          value={`${Math.round((overview?.loss_ratio ?? 0) * 100)}%`}
          tone={(overview?.loss_ratio ?? 0) > 0.8 ? 'alert' : 'default'}
        />
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Panel title="Pool Position" subtitle="Current weekly balance against paid and pending obligations.">
          <div className="grid gap-4 sm:grid-cols-2">
            <StatCard label="Week Premiums" value={`Rs ${pool?.week_premiums ?? 0}`} />
            <StatCard label="Week Payouts" value={`Rs ${pool?.week_payouts ?? 0}`} tone="warm" />
            <StatCard label="Net Pool" value={`Rs ${pool?.net_pool ?? 0}`} />
            <StatCard label="Pending Payouts" value={String(pool?.pending_payouts ?? 0)} tone="alert" />
          </div>
        </Panel>

        <Panel title="Risk Snapshot" subtitle="How the insurer book is trending right now.">
          <div className="space-y-3 text-sm text-slate-600">
            <p><span className="font-semibold text-slate-950">Pool health:</span> {overview?.pool_health ?? 'unknown'}</p>
            <p><span className="font-semibold text-slate-950">Reserve:</span> Rs {Math.round(overview?.reserve ?? 0)}</p>
            <p><span className="font-semibold text-slate-950">Reserve utilization:</span> {Math.round((overview?.reserve_utilization ?? 0) * 100)}%</p>
            <p>The current stack is wired for demo-friendly insurer operations, including synthetic data, claims generation, payout queueing, and reserve monitoring.</p>
          </div>
        </Panel>
      </div>
    </PageShell>
  )
}
