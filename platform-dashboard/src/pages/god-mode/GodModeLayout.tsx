import { CarFront, CloudRain, ThermometerSun, Wind, type LucideIcon } from 'lucide-react'
import BatchesPage from './BatchesPage'
import { useGodMode } from './state'

type FactorMeta = {
  key: 'temperature' | 'rain' | 'aqi' | 'traffic'
  title: string
  helper: string
  min: number
  max: number
  step: number
  icon: LucideIcon
}

const factorMeta: FactorMeta[] = [
  { key: 'temperature', title: 'Temperature', helper: '°C', min: 20, max: 50, step: 0.5, icon: ThermometerSun },
  { key: 'rain', title: 'Rain intensity', helper: 'mm/hour', min: 0, max: 20, step: 0.1, icon: CloudRain },
  { key: 'aqi', title: 'AQI', helper: 'Air Quality Index', min: 80, max: 350, step: 1, icon: Wind },
  { key: 'traffic', title: 'Traffic congestion', helper: '%', min: 0, max: 100, step: 1, icon: CarFront },
]

export default function GodModeLayout() {
  const {
    godModeEnabled,
    setGodModeEnabled,
    scopeLabel,
    affectedZoneIds,
    manualInputs,
    apiInputs,
    setManualInput,
    generatingBatches,
    generateBatches,
  } = useGodMode()

  const currentInputs = godModeEnabled ? manualInputs : apiInputs

  return (
    <div className="min-h-screen bg-slate-50 p-5 text-slate-900 lg:p-8">
      <div className="mx-auto max-w-[1560px] space-y-6">
        <section className="rounded-[2rem] border border-slate-200 bg-white p-6 shadow-sm">
          <div className="flex flex-col gap-5 xl:flex-row xl:items-center xl:justify-between">
            <div>
              <p className="text-[11px] uppercase tracking-[0.35em] text-sky-700">God Mode Batch Control</p>
              <h1 className="mt-2 text-4xl font-black tracking-tight text-slate-900">Factor pages with batch visibility</h1>
              <p className="mt-2 max-w-3xl text-sm leading-7 text-slate-600">
                Tune the environment factors, generate batches, and inspect available or assigned batches by zone.
              </p>
            </div>

            <div className="grid gap-3 rounded-[1.5rem] border border-slate-200 bg-slate-50 p-4 sm:grid-cols-2 xl:min-w-[640px] xl:grid-cols-[repeat(2,minmax(0,1fr))_auto]">
              <div className="flex flex-col gap-2 sm:col-span-2 xl:col-span-1">
                <span className="text-xs font-semibold uppercase tracking-[0.22em] text-slate-500">Scope</span>
                <div className="rounded-2xl border border-slate-200 bg-white px-3 py-3 text-sm font-semibold text-slate-800">
                  {scopeLabel}
                </div>
                <div className="text-[11px] uppercase tracking-[0.22em] text-slate-500">Affected zones: {affectedZoneIds.length}</div>
              </div>

              <div className="flex flex-wrap items-end gap-3 xl:justify-end">
                <button
                  type="button"
                  onClick={() => setGodModeEnabled(!godModeEnabled)}
                  className={`rounded-full border px-5 py-2 text-sm font-semibold uppercase tracking-[0.2em] transition ${godModeEnabled
                    ? 'border-rose-300 bg-rose-50 text-rose-700 hover:bg-rose-100'
                    : 'border-sky-300 bg-sky-50 text-sky-700 hover:bg-sky-100'}`}
                >
                  {godModeEnabled ? 'Disable God Mode' : 'Enable God Mode'}
                </button>

                <button
                  type="button"
                  onClick={generateBatches}
                  disabled={generatingBatches}
                  className="rounded-full border border-slate-300 bg-white px-5 py-2 text-sm font-semibold uppercase tracking-[0.2em] text-slate-800 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {generatingBatches ? 'Adding...' : 'Add Batches'}
                </button>
              </div>
            </div>
          </div>

        </section>

        <section className="grid gap-4 xl:grid-cols-2">
          {factorMeta.map((factor) => {
            const Icon = factor.icon
            const value = currentInputs[factor.key]
            const display = factor.step < 1 ? value.toFixed(2) : value.toFixed(0)

            return (
              <article key={factor.key} className="rounded-[1.75rem] border border-slate-200 bg-white p-5 shadow-sm">
                <div className="space-y-4">
                  <div>
                    <div className="flex items-center gap-2 text-sm font-semibold text-slate-900">
                      <Icon className="h-4 w-4 text-sky-700" />
                      {factor.title}
                    </div>
                    <p className="mt-2 text-sm text-slate-600">
                      {godModeEnabled
                        ? 'Manual override is enabled. This value is used for triggering disruptions.'
                        : 'This value currently mirrors the API/mock input.'}
                    </p>
                  </div>

                  <div className={`rounded-2xl border p-4 ${value > (factor.max * 0.7) ? 'border-rose-200 bg-rose-50' : 'border-slate-200 bg-slate-50'}`}>
                    <div className="flex items-center justify-between gap-3">
                      <div className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">{factor.title}</div>
                      <div className="rounded-full border border-slate-200 bg-white px-3 py-1 font-mono text-xs text-slate-700">
                        {display} {factor.helper}
                      </div>
                    </div>
                    <input
                      type="range"
                      min={factor.min}
                      max={factor.max}
                      step={factor.step}
                      value={currentInputs[factor.key]}
                      disabled={!godModeEnabled}
                      onChange={(event) => setManualInput(factor.key, Number(event.target.value))}
                      className="mt-3 h-2 w-full cursor-pointer appearance-none rounded-full bg-slate-200 accent-sky-600 disabled:cursor-not-allowed"
                    />
                    <div className="mt-2 flex justify-between text-[10px] uppercase tracking-[0.2em] text-slate-500">
                      <span>{factor.min}</span>
                      <span>{factor.max}</span>
                    </div>
                  </div>
                </div>
              </article>
            )
          })}
        </section>

        <BatchesPage />
      </div>
    </div>
  )
}
