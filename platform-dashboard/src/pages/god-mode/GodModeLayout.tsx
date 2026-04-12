import BatchesPage from './BatchesPage'

export default function GodModeLayout() {
  return (
    <div>
      <div>
        <h1 className="text-2xl font-black tracking-tight text-slate-900 dark:text-white">Batch Browser</h1>
        <p className="mt-1 text-sm text-slate-500">
          View available and assigned batches by zone.
        </p>
      </div>
      <div className="mt-8">
        <BatchesPage />
      </div>
    </div>
  )
}
