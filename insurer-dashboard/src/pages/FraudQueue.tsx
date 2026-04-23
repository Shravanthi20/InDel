import { useEffect, useState } from 'react'
import { getFraudQueue, reviewClaim, type FraudQueueRow } from '../api/insurer'
import { PageShell, Panel } from './OperationsShared'
import { useLocalization } from '../context/LocalizationContext'

export default function FraudQueue() {
  const { t } = useLocalization()
  const [rows, setRows] = useState<FraudQueueRow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const JUSTIFICATION_MAP: Record<string, string> = {
    "iso_score": "Isolation Anomaly Detected",
    "daily_earnings_drop": "Unusual inactivity during disruption",
    "historical_deviation": "Significant deviation from average earnings",
    "participation_lag": "Late signal acknowledgment detected",
    "zero_activity_suspect": "Active but zero orders reported during event",
    "extreme_order_value": "Outlier order values detected post-disruption"
  }

  const loadQueue = async () => {
    setLoading(true)
    try {
      const res = await getFraudQueue()
      setRows(res.data)
      setError(null)
    } catch (e: any) {
      setError(e?.message ?? 'Failed to load queue')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadQueue()
  }, [])

  const handleAction = async (id: number, status: string, verdict: string) => {
    try {
      await reviewClaim(id, status, verdict)
      loadQueue()
    } catch (e: any) {
      setError(e?.message ?? 'Failed to perform action')
    }
  }

  const getJustification = (signal: string) => JUSTIFICATION_MAP[signal] || signal

  return (
    <PageShell
      eyebrow={t('pages.fraudQueue.eyebrow')}
      title={t('pages.fraudQueue.title')}
      description={t('pages.fraudQueue.description')}
    >
      <div className="grid gap-8">
        <Panel title="Verification Queue" subtitle="Claims requiring human-in-the-loop validation based on ML anomaly signals.">
          {error ? <p className="mb-6 text-[10px] font-black uppercase text-rose-600 font-mono">{error}</p> : null}
          {loading ? (
            <div className="flex items-center gap-3 mb-10 text-gray-400">
               <div className="h-4 w-4 rounded-full border-2 border-brand-primary border-t-transparent animate-spin"></div>
               <p className="text-sm font-medium">Analyzing neural risk signals...</p>
            </div>
          ) : null}
          <div className="overflow-x-auto">
            <table className="w-full text-left text-[11px]">
              <thead>
                <tr className="border-b border-gray-100 dark:border-gray-800">
                  <th className="pb-5 font-black uppercase tracking-[0.2em] text-gray-400">Audit ID</th>
                  <th className="pb-5 font-black uppercase tracking-[0.2em] text-gray-400">Risk Justification</th>
                  <th className="pb-5 font-black uppercase tracking-[0.2em] text-gray-400 text-right">Anomaly Score</th>
                  <th className="pb-5 text-right font-black uppercase tracking-[0.2em] text-gray-400">Human Override</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-50 dark:divide-gray-800/50">
                {rows.map((row) => (
                  <tr key={row.claim_id} className="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-all">
                    <td className="py-6">
                      <p className="font-black text-gray-900 dark:text-white mb-1">CLAIM-{row.claim_id}</p>
                      <p className="text-[9px] font-bold text-gray-400 uppercase tracking-widest">Worker: {row.worker_id}</p>
                    </td>
                    <td className="py-6">
                      <div className="flex flex-wrap gap-2">
                        {row.violations && row.violations.length > 0 ? (
                          row.violations.map((v, i) => (
                            <span key={i} className="rounded-full bg-brand-soft px-3 py-1 text-[9px] font-black uppercase tracking-widest text-brand-dark border border-brand-primary/10">
                              {getJustification(v)}
                            </span>
                          ))
                        ) : (
                          <span className="rounded-full bg-gray-100 px-3 py-1 text-[9px] font-black uppercase tracking-widest text-gray-500">
                            {row.status}
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="py-6 text-right font-black text-gray-900 dark:text-white">
                      <div className="flex items-center justify-end gap-4">
                        <span className="text-sm font-black text-brand-primary">{(row.fraud_score ?? 0).toFixed(3)}</span>
                        <div className="h-1.5 w-24 overflow-hidden bg-gray-100 dark:bg-gray-800 rounded-full">
                          <div
                            className={`h-full rounded-full transition-all duration-1000 ${row.fraud_score > 0.7 ? 'bg-brand-primary' : 'bg-brand-primary/40'}`}
                            style={{ width: `${Math.min((row.fraud_score ?? 0) * 100, 100)}%` }}
                          />
                        </div>
                      </div>
                    </td>
                    <td className="py-6 text-right">
                      {row.fraud_verdict === 'delay' || row.status === 'manual_review' ? (
                        <div className="flex justify-end gap-3">
                          <button
                            onClick={() => handleAction(row.claim_id, 'approved', 'safe')}
                            className="rounded-full bg-gray-900 hover:bg-brand-primary px-6 py-2 text-[10px] font-black uppercase tracking-widest text-white transition-all transform hover:scale-105"
                          >
                            Approve
                          </button>
                          <button
                            onClick={() => handleAction(row.claim_id, 'rejected', 'fraud')}
                            className="rounded-full border-2 border-gray-100 hover:border-brand-primary hover:text-brand-primary px-6 py-2 text-[10px] font-black uppercase tracking-widest text-gray-500 transition-all"
                          >
                            Reject
                          </button>
                        </div>
                      ) : (
                        <span className={`px-4 py-1 rounded-full text-[10px] font-black uppercase tracking-widest border ${
                          row.fraud_verdict === 'safe' || row.fraud_verdict === 'approve' 
                            ? 'border-emerald-100 bg-emerald-50 text-emerald-600' 
                            : 'border-brand-primary/20 bg-brand-soft text-brand-primary'
                        }`}>
                          {row.fraud_verdict}
                        </span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Panel>
      </div>
    </PageShell>
  )
}
