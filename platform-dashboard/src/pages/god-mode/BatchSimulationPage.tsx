import BatchesPage from './BatchesPage'

export default function BatchSimulationPage() {
  return (
    <div>
      <div>
        <h1 className="text-2xl font-black tracking-tight text-slate-900 dark:text-white">Batch Simulation & Testing</h1>
        <p className="mt-1 text-sm text-slate-600 dark:text-slate-400">
          Generate random orders and batches for each zone. Use this tool to test batch creation, pickup, and delivery flows.
        </p>
      </div>
      <div className="mt-8">
        <BatchesPage />
      </div>
    </div>
  )
}
