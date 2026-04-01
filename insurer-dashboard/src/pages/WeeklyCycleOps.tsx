import { useState } from 'react'
import { runWeeklyCycle } from '../api/operations'
import { PageShell, Panel, ResultBox, StatCard } from './OperationsShared'

export default function WeeklyCycleOps() {
  const [result, setResult] = useState<any>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function handleRun() {
    setLoading(true)
    setError('')
    try {
      const response = await runWeeklyCycle()
      setResult(response.data.data)
    } catch (err: any) {
      setError(err?.response?.data?.error?.message || 'Weekly cycle run failed.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <PageShell
      eyebrow="Operations"
      title="Weekly Premium Cycle"
      description="Run the insurer-side weekly premium lifecycle, validate idempotency, and surface partial failure counts before payout activity starts."
    >
      <div className="grid gap-6 xl:grid-cols-[1fr_0.9fr]">
        <Panel title="Run Current Cycle" subtitle="This calls the internal weekly-cycle endpoint and returns the latest run summary.">
          <div className="space-y-5">
            <div className="rounded-2xl border border-slate-200 bg-slate-50 p-5 text-sm leading-6 text-slate-600">
              Use this after synthetic data generation or before a demo scenario to compute premiums for active policies. Repeated runs should stay idempotent unless missing data was fixed between attempts.
            </div>

            <div className="flex items-center gap-4">
              <button
                type="button"
                onClick={handleRun}
                disabled={loading}
                className="rounded-full bg-slate-950 px-6 py-3 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {loading ? 'Running...' : 'Run Weekly Cycle'}
              </button>
              {error ? <span className="text-sm font-medium text-rose-600">{error}</span> : null}
            </div>
          </div>
        </Panel>

        <Panel title="What to Watch" subtitle="The key signals that tell you whether the premium cycle is healthy.">
          <div className="grid gap-4">
            <StatCard label="Best Outcome" value="Completed" />
            <StatCard label="Recovery Signal" value="Failures drop to 0" tone="warm" />
            <StatCard label="Idempotency Check" value="Cycle ID stable" />
          </div>
        </Panel>
      </div>

      {result ? (
        <div className="grid gap-6 xl:grid-cols-[0.9fr_1.1fr]">
          <Panel title="Cycle Metrics" subtitle={`Cycle ID: ${result.cycle_id}`}>
            <div className="grid gap-4 md:grid-cols-2">
              <StatCard label="Workers Evaluated" value={String(result.workers_evaluated)} />
              <StatCard label="Premiums Computed" value={String(result.premiums_computed)} tone="warm" />
              <StatCard label="Premium Failures" value={String(result.premium_failures)} tone={result.premium_failures ? 'alert' : 'default'} />
              <StatCard label="Status" value={String(result.status)} />
            </div>
          </Panel>

          <Panel title="Run Interpretation" subtitle="Use this summary during demo narration or while debugging data quality issues.">
            <ResultBox>
              <div className="space-y-2">
                <p><span className="font-semibold text-orange-300">cycle_id:</span> {result.cycle_id}</p>
                <p><span className="font-semibold text-orange-300">status:</span> {result.status}</p>
                <p><span className="font-semibold text-orange-300">evaluated:</span> {result.workers_evaluated}</p>
                <p><span className="font-semibold text-orange-300">computed:</span> {result.premiums_computed}</p>
                <p><span className="font-semibold text-orange-300">failures:</span> {result.premium_failures}</p>
              </div>
            </ResultBox>
          </Panel>
        </div>
      ) : null}
    </PageShell>
  )
}
