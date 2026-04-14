import { useEffect, useMemo, useState } from 'react'
import { getDisruptions, getZoneHealth, getForecast } from '../api/platform'
import { BarChart3, TrendingDown, ClipboardList, AlertOctagon, ArrowDownUp, Download, BrainCircuit, Info, RefreshCw } from 'lucide-react'

type SortConfig = { key: string; direction: 'asc' | 'desc' } | null

interface ForecastPoint {
  date: string
  disruption_probability: number
}

export default function Analytics() {
  const [health, setHealth] = useState<any[]>([])
  const [disruptions, setDisruptions] = useState<any[]>([])
  const [timeFilter, setTimeFilter] = useState<'all' | 'weekly' | 'real-time'>('all')
  const [sortConfig, setSortConfig] = useState<SortConfig>(null)
  const [forecast, setForecast] = useState<ForecastPoint[]>([])
  const [forecastMeta, setForecastMeta] = useState<{ retraining_cadence?: string; scope?: string } | null>(null)
  const [forecastInference, setForecastInference] = useState<'prophet' | 'seasonal' | 'fallback' | null>(null)
  const [forecastLoading, setForecastLoading] = useState(true)
  const [forecastError, setForecastError] = useState(false)
  const [selectedZone, setSelectedZone] = useState(1)

  useEffect(() => {
    async function load() {
      const [healthRes, disruptionsRes] = await Promise.all([getZoneHealth(), getDisruptions()])
      setHealth(healthRes.data.data ?? [])
      setDisruptions(disruptionsRes.data.data ?? [])
    }

    load().catch((error) => console.error('Failed to load analytics', error))
    const timer = setInterval(() => load().catch(() => undefined), 5000)
    return () => clearInterval(timer)
  }, [])

  useEffect(() => {
    async function loadForecast() {
      setForecastLoading(true)
      setForecastError(false)
      try {
        const res = await getForecast(selectedZone)
        setForecast(res.data.forecast ?? [])
        setForecastMeta({
          retraining_cadence: res.data.retraining_cadence,
          scope: res.data.scope,
        })
        setForecastInference(res.data.inference ?? null)
      } catch {
        setForecastError(true)
      } finally {
        setForecastLoading(false)
      }
    }
    loadForecast()
  }, [selectedZone])

  const stats = useMemo(() => {
    const avgDrop = health.length
      ? Math.round((health.reduce((sum, item) => sum + (item.order_drop ?? 0), 0) / health.length) * 100)
      : 0
    const manualReview = disruptions.reduce((sum, item) => sum + (item.claims_in_review ?? 0), 0)
    const claims = disruptions.reduce((sum, item) => sum + (item.claims_generated ?? 0), 0)
    return { avgDrop, manualReview, claims }
  }, [health, disruptions])

  const filteredAndSortedDisruptions = useMemo(() => {
    let result = [...disruptions]

    const now = new Date()
    if (timeFilter === 'real-time') {
      result = result.filter(d => (now.getTime() - new Date(d.started_at).getTime()) < 10 * 60 * 1000)
    } else if (timeFilter === 'weekly') {
      result = result.filter(d => (now.getTime() - new Date(d.started_at).getTime()) < 7 * 24 * 60 * 60 * 1000)
    }

    if (sortConfig) {
      result.sort((a, b) => {
        let aVal = a[sortConfig.key]
        let bVal = b[sortConfig.key]

        if (sortConfig.key === 'disruption_id' || sortConfig.key === 'zone_id') {
          aVal = parseInt(aVal.toString().split('_')[1])
          bVal = parseInt(bVal.toString().split('_')[1])
        }

        if (aVal < bVal) return sortConfig.direction === 'asc' ? -1 : 1
        if (aVal > bVal) return sortConfig.direction === 'asc' ? 1 : -1
        return 0
      })
    }

    return result
  }, [disruptions, timeFilter, sortConfig])

  const handleSort = (key: string) => {
    let direction: 'asc' | 'desc' = 'asc'
    if (sortConfig && sortConfig.key === key && sortConfig.direction === 'asc') {
      direction = 'desc'
    }
    setSortConfig({ key, direction })
  }

  const exportCSV = () => {
    const headers = ['Event ID', 'Zone', 'Type', 'Claims', 'Status', 'Payouts', 'Time']
    const rows = filteredAndSortedDisruptions.map(d => [
      d.disruption_id,
      d.zone_id,
      d.type,
      d.claims_generated,
      d.automation_status,
      d.payout_amount_total,
      d.started_at
    ])

    const csvContent = [headers, ...rows].map(e => e.join(",")).join("\n")
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement("a")
    const url = URL.createObjectURL(blob)
    link.setAttribute("href", url)
    link.setAttribute("download", `analytics_export_${new Date().toISOString()}.csv`)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  const maxProb = Math.max(...forecast.map(f => f.disruption_probability), 0.01)

  return (
    <div className="space-y-10">
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-2xl font-black tracking-tight text-slate-900 dark:text-white">Platform Analytics</h1>
          <p className="mt-1 text-sm text-slate-500">Deep dive into historical disruption trends and automation performance.</p>
        </div>
        <button
          onClick={exportCSV}
          className="flex items-center gap-2 h-9 px-4 rounded-md border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-xs font-bold font-['Outfit'] hover:bg-slate-50 dark:hover:bg-slate-800 transition-none">
          <Download className="h-3.5 w-3.5" />
          EXPORT DATA
        </button>
      </div>

      <div className="grid gap-6 md:grid-cols-3">
        <AnalyticsCard label="Average Zone Drop" value={`${stats.avgDrop}%`} icon={TrendingDown} color="text-rose-600" />
        <AnalyticsCard label="Claims Triggered" value={stats.claims} icon={ClipboardList} color="text-orange-600" />
        <AnalyticsCard label="Manual Review Queue" value={stats.manualReview} icon={AlertOctagon} color={stats.manualReview > 0 ? "text-amber-600" : "text-emerald-600"} />
      </div>

      {/* ── Reserve Planning Forecast ─────────────────────────────────────── */}
      <div className="enterprise-panel p-8">
        <div className="flex items-start justify-between mb-6">
          <div className="flex items-center gap-3">
            <BrainCircuit className="h-5 w-5 text-violet-500" />
            <div>
              <h2 className="text-sm font-black uppercase tracking-widest text-slate-900 dark:text-white font-['Outfit']">
                7-Day Reserve Forecast
              </h2>
              <p className="text-[10px] text-slate-400 mt-0.5 uppercase tracking-widest">
                Facebook Prophet · Zone {selectedZone} · Reserve planning only
              </p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            {/* Zone selector */}
            <div className="flex gap-1">
              {[1, 2, 3, 4].map(z => (
                <button
                  key={z}
                  id={`forecast-zone-${z}`}
                  onClick={() => setSelectedZone(z)}
                  className={`px-3 py-1.5 rounded border text-[10px] font-black uppercase tracking-widest transition-none ${selectedZone === z
                      ? 'bg-violet-600 border-violet-600 text-white'
                      : 'border-slate-200 dark:border-slate-700 text-slate-500 hover:text-slate-900 dark:hover:text-white hover:border-violet-300'
                    }`}
                >
                  Zone {z}
                </button>
              ))}
            </div>
            {forecastLoading && <RefreshCw className="h-3.5 w-3.5 text-slate-400 animate-spin" />}
            {forecastInference === 'prophet' && (
              <div className="px-3 py-1.5 rounded border border-emerald-200 dark:border-emerald-500/30 bg-emerald-50 dark:bg-emerald-500/10">
                <span className="text-[9px] font-black uppercase tracking-widest text-emerald-600 dark:text-emerald-400">● Prophet Live</span>
              </div>
            )}
            {forecastInference === 'fallback' && (
              <div className="px-3 py-1.5 rounded border border-amber-200 dark:border-amber-500/30 bg-amber-50 dark:bg-amber-500/10">
                <span className="text-[9px] font-black uppercase tracking-widest text-amber-600 dark:text-amber-400">Static Fallback</span>
              </div>
            )}
            <div className="px-3 py-1.5 rounded border border-violet-200 dark:border-violet-500/30 bg-violet-50 dark:bg-violet-500/10">
              <span className="text-[9px] font-black uppercase tracking-widest text-violet-600 dark:text-violet-400">
                Reserve Planning Only
              </span>
            </div>
          </div>
        </div>

        {/* Critical guardrail notice */}
        <div className="mb-6 flex items-start gap-3 p-4 rounded-lg border border-amber-200 dark:border-amber-500/30 bg-amber-50 dark:bg-amber-500/5">
          <Info className="h-4 w-4 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5" />
          <div className="space-y-1">
            <p className="text-[11px] font-bold text-amber-800 dark:text-amber-300">
              This forecast does not influence claim approval or claim decisioning.
            </p>
            <p className="text-[10px] text-amber-700 dark:text-amber-400 leading-relaxed">
              Prophet is a per-zone seasonal model used exclusively for reserve buffer planning.
              It does not model cross-zone correlated events. Reinsurance and a catastrophic cap
              cover tail risk at portfolio level. Upgrade path: DeepAR for joint distribution modelling.
            </p>
          </div>
        </div>

        {forecastError ? (
          <div className="py-10 text-center text-xs text-slate-400 italic">
            Forecast service unavailable — reserve planning estimates offline.
          </div>
        ) : forecastLoading ? (
          <div className="py-10 text-center text-xs text-slate-400 italic">Loading forecast...</div>
        ) : (
          <>
            {/* Bar chart */}
            <div className="grid grid-cols-7 gap-2 items-end h-32 mb-3">
              {forecast.map((point) => {
                const pct = (point.disruption_probability / maxProb) * 100
                const prob = Math.round(point.disruption_probability * 100)
                const isHigh = point.disruption_probability >= 0.35
                const isMed = point.disruption_probability >= 0.20 && !isHigh
                return (
                  <div key={point.date} className="flex flex-col items-center gap-1 h-full justify-end">
                    <span className={`text-[9px] font-black ${isHigh ? 'text-rose-500' : isMed ? 'text-amber-500' : 'text-emerald-500'}`}>
                      {prob}%
                    </span>
                    <div className="w-full rounded-t overflow-hidden" style={{ height: `${Math.max(pct, 6)}%` }}>
                      <div
                        className={`w-full h-full rounded-t ${isHigh
                            ? 'bg-rose-500/80 dark:bg-rose-500/60'
                            : isMed
                              ? 'bg-amber-500/80 dark:bg-amber-500/60'
                              : 'bg-emerald-500/70 dark:bg-emerald-500/50'
                          }`}
                      />
                    </div>
                  </div>
                )
              })}
            </div>

            {/* Date labels */}
            <div className="grid grid-cols-7 gap-2 mb-6">
              {forecast.map((point) => {
                const d = new Date(point.date)
                return (
                  <div key={point.date} className="text-center">
                    <div className="text-[9px] font-bold text-slate-400 uppercase">
                      {d.toLocaleDateString('en', { weekday: 'short' })}
                    </div>
                    <div className="text-[8px] text-slate-300 dark:text-slate-600">
                      {d.toLocaleDateString('en', { month: 'short', day: 'numeric' })}
                    </div>
                  </div>
                )
              })}
            </div>

            {/* Legend + metadata */}
            <div className="flex flex-wrap items-center justify-between gap-4 pt-5 border-t border-slate-100 dark:border-slate-800">
              <div className="flex gap-4">
                <LegendDot color="bg-emerald-500" label="Low (<20%)" />
                <LegendDot color="bg-amber-500" label="Medium (20–35%)" />
                <LegendDot color="bg-rose-500" label="High (>35%)" />
              </div>
              <div className="flex flex-col items-end gap-0.5">
                {forecastMeta?.retraining_cadence && (
                  <span className="text-[9px] text-slate-400 font-bold uppercase tracking-widest">
                    Retraining: {forecastMeta.retraining_cadence}
                  </span>
                )}
                {forecastMeta?.scope && (
                  <span className="text-[9px] text-slate-300 dark:text-slate-600">
                    {forecastMeta.scope}
                  </span>
                )}
              </div>
            </div>
          </>
        )}
      </div>

      {/* ── Disruption Table ──────────────────────────────────────────────── */}
      <div className="enterprise-panel overflow-hidden">
        <div className="border-b border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/20 p-6 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <BarChart3 className="h-5 w-5 text-slate-400" />
            <h2 className="text-sm font-black uppercase tracking-widest text-slate-900 dark:text-white font-['Outfit']">Disruption Feed &amp; History</h2>
          </div>
          <div className="flex gap-2">
            {(['all', 'weekly', 'real-time'] as const).map(f => (
              <button
                key={f}
                onClick={() => setTimeFilter(f)}
                className={`px-3 py-1.5 rounded border text-[10px] font-bold transition-none uppercase ${timeFilter === f
                    ? 'bg-orange-600 border-orange-600 text-white shadow-sm'
                    : 'border-slate-200 dark:border-slate-700 text-slate-500 hover:text-slate-900 dark:hover:text-white'
                  }`}
              >
                {f}
              </button>
            ))}
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full border-collapse text-left">
            <thead>
              <tr className="border-b border-slate-100 dark:border-slate-800">
                <SortableHeader label="Event ID" sortKey="disruption_id" currentSort={sortConfig} onSort={handleSort} />
                <SortableHeader label="Zone" sortKey="zone_id" currentSort={sortConfig} onSort={handleSort} />
                <th className="px-8 py-4 text-[10px] font-black uppercase tracking-widest text-slate-400">Type</th>
                <SortableHeader label="Claims" sortKey="claims_generated" currentSort={sortConfig} onSort={handleSort} />
                <th className="px-8 py-4 text-[10px] font-black uppercase tracking-widest text-slate-400">Status</th>
                <SortableHeader label="Payouts" sortKey="payout_amount_total" currentSort={sortConfig} onSort={handleSort} />
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
              {filteredAndSortedDisruptions.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-8 py-20 text-center text-xs text-slate-400 italic">
                    No disruption records found for the selected time horizon.
                  </td>
                </tr>
              ) : filteredAndSortedDisruptions.map((item) => (
                <tr key={item.disruption_id} className="hover:bg-slate-50 dark:hover:bg-slate-800/30">
                  <td className="px-8 py-5">
                    <span className="text-xs font-black font-['Outfit'] text-orange-600 dark:text-orange-500">#{item.disruption_id.split('_')[1]}</span>
                  </td>
                  <td className="px-8 py-5">
                    <div className="text-xs font-bold text-slate-900 dark:text-white uppercase tracking-tight">Zone {item.zone_id.replace('zone_', '')}</div>
                  </td>
                  <td className="px-8 py-5">
                    <div className="text-xs text-slate-600 dark:text-slate-300 font-medium truncate max-w-[140px] uppercase tracking-tighter">{item.type}</div>
                  </td>
                  <td className="px-8 py-5">
                    <div className="text-xs font-bold text-slate-900 dark:text-white">{item.claims_generated}</div>
                  </td>
                  <td className="px-8 py-5">
                    <div className={`flex items-center gap-2 px-2 py-1 rounded w-fit border transition-all ${item.automation_status === 'paid'
                        ? 'bg-emerald-50 dark:bg-emerald-500/10 border-emerald-100 dark:border-emerald-500/20 text-emerald-600'
                        : 'bg-orange-50 dark:bg-orange-500/10 border-orange-100 dark:border-orange-500/20 text-orange-600'
                      }`}>
                      <div className={`h-1 w-1 rounded-full ${item.automation_status === 'paid' ? 'bg-emerald-500' : 'bg-orange-500'}`}></div>
                      <span className="text-[9px] font-black uppercase tracking-widest">{item.automation_status}</span>
                    </div>
                  </td>
                  <td className="px-8 py-5">
                    <div className="text-xs font-black text-slate-900 dark:text-white">Rs {Math.round(item.payout_amount_total).toLocaleString()}</div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

function LegendDot({ color, label }: { color: string; label: string }) {
  return (
    <div className="flex items-center gap-1.5">
      <div className={`h-2 w-2 rounded-full ${color}`} />
      <span className="text-[9px] text-slate-400 font-bold uppercase tracking-widest">{label}</span>
    </div>
  )
}

function SortableHeader({ label, sortKey, currentSort, onSort }: { label: string, sortKey: string, currentSort: SortConfig, onSort: (key: string) => void }) {
  const isSorted = currentSort?.key === sortKey
  return (
    <th
      onClick={() => onSort(sortKey)}
      className="px-8 py-4 text-[10px] font-black uppercase tracking-widest text-slate-400 cursor-pointer hover:text-slate-900 dark:hover:text-white transition-colors"
    >
      <div className="flex items-center gap-2">
        {label}
        <ArrowDownUp className={`h-3 w-3 ${isSorted ? 'text-orange-500' : 'text-slate-300'}`} />
      </div>
    </th>
  )
}

function AnalyticsCard({ label, value, icon: Icon, color }: { label: string; value: string | number; icon: any; color: string }) {
  return (
    <div className="enterprise-panel p-8">
      <div className="flex items-center gap-4 mb-4">
        <div className="h-12 w-12 flex items-center justify-center rounded-xl bg-slate-50 dark:bg-slate-800 border border-slate-100 dark:border-slate-800">
          <Icon className={`h-6 w-6 ${color}`} />
        </div>
        <div>
          <div className="text-[10px] font-black uppercase tracking-[0.2em] text-slate-400 font-['Outfit']">{label}</div>
          <div className="text-3xl font-black text-slate-900 dark:text-white mt-1 font-['Outfit']">{value}</div>
        </div>
      </div>
      <div className="w-full h-1 bg-slate-100 dark:bg-slate-800 rounded-full mt-6 overflow-hidden">
        <div className={`h-full ${color.replace('text-', 'bg-')} w-2/3 opacity-80`}></div>
      </div>
    </div>
  )
}
