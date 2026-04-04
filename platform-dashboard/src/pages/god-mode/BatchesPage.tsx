import { deliveryCodeFromBatchId, deliveryCodeFromOrderId, pickupCodeFromBatchId, useGodMode } from './state'

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

export default function BatchesPage() {
  const { batches, showCodes, setShowCodes, loading } = useGodMode()

  return (
    <section className="space-y-4 rounded-[1.75rem] border border-slate-200 bg-white p-5 shadow-sm">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <p className="text-[11px] uppercase tracking-[0.3em] text-slate-500">God Mode tools</p>
          <h2 className="mt-1 text-2xl font-bold text-slate-900">Batch list and verification codes</h2>
        </div>
        <label className="inline-flex items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-4 py-2 text-xs uppercase tracking-[0.2em] text-slate-700">
          <input
            type="checkbox"
            checked={showCodes}
            onChange={(event) => setShowCodes(event.target.checked)}
            className="h-4 w-4"
          />
          Show codes
        </label>
      </div>

      {loading ? (
        <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4 text-slate-600">Loading batch context...</div>
      ) : batches.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 p-4 text-slate-600">No batches available.</div>
      ) : (
        <div className="grid gap-4 xl:grid-cols-2">
          {batches.map((batch) => {
            const orderCount = batch.orderCount ?? batch.orders?.length ?? 0
            const pickupCode = pickupCodeFromBatchId(batch.batchId)
            const deliveryCode = deliveryCodeFromBatchId(batch.batchId)
            const zoneASingleStop = isZoneASingleStop(batch.zoneLevel, batch.fromCity, batch.toCity)
            const routeLabel = compactRouteLabel(batch.zoneLevel, batch.fromCity, batch.toCity)

            return (
              <article key={batch.batchId} className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
                <div className="flex items-start justify-between gap-3">
                  <div>
                      <div className="text-lg font-black text-slate-900">{batch.batchId}</div>
                    <div className="text-xs uppercase tracking-[0.2em] text-slate-500">
                      {routeLabel}
                    </div>
                  </div>
                    <div className="rounded-full border border-sky-200 bg-sky-50 px-3 py-1 text-[10px] uppercase tracking-[0.22em] text-sky-800">
                    {batch.status || 'Pending'}
                  </div>
                </div>

                  <div className="mt-3 grid gap-2 text-sm text-slate-700 md:grid-cols-4">
                  <MiniStat label="Zone" value={batch.zoneLevel || '-'} />
                  <MiniStat label="Orders" value={String(orderCount)} />
                  <MiniStat label="Weight" value={`${Number(batch.totalWeight || 0).toFixed(1)} kg`} />
                  <MiniStat label="Target" value={`${Number(batch.targetWeight || 0).toFixed(1)} kg`} />
                </div>

                  <div className="mt-3 grid gap-2 rounded-xl border border-slate-200 bg-white p-3 text-sm sm:grid-cols-2">
                  <CodeCard label="Pickup code" value={showCodes ? pickupCode : '----'} />
                  <CodeCard
                    label="Delivery code"
                    value={zoneASingleStop ? 'Per order below' : (showCodes ? deliveryCode : '----')}
                  />
                </div>

                {zoneASingleStop && (batch.orders?.length ?? 0) > 0 ? (
                  <div className="mt-3 rounded-xl border border-slate-200 bg-white p-3">
                    <div className="text-[10px] uppercase tracking-[0.2em] text-slate-500">Zone A order delivery codes</div>
                    <div className="mt-2 grid gap-2">
                      {(batch.orders || []).map((order) => {
                        const orderDeliveryCode = order.deliveryCode || deliveryCodeFromOrderId(order.orderId)
                        const orderRoute = compactRouteLabel(
                          batch.zoneLevel,
                          order.pickupArea || batch.fromCity,
                          order.dropArea || order.deliveryAddress || batch.toCity,
                        )
                        return (
                          <div key={order.orderId} className="flex flex-wrap items-center justify-between gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2">
                            <div>
                              <div className="text-xs font-semibold text-slate-900">{order.orderId}</div>
                              <div className="text-[11px] text-slate-500">{orderRoute}</div>
                            </div>
                            <div className="font-mono text-sm font-black tracking-[0.2em] text-sky-800">{showCodes ? orderDeliveryCode : '----'}</div>
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
    <div className="rounded-xl border border-slate-200 bg-white p-2">
      <div className="text-[10px] uppercase tracking-[0.2em] text-slate-500">{label}</div>
      <div className="mt-1 font-semibold text-slate-900">{value}</div>
    </div>
  )
}

function CodeCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-slate-200 bg-slate-50 px-3 py-2">
      <div className="text-[10px] uppercase tracking-[0.2em] text-slate-500">{label}</div>
      <div className="mt-1 font-mono text-lg font-black tracking-[0.2em] text-sky-800">{value}</div>
    </div>
  )
}
