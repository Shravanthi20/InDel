import { useEffect, useMemo, useState } from 'react'
import { getAssignedBatches, getAvailableBatches, getZones } from '../api/insurer'
import { PageShell, Panel } from './OperationsShared'

type ZoneRow = {
  zone_id: number
  name: string
  city: string
  state: string
}

type BatchRow = {
  batchId: string
  zoneLevel?: string
  fromCity?: string
  toCity?: string
  status?: string
  orderCount?: number
  totalWeight?: number
}

type ZoneLevel = 'A' | 'B' | 'C' | 'ALL'

function normalize(value?: string) {
  return (value || '').trim().toLowerCase()
}

function zoneLabel(zone: ZoneRow) {
  return `${zone.name}${zone.city ? ` • ${zone.city}` : ''}${zone.state ? `, ${zone.state}` : ''}`
}

function batchMatchesZone(batch: BatchRow, zone: ZoneRow) {
  const tokens = [zone.name, zone.city, zone.state].map(normalize).filter(Boolean)
  const fields = [batch.fromCity, batch.toCity, batch.zoneLevel].map(normalize).filter(Boolean)
  return tokens.some((token) => fields.some((field) => field.includes(token) || token.includes(field)))
}

export default function ViewBatches() {
  const [zones, setZones] = useState<ZoneRow[]>([])
  const [availableBatches, setAvailableBatches] = useState<BatchRow[]>([])
  const [assignedBatches, setAssignedBatches] = useState<BatchRow[]>([])
  const [zoneLevelFilter, setZoneLevelFilter] = useState<ZoneLevel | ''>('')
  const [zoneFilter, setZoneFilter] = useState('ALL')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let mounted = true

    async function loadData() {
      setLoading(true)
      setError('')

      const [zonesResult, availableResult, assignedResult] = await Promise.allSettled([
        getZones(),
        getAvailableBatches<{ batches?: BatchRow[] }>(),
        getAssignedBatches<{ batches?: BatchRow[] }>(),
      ])

      if (!mounted) {
        return
      }

      if (zonesResult.status === 'fulfilled') {
        const zonePayload = zonesResult.value?.data?.zones
        setZones(Array.isArray(zonePayload) ? zonePayload : [])
      }
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
    }

    loadData()

    return () => {
      mounted = false
    }
  }, [])

  const levelFilteredAvailable = useMemo(() => {
    if (!zoneLevelFilter || zoneLevelFilter === 'ALL') {
      return availableBatches
    }
    return availableBatches.filter((batch) => normalize(batch.zoneLevel) === normalize(zoneLevelFilter))
  }, [availableBatches, zoneLevelFilter])

  const levelFilteredAssigned = useMemo(() => {
    if (!zoneLevelFilter || zoneLevelFilter === 'ALL') {
      return assignedBatches
    }
    return assignedBatches.filter((batch) => normalize(batch.zoneLevel) === normalize(zoneLevelFilter))
  }, [assignedBatches, zoneLevelFilter])

  const levelFilteredBatches = useMemo(
    () => [...levelFilteredAvailable, ...levelFilteredAssigned],
    [levelFilteredAvailable, levelFilteredAssigned],
  )

  const zoneOptions = useMemo(() => {
    if (!zoneLevelFilter) {
      return [] as ZoneRow[]
    }
    if (zoneLevelFilter === 'ALL') {
      return zones
    }
    return zones.filter((zone) => levelFilteredBatches.some((batch) => batchMatchesZone(batch, zone)))
  }, [levelFilteredBatches, zoneLevelFilter, zones])

  const selectedZone = useMemo(() => zoneOptions.find((zone) => String(zone.zone_id) === zoneFilter) || null, [zoneOptions, zoneFilter])

  const filteredAvailable = useMemo(
    () => (selectedZone ? levelFilteredAvailable.filter((batch) => batchMatchesZone(batch, selectedZone)) : levelFilteredAvailable),
    [levelFilteredAvailable, selectedZone],
  )

  const filteredAssigned = useMemo(
    () => (selectedZone ? levelFilteredAssigned.filter((batch) => batchMatchesZone(batch, selectedZone)) : levelFilteredAssigned),
    [levelFilteredAssigned, selectedZone],
  )

  return (
    <PageShell
      eyebrow="Operations"
      title="View Batches"
      description="Choose zone level first, then zone name, and browse available and assigned batches."
    >
      <Panel title="Filters" subtitle="Select zone level, then zone name from backend zones.">
        <div className="flex flex-wrap items-center gap-3">
          <select
            value={zoneLevelFilter}
            onChange={(event) => {
              const nextLevel = event.target.value as ZoneLevel | ''
              setZoneLevelFilter(nextLevel)
              setZoneFilter('ALL')
            }}
            className="rounded-full border border-slate-200 bg-white px-4 py-2 text-sm text-slate-700 outline-none"
          >
            <option value="">Select zone level</option>
            <option value="A">Zone A</option>
            <option value="B">Zone B</option>
            <option value="C">Zone C</option>
            <option value="ALL">All levels</option>
          </select>

          <select
            value={zoneFilter}
            onChange={(event) => setZoneFilter(event.target.value)}
            disabled={!zoneLevelFilter}
            className="rounded-full border border-slate-200 bg-white px-4 py-2 text-sm text-slate-700 outline-none"
          >
            <option value="ALL">All zones</option>
            {zoneOptions.map((zone) => (
              <option key={zone.zone_id} value={String(zone.zone_id)}>{zoneLabel(zone)}</option>
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