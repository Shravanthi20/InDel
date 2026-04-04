import { useEffect, useState } from 'react'
import { getClaims, getMaintenanceChecks, getOverview } from '../api/insurer'

type PendingClaim = {
  claim_id: number
  zone_name: string
  claim_amount: number
  status: string
  fraud_verdict: string
  created_at: string
}

export default function MaintenanceChecks() {
  const [pool, setPool] = useState<any>(null)
  const [overview, setOverview] = useState<any>(null)
  const [queue, setQueue] = useState<PendingClaim[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    Promise.all([getMaintenanceChecks(), getOverview(), getClaims({ status: 'manual_review', limit: 10, page: 1 })])
      .then(([poolRes, overviewRes, queueRes]) => {
        setPool(poolRes.data?.data ?? poolRes.data)
        setOverview(overviewRes.data?.data ?? overviewRes.data)
        setQueue(queueRes.data?.data ?? [])
      })
      .catch((err) => setError(err?.response?.data?.error?.message || 'Failed to load maintenance checks.'))
  }, [])

  return (
    <div className="space-y-6 p-6">
      <div className="rounded-2xl bg-white p-6 shadow">
        <h1 className="text-2xl font-bold">Maintenance Check Queue</h1>
        <p className="mt-2 text-sm text-slate-500">Weekly pool health and claims waiting for manual verification.</p>
      </div>

      {error ? <div className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-700">{error}</div> : null}

      <div className="grid gap-4 md:grid-cols-4">
        <Card title="Week premiums" value={`₹${Number(pool?.week_premiums ?? 0).toLocaleString('en-IN')}`} />
        <Card title="Week payouts" value={`₹${Number(pool?.week_payouts ?? 0).toLocaleString('en-IN')}`} />
        <Card title="Net pool" value={`₹${Number(pool?.net_pool ?? 0).toLocaleString('en-IN')}`} />
        <Card title="Pending payouts" value={String(pool?.pending_payouts ?? 0)} />
      </div>

      <div className="grid gap-6 xl:grid-cols-[0.95fr_1.05fr]">
        <div className="rounded-2xl bg-white p-6 shadow">
          <h2 className="text-lg font-semibold">Pool health</h2>
          <div className="mt-4 grid gap-4 md:grid-cols-2">
            <Card title="Active workers" value={String(overview?.active_workers ?? 0)} />
            <Card title="Pending claims" value={String(overview?.pending_claims ?? 0)} />
            <Card title="Approved claims" value={String(overview?.approved_claims ?? 0)} />
            <Card title="Pool state" value={String(overview?.pool_health ?? 'healthy')} />
          </div>
        </div>

        <div className="rounded-2xl bg-white p-6 shadow">
          <h2 className="text-lg font-semibold">Manual review queue</h2>
          <p className="mt-2 text-sm text-slate-500">Claims that still need a human decision before payout movement.</p>
          <div className="mt-4 space-y-3">
            {queue.length === 0 ? (
              <div className="rounded-xl bg-slate-50 p-4 text-sm text-slate-500">No claims waiting for manual review.</div>
            ) : queue.map((claim) => (
              <div key={claim.claim_id} className="rounded-xl border border-slate-200 bg-slate-50 p-4">
                <div className="flex items-start justify-between gap-4">
                  <div>
                    <p className="text-sm font-semibold text-slate-950">Claim #{claim.claim_id}</p>
                    <p className="text-sm text-slate-600">{claim.zone_name}</p>
                    <p className="text-sm text-slate-500">{String(claim.created_at).slice(0, 19).replace('T', ' ')}</p>
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-semibold text-slate-950">₹{Number(claim.claim_amount).toLocaleString('en-IN')}</p>
                    <p className="text-xs uppercase tracking-[0.2em] text-orange-600">{claim.fraud_verdict}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

function Card({ title, value }: { title: string; value: string }) {
  return (
    <div className="rounded-2xl bg-slate-50 p-4">
      <p className="text-xs uppercase tracking-[0.2em] text-slate-500">{title}</p>
      <p className="mt-2 text-2xl font-black text-slate-950">{value}</p>
    </div>
  )
}
