import { useEffect, useMemo, useState } from 'react'
import { getDisruptions, getZoneHealth } from '../api/platform'

export default function Analytics() {
  const [health, setHealth] = useState<any[]>([])
  const [disruptions, setDisruptions] = useState<any[]>([])

  useEffect(() => {
    async function load() {
      const [healthRes, disruptionsRes] = await Promise.all([getZoneHealth(), getDisruptions()])
      setHealth(healthRes.data.data ?? [])
      setDisruptions(disruptionsRes.data.data ?? [])
    }

    load().catch((error) => console.error('Failed to load analytics', error))
    const timer = setInterval(() => load().catch(() => undefined), 5000)
    return () => clearInterval(timer)
  }, [])

  const stats = useMemo(() => {
    const avgDrop = health.length
      ? Math.round((health.reduce((sum, item) => sum + (item.order_drop ?? 0), 0) / health.length) * 100)
      : 0
    const manualReview = disruptions.reduce((sum, item) => sum + (item.claims_in_review ?? 0), 0)
    const claims = disruptions.reduce((sum, item) => sum + (item.claims_generated ?? 0), 0)
    return { avgDrop, manualReview, claims }
  }, [health, disruptions])

  return (
    <div className="p-6">
      <h1 className="text-3xl font-bold">Platform Analytics</h1>
      <div className="mt-6 grid gap-6 md:grid-cols-3">
        <AnalyticsCard label="Average Zone Drop" value={`${stats.avgDrop}%`} />
        <AnalyticsCard label="Claims Triggered" value={String(stats.claims)} />
        <AnalyticsCard label="Manual Review Queue" value={String(stats.manualReview)} />
      </div>

      <div className="mt-6 rounded-2xl bg-white p-6 shadow">
        <h2 className="text-lg font-semibold">Disruption Feed</h2>
        <div className="mt-4 space-y-3">
          {disruptions.map((item) => (
            <div key={item.disruption_id} className="rounded-xl border border-slate-200 p-4">
              <div className="flex items-center justify-between">
                <div className="font-semibold text-slate-900">{item.disruption_id}</div>
                <div className="text-sm uppercase text-slate-500">{item.automation_status}</div>
              </div>
              <div className="mt-2 text-sm text-slate-600">
                {item.type} in {item.zone_id} with {item.claims_generated} claims and {item.payouts_processed} payouts processed.
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

function AnalyticsCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl bg-white p-6 shadow">
      <div className="text-sm font-semibold text-slate-600">{label}</div>
      <div className="mt-2 text-3xl font-bold text-slate-950">{value}</div>
    </div>
  )
}
