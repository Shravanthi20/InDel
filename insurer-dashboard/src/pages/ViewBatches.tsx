import { useEffect, useMemo, useState } from 'react'
import { getAssignedBatches, getAvailableBatches } from '../api/insurer'
import client from '../api/client'
import { PageShell, Panel } from './OperationsShared'

type BatchRow = {
  batchId: string
  zoneLevel?: string
  fromCity?: string
  toCity?: string
  status?: string
  orderCount?: number
  totalWeight?: number
}

type ZoneRow = {
  city?: string
  state?: string
  zone_name?: string
}

type ZonePathResponse = {
  cities?: ZoneRow[]
  zones?: ZoneRow[]
}

function normalize(value?: string) {
  return (value || '').trim().toLowerCase()
}

function getZonePaths(type: 'a' | 'b' | 'c') {
  return client.get<ZonePathResponse>(`/api/v1/platform/zone-paths?type=${type}`)
}

export default function ViewBatches() {

  const [availableBatches, setAvailableBatches] = useState<BatchRow[]>([])
  const [assignedBatches, setAssignedBatches] = useState<BatchRow[]>([])
  const [zoneLevel, setZoneLevel] = useState<'a' | 'b' | 'c' | ''>('')
  const [zoneOptions, setZoneOptions] = useState<ZoneRow[]>([])
  const [zoneCache] = useState<{ [k: string]: ZoneRow[] }>({})
  const [selectedZone, setSelectedZone] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    setLoading(true)
    setError('')
    Promise.allSettled([
      getAvailableBatches<{ batches?: BatchRow[] }>(),
      getAssignedBatches<{ batches?: BatchRow[] }>(),
    ]).then(([availableResult, assignedResult]) => {
      if (availableResult.status === 'fulfilled') {
        setAvailableBatches(availableResult.value?.batches || [])
      }
      if (assignedResult.status === 'fulfilled') {
        setAssignedBatches(assignedResult.value?.batches || [])
      }
      if (availableResult.status === 'rejected' && assignedResult.status === 'rejected') {
        setError('Unable to load batches from backend.')
      }
      setLoading(false)
    })
  }, [])

  // Fetch 15 zones for selected level, cache in memory
  useEffect(() => {
    if (!zoneLevel) {
      setZoneOptions([])
      setSelectedZone('')
      return
    }
    if (zoneCache[zoneLevel]) {
      setZoneOptions(zoneCache[zoneLevel])
      setSelectedZone('')
      return
    }
    getZonePaths(zoneLevel).then((res) => {
      const cities = res.data.cities || res.data.zones || []
      setZoneOptions(cities)
      zoneCache[zoneLevel] = cities
      setSelectedZone('')
    }).catch(() => setZoneOptions([]))
  }, [zoneLevel])

  // Filter batches by selected level and zone
  const filteredAvailable = useMemo(() => {
    let filtered = availableBatches
    if (zoneLevel) {
      filtered = filtered.filter((batch) => normalize(batch.zoneLevel) === normalize(zoneLevel))
    }
    if (selectedZone) {
      filtered = filtered.filter((batch) => {
        return (batch.fromCity && batch.fromCity === selectedZone) || (batch.toCity && batch.toCity === selectedZone)
      })
    }
    return filtered
  }, [availableBatches, zoneLevel, selectedZone])

  const filteredAssigned = useMemo(() => {
    let filtered = assignedBatches
    if (zoneLevel) {
      filtered = filtered.filter((batch) => normalize(batch.zoneLevel) === normalize(zoneLevel))
    }
    if (selectedZone) {
      filtered = filtered.filter((batch) => {
        return (batch.fromCity && batch.fromCity === selectedZone) || (batch.toCity && batch.toCity === selectedZone)
      })
    }
    return filtered
  }, [assignedBatches, zoneLevel, selectedZone])

  return (
    <PageShell
      eyebrow="Operations"
      title="View Batches"
      description="Choose zone level first, then zone name, and browse available and assigned batches."
    >
      <Panel title="Filters" subtitle="Select zone level, then zone name from backend zones.">
        <div className="flex flex-wrap items-center gap-3">
          <select
            value={zoneLevel}
            onChange={(e) => setZoneLevel(e.target.value as 'a' | 'b' | 'c' | '')}
            className="rounded-full border border-slate-200 bg-white px-4 py-2 text-sm text-slate-700 outline-none"
          >
            <option value="">Select zone level</option>
            <option value="a">Zone A</option>
            <option value="b">Zone B</option>
            <option value="c">Zone C</option>
          </select>
          <select
            value={selectedZone}
            onChange={(e) => setSelectedZone(e.target.value)}
            disabled={!zoneLevel || zoneOptions.length === 0}
            className="rounded-full border border-slate-200 bg-white px-4 py-2 text-sm text-slate-700 outline-none"
          >
            <option value="">{zoneLevel ? 'Select Zone' : 'Select Level First'}</option>
            {zoneOptions.map((z, idx) => (
              <option key={z.city || z.zone_name || idx} value={z.city || z.zone_name}>
                {(z.city || z.zone_name) + (z.state ? ', ' + z.state : '')}
              </option>
            ))}
          </select>
          {loading ? <span className="text-sm text-slate-500">Loading...</span> : null}
          {error ? <span className="text-sm text-rose-600">{error}</span> : null}
        </div>
      </Panel>

      <div className="grid gap-6 xl:grid-cols-2">
        <Panel title="Available Batches" subtitle="Batches currently open for workers.">
          <BatchList rows={filteredAvailable} emptyText="No available batches for this filter." />
        </Panel>

        <Panel title="Assigned Batches" subtitle="Batches already assigned in workflow.">
          <BatchList rows={filteredAssigned} emptyText="No assigned batches for this filter." />
        </Panel>
      </div>
    </PageShell>
  )
}

function BatchList({ rows, emptyText }: { rows: BatchRow[]; emptyText: string }) {
  if (rows.length === 0) {
    return <p className="text-sm text-slate-500">{emptyText}</p>
  }

  return (
    <div className="space-y-3">
      {rows.map((row) => (
        <div key={row.batchId} className="rounded-xl border border-slate-200 bg-slate-50 p-3">
          <div className="flex items-center justify-between gap-2">
            <p className="text-sm font-semibold text-slate-900">{row.batchId}</p>
            <span className="rounded-full border border-slate-200 bg-white px-2 py-1 text-[10px] uppercase tracking-[0.2em] text-slate-600">
              {row.status || 'Pending'}
            </span>
          </div>
          <p className="mt-1 text-xs text-slate-500">
            Route: {(row.fromCity || 'Unknown')} to {(row.toCity || row.fromCity || 'Unknown')}
          </p>
          <div className="mt-2 flex flex-wrap gap-2 text-xs text-slate-600">
            <span className="rounded-full border border-slate-200 bg-white px-2 py-1">Zone {row.zoneLevel || '-'}</span>
            <span className="rounded-full border border-slate-200 bg-white px-2 py-1">Orders {row.orderCount || 0}</span>
            <span className="rounded-full border border-slate-200 bg-white px-2 py-1">Weight {(row.totalWeight || 0).toFixed(1)} kg</span>
          </div>
        </div>
      ))}
    </div>
  )
}