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
      eyebrow="Pipeline"
      title="Global Claims Stream"
      description="Inspect real-time payout requests, fraud scores, and settlement status across the ecosystem."
    >
      <Panel title="Active Stream" subtitle="Showing the most recent 20 claim events.">
        {error ? <p className="text-[10px] font-black uppercase text-rose-600 mb-6 font-bold">{error}</p> : null}
        <div className="overflow-x-auto">
          <table className="w-full text-left text-[11px]">
            <thead>
              <tr className="border-b border-slate-200 dark:border-slate-800">
                <th className="pb-4 font-black uppercase tracking-widest text-slate-400">ID</th>
                <th className="pb-4 font-black uppercase tracking-widest text-slate-400">Worker</th>
                <th className="pb-4 font-black uppercase tracking-widest text-slate-400">Zone</th>
                <th className="pb-4 font-black uppercase tracking-widest text-slate-400">Valuation</th>
                <th className="pb-4 font-black uppercase tracking-widest text-slate-400">Status</th>
                <th className="pb-4 text-right font-black uppercase tracking-widest text-slate-400">Security</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-slate-800/50">
              {rows.map((row) => (
                <tr key={row.claim_id} className="hover:bg-slate-50 dark:hover:bg-slate-800 transition-none">
                  <td className="py-4 font-bold text-slate-900 dark:text-white">#{row.claim_id}</td>
                  <td className="py-4 text-slate-500">{row.worker_id}</td>
                  <td className="py-4 font-medium text-slate-700 dark:text-slate-300">{row.zone_name}</td>
                  <td className="py-4 font-bold text-slate-900 dark:text-white">Rs {Math.round(row.claim_amount ?? 0).toLocaleString()}</td>
                  <td className="py-4">
                    <span className={`px-2 py-0.5 rounded text-[9px] font-black uppercase tracking-widest ${
                      row.status === 'approved' ? 'bg-emerald-100 text-emerald-700' : 'bg-amber-100 text-amber-700'
                    }`}>
                      {row.status}
                    </span>
                  </td>
                  <td className="py-4 text-right">
                    <span className={`px-2 py-0.5 rounded text-[9px] font-black uppercase tracking-widest leading-none ${
                      row.fraud_verdict === 'safe' ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'
                    }`}>
                      {row.fraud_verdict}
                    </span>
                  </td>
                </tr>
              ))}
              {rows.length === 0 && !error ? (
                <tr>
                  <td className="py-12 text-center text-slate-400 italic" colSpan={6}>
                    Awaiting real-time claim signals.
                  </td>
                </tr>
              ) : null}
            </tbody>
          </table>
        </div>
      </Panel>
    </PageShell>
  )
}
