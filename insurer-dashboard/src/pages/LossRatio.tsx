import { useEffect, useMemo, useState } from 'react'
import { getLossRatio, getZones } from '../api/insurer'

type LossRatioRow = {
  city: string
  zone_name: string
  premiums: number
  claims: number
  loss_ratio: number
}

type ZoneOption = {
  zone_id: number
  name: string
  city: string
}

export default function LossRatio() {
  const [zoneName, setZoneName] = useState('')
  const [rows, setRows] = useState<LossRatioRow[]>([])
  const [zones, setZones] = useState<ZoneOption[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    getZones().then((res) => setZones(res.data?.zones ?? [])).catch(() => setZones([]))
  }, [])

  useEffect(() => {
    setError('')
    getLossRatio(zoneName ? { zone_id: zoneName } : undefined)
      .then((res) => setRows(res.data?.data ?? res.data ?? []))
      .catch((err) => setError(err?.response?.data?.error?.message || 'Failed to load loss ratio.'))
  }, [zoneName])

  const maxRatio = useMemo(() => Math.max(...rows.map((row) => row.loss_ratio), 1), [rows])
  const avgRatio = useMemo(() => {
    if (!rows.length) return 0
    return rows.reduce((sum, row) => sum + row.loss_ratio, 0) / rows.length
  }, [rows])

  return (
    <div className="space-y-6 p-6">
      <div className="rounded-2xl bg-white p-6 shadow">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <h1 className="text-2xl font-bold">Loss Ratio by Zone</h1>
            <p className="mt-2 text-sm text-slate-500">Premiums versus claims, grouped by zone and city.</p>
          </div>
          <label className="space-y-2 text-sm font-medium text-slate-700">
            <span>Filter by zone</span>
            <select
              value={zoneName}
              onChange={(e) => setZoneName(e.target.value)}
              className="block w-full rounded-xl border border-slate-300 bg-white px-4 py-3 outline-none focus:border-orange-400"
            >
              <option value="">All zones</option>
              {zones.map((zone) => (
                <option key={zone.zone_id} value={zone.name}>
                  {zone.name} - {zone.city}
                </option>
              ))}
            </select>
          </label>
        </div>
      </div>

      {error ? <div className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-700">{error}</div> : null}

      <div className="grid gap-4 md:grid-cols-3">
        <div className="rounded-2xl bg-white p-5 shadow">
          <p className="text-xs uppercase tracking-[0.2em] text-slate-500">Zones loaded</p>
          <p className="mt-2 text-3xl font-black text-slate-950">{rows.length}</p>
        </div>
        <div className="rounded-2xl bg-white p-5 shadow">
          <p className="text-xs uppercase tracking-[0.2em] text-slate-500">Average loss ratio</p>
          <p className="mt-2 text-3xl font-black text-slate-950">{Math.round(avgRatio * 100)}%</p>
        </div>
        <div className="rounded-2xl bg-white p-5 shadow">
          <p className="text-xs uppercase tracking-[0.2em] text-slate-500">Highest loss ratio</p>
          <p className="mt-2 text-3xl font-black text-slate-950">{Math.round(maxRatio * 100)}%</p>
        </div>
      </div>

      <div className="grid gap-4">
        {rows.length === 0 ? (
          <div className="rounded-2xl bg-white p-6 shadow text-slate-500">No loss-ratio rows available.</div>
        ) : (
          rows.map((row) => {
            const width = Math.max(6, Math.round((row.loss_ratio / maxRatio) * 100))
            return (
              <div key={`${row.zone_name}-${row.city}`} className="rounded-2xl bg-white p-6 shadow">
                <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                  <div>
                    <p className="text-xs uppercase tracking-[0.2em] text-slate-500">{row.city}</p>
                    <h3 className="text-lg font-bold text-slate-950">{row.zone_name}</h3>
                  </div>
                  <div className="text-sm text-slate-600">
                    Premiums ₹{Number(row.premiums).toLocaleString('en-IN')} · Claims ₹{Number(row.claims).toLocaleString('en-IN')}
                  </div>
                </div>
                <div className="mt-4 h-3 rounded-full bg-slate-100">
                  <div className="h-3 rounded-full bg-gradient-to-r from-orange-400 to-rose-500" style={{ width: `${width}%` }} />
                </div>
                <div className="mt-3 flex items-center justify-between text-sm text-slate-600">
                  <span>Loss ratio</span>
                  <span className="font-semibold text-slate-900">{Math.round(row.loss_ratio * 100)}%</span>
                </div>
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}
