import { useMemo, useState } from 'react'
import { getPayoutReconciliation } from '../api/operations'
import { PageShell, Panel, ResultBox, StatCard } from './OperationsShared'

function todayDate() {
  return new Date().toISOString().slice(0, 10)
}

function previousWeekDate() {
  const date = new Date()
  date.setDate(date.getDate() - 7)
  return date.toISOString().slice(0, 10)
}

export default function ReconciliationOps() {
  const [from, setFrom] = useState(previousWeekDate())
  const [to, setTo] = useState(todayDate())
  const [result, setResult] = useState<any>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function handleLoad() {
    setLoading(true)
    setError('')
    try {
      const response = await getPayoutReconciliation({ from, to })
      setResult(response.data.data)
    } catch (err: any) {
      setError(err?.response?.data?.error?.message || 'Reconciliation lookup failed.')
    } finally {
      setLoading(false)
    }
  }

  const healthTone = useMemo(() => {
    if (!result) return 'default'
    return result.reconciliation_ok ? 'warm' : 'alert'
  }, [result])

  return (
    <PageShell
      eyebrow="Operations"
      title="Payout Reconciliation"
      description="Inspect payout totals, queue states, and mismatch counts across a selected date window. This is the fastest control point for spotting payout drift during demos."
    >
      <div className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
        <Panel title="Date Window" subtitle="Choose the reconciliation window and load the insurer-side payout summary.">
          <div className="grid gap-4 md:grid-cols-2">
            <label className="space-y-2">
              <span className="text-sm font-semibold text-slate-700">From</span>
              <input
                type="date"
                value={from}
                onChange={(e) => setFrom(e.target.value)}
                className="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-slate-900 outline-none transition focus:border-orange-400"
              />
            </label>

            <label className="space-y-2">
              <span className="text-sm font-semibold text-slate-700">To</span>
              <input
                type="date"
                value={to}
                onChange={(e) => setTo(e.target.value)}
                className="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-slate-900 outline-none transition focus:border-orange-400"
              />
            </label>
          </div>

          <div className="mt-5 flex items-center gap-4">
            <button
              type="button"
              onClick={handleLoad}
              disabled={loading}
              className="rounded-full bg-slate-950 px-6 py-3 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {loading ? 'Loading...' : 'Load Reconciliation'}
            </button>
            {error ? <span className="text-sm font-medium text-rose-600">{error}</span> : null}
          </div>
        </Panel>

        <Panel title="Signal to Watch" subtitle="Healthy reconciliation means payout records and claim statuses agree.">
          <div className="grid gap-4">
            <StatCard
              label="Target"
              value={result ? (result.reconciliation_ok ? 'Healthy' : 'Mismatch') : 'Awaiting run'}
              tone={healthTone as 'default' | 'warm' | 'alert'}
            />
            <StatCard label="Mismatch Count" value={String(result?.mismatch_count ?? 0)} tone={result?.mismatch_count ? 'alert' : 'default'} />
          </div>
        </Panel>
      </div>

      {result ? (
        <div className="grid gap-6 xl:grid-cols-[1fr_1fr]">
          <Panel title="Status Totals">
            <div className="grid gap-4 md:grid-cols-2">
              <StatCard label="Queued" value={String(result.counts?.queued ?? 0)} />
              <StatCard label="Processed" value={String(result.counts?.processed ?? 0)} tone="warm" />
              <StatCard label="Retry Pending" value={String(result.counts?.retry_pending ?? 0)} tone={result.counts?.retry_pending ? 'alert' : 'default'} />
              <StatCard label="Failed" value={String(result.counts?.failed ?? 0)} tone={result.counts?.failed ? 'alert' : 'default'} />
            </div>
          </Panel>

          <Panel title="Amount Totals">
            <div className="grid gap-4 md:grid-cols-2">
              <StatCard label="Queued Amount" value={`Rs ${result.totals?.queued_amount ?? 0}`} />
              <StatCard label="Processed Amount" value={`Rs ${result.totals?.processed_amount ?? 0}`} tone="warm" />
              <StatCard label="Retry Amount" value={`Rs ${result.totals?.retry_amount ?? 0}`} tone={result.totals?.retry_amount ? 'alert' : 'default'} />
              <StatCard label="Window" value={`${from} to ${to}`} />
            </div>
          </Panel>

          <Panel title="Audit Snapshot" subtitle="Compact payload summary for reviews and walkthroughs.">
            <ResultBox>
              <p><span className="font-semibold text-orange-300">from:</span> {result.from}</p>
              <p><span className="font-semibold text-orange-300">to:</span> {result.to}</p>
              <p><span className="font-semibold text-orange-300">reconciliation_ok:</span> {String(result.reconciliation_ok)}</p>
              <p><span className="font-semibold text-orange-300">mismatch_count:</span> {result.mismatch_count}</p>
            </ResultBox>
          </Panel>
        </div>
      ) : null}
    </PageShell>
  )
}
