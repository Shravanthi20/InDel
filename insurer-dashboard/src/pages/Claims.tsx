import { useEffect, useMemo, useState } from 'react'
import { getClaimDetail, getClaims, reviewClaim } from '../api/insurer'

type ClaimListItem = {
  claim_id: number
  disruption_id?: number
  worker_id?: number
  zone_name: string
  claim_amount: number
  status: string
  fraud_verdict: string
  created_at: string
}

type ClaimDetail = {
  claim_id: string
  worker_id: string
  zone_id: string
  disruption_id: string
  loss_amount: number
  recommended_payout: number
  status: string
  fraud_verdict: string
  fraud_score: number
  factors: Array<{ name: string; impact: number }>
  created_at: string
}

export default function Claims() {
  const [claims, setClaims] = useState<ClaimListItem[]>([])
  const [selectedClaimId, setSelectedClaimId] = useState<string>('')
  const [selectedClaim, setSelectedClaim] = useState<ClaimDetail | null>(null)
  const [pagination, setPagination] = useState({ page: 1, limit: 10, total: 0, hasNext: false })
  const [status, setStatus] = useState('')
  const [fraudVerdict, setFraudVerdict] = useState('')
  const [loading, setLoading] = useState(false)
  const [detailLoading, setDetailLoading] = useState(false)
  const [error, setError] = useState('')
  const [reviewStatus, setReviewStatus] = useState('approved')
  const [reviewVerdict, setReviewVerdict] = useState('clear')
  const [reviewNotes, setReviewNotes] = useState('')
  const [reviewMessage, setReviewMessage] = useState('')

  useEffect(() => {
    setLoading(true)
    setError('')
    getClaims({ page: pagination.page, limit: pagination.limit, status: status || undefined, fraud_verdict: fraudVerdict || undefined })
      .then((res) => {
        const body = res.data
        setClaims(body.data ?? [])
        setPagination((prev) => ({
          ...prev,
          page: body.pagination?.page ?? prev.page,
          limit: body.pagination?.limit ?? prev.limit,
          total: body.pagination?.total ?? 0,
          hasNext: body.pagination?.has_next ?? false,
        }))
        const nextClaimId = body.data?.[0]?.claim_id ? String(body.data[0].claim_id) : ''
        if (nextClaimId && !selectedClaimId) {
          setSelectedClaimId(nextClaimId)
        }
      })
      .catch((err) => setError(err?.response?.data?.error?.message || 'Failed to load claims.'))
      .finally(() => setLoading(false))
  }, [pagination.page, pagination.limit, status, fraudVerdict])

  useEffect(() => {
    if (!selectedClaimId) {
      setSelectedClaim(null)
      return
    }
    setDetailLoading(true)
    getClaimDetail(selectedClaimId)
      .then((res) => setSelectedClaim(res.data?.data ?? res.data ?? null))
      .catch((err) => setError(err?.response?.data?.error?.message || 'Failed to load claim detail.'))
      .finally(() => setDetailLoading(false))
  }, [selectedClaimId])

  const summary = useMemo(() => {
    return {
      pending: claims.filter((claim) => claim.status === 'pending').length,
      review: claims.filter((claim) => claim.status === 'manual_review').length,
      approved: claims.filter((claim) => claim.status === 'approved').length,
    }
  }, [claims])

  async function submitReview() {
    if (!selectedClaimId) return
    setReviewMessage('')
    try {
      await reviewClaim(selectedClaimId, {
        status: reviewStatus,
        fraud_verdict: reviewVerdict,
        notes: reviewNotes,
      })
      setReviewMessage('Claim review submitted.')
      setReviewNotes('')
      const refreshed = await getClaimDetail(selectedClaimId)
      setSelectedClaim(refreshed.data?.data ?? refreshed.data ?? null)
      const claimsRes = await getClaims({ page: pagination.page, limit: pagination.limit, status: status || undefined, fraud_verdict: fraudVerdict || undefined })
      setClaims(claimsRes.data?.data ?? [])
    } catch (err: any) {
      setReviewMessage(err?.response?.data?.error?.message || 'Claim review failed.')
    }
  }

  return (
    <div className="space-y-6 p-6">
      <div className="rounded-2xl bg-white p-6 shadow">
        <h1 className="text-2xl font-bold">Claims Pipeline</h1>
        <p className="mt-2 text-sm text-slate-500">Filter, inspect, and review insurer claims without leaving the pipeline view.</p>
        <div className="mt-4 grid gap-3 md:grid-cols-3">
          <Stat label="Pending" value={String(summary.pending)} />
          <Stat label="Manual review" value={String(summary.review)} />
          <Stat label="Approved" value={String(summary.approved)} />
        </div>
      </div>

      {error ? <div className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-700">{error}</div> : null}

      <div className="grid gap-6 xl:grid-cols-[1.2fr_0.8fr]">
        <div className="rounded-2xl bg-white p-6 shadow">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
            <div>
              <h2 className="text-lg font-semibold">Claims</h2>
              <p className="text-sm text-slate-500">Page {pagination.page} of {Math.max(1, Math.ceil(pagination.total / pagination.limit))}</p>
            </div>
            <div className="flex flex-wrap gap-3">
              <label className="space-y-1 text-sm">
                <span className="font-medium text-slate-700">Status</span>
                <select value={status} onChange={(e) => { setPagination((p) => ({ ...p, page: 1 })); setStatus(e.target.value) }} className="rounded-xl border border-slate-300 bg-white px-3 py-2">
                  <option value="">All</option>
                  <option value="pending">Pending</option>
                  <option value="manual_review">Manual review</option>
                  <option value="approved">Approved</option>
                  <option value="processed">Processed</option>
                  <option value="paid">Paid</option>
                </select>
              </label>
              <label className="space-y-1 text-sm">
                <span className="font-medium text-slate-700">Fraud verdict</span>
                <select value={fraudVerdict} onChange={(e) => { setPagination((p) => ({ ...p, page: 1 })); setFraudVerdict(e.target.value) }} className="rounded-xl border border-slate-300 bg-white px-3 py-2">
                  <option value="">All</option>
                  <option value="pending">Pending</option>
                  <option value="clear">Clear</option>
                  <option value="review">Review</option>
                  <option value="flagged">Flagged</option>
                  <option value="manual_review">Manual review</option>
                </select>
              </label>
            </div>
          </div>

          <div className="mt-5 overflow-hidden rounded-2xl border border-slate-200">
            <table className="min-w-full divide-y divide-slate-200 text-sm">
              <thead className="bg-slate-50 text-left text-xs uppercase tracking-[0.2em] text-slate-500">
                <tr>
                  <th className="px-4 py-3">Claim</th>
                  <th className="px-4 py-3">Zone</th>
                  <th className="px-4 py-3">Amount</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3">Verdict</th>
                  <th className="px-4 py-3">Created</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-200 bg-white">
                {loading ? (
                  <tr><td className="px-4 py-6 text-slate-500" colSpan={6}>Loading claims...</td></tr>
                ) : claims.length === 0 ? (
                  <tr><td className="px-4 py-6 text-slate-500" colSpan={6}>No claims found.</td></tr>
                ) : (
                  claims.map((claim) => (
                    <tr key={claim.claim_id} className={selectedClaimId === String(claim.claim_id) ? 'bg-orange-50' : 'hover:bg-slate-50'} onClick={() => setSelectedClaimId(String(claim.claim_id))}>
                      <td className="px-4 py-3 font-medium text-slate-900">#{claim.claim_id}</td>
                      <td className="px-4 py-3 text-slate-600">{claim.zone_name}</td>
                      <td className="px-4 py-3 text-slate-900">₹{Number(claim.claim_amount).toLocaleString('en-IN')}</td>
                      <td className="px-4 py-3"><Badge value={claim.status} /></td>
                      <td className="px-4 py-3"><Badge value={claim.fraud_verdict} tone="soft" /></td>
                      <td className="px-4 py-3 text-slate-600">{String(claim.created_at).slice(0, 19).replace('T', ' ')}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          <div className="mt-4 flex items-center justify-between">
            <button type="button" disabled={pagination.page <= 1} onClick={() => setPagination((p) => ({ ...p, page: p.page - 1 }))} className="rounded-full border border-slate-300 px-4 py-2 text-sm disabled:opacity-50">Previous</button>
            <button type="button" disabled={!pagination.hasNext} onClick={() => setPagination((p) => ({ ...p, page: p.page + 1 }))} className="rounded-full border border-slate-300 px-4 py-2 text-sm disabled:opacity-50">Next</button>
          </div>
        </div>

        <div className="space-y-6">
          <div className="rounded-2xl bg-white p-6 shadow">
            <h2 className="text-lg font-semibold">Claim detail</h2>
            {detailLoading ? (
              <p className="mt-4 text-sm text-slate-500">Loading detail...</p>
            ) : selectedClaim ? (
              <div className="mt-4 space-y-4 text-sm text-slate-700">
                <div className="rounded-xl bg-slate-50 p-4">
                  <p className="text-xs uppercase tracking-[0.2em] text-slate-500">Claim</p>
                  <p className="mt-1 font-semibold text-slate-950">{selectedClaim.claim_id}</p>
                  <p className="mt-2">Worker: {selectedClaim.worker_id}</p>
                  <p>Zone: {selectedClaim.zone_id}</p>
                  <p>Disruption: {selectedClaim.disruption_id}</p>
                </div>
                <div className="grid gap-3 md:grid-cols-2">
                  <Stat label="Loss amount" value={`₹${Number(selectedClaim.loss_amount).toLocaleString('en-IN')}`} />
                  <Stat label="Recommended payout" value={`₹${Number(selectedClaim.recommended_payout).toLocaleString('en-IN')}`} />
                  <Stat label="Status" value={selectedClaim.status} />
                  <Stat label="Fraud verdict" value={selectedClaim.fraud_verdict} />
                </div>
                <div>
                  <p className="text-xs uppercase tracking-[0.2em] text-slate-500">Fraud factors</p>
                  <div className="mt-2 flex flex-wrap gap-2">
                    {selectedClaim.factors?.length ? selectedClaim.factors.map((factor) => (
                      <span key={factor.name} className="rounded-full bg-slate-100 px-3 py-1 text-xs font-medium text-slate-700">
                        {factor.name}: {Math.round(factor.impact * 100)}%
                      </span>
                    )) : <span className="text-sm text-slate-500">No factor breakdown available.</span>}
                  </div>
                </div>
                <div className="rounded-xl border border-orange-200 bg-orange-50 p-4">
                  <p className="text-xs uppercase tracking-[0.2em] text-orange-700">Review claim</p>
                  <div className="mt-3 grid gap-3">
                    <div className="grid gap-3 md:grid-cols-2">
                      <label className="space-y-1 text-sm">
                        <span className="font-medium text-slate-700">Status</span>
                        <select value={reviewStatus} onChange={(e) => setReviewStatus(e.target.value)} className="w-full rounded-xl border border-slate-300 bg-white px-3 py-2">
                          <option value="approved">Approved</option>
                          <option value="denied">Denied</option>
                          <option value="manual_review">Manual review</option>
                        </select>
                      </label>
                      <label className="space-y-1 text-sm">
                        <span className="font-medium text-slate-700">Fraud verdict</span>
                        <select value={reviewVerdict} onChange={(e) => setReviewVerdict(e.target.value)} className="w-full rounded-xl border border-slate-300 bg-white px-3 py-2">
                          <option value="clear">Clear</option>
                          <option value="review">Review</option>
                          <option value="flagged">Flagged</option>
                        </select>
                      </label>
                    </div>
                    <textarea value={reviewNotes} onChange={(e) => setReviewNotes(e.target.value)} placeholder="Review notes" className="min-h-28 rounded-xl border border-slate-300 bg-white px-3 py-2 outline-none focus:border-orange-400" />
                    <button type="button" onClick={submitReview} className="rounded-full bg-slate-950 px-4 py-2 text-sm font-semibold text-white hover:bg-slate-800">Submit review</button>
                    {reviewMessage ? <p className="text-sm text-slate-600">{reviewMessage}</p> : null}
                  </div>
                </div>
              </div>
            ) : (
              <p className="mt-4 text-sm text-slate-500">Select a claim to inspect details.</p>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl bg-slate-50 p-4">
      <p className="text-xs uppercase tracking-[0.2em] text-slate-500">{label}</p>
      <p className="mt-2 text-lg font-bold text-slate-950">{value}</p>
    </div>
  )
}

function Badge({ value, tone = 'default' }: { value: string; tone?: 'default' | 'soft' }) {
  const isSoft = tone === 'soft'
  return (
    <span className={[
      'rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-[0.2em]',
      isSoft ? 'bg-slate-100 text-slate-700' : 'bg-orange-100 text-orange-700',
    ].join(' ')}>
      {value.replace('_', ' ')}
    </span>
  )
}
