import { useEffect, useState } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Cell } from 'recharts'
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
      .then((payload) => setRows(Array.isArray(payload.data) ? payload.data : []))
      .catch((err) => setError(err?.message ?? 'Failed to load loss ratio'))
  }, [])

  const chartData = rows.map(r => ({
    name: r.zone_name,
    ratio: Math.round(r.loss_ratio * 100),
    premiums: r.premiums,
    claims: r.claims
  })).sort((a, b) => b.ratio - a.ratio).slice(0, 8)

  return (
    <PageShell
      eyebrow="Analysis"
      title="Loss Ratio Distribution"
      description="Deep dive into zone performance and risk concentration across the active insurer book."
    >
      <div className="grid gap-8 xl:grid-cols-[1fr_0.4fr]">
        <Panel title="Zone Metrics" subtitle="Variance across active operational zones.">
          <div className="h-[350px] w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#e2e8f0" dark-stroke="#1e293b" />
                <XAxis dataKey="name" axisLine={false} tickLine={false} tick={{ fontSize: 10, fill: '#64748b' }} dy={10} />
                <YAxis axisLine={false} tickLine={false} tick={{ fontSize: 10, fill: '#64748b' }} />
                <Tooltip 
                  cursor={{ fill: 'rgba(0,0,0,0.05)' }}
                  contentStyle={{ borderRadius: '4px', border: '1px solid #e2e8f0', fontSize: '11px' }}
                />
                <Bar dataKey="ratio" radius={[2, 2, 0, 0]} barSize={32}>
                  {chartData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.ratio > 80 ? '#ef4444' : '#f59e0b'} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </div>
        </Panel>

        <Panel title="Summary Insights" subtitle="Critical risk focuses.">
          <div className="space-y-4">
             <div className="p-4 rounded border border-orange-200 bg-orange-50 text-xs text-orange-800 dark:border-orange-950 dark:bg-orange-900/10 dark:text-orange-400">
                <p className="font-bold mb-1 uppercase tracking-widest text-[9px]">Exposure Alert</p>
                <p className="leading-relaxed">High variance detected in industrial zones. Suggested adjustment for ratios &gt; 80%.</p>
             </div>
             <div className="p-4 rounded border border-emerald-200 bg-emerald-50 text-xs text-emerald-800 dark:border-emerald-950 dark:bg-emerald-900/10 dark:text-emerald-400">
                <p className="font-bold mb-1 uppercase tracking-widest text-[9px]">Growth Opportunity</p>
                <p className="leading-relaxed">Zone scaling successful where loss ratio remains below 15% threshold.</p>
             </div>
          </div>
        </Panel>
      </div>

      <Panel title="Data Grid" className="mt-8">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead>
              <tr className="border-b border-slate-200 dark:border-slate-800">
                <th className="pb-4 font-black uppercase tracking-widest text-slate-400">Zone</th>
                <th className="pb-4 font-black uppercase tracking-widest text-slate-400">Premiums</th>
                <th className="pb-4 font-black uppercase tracking-widest text-slate-400">Claims</th>
                <th className="pb-4 font-black uppercase tracking-widest text-slate-400 text-right">Ratio</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-slate-800/50">
              {rows.map((row) => (
                <tr key={`${row.city}-${row.zone_name}`} className="hover:bg-slate-50 dark:hover:bg-slate-800 transition-none">
                  <td className="py-4">
                    <p className="font-bold text-slate-900 dark:text-white">{row.zone_name}</p>
                    <p className="text-[9px] font-black uppercase tracking-widest text-slate-400">{row.city}</p>
                  </td>
                  <td className="py-4 text-slate-500">Rs {Math.round(row.premiums).toLocaleString()}</td>
                  <td className="py-4 text-slate-500">Rs {Math.round(row.claims).toLocaleString()}</td>
                  <td className="py-4 text-right">
                    <span className={`px-2 py-0.5 rounded text-[10px] font-black uppercase tracking-widest ${
                      row.loss_ratio > 0.8 ? 'bg-rose-500/10 text-rose-600' : 'bg-orange-500/10 text-orange-600'
                    }`}>
                      {Math.round(row.loss_ratio * 100)}%
                    </span>
                  </td>
                </tr>
              ))}
              {rows.length === 0 && !error ? (
                <tr>
                  <td className="py-12 text-center text-slate-400 italic" colSpan={4}>
                     No zone metrics currently streaming.
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
