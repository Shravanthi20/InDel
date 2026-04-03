import { useEffect, useMemo, useState } from 'react'
import { getDisruptions, getWorkers, getZoneHealth, getZones } from '../api/platform'

export default function Overview() {
  const [workers, setWorkers] = useState<any[]>([])
  const [zones, setZones] = useState<any[]>([])
  const [health, setHealth] = useState<any[]>([])
  const [disruptions, setDisruptions] = useState<any[]>([])

  useEffect(() => {
    async function load() {
      const [workersRes, zonesRes, healthRes, disruptionsRes] = await Promise.all([
        getWorkers(),
        getZones(),
        getZoneHealth(),
        getDisruptions(),
      ])
      setWorkers(workersRes.data.workers ?? [])
      setZones(zonesRes.data.zones ?? [])
      setHealth(healthRes.data.data ?? [])
      setDisruptions(disruptionsRes.data.data ?? [])
    }

    load().catch((error) => console.error('Failed to load platform overview', error))
    const timer = setInterval(() => load().catch(() => undefined), 5000)
    return () => clearInterval(timer)
  }, [])

  const totals = useMemo(() => {
    const liveOrders = health.reduce((sum, item) => sum + (item.current_orders ?? 0), 0)
    const baselineOrders = health.reduce((sum, item) => sum + (item.baseline_orders ?? 0), 0)
    const disruptedZones = health.filter((item) => item.status === 'disrupted').length
    const paid = disruptions.reduce((sum, item) => sum + (item.payout_amount_total ?? 0), 0)
    return { liveOrders, baselineOrders, disruptedZones, paid }
  }, [health, disruptions])

  return (
    <div className="p-6">
      <h1 className="text-3xl font-bold">Platform Overview</h1>
      <p className="mt-2 text-slate-600">Live worker coverage, zone telemetry, and disruption automation in one glance.</p>

      <div className="mt-6 grid gap-6 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard label="Active Workers" value={String(workers.length)} />
        <MetricCard label="Tracked Zones" value={String(zones.length)} />
        <MetricCard label="Live Orders" value={String(totals.liveOrders)} />
        <MetricCard label="Disrupted Zones" value={String(totals.disruptedZones)} />
      </div>

      <div className="mt-6 grid gap-6 xl:grid-cols-[1.15fr_0.85fr]">
        <div className="rounded-2xl bg-white p-6 shadow">
          <h2 className="text-lg font-semibold">Current Zone Pressure</h2>
          <div className="mt-4 space-y-3">
            {health.map((item) => (
              <div key={item.zone_id} className="rounded-xl border border-slate-200 p-4">
                <div className="flex items-center justify-between">
                  <div className="font-semibold text-slate-900">Zone {item.zone_id}</div>
                  <div className="text-sm uppercase tracking-wide text-slate-500">{item.status}</div>
                </div>
                <div className="mt-2 text-sm text-slate-600">
                  {item.current_orders} live orders against {Math.round(item.baseline_orders ?? 0)} expected.
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="rounded-2xl bg-white p-6 shadow">
          <h2 className="text-lg font-semibold">Automation Snapshot</h2>
          <div className="mt-4 space-y-4 text-sm text-slate-600">
            <div className="rounded-xl bg-slate-50 p-4">Baseline order capacity: {Math.round(totals.baselineOrders)}</div>
            <div className="rounded-xl bg-slate-50 p-4">Confirmed disruptions: {disruptions.length}</div>
            <div className="rounded-xl bg-slate-50 p-4">Payout amount processed: Rs {Math.round(totals.paid)}</div>
            <div className="rounded-xl bg-slate-50 p-4">Workers visible to platform ops: {workers.length}</div>
          </div>
        </div>
      </div>
    </div>
  )
}

function MetricCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl bg-white p-6 shadow">
      <div className="text-sm font-semibold text-slate-600">{label}</div>
      <div className="mt-2 text-4xl font-bold text-slate-950">{value}</div>
    </div>
  )
}
