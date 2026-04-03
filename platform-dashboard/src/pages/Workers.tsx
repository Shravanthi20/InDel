import { useEffect, useState } from 'react'
import { getWorkers } from '../api/platform'

export default function Workers() {
  const [workers, setWorkers] = useState<any[]>([])

  useEffect(() => {
    async function load() {
      const response = await getWorkers()
      setWorkers(response.data.workers ?? [])
    }

    load().catch((error) => console.error('Failed to load workers', error))
  }, [])

  return (
    <div className="p-6">
      <h1 className="text-3xl font-bold">Worker Management</h1>
      <div className="mt-6 rounded-2xl bg-white p-6 shadow">
        <div className="mb-4 text-sm text-slate-600">Workers visible to platform operations and linked to covered zones.</div>
        <div className="space-y-3">
          {workers.map((worker) => (
            <div key={worker.worker_id} className="rounded-xl border border-slate-200 p-4">
              <div className="font-semibold text-slate-900">{worker.name}</div>
              <div className="mt-1 text-sm text-slate-600">{worker.phone}</div>
              <div className="mt-1 text-sm text-slate-500">{worker.zone}</div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
