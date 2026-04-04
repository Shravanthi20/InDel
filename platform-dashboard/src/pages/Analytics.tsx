import { useEffect, useMemo, useState } from 'react'
import { getDisruptions, getZoneHealth } from '../api/platform'
import { BarChart3, TrendingDown, ClipboardList, CheckCircle2, AlertOctagon, ArrowDown } from 'lucide-react'

export default function Analytics() {
  const [health, setHealth] = useState<any[]>([])
  const [disruptions, setDisruptions] = useState<any[]>([])

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

  return (
    <div className="space-y-10">
      <div>
        <h1 className="text-2xl font-black tracking-tight text-slate-900 dark:text-white">Platform Analytics</h1>
        <p className="mt-1 text-sm text-slate-500">Deep dive into historical disruption trends and automation performance.</p>
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
              <h2 className="text-sm font-black uppercase tracking-widest text-slate-900 dark:text-white">Disruption Feed & Automation History</h2>
           </div>
           <div className="flex gap-2">
              <button className="px-3 py-1.5 rounded border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-[10px] font-bold text-slate-500 hover:text-slate-900 dark:hover:text-white transition-none">
                 WEEKLY
              </button>
              <button className="px-3 py-1.5 rounded border border-slate-200 dark:border-slate-700 bg-orange-600 text-white text-[10px] font-bold transition-none">
                 REAL-TIME
              </button>
           </div>
        </div>

        <div className="p-0">
          <div className="overflow-x-auto">
            <table className="w-full border-collapse text-left">
              <thead>
                <tr className="border-b border-slate-100 dark:border-slate-800">
                  <th className="px-8 py-4 text-[10px] font-black uppercase tracking-widest text-slate-400">Event ID</th>
                  <th className="px-8 py-4 text-[10px] font-black uppercase tracking-widest text-slate-400">Zone</th>
                  <th className="px-8 py-4 text-[10px] font-black uppercase tracking-widest text-slate-400">Type</th>
                  <th className="px-8 py-4 text-[10px] font-black uppercase tracking-widest text-slate-400">Claims</th>
                  <th className="px-8 py-4 text-[10px] font-black uppercase tracking-widest text-slate-400">Status</th>
                  <th className="px-8 py-4 text-[10px] font-black uppercase tracking-widest text-slate-400">Payouts</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                {disruptions.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="px-8 py-12 text-center text-xs text-slate-400 italic">No historical data available.</td>
                  </tr>
                ) : disruptions.map((item) => (
                  <tr key={item.disruption_id} className="hover:bg-slate-50 dark:hover:bg-slate-800/30">
                    <td className="px-8 py-5">
                       <span className="text-xs font-black font-['Outfit'] text-orange-600 dark:text-orange-500">#{item.disruption_id.substr(0, 10)}</span>
                    </td>
                    <td className="px-8 py-5">
                       <div className="text-xs font-bold text-slate-900 dark:text-white">Zone {item.zone_id}</div>
                    </td>
                    <td className="px-8 py-5">
                       <div className="text-xs text-slate-600 dark:text-slate-300 font-medium truncate max-w-[140px] uppercase tracking-tighter">{item.type}</div>
                    </td>
                    <td className="px-8 py-5">
                       <div className="text-xs font-bold text-slate-900 dark:text-white">{item.claims_generated}</div>
                    </td>
                    <td className="px-8 py-5">
                       <div className={`flex items-center gap-2 px-2 py-1 rounded w-fit border ${
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
    </div>
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
          <div className="text-[10px] font-black uppercase tracking-[0.2em] text-slate-400">{label}</div>
          <div className="text-3xl font-black text-slate-900 dark:text-white mt-1">{value}</div>
        </div>
      </div>
      <div className="w-full h-1 bg-slate-100 dark:bg-slate-800 rounded-full mt-6 overflow-hidden">
         <div className={`h-full ${color.replace('text-', 'bg-')} w-2/3`}></div>
      </div>
    </div>
  )
}
