import { useEffect, useMemo, useState } from 'react'
import { getDisruptions, getZoneHealth } from '../api/platform'
import { BarChart3, TrendingDown, ClipboardList, AlertOctagon, ArrowDownUp, Download } from 'lucide-react'

type SortConfig = { key: string; direction: 'asc' | 'desc' } | null

export default function Analytics() {
  const [health, setHealth] = useState<any[]>([])
  const [disruptions, setDisruptions] = useState<any[]>([])
  const [timeFilter, setTimeFilter] = useState<'all' | 'weekly' | 'real-time'>('all')
  const [sortConfig, setSortConfig] = useState<SortConfig>(null)

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

    // Time filtering
    const now = new Date()
    if (timeFilter === 'real-time') {
      result = result.filter(d => (now.getTime() - new Date(d.started_at).getTime()) < 10 * 60 * 1000) // Last 10 mins
    } else if (timeFilter === 'weekly') {
      result = result.filter(d => (now.getTime() - new Date(d.started_at).getTime()) < 7 * 24 * 60 * 60 * 1000) // Last 7 days
    }

    // Sorting
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

      <div className="enterprise-panel overflow-hidden">
        <div className="border-b border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/20 p-6 flex items-center justify-between">
           <div className="flex items-center gap-3">
              <BarChart3 className="h-5 w-5 text-slate-400" />
              <h2 className="text-sm font-black uppercase tracking-widest text-slate-900 dark:text-white font-['Outfit']">Disruption Feed & History</h2>
           </div>
           <div className="flex gap-2">
              {(['all', 'weekly', 'real-time'] as const).map(f => (
                <button
                  key={f}
                  onClick={() => setTimeFilter(f)}
                  className={`px-3 py-1.5 rounded border text-[10px] font-bold transition-none uppercase ${
                    timeFilter === f 
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
                     <div className={`flex items-center gap-2 px-2 py-1 rounded w-fit border transition-all ${
                       item.automation_status === 'paid' 
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
