import { useEffect, useState } from 'react'
import { getForecast } from '../api/insurer'
import { PageShell, Panel } from './OperationsShared'

type ForecastRow = {
  city: string
  zone: string
  date: string
  probability: number
}

export default function Forecast() {
  const [rows, setRows] = useState<ForecastRow[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    getForecast()
      .then((data) => setRows(Array.isArray(data) ? data : []))
      .catch((err) => setError(err?.message ?? 'Failed to load forecast'))
  }, [])

  return (
    <PageShell
      eyebrow="Forecast"
      title="7-Day Forecast"
      description="Surface near-term disruption probability so reserve planning can happen before claims arrive."
    >
      <Panel title="Upcoming Disruption Risk">
        {error ? <p className="text-sm text-rose-600">{error}</p> : null}
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {rows.map((row) => (
            <div key={`${row.zone}-${row.date}`} className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
              <p className="text-xs uppercase tracking-[0.22em] text-slate-500">{row.date}</p>
              <p className="mt-2 text-lg font-bold text-slate-950">{row.zone}, {row.city}</p>
              <p className="mt-3 text-sm text-slate-600">Disruption probability</p>
              <p className="text-3xl font-black text-slate-950">{Math.round((row.probability ?? 0) * 100)}%</p>
            </div>
          ))}
          {rows.length === 0 && !error ? (
            <p className="text-sm text-slate-500">No forecast outputs available yet.</p>
          ) : null}
        </div>
      </Panel>
    </PageShell>
  )
}
