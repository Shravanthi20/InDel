import { useState } from 'react'
import { generateClaimsForDisruption, processQueuedPayouts, queuePayoutForClaim } from '../api/operations'
import { PageShell, Panel, ResultBox, StatCard } from './OperationsShared'

export default function PayoutOps() {
  const [disruptionId, setDisruptionId] = useState('')
  const [claimId, setClaimId] = useState('')
  const [claimsResult, setClaimsResult] = useState<any>(null)
  const [queueResult, setQueueResult] = useState<any>(null)
  const [processResult, setProcessResult] = useState<any>(null)
  const [error, setError] = useState('')
  const [loadingAction, setLoadingAction] = useState('')

  async function handleGenerateClaims() {
    setLoadingAction('claims')
    setError('')
    try {
      const response = await generateClaimsForDisruption(disruptionId.trim())
      setClaimsResult(response.data.data)
    } catch (err: any) {
      setError(err?.response?.data?.error?.message || 'Claim generation failed.')
    } finally {
      setLoadingAction('')
    }
  }

  async function handleQueuePayout() {
    setLoadingAction('queue')
    setError('')
    try {
      const response = await queuePayoutForClaim(claimId.trim())
      setQueueResult(response.data.data)
    } catch (err: any) {
      setError(err?.response?.data?.error?.message || 'Payout queueing failed.')
    } finally {
      setLoadingAction('')
    }
  }

  async function handleProcessPayouts() {
    setLoadingAction('process')
    setError('')
    try {
      const response = await processQueuedPayouts()
      setProcessResult(response.data.data)
    } catch (err: any) {
      setError(err?.response?.data?.error?.message || 'Payout processing failed.')
    } finally {
      setLoadingAction('')
    }
  }

  return (
    <PageShell
      eyebrow="Operations"
      title="Claims and Payout Orchestration"
      description="Drive the Part 4 claims-to-payout path: generate claims for a confirmed disruption, queue individual payouts, and process the payout backlog with retry visibility."
    >
      <div className="grid gap-6 xl:grid-cols-2">
        <Panel title="Generate Claims for Disruption" subtitle="Use a seeded or real disruption id from the backend. Numeric ids work best, for example `1`.">
          <div className="space-y-4">
            <input
              type="text"
              value={disruptionId}
              onChange={(e) => setDisruptionId(e.target.value)}
              placeholder="disruption id"
              className="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-slate-900 outline-none transition focus:border-orange-400"
            />
            <button
              type="button"
              onClick={handleGenerateClaims}
              disabled={loadingAction === 'claims' || !disruptionId.trim()}
              className="rounded-full bg-slate-950 px-6 py-3 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {loadingAction === 'claims' ? 'Generating...' : 'Generate Claims'}
            </button>
          </div>
        </Panel>

        <Panel title="Queue Payout for Claim" subtitle="Queue a specific claim into the payout pipeline with idempotent behavior on repeat calls.">
          <div className="space-y-4">
            <input
              type="text"
              value={claimId}
              onChange={(e) => setClaimId(e.target.value)}
              placeholder="claim id"
              className="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-slate-900 outline-none transition focus:border-orange-400"
            />
            <button
              type="button"
              onClick={handleQueuePayout}
              disabled={loadingAction === 'queue' || !claimId.trim()}
              className="rounded-full bg-orange-500 px-6 py-3 text-sm font-semibold text-white transition hover:bg-orange-400 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {loadingAction === 'queue' ? 'Queueing...' : 'Queue Payout'}
            </button>
          </div>
        </Panel>
      </div>

      <div className="grid gap-6 xl:grid-cols-[1.2fr_0.8fr]">
        <Panel title="Process Payout Backlog" subtitle="Run the payout worker behavior from the dashboard to surface successes, retries, and transient failures.">
          <div className="flex flex-wrap items-center gap-4">
            <button
              type="button"
              onClick={handleProcessPayouts}
              disabled={loadingAction === 'process'}
              className="rounded-full bg-emerald-600 px-6 py-3 text-sm font-semibold text-white transition hover:bg-emerald-500 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {loadingAction === 'process' ? 'Processing...' : 'Process Queued Payouts'}
            </button>
            {error ? <span className="text-sm font-medium text-rose-600">{error}</span> : null}
          </div>
        </Panel>

        <Panel title="Flow Summary" subtitle="Recommended demo sequence">
          <div className="space-y-3 text-sm leading-6 text-slate-600">
            <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">1. Generate synthetic data.</div>
            <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">2. Trigger claims for a disruption id.</div>
            <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">3. Queue one or more claims for payout.</div>
            <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">4. Process payouts and inspect retries.</div>
          </div>
        </Panel>
      </div>

      <div className="grid gap-6 xl:grid-cols-3">
        <Panel title="Claim Generation Result">
          {claimsResult ? (
            <ResultBox>
              <p><span className="font-semibold text-orange-300">disruption:</span> {claimsResult.disruption_id}</p>
              <p><span className="font-semibold text-orange-300">workers checked:</span> {claimsResult.workers_checked}</p>
              <p><span className="font-semibold text-orange-300">claims generated:</span> {claimsResult.claims_generated}</p>
              <p><span className="font-semibold text-orange-300">claims skipped:</span> {claimsResult.claims_skipped}</p>
            </ResultBox>
          ) : (
            <p className="text-sm text-slate-500">No claim-generation run yet.</p>
          )}
        </Panel>

        <Panel title="Queue Result">
          {queueResult ? (
            <ResultBox>
              <p><span className="font-semibold text-orange-300">payout:</span> {queueResult.payout_id}</p>
              <p><span className="font-semibold text-orange-300">claim:</span> {queueResult.claim_id}</p>
              <p><span className="font-semibold text-orange-300">status:</span> {queueResult.status}</p>
              <p><span className="font-semibold text-orange-300">idempotency:</span> {queueResult.idempotency_key}</p>
            </ResultBox>
          ) : (
            <p className="text-sm text-slate-500">No payout queued yet.</p>
          )}
        </Panel>

        <Panel title="Processing Result">
          {processResult ? (
            <div className="grid gap-4">
              <StatCard label="Processed" value={String(processResult.processed)} />
              <StatCard label="Succeeded" value={String(processResult.succeeded)} tone="warm" />
              <StatCard label="Failed" value={String(processResult.failed)} tone={processResult.failed ? 'alert' : 'default'} />
              <StatCard label="Retried" value={String(processResult.retried)} />
            </div>
          ) : (
            <p className="text-sm text-slate-500">No payout-processing run yet.</p>
          )}
        </Panel>
      </div>
    </PageShell>
  )
}
