import { CarFront, CloudRain, ThermometerSun, Wind, type LucideIcon } from 'lucide-react'
import { useGodMode, formatCurrency, type EnvInputs } from './state'

type FactorMeta = {
  key: keyof EnvInputs
  title: string
  helper: string
  min: number
  max: number
  step: number
  icon: LucideIcon
}

const factorMeta: Record<keyof EnvInputs, FactorMeta> = {
  temperature: { key: 'temperature', title: 'Temperature', helper: '°C', min: 20, max: 50, step: 0.5, icon: ThermometerSun },
  rain: { key: 'rain', title: 'Rain intensity', helper: 'mm/hour', min: 0, max: 20, step: 0.1, icon: CloudRain },
  aqi: { key: 'aqi', title: 'AQI', helper: 'Air Quality Index', min: 80, max: 350, step: 1, icon: Wind },
  traffic: { key: 'traffic', title: 'Traffic congestion', helper: '%', min: 0, max: 100, step: 1, icon: CarFront },
}

export default function FactorPage({ factorKey }: { factorKey: keyof EnvInputs }) {
  const {
    godModeEnabled,
    apiInputs,
    manualInputs,
    policyInputs,
    setManualInput,
    setPolicyInput,
    result,
  } = useGodMode()

  const meta = factorMeta[factorKey]
  const Icon = meta.icon

  const currentInputs = godModeEnabled ? manualInputs : apiInputs
  const value = currentInputs[factorKey]
  const display = meta.step < 1 ? value.toFixed(2) : value.toFixed(0)

  return (
    <section className="rounded-[1.75rem] border border-slate-200 bg-white p-5 shadow-sm">
      <div className="grid gap-6 xl:grid-cols-[1fr_1fr]">
        <div className="space-y-4">
          <div>
            <p className="text-[11px] uppercase tracking-[0.3em] text-slate-500">Factor Control</p>
            <h2 className="mt-1 text-2xl font-bold text-slate-900">{meta.title}</h2>
            <p className="mt-2 text-sm text-slate-600">
              {godModeEnabled
                ? 'Manual override is enabled. This slider will be used by simulation.'
                : 'God Mode is OFF. This value currently mirrors API/mock input.'}
            </p>
          </div>

          <div className={`rounded-2xl border p-4 ${value > (meta.max * 0.7) ? 'border-rose-200 bg-rose-50' : 'border-slate-200 bg-slate-50'}`}>
            <div className="flex items-center justify-between gap-3">
              <div className="flex items-center gap-2 text-sm font-semibold text-slate-900">
                <Icon className="h-4 w-4 text-sky-700" />
                {meta.title}
              </div>
              <div className="rounded-full border border-slate-200 bg-white px-3 py-1 font-mono text-xs text-slate-700">
                {display} {meta.helper}
              </div>
            </div>
            <input
              type="range"
              min={meta.min}
              max={meta.max}
              step={meta.step}
              value={manualInputs[factorKey]}
              disabled={!godModeEnabled}
              onChange={(event) => setManualInput(factorKey, Number(event.target.value))}
              className="mt-3 h-2 w-full cursor-pointer appearance-none rounded-full bg-slate-200 accent-sky-600 disabled:cursor-not-allowed"
            />
            <div className="mt-2 flex justify-between text-[10px] uppercase tracking-[0.2em] text-slate-500">
              <span>{meta.min}</span>
              <span>{meta.max}</span>
            </div>
          </div>

          <div className="grid gap-3 md:grid-cols-2">
            <PolicyCard
              label="Max payout per day"
              value={`${Math.round(policyInputs.maxPayoutPerDay)}`}
              min={500}
              max={5000}
              step={50}
              onChange={(value) => setPolicyInput('maxPayoutPerDay', value)}
            />
            <PolicyCard
              label="Coverage ratio"
              value={policyInputs.coverageRatio.toFixed(2)}
              min={0}
              max={1}
              step={0.01}
              onChange={(value) => setPolicyInput('coverageRatio', value)}
            />
          </div>
        </div>

        <div className="space-y-4">
          <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
            <div className="text-[11px] uppercase tracking-[0.3em] text-slate-500">Live simulation snapshot</div>
            <div className="mt-2 text-sm text-slate-300">Run Simulation to commit this state. Snapshot below reflects last run.</div>
          </div>

          <div className="grid gap-3 sm:grid-cols-2">
            <ResultTile label="Risk score" value={result.riskScore.toFixed(2)} critical={result.riskScore >= 0.75} normal={result.riskScore < 0.3} />
            <ResultTile label="Payout" value={formatCurrency(result.payout)} critical={result.payout > 1200} normal={result.payout === 0} />
            <ResultTile label="Status" value={result.status} critical={result.status === 'Severe'} normal={result.status === 'Normal'} />
            <ResultTile label="Dominant reason" value={result.reason} critical={false} normal={false} />
          </div>
        </div>
      </div>
    </section>
  )
}

function PolicyCard({
  label,
  value,
  min,
  max,
  step,
  onChange,
}: {
  label: string
  value: string
  min: number
  max: number
  step: number
  onChange: (value: number) => void
}) {
  return (
    <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
      <div className="text-[11px] uppercase tracking-[0.22em] text-slate-500">{label}</div>
      <div className="mt-1 text-xl font-bold text-slate-900">{value}</div>
      <input
        type="range"
        min={min}
        max={max}
        step={step}
        value={Number(value)}
        onChange={(event) => onChange(Number(event.target.value))}
        className="mt-3 h-2 w-full cursor-pointer appearance-none rounded-full bg-slate-200 accent-sky-600"
      />
    </div>
  )
}

function ResultTile({ label, value, critical, normal }: { label: string; value: string; critical: boolean; normal: boolean }) {
  const tone = critical
    ? 'border-rose-200 bg-rose-50 text-rose-800'
    : normal
      ? 'border-emerald-200 bg-emerald-50 text-emerald-800'
      : 'border-sky-200 bg-sky-50 text-sky-800'
  return (
    <div className={`rounded-2xl border p-3 ${tone}`}>
      <div className="text-[10px] uppercase tracking-[0.25em] opacity-70">{label}</div>
      <div className="mt-1 text-lg font-semibold text-slate-900">{value}</div>
    </div>
  )
}
