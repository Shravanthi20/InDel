import { useState } from 'react'
import { generateSyntheticData, type SyntheticScenario } from '../../api/operations'
import { PageShell, Panel, ResultBox, StatCard } from '../OperationsShared'

const scenarios: { label: string; value: SyntheticScenario }[] = [
  { label: 'Normal Week', value: 'normal_week' },
  { label: 'Mild Disruption', value: 'mild_disruption' },
  { label: 'Severe Disruption', value: 'severe_disruption' },
  { label: 'Fraud Burst', value: 'fraud_burst' },
]

export default function SyntheticGeneration() {
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
      eyebrow="Test Tools"
      title="Synthetic Signal Control"
      description="Seed deterministic backend data for the full InDel pipeline and validate system behavior."
    >
      <div className="grid gap-10 xl:grid-cols-[1.2fr_0.8fr]">
        <Panel title="Simulation Vector" subtitle="Configure the deterministic seed and scenario target.">
          <div className="space-y-8">
            <div className="grid gap-8 md:grid-cols-2">
              <label className="space-y-3">
                <span className="text-[9px] font-black uppercase tracking-[0.3em] text-slate-400 dark:text-slate-500">Entropy Seed</span>
                <input
                  type="number"
                  value={seed}
                  onChange={(e) => setSeed(Number(e.target.value))}
                  className="w-full rounded border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 px-6 py-4 text-xl font-black text-slate-900 dark:text-white outline-none focus:border-blue-600 transition-none"
                />
              </label>

              <label className="space-y-3">
                <span className="text-[9px] font-black uppercase tracking-[0.3em] text-slate-400 dark:text-slate-500">Namespace</span>
                <input
                  type="text"
                  value={outputDir}
                  onChange={(e) => setOutputDir(e.target.value)}
                  placeholder="synthetic/run-primary"
                  className="w-full rounded border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 px-6 py-4 text-lg font-bold text-slate-900 dark:text-white outline-none focus:border-blue-600 transition-none"
                />
              </label>
            </div>

            <div className="space-y-4">
              <p className="text-[9px] font-black uppercase tracking-[0.3em] text-slate-400 dark:text-slate-500 text-center">Scenario Selection</p>
              <div className="grid gap-4 md:grid-cols-2">
                {scenarios.map((item) => (
                  <button
                    key={item.value}
                    type="button"
                    onClick={() => setScenario(item.value)}
                    className={`rounded-lg border p-6 text-left transition-none ${
                        scenario === item.value
                          ? 'border-blue-600 bg-blue-600 text-white shadow-xl dark:border-blue-500 dark:bg-blue-500'
                          : 'border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-slate-500 dark:text-slate-400 hover:border-slate-300 dark:hover:border-slate-600'
                    }`}
                  >
                    <p className={`text-lg font-black tracking-tight ${scenario === item.value ? 'text-white' : 'text-slate-900 dark:text-slate-100'}`}>{item.label}</p>
                    <p className={`mt-2 text-[11px] leading-relaxed ${scenario === item.value ? 'text-blue-100' : 'text-slate-500 dark:text-slate-500'}`}>
                      {item.value.replace('_', ' ')}
                    </p>
                  </button>
                ))}
              </div>
            </div>

            <div className="flex items-center gap-10 pt-4 border-t border-slate-100 dark:border-slate-800">
              <button
                type="button"
                onClick={handleGenerate}
                disabled={loading}
                className="rounded bg-slate-900 dark:bg-white px-10 py-4 text-[11px] font-black uppercase tracking-widest text-white dark:text-slate-900 shadow-xl transition-none disabled:opacity-50"
              >
                {loading ? 'Scaling Signal...' : 'Trigger Simulation'}
              </button>
              {error ? <span className="text-[10px] font-bold text-rose-600 uppercase tracking-widest">{error}</span> : null}
            </div>
          </div>
        </Panel>

        <Panel title="Projections" subtitle="Projected impact of the current simulation.">
          <div className="grid gap-6">
            <StatCard label="Selected Seed" value={String(seed)} />
            <StatCard label="Scenario" value={scenario.replace('_', ' ')} tone="warm" />
            <div className="p-6 rounded border border-slate-100 dark:border-slate-800 bg-slate-50 dark:bg-slate-950 text-[11px] leading-relaxed text-slate-500 dark:text-slate-400">
               <p className="font-black text-slate-900 dark:text-white mb-2 uppercase tracking-widest text-[9px]">Simulation Model</p>
               Current backend request targets <span className="text-blue-600 dark:text-blue-300 font-bold uppercase">{scenario.replace('_', ' ')}</span> with seed <span className="font-bold text-slate-900 dark:text-slate-100">{seed}</span>.
            </div>
          </div>
        </Panel>
      </div>

      {result ? (
        <div className="grid gap-8 xl:grid-cols-[0.9fr_1.1fr] mt-10">
          <Panel title="Payload" subtitle={`Execution GUID: ${result.run_id}`}>
            <div className="grid gap-6 md:grid-cols-2">
              <StatCard label="Entity Map" value={String(counts?.workers ?? 0)} />
              <StatCard label="Primary Nodes" value={String(counts?.zones ?? 0)} />
              <StatCard label="Flagged Anomalies" value={String(counts?.disruptions ?? 0)} tone="warm" />
              <StatCard label="System Events" value={String(counts?.claims ?? 0)} tone="alert" />
            </div>
          </Panel>

          <Panel title="Artifacts" subtitle="Automated ETL export paths for analysis.">
            <ResultBox>
              <div className="space-y-4">
                {[
                  ['Entity Stream', artifacts?.workers_csv],
                  ['Anomaly Set', artifacts?.claims_csv],
                  ['Payout Vector', artifacts?.payouts_csv],
                  ['Entropy Seed', artifacts?.seed_sql]
                ].map(([label, path]) => (
                  <div key={label}>
                    <p className="text-[9px] font-black uppercase tracking-widest text-slate-500 dark:text-slate-500 mb-1">{label}</p>
                    <p className="truncate text-slate-900 dark:text-slate-300 font-mono text-[11px]">{path}</p>
                  </div>
                ))}
              </div>
            </ResultBox>
          </Panel>
        </div>
      ) : null}
    </PageShell>
  )
}
