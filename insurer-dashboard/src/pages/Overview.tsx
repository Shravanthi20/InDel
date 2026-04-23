import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar, Cell } from 'recharts'
import { getMoneyExchange, getOverview, getPoolHealth, getLedger } from '../api/insurer'
import { PageShell, Panel, StatCard, ImpactCard } from './OperationsShared'
import { useLocalization } from '../context/LocalizationContext'

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
  const { t } = useLocalization()
  const [levelFilter, setLevelFilter] = useState('ALL')
  const [zoneFilter, setZoneFilter] = useState('')
  const [refreshTick, setRefreshTick] = useState(0)

  const params = {
    level: levelFilter === 'ALL' ? '' : levelFilter,
    zone: zoneFilter.trim(),
  }

  const { data: overview, error: overviewErr, refetch: refetchOverview } = useQuery<OverviewData>({ queryKey: ['overview'], queryFn: getOverview })
  const { data: pool, error: poolErr, refetch: refetchPool } = useQuery<PoolHealth>({ queryKey: ['poolHealth'], queryFn: getPoolHealth })
  const { data: moneyExchange, error: mxErr, refetch: refetchMx } = useQuery<MoneyExchangeSummary>({ queryKey: ['moneyExchange', params], queryFn: () => getMoneyExchange(params) })
  const { data: ledgerData, refetch: refetchLedger } = useQuery({ 
    queryKey: ['ledger', refreshTick], 
    queryFn: () => getLedger({ limit: 10 }) 
  })

  const handleRefresh = () => {
    refetchOverview()
    refetchPool()
    refetchMx()
    refetchLedger()
    setRefreshTick((v) => v + 1)
  }

  const errorObj = overviewErr || poolErr || mxErr
  const error = errorObj ? (errorObj as Error).message : null

  const claimsDistribution = [
    { name: t('pages.overview.pending'), value: overview?.pending_claims ?? 0, color: '#F472B6' },
    { name: t('pages.overview.approved'), value: overview?.approved_claims ?? 0, color: '#10b981' },
    { name: t('pages.overview.flagged'), value: 12, color: '#BE185D' },
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
      eyebrow={t('pages.overview.eyebrow')}
      title={t('pages.overview.title')}
      description={t('pages.overview.description')}
    >
      {error ? <div className="mb-8 p-4 rounded bg-rose-50 text-rose-700 dark:bg-rose-950 dark:text-rose-400 border border-rose-200 dark:border-rose-900 font-bold uppercase text-[10px] tracking-widest">{error}</div> : null}
      <Panel title={t('pages.overview.dynamicControls')} subtitle={t('pages.overview.controlsSubtitle')}>
        <div className="grid gap-4 md:grid-cols-[180px_1fr_auto]">
          <label className="space-y-2">
            <span className="text-[9px] font-black uppercase tracking-[0.3em] text-slate-400">{t('pages.overview.zoneLevel')}</span>
            <select
              value={levelFilter}
              onChange={(e) => setLevelFilter(e.target.value)}
              className="w-full rounded border border-slate-200 bg-white px-3 py-2 text-sm font-bold text-slate-900 outline-none"
            >
              <option value="ALL">{t('pages.overview.allZones')}</option>
              <option value="A">{t('pages.overview.levelA')}</option>
              <option value="B">{t('pages.overview.levelB')}</option>
              <option value="C">{t('pages.overview.levelC')}</option>
            </select>
          </label>
          <label className="space-y-2">
            <span className="text-[9px] font-black uppercase tracking-[0.3em] text-slate-400">{t('pages.overview.zoneSearch')}</span>
            <input
              value={zoneFilter}
              onChange={(e) => setZoneFilter(e.target.value)}
              placeholder={t('pages.overview.searchPlaceholder')}
              className="w-full rounded border border-slate-200 bg-white px-3 py-2 text-sm font-bold text-slate-900 outline-none"
            />
          </label>
          <div className="flex items-end">
            <button
              type="button"
              onClick={handleRefresh}
              className="rounded bg-slate-900 px-4 py-2 text-xs font-black uppercase tracking-widest text-white"
            >
              {t('pages.overview.refresh')}
            </button>
          </div>
        </div>
      </Panel>
      
      <div className="grid gap-8 md:grid-cols-2 xl:grid-cols-4">
        <StatCard label="Economic Inflow (Premiums)" value={`₹ ${Math.round(moneyExchange?.premium_pool ?? 0).toLocaleString()}`} tone="brand" />
        <StatCard label="Active Worker Polices" value={String(moneyExchange?.total_subscribed ?? Math.round(overview?.active_workers ?? 0))} />
        <StatCard label="Disruptions Verified" value={String(moneyExchange?.total_claims ?? 0)} tone="warm" />
        <StatCard
          label="Adjusted Reserve Delta"
          value={`₹ ${Math.round((moneyExchange?.total_payouts ?? 0) - (moneyExchange?.premium_pool ?? 0)).toLocaleString()}`}
          tone={(moneyExchange?.net_pool ?? 0) < 0 ? 'alert' : 'default'}
        />
      </div>

      <div className="grid gap-8 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <ImpactCard 
            loss={moneyExchange?.total_claim_amount ?? 0} 
            payout={moneyExchange?.total_payouts ?? 0} 
          />
        </div>
        <Panel title="Pool Health Meter" className="flex flex-col items-center justify-center text-center">
          <div className="relative h-48 w-48 flex items-center justify-center">
            <svg className="absolute inset-0 w-full h-full -rotate-90">
              <circle cx="96" cy="96" r="80" stroke="currentColor" strokeWidth="12" fill="transparent" className="text-gray-100 dark:text-gray-800" />
              <circle 
                cx="96" cy="96" r="80" stroke="currentColor" strokeWidth="12" fill="transparent" 
                strokeDasharray={`${Math.PI * 160}`}
                strokeDashoffset={`${Math.PI * 160 * (1 - (overview?.reserve_utilization ?? 0))}`}
                className="text-brand-primary"
              />
            </svg>
            <div className="text-center">
              <p className="text-4xl font-black text-gray-900 dark:text-white leading-none">
                {Math.round((overview?.reserve_utilization ?? 0) * 100)}%
              </p>
              <p className="text-[10px] font-black uppercase tracking-widest text-gray-400 mt-2">Utilization</p>
            </div>
          </div>
        </Panel>
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
                  cursor={{ stroke: '#EC4899', strokeWidth: 1 }}
                  contentStyle={{ borderRadius: '12px', border: '1px solid #FDF2F8', backgroundColor: '#fff', fontSize: '12px', fontWeight: 'bold' }}
                />
                <Area type="monotone" dataKey={zoneChartData.length > 0 ? 'netFlow' : 'exposure'} stroke="#EC4899" strokeWidth={3} fillOpacity={0.15} fill="#EC4899" />
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
      <Panel 
        title="Audit-Grade Financial Ledger" 
        subtitle="Immutable ledger of economic inflows (premiums) and verified disbursements (payouts)."
      >
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead>
              <tr className="border-b border-gray-100 dark:border-gray-800">
                <th className="pb-4 font-black uppercase tracking-[0.2em] text-gray-400">Timestamp</th>
                <th className="pb-4 font-black uppercase tracking-[0.2em] text-gray-400">Event</th>
                <th className="pb-4 font-black uppercase tracking-[0.2em] text-gray-400">Zone</th>
                <th className="pb-4 font-black uppercase tracking-[0.2em] text-gray-400">Amount</th>
                <th className="pb-4 font-black uppercase tracking-[0.2em] text-gray-400">Status</th>
                <th className="pb-4 font-black uppercase tracking-[0.2em] text-gray-400 text-right">Reference</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50 dark:divide-gray-800/50">
              {(ledgerData?.data ?? []).map((item) => (
                <tr 
                  key={`${item.event_type}-${item.reference_id}`} 
                  className={`hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-all ${
                    item.event_type === 'payout' ? 'bg-pink-50/20' : ''
                  }`}
                >
                  <td className="py-4 text-gray-500 font-medium">
                    {new Date(item.timestamp).toLocaleString(undefined, { 
                      month: 'short', day: '2-digit', hour: '2-digit', minute: '2-digit' 
                    })}
                  </td>
                  <td className="py-4">
                    <span className={`px-2 py-1 rounded-full text-[9px] font-black uppercase tracking-widest ${
                      item.event_type === 'premium' 
                        ? 'bg-emerald-50 text-emerald-600' 
                        : 'bg-brand-primary text-white shadow-sm shadow-brand-primary/20'
                    }`}>
                      {item.event_type}
                    </span>
                  </td>
                  <td className="py-4 font-bold text-gray-900 dark:text-gray-100">{item.zone}</td>
                  <td className="py-4 font-black text-sm">₹{Math.round(item.amount).toLocaleString()}</td>
                  <td className="py-4">
                    <span className="text-[10px] font-bold text-gray-400 uppercase tracking-tight">{item.status}</span>
                  </td>
                  <td className="py-4 text-right font-mono text-[10px] text-gray-400">#RES-{item.reference_id}</td>
                </tr>
              ))}
              {(ledgerData?.data ?? []).length === 0 && (
                <tr>
                  <td colSpan={6} className="py-12 text-center text-gray-400 italic">No transactions recorded in the audit ledger yet.</td>
                </tr>
              )}
            </tbody>
          </table>
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
