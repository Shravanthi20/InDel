import { useEffect, useState } from 'react'
import { getLossRatio } from '../api/insurer'
import { PageShell, Panel } from './OperationsShared'

type LossRatioRow = {
  city: string
  zone_name: string
  premiums: number
  claims: number
  loss_ratio: number
}

export default function LossRatio() {
  const [rows, setRows] = useState<LossRatioRow[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    getLossRatio()
      .then((data) => setRows(Array.isArray(data) ? data : []))
      .catch((err) => setError(err?.message ?? 'Failed to load loss ratio'))
  }, [])

  return (
    <PageShell
      eyebrow="Portfolio"
      title="Loss Ratio by Zone"
      description="Compare premium inflow and claim outflow by zone to spot pockets of concentrated exposure."
    >
      <Panel title="Zone Performance">
        {error ? <p className="text-sm text-rose-600">{error}</p> : null}
        <div className="overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead className="text-left text-slate-500">
              <tr>
                <th className="pb-3 pr-4">City</th>
                <th className="pb-3 pr-4">Zone</th>
                <th className="pb-3 pr-4">Premiums</th>
                <th className="pb-3 pr-4">Claims</th>
                <th className="pb-3 pr-4">Loss Ratio</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((row) => (
                <tr key={`${row.city}-${row.zone_name}`} className="border-t border-slate-100">
                  <td className="py-3 pr-4">{row.city}</td>
                  <td className="py-3 pr-4 font-medium text-slate-900">{row.zone_name}</td>
                  <td className="py-3 pr-4">Rs {Math.round(row.premiums ?? 0)}</td>
                  <td className="py-3 pr-4">Rs {Math.round(row.claims ?? 0)}</td>
                  <td className="py-3 pr-4">{Math.round((row.loss_ratio ?? 0) * 100)}%</td>
                </tr>
              ))}
              {rows.length === 0 && !error ? (
                <tr>
                  <td className="py-6 text-slate-500" colSpan={5}>No zone data available yet.</td>
                </tr>
              ) : null}
            </tbody>
          </table>
        </div>
      </Panel>
    </PageShell>
  )
}
