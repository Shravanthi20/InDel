import { useEffect, useState } from 'react'
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar, Cell } from 'recharts'
import { getMoneyExchange, getOverview, getPoolHealth } from '../api/insurer'
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

type ZoneMoneyExchange = {
  zone_id: number
  zone_name: string
  city: string
  state: string
  level: string
  subscribed_workers: number
  claims_count: number
  premiums_collected: number
  claims_amount: number
  payouts_processed: number
  net_flow: number
}

type MoneyExchangeSummary = {
  premium_pool: number
  total_subscribed: number
  total_claims: number
  total_claim_amount: number
  total_payouts: number
  net_pool: number
  pending_payouts: number
  zone_breakdown: ZoneMoneyExchange[]
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
  const [moneyExchange, setMoneyExchange] = useState<MoneyExchangeSummary | null>(null)
  const [levelFilter, setLevelFilter] = useState('ALL')
  const [zoneFilter, setZoneFilter] = useState('')
  const [refreshTick, setRefreshTick] = useState(0)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const params = {
      level: levelFilter === 'ALL' ? '' : levelFilter,
      zone: zoneFilter.trim(),
    }

    Promise.all([getOverview(), getPoolHealth(), getMoneyExchange(params)])
      .then(([overviewData, poolHealth, moneyExchangeData]) => {
        setOverview(overviewData)
        setPool(poolHealth)
        setMoneyExchange(moneyExchangeData)
        setError(null)
      })
      .catch((err) => setError(err?.message ?? 'Failed to load overview'))
  }, [levelFilter, zoneFilter, refreshTick])

  const claimsDistribution = [
    { name: 'Pending', value: overview?.pending_claims ?? 0, color: '#f97316' },
    { name: 'Approved', value: overview?.approved_claims ?? 0, color: '#10b981' },
    { name: 'Flagged', value: 12, color: '#f43f5e' },
  ]

  const zoneRows = moneyExchange?.zone_breakdown ?? []
  const zoneChartData = zoneRows
    .map((row) => ({
      name: `${row.zone_name}`,
      netFlow: Math.round(row.net_flow),
    }))
    .slice(0, 12)

  return (
    <PageShell
      eyebrow="Console"
      title="Global Portfolio Operations"
      description="Track real-time worker coverage, enterprise claims pressure, and reserve posture across the ecosystem."
    >
      {error ? <div className="mb-8 p-4 rounded bg-rose-50 text-rose-700 dark:bg-rose-950 dark:text-rose-400 border border-rose-200 dark:border-rose-900 font-bold uppercase text-[10px] tracking-widest">{error}</div> : null}
      <Panel title="Dynamic Controls" subtitle="Slice by zone level/name and refresh after each scenario change.">
        <div className="grid gap-4 md:grid-cols-[180px_1fr_auto]">
          <label className="space-y-2">
            <span className="text-[9px] font-black uppercase tracking-[0.3em] text-slate-400">Zone Level</span>
            <select
              value={levelFilter}
              onChange={(e) => setLevelFilter(e.target.value)}
              className="w-full rounded border border-slate-200 bg-white px-3 py-2 text-sm font-bold text-slate-900 outline-none"
            >
              <option value="ALL">ALL</option>
              <option value="A">A</option>
              <option value="B">B</option>
              <option value="C">C</option>
            </select>
          </label>
          <label className="space-y-2">
            <span className="text-[9px] font-black uppercase tracking-[0.3em] text-slate-400">Zone / City Search</span>
            <input
              value={zoneFilter}
              onChange={(e) => setZoneFilter(e.target.value)}
              placeholder="Filter by zone or city"
              className="w-full rounded border border-slate-200 bg-white px-3 py-2 text-sm font-bold text-slate-900 outline-none"
            />
          </label>
          <div className="flex items-end">
            <button
              type="button"
              onClick={() => setRefreshTick((value) => value + 1)}
              className="rounded bg-slate-900 px-4 py-2 text-xs font-black uppercase tracking-widest text-white"
            >
              Refresh
            </button>
          </div>
        </div>
      </Panel>
      
      <div className="grid gap-8 md:grid-cols-2 xl:grid-cols-4">
        <StatCard label="Premium Pool" value={`Rs ${Math.round(moneyExchange?.premium_pool ?? 0).toLocaleString()}`} />
        <StatCard label="Subscribed On Plan" value={String(moneyExchange?.total_subscribed ?? Math.round(overview?.active_workers ?? 0))} />
        <StatCard label="Claims Happened" value={String(moneyExchange?.total_claims ?? 0)} tone="alert" />
        <StatCard
          label="Overall Money Exchange"
          value={`Rs ${Math.round((moneyExchange?.total_payouts ?? 0) - (moneyExchange?.premium_pool ?? 0)).toLocaleString()}`}
          tone={(moneyExchange?.net_pool ?? 0) < 0 ? 'alert' : 'default'}
        />
      </div>

      <div className="grid gap-8 xl:grid-cols-3">
        <Panel title="Zone Net Flow" subtitle="Net premium minus payout by zone for the selected filters." className="xl:col-span-2">
          <div className="h-[300px] w-full">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={zoneChartData.length > 0 ? zoneChartData : portfolioTrend}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#e2e8f0" dark-stroke="#1e293b" />
                <XAxis dataKey="name" axisLine={false} tickLine={false} tick={{ fontSize: 10, fill: '#64748b' }} dy={10} />
                <YAxis axisLine={false} tickLine={false} tick={{ fontSize: 10, fill: '#64748b' }} />
                <Tooltip 
                  cursor={{ stroke: '#f97316', strokeWidth: 1 }}
                  contentStyle={{ borderRadius: '4px', border: '1px solid #e2e8f0', backgroundColor: '#fff', fontSize: '12px' }}
                />
                <Area type="monotone" dataKey={zoneChartData.length > 0 ? 'netFlow' : 'exposure'} stroke="#f97316" strokeWidth={2} fillOpacity={0.1} fill="#f97316" />
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
            
            <p className="text-[10px] leading-relaxed text-slate-400 dark:text-slate-500 border-t border-slate-100 dark:border-slate-800 pt-4">
              Premiums: Rs {Math.round(moneyExchange?.premium_pool ?? 0).toLocaleString()} | Claims: Rs {Math.round(moneyExchange?.total_claim_amount ?? 0).toLocaleString()} | Payouts: Rs {Math.round(moneyExchange?.total_payouts ?? 0).toLocaleString()}.
            </p>
          </div>
        </Panel>
      </div>

      <Panel title="Zone Money Exchange" subtitle="Check each zone after every change in disruption/claims/payouts.">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead>
              <tr className="border-b border-slate-200 dark:border-slate-800">
                <th className="pb-3 font-black uppercase tracking-widest text-slate-400">Zone</th>
                <th className="pb-3 font-black uppercase tracking-widest text-slate-400">Lvl</th>
                <th className="pb-3 font-black uppercase tracking-widest text-slate-400">Subscribed</th>
                <th className="pb-3 font-black uppercase tracking-widest text-slate-400">Claims</th>
                <th className="pb-3 font-black uppercase tracking-widest text-slate-400">Premiums</th>
                <th className="pb-3 font-black uppercase tracking-widest text-slate-400">Payouts</th>
                <th className="pb-3 font-black uppercase tracking-widest text-slate-400 text-right">Net</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-slate-800/50">
              {zoneRows.map((row) => (
                <tr key={row.zone_id} className="hover:bg-slate-50 dark:hover:bg-slate-800 transition-none">
                  <td className="py-3">
                    <p className="font-bold text-slate-900 dark:text-white">{row.zone_name}</p>
                    <p className="text-[9px] font-black uppercase tracking-widest text-slate-400">{row.city}, {row.state}</p>
                  </td>
                  <td className="py-3 text-slate-600">{row.level || '-'}</td>
                  <td className="py-3 text-slate-600">{row.subscribed_workers}</td>
                  <td className="py-3 text-slate-600">{row.claims_count}</td>
                  <td className="py-3 text-slate-600">Rs {Math.round(row.premiums_collected).toLocaleString()}</td>
                  <td className="py-3 text-slate-600">Rs {Math.round(row.payouts_processed).toLocaleString()}</td>
                  <td className="py-3 text-right">
                    <span className={`px-2 py-0.5 rounded text-[10px] font-black uppercase tracking-widest ${
                      row.net_flow < 0 ? 'bg-rose-500/10 text-rose-600' : 'bg-emerald-500/10 text-emerald-600'
                    }`}>
                      Rs {Math.round(row.net_flow).toLocaleString()}
                    </span>
                  </td>
                </tr>
              ))}
              {zoneRows.length === 0 ? (
                <tr>
                  <td className="py-10 text-center text-slate-400 italic" colSpan={7}>
                    No zone rows found for the selected filters.
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
