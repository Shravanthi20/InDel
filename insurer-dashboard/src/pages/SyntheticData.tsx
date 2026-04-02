import { useState } from 'react'
import { generateSyntheticData, type SyntheticScenario } from '../api/operations'
import { PageShell, Panel, ResultBox, StatCard } from './OperationsShared'

const scenarios: { label: string; value: SyntheticScenario; note: string }[] = [
  { label: 'Normal Week', value: 'normal_week', note: 'Balanced baseline for demo seeding and happy-path portfolio flows.' },
  { label: 'Mild Disruption', value: 'mild_disruption', note: 'Small pressure on payouts without overwhelming portfolio health.' },
  { label: 'Severe Disruption', value: 'severe_disruption', note: 'High-loss week useful for claims and payout surge demos.' },
  { label: 'Fraud Burst', value: 'fraud_burst', note: 'Heavier flagged-claim mix for manual review and payout controls.' },
]

export default function SyntheticData() {
  const [seed, setSeed] = useState(42)
  const [scenario, setScenario] = useState<SyntheticScenario>('normal_week')
  const [outputDir, setOutputDir] = useState('')
  const [result, setResult] = useState<any>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function handleGenerate() {
    setLoading(true)
    setError('')
    try {
      const response = await generateSyntheticData({
        seed,
        scenario,
        output_dir: outputDir.trim() || undefined,
      })
      setResult(response.data.data)
    } catch (err: any) {
      setError(err?.response?.data?.error?.message || 'Synthetic generation failed.')
    } finally {
      setLoading(false)
    }
  }

  const counts = result?.counts
  const artifacts = result?.artifacts

  return (
    <PageShell
      eyebrow="Operations"
      title="Synthetic Data Control"
      description="Seed deterministic demo data for the full Part 4 pipeline. This is the fastest way to stand up a realistic insurer operations scenario without waiting on live platform data."
    >
      <div className="grid gap-6 xl:grid-cols-[1.2fr_0.8fr]">
        <Panel title="Generation Inputs" subtitle="Set the deterministic seed, choose the scenario, and optionally override the export directory.">
          <div className="space-y-5">
            <div className="grid gap-4 md:grid-cols-2">
              <label className="space-y-2">
                <span className="text-sm font-semibold text-slate-700">Seed</span>
                <input
                  type="number"
                  value={seed}
                  onChange={(e) => setSeed(Number(e.target.value))}
                  className="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-slate-900 outline-none ring-0 transition focus:border-orange-400"
                />
              </label>

              <label className="space-y-2">
                <span className="text-sm font-semibold text-slate-700">Output Directory</span>
                <input
                  type="text"
                  value={outputDir}
                  onChange={(e) => setOutputDir(e.target.value)}
                  placeholder="generated/synthetic/custom-run"
                  className="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-slate-900 outline-none transition focus:border-orange-400"
                />
              </label>
            </div>

            <div className="space-y-3">
              <p className="text-sm font-semibold text-slate-700">Scenario Preset</p>
              <div className="grid gap-3 md:grid-cols-2">
                {scenarios.map((item) => (
                  <button
                    key={item.value}
                    type="button"
                    onClick={() => setScenario(item.value)}
                    className={`rounded-2xl border p-4 text-left transition ${
                      scenario === item.value
                        ? 'border-orange-400 bg-orange-50 shadow-[0_10px_30px_rgba(251,146,60,0.18)]'
                        : 'border-slate-200 bg-slate-50 hover:border-slate-300 hover:bg-white'
                    }`}
                  >
                    <p className="font-bold text-slate-950">{item.label}</p>
                    <p className="mt-2 text-sm leading-6 text-slate-600">{item.note}</p>
                  </button>
                ))}
              </div>
            </div>

            <div className="flex items-center gap-4">
              <button
                type="button"
                onClick={handleGenerate}
                disabled={loading}
                className="rounded-full bg-slate-950 px-6 py-3 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {loading ? 'Generating...' : 'Generate Synthetic Data'}
              </button>
              {error ? <span className="text-sm font-medium text-rose-600">{error}</span> : null}
            </div>
          </div>
        </Panel>

        <Panel title="Expected Demo Outcome" subtitle="A quick map of what each seed run is meant to unlock in the rest of the operations UI.">
          <div className="grid gap-4">
            <StatCard label="Workers" value="500" />
            <StatCard label="Claims" value="2,000" tone="warm" />
            <StatCard label="Payouts" value="1,000+" />
            <StatCard label="Use Case" value={scenario.replace('_', ' ')} tone="alert" />
          </div>
        </Panel>
      </div>

      {result ? (
        <div className="grid gap-6 xl:grid-cols-[0.9fr_1.1fr]">
          <Panel title="Generation Summary" subtitle={`Run ID: ${result.run_id}`}>
            <div className="grid gap-4 md:grid-cols-2">
              <StatCard label="Workers" value={String(counts?.workers ?? 0)} />
              <StatCard label="Zones" value={String(counts?.zones ?? 0)} />
              <StatCard label="Disruptions" value={String(counts?.disruptions ?? 0)} tone="warm" />
              <StatCard label="Claims" value={String(counts?.claims ?? 0)} tone="alert" />
            </div>
          </Panel>

          <Panel title="Artifacts and Integration Notes" subtitle="Export paths for data consumers and the current Part 3 placeholders.">
            <ResultBox>
              <div className="space-y-3">
                <p><span className="font-semibold text-orange-300">workers_csv:</span> {artifacts?.workers_csv}</p>
                <p><span className="font-semibold text-orange-300">claims_csv:</span> {artifacts?.claims_csv}</p>
                <p><span className="font-semibold text-orange-300">payouts_csv:</span> {artifacts?.payouts_csv}</p>
                <p><span className="font-semibold text-orange-300">seed_sql:</span> {artifacts?.seed_sql}</p>
              </div>
            </ResultBox>
            <div className="mt-4 grid gap-3 text-sm text-slate-600">
              <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
                <p className="font-semibold text-slate-950">Premium Model Integration</p>
                <p className="mt-1">{result?.integration?.premium_service}</p>
              </div>
              <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
                <p className="font-semibold text-slate-950">Fraud Model Integration</p>
                <p className="mt-1">{result?.integration?.fraud_service}</p>
              </div>
            </div>
          </Panel>
        </div>
      ) : null}
    </PageShell>
  )
}
