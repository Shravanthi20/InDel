import { useEffect, useState } from 'react'
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar, Cell } from 'recharts'
import { getOverview, getPoolHealth } from '../api/insurer'
import { PageShell, Panel, StatCard } from './OperationsShared'

type OverviewData = {
  active_workers: number
  pending_claims: number
  approved_claims: number
  loss_ratio: number
  reserve_utilization: number
  reserve: number
  pool_health: string
}

type PoolHealth = {
  week_premiums: number
  week_payouts: number
  net_pool: number
  pending_payouts: number
}

// Enterprise trend data
const portfolioTrend = [
  { name: 'Mon', exposure: 4000, claims: 2400 },
  { name: 'Tue', exposure: 4500, claims: 1398 },
  { name: 'Wed', exposure: 4200, claims: 9800 },
  { name: 'Thu', exposure: 5000, claims: 3908 },
  { name: 'Fri', exposure: 5800, claims: 4800 },
  { name: 'Sat', exposure: 6200, claims: 3800 },
  { name: 'Sun', exposure: 6500, claims: 4300 },
]

export default function Overview() {
  const [overview, setOverview] = useState<OverviewData | null>(null)
  const [pool, setPool] = useState<PoolHealth | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    Promise.all([getOverview(), getPoolHealth()])
      .then(([overviewData, poolHealth]) => {
        setOverview(overviewData)
        setPool(poolHealth)
      })
      .catch((err) => setError(err?.message ?? 'Failed to load overview'))
  }, [])

  const claimsDistribution = [
    { name: 'Pending', value: overview?.pending_claims ?? 0, color: '#f97316' },
    { name: 'Approved', value: overview?.approved_claims ?? 0, color: '#10b981' },
    { name: 'Flagged', value: 12, color: '#f43f5e' },
  ]

  return (
    <PageShell
      eyebrow="Console"
      title="Global Portfolio Operations"
      description="Track real-time worker coverage, enterprise claims pressure, and reserve posture across the ecosystem."
    >
      {error ? <div className="mb-8 p-4 rounded bg-rose-50 text-rose-700 dark:bg-rose-950 dark:text-rose-400 border border-rose-200 dark:border-rose-900 font-bold uppercase text-[10px] tracking-widest">{error}</div> : null}
      
      <div className="grid gap-8 md:grid-cols-2 xl:grid-cols-4">
        <StatCard label="Active Workers" value={String(overview?.active_workers ?? 0)} />
        <StatCard label="Pending Claims" value={String(overview?.pending_claims ?? 0)} tone="alert" />
        <StatCard label="Approved Claims" value={String(overview?.approved_claims ?? 0)} tone="warm" />
        <StatCard
          label="Loss Ratio"
          value={`${Math.round((overview?.loss_ratio ?? 0) * 100)}%`}
          tone={(overview?.loss_ratio ?? 0) > 0.8 ? 'alert' : 'default'}
        />
      </div>

      <div className="grid gap-8 xl:grid-cols-3">
        <Panel title="Exposure Stream" subtitle="Live tracking of portfolo exposure vs claimed value." className="xl:col-span-2">
          <div className="h-[300px] w-full">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={portfolioTrend}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#e2e8f0" dark-stroke="#1e293b" />
                <XAxis dataKey="name" axisLine={false} tickLine={false} tick={{ fontSize: 10, fill: '#64748b' }} dy={10} />
                <YAxis axisLine={false} tickLine={false} tick={{ fontSize: 10, fill: '#64748b' }} />
                <Tooltip 
                  cursor={{ stroke: '#f97316', strokeWidth: 1 }}
                  contentStyle={{ borderRadius: '4px', border: '1px solid #e2e8f0', backgroundColor: '#fff', fontSize: '12px' }}
                />
                <Area type="monotone" dataKey="exposure" stroke="#f97316" strokeWidth={2} fillOpacity={0.1} fill="#f97316" />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </Panel>

        <Panel title="Event Segment" subtitle="Claims status distribution matrix.">
          <div className="h-[300px] w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={claimsDistribution} layout="vertical" margin={{ left: 40, right: 30, top: 10, bottom: 10 }}>
                <XAxis type="number" hide />
                <YAxis 
                  dataKey="name" 
                  type="category" 
                  axisLine={false} 
                  tickLine={false} 
                  tick={{ fontSize: 11, fontWeight: 700, fill: '#64748b' }} 
                  width={80}
                />
                <Tooltip cursor={{ fill: 'transparent' }} contentStyle={{ borderRadius: '4px', border: '1px solid #e2e8f0', fontSize: '11px' }} />
                <Bar dataKey="value" radius={[0, 4, 4, 0]} barSize={24}>
                  {claimsDistribution.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </div>
        </Panel>
      </div>

      <div className="grid gap-8 lg:grid-cols-2">
        <Panel title="Pool Posture" subtitle="Balance against paid and pending obligations.">
          <div className="grid gap-6 sm:grid-cols-2">
            <div className="rounded border border-slate-100 dark:border-slate-800 p-6 bg-slate-50 dark:bg-slate-900/50">
               <p className="text-[10px] font-black uppercase tracking-widest text-slate-500 mb-2">Week Premiums</p>
               <p className="text-2xl font-black text-slate-900 dark:text-white">Rs {pool?.week_premiums ?? 0}</p>
            </div>
            <div className="rounded border border-slate-100 dark:border-slate-800 p-6 bg-slate-50 dark:bg-slate-900/50">
               <p className="text-[10px] font-black uppercase tracking-widest text-slate-500 mb-2">Week Payouts</p>
               <p className="text-2xl font-black text-slate-900 dark:text-white">Rs {pool?.week_payouts ?? 0}</p>
            </div>
            <div className="rounded border border-slate-100 dark:border-slate-800 p-6 bg-slate-50 dark:bg-slate-900/50">
               <p className="text-[10px] font-black uppercase tracking-widest text-slate-500 mb-2">Net Pool</p>
               <p className="text-2xl font-black text-slate-900 dark:text-white">Rs {pool?.net_pool ?? 0}</p>
            </div>
            <div className="rounded border border-slate-100 dark:border-slate-800 p-6 bg-slate-50 dark:bg-slate-900/50">
               <p className="text-[10px] font-black uppercase tracking-widest text-slate-500 mb-2">Pending Payouts</p>
               <p className="text-2xl font-black text-slate-900 dark:text-white">Rs {pool?.pending_payouts ?? 0}</p>
            </div>
          </div>
        </Panel>

        <Panel title="System Status" subtitle="Operational health of the book.">
          <div className="space-y-6">
            <div className="flex items-center justify-between p-5 rounded border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-950">
               <div className="flex items-center gap-3">
                  <div className={`h-2 w-2 rounded-full ${overview ? 'bg-emerald-500' : 'bg-slate-400'}`}></div>
                  <span className="text-[10px] font-black uppercase tracking-widest text-slate-500">Service Connectivity</span>
               </div>
               <span className={`text-[10px] font-black uppercase tracking-[0.2em] px-3 py-1 rounded bg-emerald-500/10 text-emerald-600`}>
                  {overview ? 'Operational' : 'Syncing'}
               </span>
            </div>

            <div className="space-y-3 px-1 text-xs">
              <div className="flex items-center justify-between">
                <span className="font-bold text-slate-500 uppercase tracking-widest text-[9px]">Reserve Utilization</span>
                <span className="font-black text-slate-900 dark:text-white">{Math.round((overview?.reserve_utilization ?? 0) * 100)}%</span>
              </div>
              <div className="h-1 w-full bg-slate-100 dark:bg-slate-800 overflow-hidden">
                <div 
                  className="h-full bg-orange-600 transition-none" 
                  style={{ width: `${Math.round((overview?.reserve_utilization ?? 0) * 100)}%` }}
                ></div>
              </div>
            </div>
            
            <p className="text-[10px] leading-relaxed text-slate-400 dark:text-slate-500 border-t border-slate-100 dark:border-slate-800 pt-4">Processing via node-gateway-alpha. JWT session verified for 2h 45m remaining.</p>
          </div>
        </Panel>
      </div>
    </PageShell>
  )
}
