import { useEffect, useState } from 'react'
import { getClaims } from '../api/insurer'
import { PageShell, Panel } from './OperationsShared'

type ClaimRow = {
  claim_id: number
  worker_id: number
  zone_name: string
  claim_amount: number
  status: string
  fraud_verdict: string
  created_at: string
}

export default function Claims() {
  const [rows, setRows] = useState<ClaimRow[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    getClaims({ page: 1, limit: 20 })
      .then((payload) => setRows(Array.isArray(payload.data) ? payload.data : []))
      .catch((err) => setError(err?.message ?? 'Failed to load claims'))
  }, [])

  return (
    <PageShell
      eyebrow="Claims"
      title="Claims Pipeline"
      description="Inspect the current automated claim stream, including status and fraud verdict at a glance."
    >
      <Panel title="Recent Claims">
        {error ? <p className="text-sm text-rose-600">{error}</p> : null}
        <div className="overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead className="text-left text-slate-500">
              <tr>
                <th className="pb-3 pr-4">Claim</th>
                <th className="pb-3 pr-4">Worker</th>
                <th className="pb-3 pr-4">Zone</th>
                <th className="pb-3 pr-4">Amount</th>
                <th className="pb-3 pr-4">Status</th>
                <th className="pb-3 pr-4">Fraud</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((row) => (
                <tr key={row.claim_id} className="border-t border-slate-100">
                  <td className="py-3 pr-4 font-medium text-slate-900">#{row.claim_id}</td>
                  <td className="py-3 pr-4">{row.worker_id}</td>
                  <td className="py-3 pr-4">{row.zone_name}</td>
                  <td className="py-3 pr-4">Rs {Math.round(row.claim_amount ?? 0)}</td>
                  <td className="py-3 pr-4">{row.status}</td>
                  <td className="py-3 pr-4">{row.fraud_verdict}</td>
                </tr>
              ))}
              {rows.length === 0 && !error ? (
                <tr>
                  <td className="py-6 text-slate-500" colSpan={6}>No claims available yet.</td>
                </tr>
              ) : null}
            </tbody>
          </table>
        </div>
      </Panel>
    </PageShell>
  )
}
