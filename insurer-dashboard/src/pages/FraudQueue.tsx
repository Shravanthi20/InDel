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
      eyebrow="Security"
      title="Fraud Analysis Queue"
      description="Review high-risk claims routed for manual verification by the ML scoring engine."
    >
      <div className="grid gap-8">
        <Panel title="Security Flags" subtitle="Sorted by anomaly priority for manual review.">
          {error ? <p className="text-[10px] font-black uppercase text-rose-600 mb-6 font-bold">{error}</p> : null}
          <div className="overflow-x-auto">
            <table className="w-full text-left text-[11px]">
              <thead>
                <tr className="border-b border-slate-200 dark:border-slate-800">
                  <th className="pb-4 font-black uppercase tracking-widest text-slate-400">ID</th>
                  <th className="pb-4 font-black uppercase tracking-widest text-slate-400">Signal</th>
                  <th className="pb-4 font-black uppercase tracking-widest text-slate-400 text-right">Anomaly Score</th>
                  <th className="pb-4 text-right font-black uppercase tracking-widest text-slate-400">Verification</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 dark:divide-slate-800/50">
                {rows.map((row) => (
                  <tr key={row.claim_id} className="hover:bg-slate-50 dark:hover:bg-slate-800 transition-none">
                    <td className="py-5 font-bold text-slate-900 dark:text-white">#{row.claim_id}</td>
                    <td className="py-5">
                       <span className="px-2 py-0.5 rounded text-[9px] font-black uppercase tracking-widest bg-slate-100 dark:bg-slate-800 text-slate-500">
                          {row.status}
                       </span>
                    </td>
                    <td className="py-5 text-right font-black text-slate-900 dark:text-white">
                       <div className="flex items-center justify-end gap-3">
                          <span className="text-sm font-black">{(row.fraud_score ?? 0).toFixed(2)}</span>
                          <div className="h-1 w-16 bg-slate-100 dark:bg-slate-800 overflow-hidden">
                             <div 
                                className={`h-full ${row.fraud_score > 0.7 ? 'bg-rose-500' : 'bg-orange-500'}`}
                                style={{ width: `${Math.min((row.fraud_score ?? 0) * 100, 100)}%` }}
                             ></div>
                          </div>
                       </div>
                    </td>
                    <td className="py-5 text-right font-black tracking-[0.2em] uppercase text-[9px]">
                       <span className={row.fraud_verdict === 'safe' ? 'text-emerald-600' : 'text-rose-600'}>
                          {row.fraud_verdict}
                       </span>
                    </td>
                  </tr>
                ))}
                {rows.length === 0 && !error ? (
                  <tr>
                    <td className="py-12 text-center text-slate-400 italic" colSpan={4}>
                       Security clearance complete. No flags.
                    </td>
                  </tr>
                ) : null}
              </tbody>
            </table>
          </div>
        </Panel>
      </div>
    </PageShell>
  )
}
