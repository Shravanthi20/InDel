import { Activity } from 'lucide-react'
import { formatCurrency, useGodMode } from './state'

function statusBadge(status: string) {
  if (status === 'ok') return 'border-emerald-200 bg-emerald-50 text-emerald-800'
  if (status === 'failed') return 'border-rose-200 bg-rose-50 text-rose-800'
  return 'border-amber-200 bg-amber-50 text-amber-800'
}

export default function ResultsPage() {
  const { result, endpointStatus, error, addDisruption, addingDisruption, loading } = useGodMode()

  const statusTone = result.status === 'Severe'
    ? 'text-rose-800 border-rose-200 bg-rose-50'
    : result.status === 'Critical'
      ? 'text-orange-800 border-orange-200 bg-orange-50'
      : result.status === 'Reduced Operation'
        ? 'text-amber-800 border-amber-200 bg-amber-50'
        : 'text-emerald-800 border-emerald-200 bg-emerald-50'

  return (
    <section className="space-y-5 rounded-[1.75rem] border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-950 dark:shadow-black/20">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-[11px] uppercase tracking-[0.3em] text-slate-500 dark:text-slate-400">Simulation results</p>
          <h2 className="mt-1 text-2xl font-bold text-slate-900 dark:text-white">Risk, claims, and cash outflow</h2>
          <p className="mt-2 text-sm font-medium text-slate-600 dark:text-slate-400">Scope: {result.scopeLabel}</p>
        </div>
        <div className={`rounded-full border px-4 py-1 text-xs font-semibold uppercase tracking-[0.25em] ${statusTone}`}>
          {result.status}
        </div>
      </div>

      {error ? (
        <div className="rounded-2xl border border-rose-200 bg-rose-50 p-4 text-rose-800 dark:border-rose-500/30 dark:bg-rose-500/10 dark:text-rose-300">{error}</div>
      ) : null}

      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <ScoreTile label="Risk score" value={result.riskScore.toFixed(2)} critical={result.riskScore >= 0.75} normal={result.riskScore < 0.3} />
        <ScoreTile label="AQI_score" value={result.aqiScore.toFixed(2)} critical={result.aqiScore >= 0.75} normal={result.aqiScore < 0.3} />
        <ScoreTile label="Temp_score" value={result.tempScore.toFixed(2)} critical={result.tempScore >= 0.75} normal={result.tempScore < 0.3} />
        <ScoreTile label="Rain_score" value={result.rainScore.toFixed(2)} critical={result.rainScore >= 0.75} normal={result.rainScore < 0.3} />
        <ScoreTile label="Traffic_score" value={result.trafficScore.toFixed(2)} critical={result.trafficScore >= 0.75} normal={result.trafficScore < 0.3} />
        <ScoreTile label="Cash outflow" value={formatCurrency(result.cashOutflow)} critical={result.cashOutflow > 1200} normal={result.cashOutflow === 0} />
      </div>

      <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-700 dark:bg-slate-900">
        <div className="text-[11px] uppercase tracking-[0.25em] text-slate-500 dark:text-slate-400">Disruption reason</div>
        <div className="mt-2 text-sm font-semibold text-slate-900 dark:text-slate-100">{result.reason}</div>
        <div className="mt-2 text-sm text-slate-600 dark:text-slate-300">Action taken: {result.actionTaken}</div>
      </div>

      <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-700 dark:bg-slate-900">
        <div className="flex items-center gap-2 text-sm font-semibold text-slate-900 dark:text-slate-100">
          <Activity className="h-4 w-4 text-sky-700" />
          System actions (simulated)
        </div>
        <ul className="mt-3 space-y-2 text-sm text-slate-700 dark:text-slate-300">
          <li className="rounded-xl border border-slate-200 bg-white px-3 py-2 dark:border-slate-700 dark:bg-slate-950">
            Notification to worker app: <span className="font-semibold text-sky-700">"{result.notification}"</span>
          </li>
          <li className="rounded-xl border border-slate-200 bg-white px-3 py-2 dark:border-slate-700 dark:bg-slate-950">
            Cash leaving the system: <span className="font-semibold text-rose-700">{formatCurrency(result.cashOutflow)}</span>
          </li>
          <li className="rounded-xl border border-slate-200 bg-white px-3 py-2 dark:border-slate-700 dark:bg-slate-950">
            Reduce batch size: <span className="font-semibold text-amber-700">{result.reducedBatchSizeBy}%</span>
          </li>
          <li className="rounded-xl border border-slate-200 bg-white px-3 py-2 dark:border-slate-700 dark:bg-slate-950">
            Increase delivery buffer: <span className="font-semibold text-amber-700">+{result.deliveryBufferMinutes} min</span>
          </li>
          <li className={`rounded-xl border px-3 py-2 ${result.zoneOperational ? 'border-emerald-200 bg-emerald-50 text-emerald-800' : 'border-rose-200 bg-rose-50 text-rose-800'}`}>
            Zone state: <span className="font-semibold">{result.zoneOperational ? 'Operational' : 'Non-operational'}</span>
          </li>
          <li className="rounded-xl border border-slate-200 bg-white px-3 py-2 dark:border-slate-700 dark:bg-slate-950">
            Affected zones: <span className="font-semibold text-slate-900 dark:text-slate-100">{result.affectedZoneCount}</span>
          </li>
        </ul>
      </div>

      <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-700 dark:bg-slate-900">
        <div className="text-[11px] uppercase tracking-[0.25em] text-slate-500 dark:text-slate-400">Backend endpoint monitor</div>
        <div className="mt-3 grid gap-2 sm:grid-cols-3">
          <EndpointCard label="GET /api/v1/platform/zones/health" status={endpointStatus.zoneHealth} />
          <EndpointCard label="GET /api/v1/worker/batches" status={endpointStatus.availableBatches} />
          <EndpointCard label="GET /api/v1/worker/batches/assigned" status={endpointStatus.assignedBatches} />
        </div>
        <button
          type="button"
          onClick={addDisruption}
          disabled={addingDisruption || loading}
          className="mt-4 inline-flex items-center justify-center rounded-full border border-rose-300 bg-rose-50 px-5 py-2 text-sm font-semibold uppercase tracking-[0.2em] text-rose-700 transition hover:bg-rose-100 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {addingDisruption ? 'Triggering all-zone disruption...' : 'Trigger disruption (all factors, all zones)'}
        </button>
      </div>
    </section>
  )
}

function ScoreTile({ label, value, critical, normal }: { label: string; value: string; critical: boolean; normal: boolean }) {
  const tone = critical
    ? 'border-rose-200 bg-rose-50 text-rose-800'
    : normal
      ? 'border-emerald-200 bg-emerald-50 text-emerald-800'
      : 'border-sky-200 bg-sky-50 text-sky-800'

  return (
    <div className={`rounded-2xl border p-3 ${tone}`}>
      <div className="text-[10px] uppercase tracking-[0.25em] opacity-75">{label}</div>
      <div className="mt-1 text-2xl font-black text-slate-900 dark:text-slate-100">{value}</div>
    </div>
  )
}

function EndpointCard({ label, status }: { label: string; status: 'ok' | 'failed' | 'pending' }) {
  return (
    <div className={`rounded-xl border px-3 py-2 ${statusBadge(status)}`}>
      <div className="text-[10px] uppercase tracking-[0.2em] opacity-75">Endpoint</div>
      <div className="mt-1 text-xs font-semibold">{label}</div>
      <div className="mt-2 text-[10px] uppercase tracking-[0.22em]">{status}</div>
    </div>
  )
}
