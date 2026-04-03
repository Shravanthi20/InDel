import { useEffect, useState } from 'react'
import { getMaintenanceChecks, respondToCheck } from '../api/insurer'
import { PageShell, Panel, ResultBox } from './OperationsShared'

type MaintenanceRow = {
  id: number
  claim_id: number
  worker_id: number
  zone_name: string
  city: string
  status: string
  fraud_verdict: string
  claim_amount: number
  initiated_at: string
  response_at?: string
  findings: string
}

export default function MaintenanceChecks() {
  const [rows, setRows] = useState<MaintenanceRow[]>([])
  const [drafts, setDrafts] = useState<Record<number, string>>({})
  const [result, setResult] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const load = () => {
    getMaintenanceChecks({ page: 1, limit: 20 })
      .then((payload) => setRows(Array.isArray(payload.data) ? payload.data : []))
      .catch((err) => setError(err?.message ?? 'Failed to load maintenance checks'))
  }

  useEffect(() => {
    load()
  }, [])

  const submit = async (id: number) => {
    const findings = drafts[id]?.trim()
    if (!findings) {
      setResult('Add findings before submitting a response.')
      return
    }
    try {
      await respondToCheck(String(id), { findings })
      setResult(`Maintenance check ${id} updated.`)
      setDrafts((prev) => ({ ...prev, [id]: '' }))
      load()
    } catch (err: any) {
      setResult(err?.message ?? 'Failed to submit maintenance response.')
    }
  }

  return (
    <PageShell
      eyebrow="Maintenance"
      title="Maintenance Check Queue"
      description="Review worker disputes and record insurer-side findings against the linked claim."
    >
      {result ? <ResultBox>{result}</ResultBox> : null}
      <Panel title="Open Checks" subtitle="These cases are ready for insurer review and response.">
        {error ? <p className="text-sm text-rose-600">{error}</p> : null}
        <div className="space-y-4">
          {rows.map((row) => (
            <div key={row.id} className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
              <div className="flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
                <div>
                  <p className="text-sm font-semibold text-slate-950">
                    Check #{row.id} for claim #{row.claim_id}
                  </p>
                  <p className="text-sm text-slate-600">
                    Worker {row.worker_id} in {row.zone_name}, {row.city} • Rs {Math.round(row.claim_amount ?? 0)}
                  </p>
                  <p className="text-xs uppercase tracking-[0.22em] text-slate-500">
                    status {row.status} • fraud {row.fraud_verdict}
                  </p>
                </div>
                <p className="text-xs text-slate-500">Opened {row.initiated_at}</p>
              </div>

              {row.findings ? (
                <div className="mt-3 rounded-xl border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-950">
                  <span className="font-semibold">Existing findings:</span> {row.findings}
                </div>
              ) : null}

              <textarea
                className="mt-4 min-h-28 w-full rounded-2xl border border-slate-200 bg-white p-4 text-sm outline-none"
                placeholder="Record what the insurer reviewer found and what action was taken."
                value={drafts[row.id] ?? ''}
                onChange={(event) =>
                  setDrafts((prev) => ({ ...prev, [row.id]: event.target.value }))
                }
              />
              <div className="mt-3 flex items-center justify-between">
                <span className="text-xs text-slate-500">
                  {row.response_at ? `Responded at ${row.response_at}` : 'Awaiting response'}
                </span>
                <button
                  className="rounded-full bg-slate-950 px-4 py-2 text-sm font-semibold text-white"
                  onClick={() => submit(row.id)}
                >
                  Save Response
                </button>
              </div>
            </div>
          ))}
          {rows.length === 0 && !error ? (
            <p className="text-sm text-slate-500">No maintenance checks queued right now.</p>
          ) : null}
        </div>
      </Panel>
    </PageShell>
  )
}
