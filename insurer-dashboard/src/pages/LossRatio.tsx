import { useEffect, useMemo, useState } from 'react'
import { 
  BarChart, 
  Bar, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer, 
  Cell, 
  ReferenceLine,
  LabelList
} from 'recharts'
import { getLossRatio, type LossRatioRow } from '../api/insurer'
import { PageShell, Panel } from './OperationsShared'
import { useLocalization } from '../context/LocalizationContext'
import { AlertCircle, TrendingDown, TrendingUp, Info } from 'lucide-react'

export default function LossRatio() {
  const { t } = useLocalization()
  const [rows, setRows] = useState<LossRatioRow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setLoading(true)
    getLossRatio()
      .then((payload) => setRows(Array.isArray(payload) ? payload : []))
      .catch((err) => setError(err?.message ?? 'Failed to load loss ratio'))
      .finally(() => setLoading(false))
  }, [])

  const TARGET_RATIO = 50
  const VISUAL_CEILING = 150
  const MAX_DIVERGENCE = VISUAL_CEILING - TARGET_RATIO

  const chartData = useMemo(() => {
    return rows
      .map((row) => {
        const ratio = Math.round(row.loss_ratio * 100)
        const divergence = ratio - TARGET_RATIO
        return {
          name: row.zone_name,
          city: row.city,
          ratio: ratio,
          divergence: divergence,
          // Cap at 100% above target for visualization clarity
          displayDivergence: Math.min(Math.max(divergence, -50), MAX_DIVERGENCE),
          isCapped: divergence > MAX_DIVERGENCE,
          premiums: row.premiums,
          claims: row.claims,
        }
      })
      .sort((a, b) => b.divergence - a.divergence)
      .slice(0, 10)
  }, [rows])

  const topInsight = useMemo(() => {
    if (rows.length === 0) return null
    const worstZone = [...rows].sort((a, b) => b.claims - a.claims)[0]
    const totalClaims = rows.reduce((sum, r) => sum + r.claims, 0)
    const contribution = totalClaims > 0 ? Math.round((worstZone.claims / totalClaims) * 100) : 0
    
    return {
      name: worstZone.zone_name,
      city: worstZone.city,
      contribution: contribution,
      ratio: Math.round(worstZone.loss_ratio * 100)
    }
  }, [rows])

  const getBarColor = (ratio: number) => {
    if (ratio < 30) return '#065f46' // Deep Emerald (Highly Efficient)
    if (ratio < 45) return '#10b981' // Emerald (Efficient)
    if (ratio <= 55) return '#f59e0b' // Amber (Monitoring)
    return '#f43f5e' // Rose (Critical)
  }

  return (
    <PageShell
      eyebrow={t('pages.lossRatio.eyebrow')}
      title={t('pages.lossRatio.title')}
      description={t('pages.lossRatio.description')}
    >
      {error ? (
        <div className="mb-6 rounded-2xl border border-rose-200 bg-rose-50 p-6 text-xs font-black uppercase tracking-widest text-rose-700 shadow-[0_20px_50px_rgba(244,63,94,0.1)]">
          <div className="flex items-center gap-3">
            <AlertCircle size={16} />
            {error}
          </div>
        </div>
      ) : null}

      <div className="grid gap-10 xl:grid-cols-[1fr_0.4fr]">
        <Panel title={t('pages.lossRatio.zoneMetrics')} subtitle={`${t('pages.lossRatio.zoneMetricsSubtitle')} vs Target Performance`}>
          
          {/* Insight Line */}
          {!loading && topInsight && (
            <div className="mb-10 flex items-center gap-5 p-6 rounded-2xl bg-slate-900 dark:bg-brand-primary/5 border-l-8 border-brand-primary shadow-2xl animate-in fade-in slide-in-from-top-4 duration-700">
              <div className="h-12 w-12 rounded-xl bg-brand-primary/10 flex items-center justify-center text-brand-primary shrink-0 border border-brand-primary/20">
                <TrendingUp size={24} strokeWidth={3} />
              </div>
              <div className="flex-1">
                <p className="text-[10px] font-black uppercase tracking-[0.3em] text-brand-primary mb-1 opacity-80">System Priority Alert</p>
                <p className="text-[15px] font-bold text-white leading-snug">
                  <span className="text-brand-primary font-black">{topInsight.name} ({topInsight.city})</span> is currently generating <span className="font-black bg-brand-primary/20 px-1 rounded">{topInsight.contribution}%</span> of platform-wide claims. 
                  <span className="ml-2 text-slate-400 font-medium italic opacity-60 italic">Reviewing pricing multipliers for this sector.</span>
                </p>
              </div>
            </div>
          )}

          <div className="h-[500px] w-full">
            {loading ? (
              <div className="flex h-full items-center justify-center text-[10px] text-slate-400 font-black uppercase tracking-[0.4em] animate-pulse">Synchronizing actuarial feeds...</div>
            ) : chartData.length > 0 ? (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={chartData} margin={{ top: 60, right: 80, left: 20, bottom: 20 }} layout="vertical">
                  <CartesianGrid strokeDasharray="12 12" horizontal={false} stroke="rgba(148, 163, 184, 0.04)" />
                  <XAxis 
                    type="number" 
                    domain={[-50, MAX_DIVERGENCE]} 
                    axisLine={false} 
                    tickLine={false} 
                    tick={{ fontSize: 10, fill: '#94a3b8', fontWeight: 900 }} 
                    tickFormatter={(val) => `${val + TARGET_RATIO}%`}
                  />
                  <YAxis 
                    dataKey="name" 
                    type="category" 
                    axisLine={false} 
                    tickLine={false} 
                    tick={{ fontSize: 10, fill: '#475569', fontWeight: 900 }} 
                    width={110}
                  />
                  <Tooltip
                    cursor={{ fill: 'rgba(236, 72, 153, 0.03)' }}
                    content={({ active, payload }) => {
                      if (active && payload && payload.length) {
                        const data = payload[0].payload
                        const color = getBarColor(data.ratio)
                        return (
                          <div className="bg-white/95 dark:bg-slate-900/95 backdrop-blur-3xl border border-slate-200 dark:border-slate-800 p-6 rounded-3xl shadow-[0_30px_60px_rgba(0,0,0,0.15)]">
                            <p className="text-[10px] font-black uppercase tracking-[0.2em] text-slate-400 mb-4">{data.name} · {data.city}</p>
                            <div className="space-y-4">
                              <div className="flex items-end gap-3">
                                <p className="text-4xl font-black text-slate-900 dark:text-white leading-none tracking-tighter">{data.ratio}%</p>
                                <p className="text-[11px] font-black text-slate-500 uppercase mb-1 opacity-60">Loss Index</p>
                              </div>
                              <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-slate-50 dark:bg-slate-800 w-fit">
                                <div className="h-2 w-2 rounded-full" style={{ backgroundColor: color }} />
                                <p className="text-[11px] font-black uppercase tracking-widest" style={{ color: color }}>
                                  {data.divergence > 0 ? `+${data.divergence}% VS TARGET` : `${Math.abs(data.divergence)}% EFFICIENT`}
                                </p>
                              </div>
                              {data.isCapped && (
                                <p className="text-[10px] font-bold text-slate-400 italic pt-3 border-t border-slate-100 dark:border-slate-800 uppercase tracking-tighter">
                                  Exact Actuarial Index: {data.ratio}%
                                </p>
                              )}
                            </div>
                          </div>
                        )
                      }
                      return null
                    }}
                  />

                  {/* Target Line */}
                  <ReferenceLine 
                    x={0} 
                    stroke="#EC4899" 
                    strokeWidth={2} 
                    strokeDasharray="6 6"
                    label={{ 
                      value: 'TARGET 50%', 
                      position: 'top', 
                      fill: '#EC4899', 
                      fontSize: 10, 
                      fontWeight: 900,
                      offset: 20 
                    }} 
                  />

                  <Bar dataKey="displayDivergence" barSize={32} radius={[0, 6, 6, 0]}>
                    <LabelList 
                      dataKey="ratio" 
                      content={(props: any) => {
                        const { x, y, width, height, value } = props
                        const isPositive = (value as number) > TARGET_RATIO
                        const labelText = (value as number) > 150 ? '150%+' : `${value}%`
                        // For positive bars, place to the right of the bar end (x + width)
                        // For negative bars, place to the left of the bar end (x)
                        const labelX = isPositive ? x + width + 10 : x - 10
                        
                        return (
                          <text 
                            x={labelX} 
                            y={y + height / 2} 
                            fill={isPositive ? '#f43f5e' : '#065f46'} 
                            fontSize={11} 
                            fontWeight={900} 
                            textAnchor={isPositive ? 'start' : 'end'} 
                            dominantBaseline="middle"
                            className="font-['Outfit'] italic shadow-sm"
                          >
                            {labelText}
                          </text>
                        )
                      }}
                    />
                    {chartData.map((entry, index) => {
                      const isTarget = topInsight && entry.name === topInsight.name
                      const baseColor = getBarColor(entry.ratio)
                      return (
                        <Cell 
                          key={`cell-${index}`} 
                          fill={baseColor}
                          radius={entry.displayDivergence > 0 ? [0, 8, 8, 0] as any : [8, 0, 0, 8] as any}
                          stroke={isTarget ? '#EC4899' : 'none'}
                          strokeWidth={isTarget ? 4 : 0}
                          className={isTarget ? 'drop-shadow-[0_0_15px_rgba(236,72,153,0.5)]' : ''}
                          opacity={isTarget ? 1 : 0.85}
                        />
                      )
                    })}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="flex h-full items-center justify-center text-sm text-slate-500 italic opacity-50 uppercase tracking-[0.2em]">No zone metrics currently streaming.</div>
            )}
          </div>
        </Panel>

        <Panel title={t('pages.lossRatio.summaryInsights')} subtitle="Operational Guidance">
          <div className="space-y-6">
            <div className="rounded-2xl border border-orange-100 dark:border-orange-900/20 bg-orange-50/20 dark:bg-orange-900/5 p-6 shadow-sm">
              <div className="flex items-center gap-2 mb-3">
                <AlertCircle className="h-4 w-4 text-orange-500" />
                <p className="text-[10px] font-black uppercase tracking-widest text-orange-800 dark:text-orange-400">{t('pages.lossRatio.exposureAlert')}</p>
              </div>
              <p className="text-xs leading-relaxed font-medium text-orange-800/80 dark:text-orange-400/80">{t('pages.lossRatio.exposureAlertDesc')}</p>
            </div>
            <div className="rounded-2xl border border-emerald-100 dark:border-emerald-900/20 bg-emerald-50/20 dark:bg-emerald-900/5 p-6 shadow-sm">
              <div className="flex items-center gap-2 mb-3">
                <TrendingDown className="h-4 w-4 text-emerald-500" />
                <p className="text-[10px] font-black uppercase tracking-widest text-emerald-800 dark:text-emerald-400">{t('pages.lossRatio.growthOpportunity')}</p>
              </div>
              <p className="text-xs leading-relaxed font-medium text-emerald-800/80 dark:text-emerald-400/80">{t('pages.lossRatio.growthOpportunityDesc')}</p>
            </div>

            <div className="pt-4 border-t border-slate-50 dark:border-slate-800 mt-4">
               <div className="flex items-start gap-4 p-4 rounded-xl bg-slate-50 dark:bg-slate-900/50">
                  <Info className="h-4 w-4 text-slate-400 shrink-0 mt-0.5" />
                  <p className="text-[10px] text-slate-500 font-medium leading-relaxed italic">
                    Zones appearing in <span className="font-black text-rose-500 uppercase tracking-tighter">Rose</span> have exceeded the 55% loss threshold and are prioritized for real-time order pricing adjustment.
                  </p>
               </div>
            </div>
          </div>
        </Panel>
      </div>

      <Panel title={t('pages.lossRatio.dataGrid')} className="mt-8 overflow-hidden">
        <div className="overflow-x-auto -mx-10 px-10">
          <table className="w-full text-left">
            <thead>
              <tr className="border-b border-slate-100 dark:border-slate-800">
                <th className="pb-5 px-4 font-black uppercase tracking-[0.2em] text-[10px] text-slate-400 font-['Outfit']">{t('pages.lossRatio.headerZone')}</th>
                <th className="pb-5 px-4 font-black uppercase tracking-[0.2em] text-[10px] text-slate-400 font-['Outfit']">{t('pages.lossRatio.headerPremiums')}</th>
                <th className="pb-5 px-4 font-black uppercase tracking-[0.2em] text-[10px] text-slate-400 font-['Outfit']">{t('pages.lossRatio.headerClaims')}</th>
                <th className="pb-5 px-4 font-black uppercase tracking-[0.2em] text-[10px] text-slate-400 font-['Outfit'] text-right">{t('pages.lossRatio.headerRatio')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-50 dark:divide-slate-800/50">
              {rows.map((row) => {
                 const ratio = Math.round(row.loss_ratio * 100)
                 const color = getBarColor(ratio)
                 return (
                  <tr key={`${row.city}-${row.zone_name}`} className="hover:bg-slate-50/50 dark:hover:bg-slate-800/30 transition-colors group">
                    <td className="py-5 px-4">
                      <p className="font-black text-slate-900 dark:text-white uppercase tracking-tight text-xs">{row.zone_name}</p>
                      <p className="text-[9px] font-black uppercase tracking-[0.2em] text-slate-400 mt-0.5">{row.city}</p>
                    </td>
                    <td className="py-5 px-4 text-xs font-bold text-slate-500">₹{Math.round(row.premiums).toLocaleString()}</td>
                    <td className="py-5 px-4 text-xs font-bold text-slate-500">₹{Math.round(row.claims).toLocaleString()}</td>
                    <td className="py-5 px-4 text-right">
                      <span
                        className="rounded-full px-4 py-1.5 text-[10px] font-black uppercase tracking-widest border transition-all"
                        style={{ 
                          backgroundColor: `${color}10`, 
                          borderColor: `${color}30`,
                          color: color 
                        }}
                      >
                        {ratio}%
                      </span>
                    </td>
                  </tr>
                 )
              })}
              {rows.length === 0 && !error ? (
                <tr>
                  <td className="py-24 text-center italic text-slate-400 font-medium uppercase tracking-widest text-[10px]" colSpan={4}>
                    Awaiting real-time actuarial streams...
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
