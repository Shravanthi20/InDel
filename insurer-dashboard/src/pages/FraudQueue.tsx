import { useEffect, useState } from 'react'
import { getFraudQueue } from '../api/insurer'
import { PageShell, Panel } from './OperationsShared'

type FraudRow = {
  claim_id: number
  status: string
  fraud_verdict: string
  fraud_score: number
  created_at: string
}

export default function FraudQueue() {
  const [rows, setRows] = useState<FraudRow[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    getFraudQueue({ page: 1, limit: 20 })
      .then((payload) => setRows(Array.isArray(payload.data) ? payload.data : []))
      .catch((err) => setError(err?.message ?? 'Failed to load fraud queue'))
  }, [])

  return (
    <PageShell
      eyebrow="Fraud"
      title="Fraud Queue"
      description="Review the claims that the scoring layer has routed for manual attention."
    >
      <Panel title="Flagged Claims">
        {error ? <p className="text-sm text-rose-600">{error}</p> : null}
        <div className="overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead className="text-left text-slate-500">
              <tr>
                <th className="pb-3 pr-4">Claim</th>
                <th className="pb-3 pr-4">Status</th>
                <th className="pb-3 pr-4">Verdict</th>
                <th className="pb-3 pr-4">Score</th>
                <th className="pb-3 pr-4">Created</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((row) => (
                <tr key={row.claim_id} className="border-t border-slate-100">
                  <td className="py-3 pr-4 font-medium text-slate-900">#{row.claim_id}</td>
                  <td className="py-3 pr-4">{row.status}</td>
                  <td className="py-3 pr-4">{row.fraud_verdict}</td>
                  <td className="py-3 pr-4">{(row.fraud_score ?? 0).toFixed(2)}</td>
                  <td className="py-3 pr-4">{row.created_at}</td>
                </tr>
              ))}
              {rows.length === 0 && !error ? (
                <tr>
                  <td className="py-6 text-slate-500" colSpan={5}>No flagged claims currently.</td>
                </tr>
              ) : null}
            </tbody>
          </table>
        </div>
      </Panel>
    </PageShell>
  )
}
