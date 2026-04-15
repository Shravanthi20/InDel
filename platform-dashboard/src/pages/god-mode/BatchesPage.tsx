import { useMemo, useState } from 'react'
import { useGodMode, type BatchRow, type ZoneRecord } from './state'

function compactRouteLabel(zoneLevel?: string, fromCity?: string, toCity?: string) {
  const from = (fromCity || '').trim()
  const to = (toCity || '').trim()
  if (zoneLevel?.toUpperCase() === 'A' && from && from.toLowerCase() === to.toLowerCase()) {
    return from
  }
  if (!from && !to) {
    return 'Unknown'
  }
  if (!from) {
    return to
  }
  if (!to) {
    return from
  }
  return `${from} -> ${to}`
}

function isZoneASingleStop(zoneLevel?: string, fromCity?: string, toCity?: string) {
  const from = (fromCity || '').trim()
  const to = (toCity || '').trim()
  return zoneLevel?.toUpperCase() === 'A' && from !== '' && from.toLowerCase() === to.toLowerCase()
}

function normalize(value: string | undefined) {
  return (value || '').trim().toLowerCase()
}

function batchMatchesZone(batch: BatchRow, zone: ZoneRecord) {
  const zoneTokens = [zone.name, zone.city, zone.state].map(normalize).filter(Boolean)
  if (zoneTokens.length === 0) {
    return false
  }

  const batchTokens = [
    batch.fromCity,
    batch.toCity,
    batch.zoneLevel,
    ...(batch.orders || []).flatMap((order) => [order.pickupArea, order.dropArea, order.deliveryAddress]),
  ]
    .map(normalize)
    .filter(Boolean)

  return zoneTokens.some((zoneToken) => batchTokens.some((batchToken) => batchToken.includes(zoneToken) || zoneToken.includes(batchToken)))
}

function zoneLabel(zone: ZoneRecord) {
  return `${zone.name}${zone.city ? ` • ${zone.city}` : ''}${zone.state ? `, ${zone.state}` : ''}`
}

function batchZoneNames(batch: BatchRow, zones: ZoneRecord[]) {
  const matches = zones.filter((zone) => batchMatchesZone(batch, zone))
  if (matches.length > 0) {
    return Array.from(new Set(matches.map((zone) => zone.name)))
  }

  const fallback = [batch.fromCity, batch.toCity].map((value) => value?.trim()).filter(Boolean) as string[]
  return fallback.length > 0 ? fallback : ['Unmapped']
}

