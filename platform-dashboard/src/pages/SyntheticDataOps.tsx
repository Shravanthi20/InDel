import SyntheticGeneration from './god-mode/SyntheticGeneration'

export default function SyntheticDataOps() {
  return (
    <div>
      <h1 className="text-2xl font-black tracking-tight text-slate-900 dark:text-white">Synthetic Data Studio</h1>
      <p className="mt-1 text-sm text-slate-500">
        Generate controlled datasets and artifacts for disruption, claims, and payout validation.
      </p>
      <div className="mt-8">
        <SyntheticGeneration />
      </div>
    </div>
  )
}
