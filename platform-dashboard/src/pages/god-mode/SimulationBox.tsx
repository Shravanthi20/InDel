import { Activity, ArrowRight, SlidersHorizontal } from 'lucide-react'
import { formatCurrency, useGodMode } from './state'

function statusBadge(status: string) {
  if (status === 'Severe') return 'border-rose-200 bg-rose-50 text-rose-800'
  if (status === 'Critical') return 'border-orange-200 bg-orange-50 text-orange-800'
  if (status === 'Reduced Operation') return 'border-amber-200 bg-amber-50 text-amber-800'
  return 'border-emerald-200 bg-emerald-50 text-emerald-800'
}

export default function SimulationBox() {
  const {
    previewResult,
    policyInputs,
    setPolicyInput,
    runSimulation,
    running,
    loading,
  } = useGodMode()

  const maxPayoutDisplay = formatCurrency(policyInputs.maxPayoutPerDay)
  const coveragePercent = Math.round(policyInputs.coverageRatio * 100)
  const coverageCap = formatCurrency(policyInputs.maxPayoutPerDay * policyInputs.coverageRatio)

  return (
    <section className="rounded-[1.75rem] border border-slate-200 bg-white p-5 shadow-sm">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="text-[11px] uppercase tracking-[0.3em] text-slate-500">Simulation box</p>
          <h2 className="mt-1 text-2xl font-bold text-slate-900">Preview payout before triggering</h2>
          <p className="mt-2 max-w-3xl text-sm leading-6 text-slate-600">
            Move the payout and coverage sliders, then run the simulation to see the risk score, projected payout, and zone impact.
          </p>
        </div>

        <div className={`rounded-full border px-4 py-1 text-xs font-semibold uppercase tracking-[0.25em] ${statusBadge(previewResult.status)}`}>
          {previewResult.status}
        </div>
      </div>

      <div className="mt-5 grid gap-4 xl:grid-cols-[1.2fr_0.8fr]">
        <div className="space-y-4 rounded-2xl border border-slate-200 bg-slate-50 p-4">
          <div className="flex items-center gap-2 text-sm font-semibold text-slate-900">
            <SlidersHorizontal className="h-4 w-4 text-sky-700" />
            Policy sliders
          </div>

          <div className="space-y-4">
            <label className="grid gap-2 text-sm font-medium text-slate-700">
              <div className="flex items-center justify-between gap-3">
                <span>Maximum payout per day</span>
                <span className="rounded-full border border-slate-200 bg-white px-3 py-1 text-xs font-semibold text-slate-800">
                  {maxPayoutDisplay}
                </span>
              </div>
              <input
                type="range"
                min={500}
                max={10000}
                step={100}
                value={policyInputs.maxPayoutPerDay}
                onChange={(event) => setPolicyInput('maxPayoutPerDay', Number(event.target.value))}
                className="h-2 w-full cursor-pointer appearance-none rounded-full bg-slate-200 accent-sky-600"
              />
            </label>

            <label className="grid gap-2 text-sm font-medium text-slate-700">
              <div className="flex items-center justify-between gap-3">
                <span>Coverage ratio</span>
                <span className="rounded-full border border-slate-200 bg-white px-3 py-1 text-xs font-semibold text-slate-800">
                  {coveragePercent}%
                </span>
              </div>
              <input
                type="range"
                min={0}
                max={1}
                step={0.01}
                value={policyInputs.coverageRatio}
                onChange={(event) => setPolicyInput('coverageRatio', Number(event.target.value))}
                className="h-2 w-full cursor-pointer appearance-none rounded-full bg-slate-200 accent-sky-600"
              />
            </label>

            <div className="rounded-2xl border border-dashed border-slate-300 bg-white p-4 text-sm text-slate-600">
              Live preview uses the zone sliders above and these policy values.
            </div>
          </div>

          <button
            type="button"
            onClick={runSimulation}
            disabled={running || loading}
            className="inline-flex items-center justify-center gap-2 rounded-full border border-slate-300 bg-white px-5 py-2 text-sm font-semibold uppercase tracking-[0.2em] text-slate-800 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {running ? 'Running...' : 'Run simulation'}
            <ArrowRight className="h-4 w-4" />
          </button>
        </div>

        <div className="space-y-4 rounded-2xl border border-slate-200 bg-slate-50 p-4">
          <div className="flex items-center gap-2 text-sm font-semibold text-slate-900">
            <Activity className="h-4 w-4 text-sky-700" />
            Live payout preview
          </div>

          <div className="grid gap-3 sm:grid-cols-2">
            <PreviewTile label="Risk score" value={previewResult.riskScore.toFixed(2)} />
            <PreviewTile label="Projected payout" value={formatCurrency(previewResult.payout)} />
            <PreviewTile label="Cash outflow" value={formatCurrency(previewResult.cashOutflow)} />
            <PreviewTile label="Coverage ratio" value={`${coveragePercent}%`} />
          </div>

          <div className="rounded-2xl border border-slate-200 bg-white p-4">
            <div className="text-[11px] uppercase tracking-[0.25em] text-slate-500">What the sliders mean</div>
            <div className="mt-2 space-y-2 text-sm text-slate-700">
              <div>Current maximum payout: <span className="font-semibold text-slate-900">{maxPayoutDisplay}</span></div>
              <div>Current coverage cap: <span className="font-semibold text-slate-900">{coverageCap}</span></div>
              <div>Risk response: <span className="font-semibold text-slate-900">{previewResult.reason}</span></div>
            </div>
          </div>

          <div className={`rounded-2xl border px-4 py-3 text-sm ${statusBadge(previewResult.status)}`}>
            <div className="font-semibold uppercase tracking-[0.2em]">Projected action</div>
            <div className="mt-2">{previewResult.notification}</div>
          </div>
        </div>
      </div>
    </section>
  )
}

function PreviewTile({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-3">
      <div className="text-[10px] uppercase tracking-[0.25em] text-slate-500">{label}</div>
      <div className="mt-1 text-xl font-black text-slate-900">{value}</div>
    </div>
  )
}