export default function BatchesPage() {
  const {
    availableBatches,
    assignedBatches,
    zones,
    generatingBatches,
    generateBatches,
    showCodes,
    setShowCodes,
    loading,
    checkingBatchFlow,
    runBatchFlowCheck,
    lastBatchFlowCheck,
    checkingIntegrationSelfTest,
    runIntegrationSelfTest,
    integrationSelfTestResult,
  } = useGodMode()
  const [zoneFilter, setZoneFilter] = useState('ALL')

  const zoneOptions = useMemo(() => {
    const options = zones
      .slice()
      .sort((left, right) => {
        const leftLabel = `${left.name} ${left.city} ${left.state}`.toLowerCase()
        const rightLabel = `${right.name} ${right.city} ${right.state}`.toLowerCase()
        return leftLabel.localeCompare(rightLabel)
      })

    return [{ value: 'ALL', label: 'All zones' }, ...options.map((zone) => ({ value: String(zone.zone_id), label: zoneLabel(zone) }))]
  }, [zones])

  const selectedZone = useMemo(() => zones.find((zone) => String(zone.zone_id) === zoneFilter) ?? null, [zones, zoneFilter])

  const filteredAvailable = useMemo(
    () => (selectedZone ? availableBatches.filter((batch) => batchMatchesZone(batch, selectedZone)) : availableBatches),
    [availableBatches, selectedZone],
  )

  const filteredAssigned = useMemo(
    () => (selectedZone ? assignedBatches.filter((batch) => batchMatchesZone(batch, selectedZone)) : assignedBatches),
    [assignedBatches, selectedZone],
  )

  return (
    <section className="space-y-5 rounded-[1.75rem] border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-950 dark:shadow-black/20">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <p className="text-[11px] uppercase tracking-[0.3em] text-slate-500 dark:text-slate-400">God Mode tools</p>
          <h2 className="mt-1 text-2xl font-bold text-slate-900 dark:text-white">Batch browser by zone</h2>
          <p className="mt-2 text-sm text-slate-600 dark:text-slate-400">
            Generate simulation orders, inspect every batch state, and reveal pickup or delivery codes only in admin mode.
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          <button
            type="button"
            onClick={() => void runIntegrationSelfTest()}
            disabled={loading || checkingIntegrationSelfTest}
            className="rounded-full border border-emerald-300 bg-emerald-50 px-4 py-2 text-xs font-bold uppercase tracking-[0.2em] text-emerald-700 transition hover:bg-emerald-100 disabled:cursor-not-allowed disabled:border-slate-200 disabled:bg-slate-100 disabled:text-slate-500 dark:border-emerald-500/30 dark:bg-emerald-500/10 dark:text-emerald-300 dark:hover:bg-emerald-500/20 dark:disabled:border-slate-700 dark:disabled:bg-slate-800 dark:disabled:text-slate-500"
          >
            {checkingIntegrationSelfTest ? 'Running integration test...' : 'Run integration self-test'}
          </button>

          <button
            type="button"
            onClick={() => void generateBatches()}
            disabled={loading || generatingBatches}
            className="rounded-full border border-amber-300 bg-amber-50 px-4 py-2 text-xs font-bold uppercase tracking-[0.2em] text-amber-800 transition hover:bg-amber-100 disabled:cursor-not-allowed disabled:border-slate-200 disabled:bg-slate-100 disabled:text-slate-500 dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-300 dark:hover:bg-amber-500/20 dark:disabled:border-slate-700 dark:disabled:bg-slate-800 dark:disabled:text-slate-500"
          >
            {generatingBatches ? 'Generating fake orders...' : 'Generate Fake Orders'}
          </button>

          <button
            type="button"
            onClick={() => void runBatchFlowCheck()}
            disabled={loading || checkingBatchFlow}
            className="rounded-full bg-sky-600 px-4 py-2 text-xs font-bold uppercase tracking-[0.2em] text-white transition hover:bg-sky-700 disabled:cursor-not-allowed disabled:bg-slate-300 dark:disabled:bg-slate-700"
          >
            {checkingBatchFlow ? 'Running check...' : 'Run batch flow check'}
          </button>

          <label className="inline-flex items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-4 py-2 text-xs uppercase tracking-[0.2em] text-slate-700 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300">
            <input
              type="checkbox"
              checked={showCodes}
              onChange={(event) => setShowCodes(event.target.checked)}
              className="h-4 w-4"
            />
            Show codes
          </label>

          <label className="flex items-center gap-3 rounded-full border border-slate-200 bg-slate-50 px-4 py-2 text-xs uppercase tracking-[0.2em] text-slate-700 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300">
            <span>Zone filter</span>
            <select
              value={zoneFilter}
              onChange={(event) => setZoneFilter(event.target.value)}
              className="rounded-full border border-slate-200 bg-white px-3 py-1 text-xs uppercase tracking-[0.2em] text-slate-700 outline-none dark:border-slate-700 dark:bg-slate-950 dark:text-slate-300"
            >
              {zoneOptions.map((zone) => (
                <option key={zone.value} value={zone.value}>
                  {zone.label}
                </option>
              ))}
            </select>
          </label>
        </div>
      </div>

      {lastBatchFlowCheck ? (
        <div className={`rounded-2xl border p-4 ${lastBatchFlowCheck.status === 'success' ? 'border-emerald-200 bg-emerald-50 dark:border-emerald-500/30 dark:bg-emerald-500/10' : 'border-rose-200 bg-rose-50 dark:border-rose-500/30 dark:bg-rose-500/10'}`}>
          <div className="flex flex-wrap items-center justify-between gap-3">
            <p className={`text-xs font-bold uppercase tracking-[0.2em] ${lastBatchFlowCheck.status === 'success' ? 'text-emerald-700 dark:text-emerald-300' : 'text-rose-700 dark:text-rose-300'}`}>
              Last batch flow check: {lastBatchFlowCheck.status}
            </p>
            <p className="text-xs text-slate-600 dark:text-slate-400">
              {new Date(lastBatchFlowCheck.checkedAt).toLocaleString()}
            </p>
          </div>

          <p className="mt-2 text-sm text-slate-800 dark:text-slate-200">{lastBatchFlowCheck.detail}</p>

          {(lastBatchFlowCheck.batchId || lastBatchFlowCheck.pickupMessage || lastBatchFlowCheck.deliveryMessage) ? (
            <div className="mt-3 grid gap-2 text-xs text-slate-700 dark:text-slate-300 md:grid-cols-3">
              <div className="rounded-lg border border-slate-200 bg-white px-3 py-2 dark:border-slate-700 dark:bg-slate-900">
                <span className="font-semibold">Batch:</span> {lastBatchFlowCheck.batchId || '-'}
              </div>
              <div className="rounded-lg border border-slate-200 bg-white px-3 py-2 dark:border-slate-700 dark:bg-slate-900">
                <span className="font-semibold">Pickup:</span> {lastBatchFlowCheck.pickupMessage || '-'}
              </div>
              <div className="rounded-lg border border-slate-200 bg-white px-3 py-2 dark:border-slate-700 dark:bg-slate-900">
                <span className="font-semibold">Delivery:</span> {lastBatchFlowCheck.deliveryMessage || '-'}
              </div>
            </div>
          ) : null}
        </div>
      ) : null}

      {integrationSelfTestResult ? (
        <div className={`rounded-2xl border p-4 ${integrationSelfTestResult.failed === 0 ? 'border-emerald-200 bg-emerald-50 dark:border-emerald-500/30 dark:bg-emerald-500/10' : 'border-rose-200 bg-rose-50 dark:border-rose-500/30 dark:bg-rose-500/10'}`}>
          <div className="flex flex-wrap items-center justify-between gap-3">
            <p className={`text-xs font-bold uppercase tracking-[0.2em] ${integrationSelfTestResult.failed === 0 ? 'text-emerald-700 dark:text-emerald-300' : 'text-rose-700 dark:text-rose-300'}`}>
              Integration self-test
            </p>
            <p className="text-xs text-slate-600 dark:text-slate-400">
              {new Date(integrationSelfTestResult.checkedAt).toLocaleString()}
            </p>
          </div>

          <div className="mt-2 flex flex-wrap gap-2 text-xs">
            <span className="rounded-full border border-emerald-300 bg-white px-3 py-1 text-emerald-700 dark:border-emerald-500/30 dark:bg-slate-900 dark:text-emerald-300">Passed: {integrationSelfTestResult.passed}</span>
            <span className="rounded-full border border-rose-300 bg-white px-3 py-1 text-rose-700 dark:border-rose-500/30 dark:bg-slate-900 dark:text-rose-300">Failed: {integrationSelfTestResult.failed}</span>
            <span className="rounded-full border border-amber-300 bg-white px-3 py-1 text-amber-700 dark:border-amber-500/30 dark:bg-slate-900 dark:text-amber-300">Skipped: {integrationSelfTestResult.skipped}</span>
          </div>

          <div className="mt-3 max-h-52 space-y-2 overflow-auto pr-1">
            {integrationSelfTestResult.checks.map((check) => (
              <div key={check.name} className="rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs text-slate-700 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <span className="font-semibold text-slate-900 dark:text-slate-100">{check.name}</span>
                  <span className={`rounded-full px-2 py-0.5 text-[10px] font-bold uppercase tracking-[0.14em] ${check.status === 'pass' ? 'bg-emerald-100 text-emerald-700' : check.status === 'fail' ? 'bg-rose-100 text-rose-700' : 'bg-amber-100 text-amber-700'}`}>
                    {check.status}
                  </span>
                </div>
                <div className="mt-1 text-slate-500 dark:text-slate-400">{check.detail}</div>
              </div>
            ))}
          </div>
        </div>
      ) : null}

      {loading ? (
        <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4 text-slate-600 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-400">Loading batch context...</div>
      ) : (
        <div className="space-y-5">
          <BatchGroup
            title="Available batches"
            subtitle="Orders that can move from the worker's zone to other zones."
            batches={filteredAvailable}
            zones={zones}
            showCodes={showCodes}
          />

          <BatchGroup
            title="Assigned batches"
            subtitle="Orders already grouped for the same zone."
            batches={filteredAssigned}
            zones={zones}
            showCodes={showCodes}
          />
        </div>
      )}
    </section>
  )
}

