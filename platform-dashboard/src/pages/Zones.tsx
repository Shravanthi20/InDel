import { useEffect, useState } from 'react'
import { getZoneHealth, getZones } from '../api/platform'

export default function Zones() {
  const [zones, setZones] = useState<any[]>([])
  const [health, setHealth] = useState<any[]>([])

  useEffect(() => {
    async function load() {
      const [zonesRes, healthRes] = await Promise.all([getZones(), getZoneHealth()])
      setZones(zonesRes.data.zones ?? [])
      setHealth(healthRes.data.data ?? [])
    }

    load().catch((error) => console.error('Failed to load zones', error))
    const timer = setInterval(() => load().catch(() => undefined), 5000)
    return () => clearInterval(timer)
  }, [])

  return (
    <div className="p-6">
      <h1 className="text-3xl font-bold">Zone Monitoring</h1>
      <div className="mt-6 grid gap-4">
        {zones.map((zone) => {
          const zoneHealth = health.find((item) => item.zone_id === zone.zone_id)
          return (
            <div key={zone.zone_id} className="rounded-2xl bg-white p-6 shadow">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-lg font-semibold text-slate-900">{zone.name}, {zone.city}</h2>
                  <p className="text-sm text-slate-500">{zone.state} • risk rating {zone.risk_rating}</p>
                </div>
                <div className="text-sm uppercase tracking-wide text-slate-500">
                  {zoneHealth?.status ?? 'healthy'}
                </div>
              </div>
              <div className="mt-4 grid gap-3 md:grid-cols-3">
                <div className="rounded-xl bg-slate-50 p-4 text-sm text-slate-700">Workers: {zone.active_workers}</div>
                <div className="rounded-xl bg-slate-50 p-4 text-sm text-slate-700">Current orders: {zoneHealth?.current_orders ?? 0}</div>
                <div className="rounded-xl bg-slate-50 p-4 text-sm text-slate-700">Drop: {Math.round((zoneHealth?.order_drop ?? 0) * 100)}%</div>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