function BatchGroup({
  title,
  subtitle,
  batches,
  zones,
  showCodes,
}: {
  title: string
  subtitle: string
  batches: BatchRow[]
  zones: ZoneRecord[]
  showCodes: boolean
}) {
  return (
    <section className="space-y-3">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h3 className="text-lg font-bold text-slate-900 dark:text-white">{title}</h3>
          <p className="text-sm text-slate-500 dark:text-slate-400">{subtitle}</p>
        </div>
        <div className="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs uppercase tracking-[0.2em] text-slate-600 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300">
          {batches.length} batches
        </div>
      </div>

      {batches.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 p-4 text-slate-600 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-400">No batches match this zone.</div>
      ) : (
        <div className="grid gap-4 xl:grid-cols-2">
          {batches.map((batch) => {
            const orderCount = batch.orderCount ?? batch.orders?.length ?? 0
            const pickupCode = batch.pickupCode || '----'
            const deliveryCode = batch.deliveryCode || '----'
            const zoneASingleStop = isZoneASingleStop(batch.zoneLevel, batch.fromCity, batch.toCity)
            const routeLabel = compactRouteLabel(batch.zoneLevel, batch.fromCity, batch.toCity)
            const zoneNames = batchZoneNames(batch, zones)

            return (
              <article key={batch.batchId} className="rounded-2xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-700 dark:bg-slate-900">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <div className="text-lg font-black text-slate-900 dark:text-white">{batch.batchId}</div>
                    <div className="text-xs uppercase tracking-[0.2em] text-slate-500 dark:text-slate-400">{routeLabel}</div>
                  </div>
                  <div className="rounded-full border border-sky-200 bg-sky-50 px-3 py-1 text-[10px] uppercase tracking-[0.22em] text-sky-800">
                    {batch.status || 'Pending'}
                  </div>
                </div>

                <div className="mt-3 flex flex-wrap gap-2 text-[10px] uppercase tracking-[0.2em] text-slate-500 dark:text-slate-400">
                  <span className="rounded-full border border-slate-200 bg-white px-3 py-1 dark:border-slate-700 dark:bg-slate-950">Zones: {zoneNames.join(', ')}</span>
                  <span className="rounded-full border border-slate-200 bg-white px-3 py-1 dark:border-slate-700 dark:bg-slate-950">Zone level: {batch.zoneLevel || '-'}</span>
                </div>

                <div className="mt-3 grid gap-2 text-sm text-slate-700 dark:text-slate-300 md:grid-cols-4">
                  <MiniStat label="Orders" value={String(orderCount)} />
                  <MiniStat label="Weight" value={`${Number(batch.totalWeight || 0).toFixed(1)} kg`} />
                  <MiniStat label="Target" value={`${Number(batch.targetWeight || 0).toFixed(1)} kg`} />
                  <MiniStat label="Status" value={batch.status || 'Pending'} />
                </div>

                <div className="mt-3 grid gap-2 rounded-xl border border-slate-200 bg-white p-3 text-sm sm:grid-cols-2 dark:border-slate-700 dark:bg-slate-950">
                  <CodeCard label="Pickup code" value={showCodes ? pickupCode : '----'} />
                  <CodeCard
                    label="Delivery code"
                    value={zoneASingleStop ? 'Per order below' : (showCodes ? deliveryCode : '----')}
                  />
                </div>

                {zoneASingleStop && (batch.orders?.length ?? 0) > 0 ? (
                  <div className="mt-3 rounded-xl border border-slate-200 bg-white p-3 dark:border-slate-700 dark:bg-slate-950">
                    <div className="text-[10px] uppercase tracking-[0.2em] text-slate-500 dark:text-slate-400">Zone A order delivery codes</div>
                    <div className="mt-2 grid gap-2">
                      {(batch.orders || []).map((order) => {
                        const orderDeliveryCode = order.deliveryCode || '----'
                        const orderRoute = compactRouteLabel(
                          batch.zoneLevel,
                          order.pickupArea || batch.fromCity,
                          order.dropArea || order.deliveryAddress || batch.toCity,
                        )
                        return (
                          <div key={order.orderId} className="flex flex-wrap items-center justify-between gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 dark:border-slate-700 dark:bg-slate-900">
                            <div>
                              <div className="text-xs font-semibold text-slate-900 dark:text-slate-100">{order.orderId}</div>
                              <div className="text-[11px] text-slate-500 dark:text-slate-400">{orderRoute}</div>
                            </div>
                            <div className="font-mono text-sm font-black tracking-[0.2em] text-sky-800">
                              {showCodes ? orderDeliveryCode : '----'}
                            </div>
                          </div>
                        )
                      })}
                    </div>
                  </div>
                ) : null}
              </article>
            )
          })}
        </div>
      )}
    </section>
  )
}

function MiniStat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-2 dark:border-slate-700 dark:bg-slate-950">
      <div className="text-[10px] uppercase tracking-[0.2em] text-slate-500 dark:text-slate-400">{label}</div>
      <div className="mt-1 font-semibold text-slate-900 dark:text-slate-100">{value}</div>
    </div>
  )
}

function CodeCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 dark:border-slate-700 dark:bg-slate-900">
      <div className="text-[10px] uppercase tracking-[0.2em] text-slate-500 dark:text-slate-400">{label}</div>
      <div className="mt-1 font-mono text-lg font-black tracking-[0.2em] text-sky-800 dark:text-sky-300">{value}</div>
    </div>
  )
}




