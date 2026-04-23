import { useEffect, useMemo, useRef, useState } from 'react'
import { getDisruptions, getWorkers, getZoneHealth, getZones } from '../api/platform'
import { Users, MapPin, ShoppingBag, AlertTriangle, ArrowUpRight, TrendingUp, ShieldCheck } from 'lucide-react'
import { useLocalization } from '../context/LocalizationContext'

export default function Overview() {
  const { t } = useLocalization()
  const [workers, setWorkers] = useState<any[]>([])
  const [previousWorkerCount, setPreviousWorkerCount] = useState<number | null>(null)
  const lastWorkerCountRef = useRef<number | null>(null)
  const [zones, setZones] = useState<any[]>([])
  const [health, setHealth] = useState<any[]>([])
  const [disruptions, setDisruptions] = useState<any[]>([])

  useEffect(() => {
    async function load() {
      const [workersRes, zonesRes, healthRes, disruptionsRes] = await Promise.all([
        getWorkers(),
        getZones(),
        getZoneHealth(),
        getDisruptions(),
      ])
      const nextWorkerCount = workersRes.data.workers?.length ?? 0
      setPreviousWorkerCount(lastWorkerCountRef.current)
      lastWorkerCountRef.current = nextWorkerCount
      setWorkers(workersRes.data.workers ?? [])
      setZones(zonesRes.data.zones ?? [])
      setHealth(healthRes.data.data ?? [])
      setDisruptions(disruptionsRes.data.data ?? [])
    }

    load().catch((error) => console.error('Failed to load platform overview', error))
    const timer = setInterval(() => load().catch(() => undefined), 2000) // Lower polling to 2s for "snappy" dynamic feel
    return () => clearInterval(timer)
  }, [])

  const totals = useMemo(() => {
    const liveOrders = health.reduce((sum, item) => sum + (item.current_orders ?? 0), 0)
    const baselineOrders = health.reduce((sum, item) => sum + (item.baseline_orders ?? 0), 0)
    const disruptedZones = health.filter((item) => item.status === 'disrupted').length
    const paid = disruptions.reduce((sum, item) => sum + (item.payout_amount_total ?? 0), 0)
    const ordersTrend = baselineOrders > 0 ? ((liveOrders - baselineOrders) / baselineOrders * 100).toFixed(1) : '0.0'
    const healthyZones = health.filter((item) => item.status === 'healthy').length
    const workerDelta = previousWorkerCount === null ? null : workers.length - previousWorkerCount
    const workersTrend = workerDelta === null
      ? 'Loading'
      : workerDelta === 0
        ? t('pages.overview.stable')
        : `${workerDelta > 0 ? '+' : ''}${workerDelta} ${t('pages.overview.refreshed')}`
    const coverage = health.length > 0 ? Math.round((healthyZones / health.length) * 100) : 0

    return { liveOrders, baselineOrders, disruptedZones, paid, ordersTrend, workersTrend, coverage }
  }, [health, disruptions, workers, previousWorkerCount])

  return (
    <div className="space-y-12">
      <div>
        <p className="text-[10px] font-black uppercase tracking-[0.5em] text-brand-primary mb-3">Institutional Overview</p>
        <h1 className="text-5xl font-black tracking-tighter text-gray-900 dark:text-white font-['Outfit'] italic">Market <span className="text-brand-primary">Pulse</span></h1>
        <p className="mt-4 text-lg text-gray-500 max-w-2xl leading-relaxed">{t('pages.overview.description')}</p>
      </div>

      <div className="grid gap-8 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard label={t('pages.overview.activeWorkers')} value={workers.length} icon={Users} trend={totals.workersTrend} color="text-brand-primary" />
        <MetricCard label={t('pages.overview.trackedZones')} value={zones.length} icon={MapPin} trend={t('pages.overview.stable')} color="text-gray-600" />
        <MetricCard label={t('pages.overview.liveOrders')} value={totals.liveOrders} icon={ShoppingBag} trend={`${totals.ordersTrend}%`} color={parseFloat(totals.ordersTrend) < -30 ? "text-brand-primary" : "text-gray-900"} />
        <MetricCard label={t('pages.overview.disrupted')} value={totals.disruptedZones} icon={AlertTriangle} trend={totals.disruptedZones > 0 ? t('pages.overview.critical') : t('pages.overview.none')} color={totals.disruptedZones > 0 ? "text-brand-primary" : "text-emerald-600"} />
      </div>

      <div className="grid gap-10 lg:grid-cols-2">
        <div className="enterprise-panel p-10">
          <div className="flex items-center justify-between mb-10">
             <h2 className="text-sm font-black uppercase tracking-[0.3em] text-gray-400 font-['Outfit']">{t('pages.overview.zonePressure')}</h2>
             <TrendingUp className="h-4 w-4 text-brand-primary opacity-30" />
          </div>
          <div className="space-y-6">
            {health.length === 0 ? (
               <div className="py-12 text-center text-xs text-brand-primary italic opacity-60 animate-pulse">{t('pages.overview.connecting')}</div>
            ) : health.map((item) => (
              <div key={item.zone_id} className="group flex items-center justify-between p-6 rounded-2xl bg-gray-50 dark:bg-gray-900/50 border border-gray-100 dark:border-gray-800 hover:border-brand-primary/20 transition-all cursor-pointer">
                <div>
                   <div className="text-base font-black text-gray-900 dark:text-white font-['Outfit'] tracking-tight">Zone {item.zone_id}</div>
                   <div className="text-[10px] text-gray-400 font-black uppercase tracking-widest mt-1 font-['Outfit']">
                      NODE: {Math.round(item.current_orders)} / BASELINE: {Math.round(item.baseline_orders)}
                   </div>
                </div>
                <div className="flex items-center gap-6">
                   <div className="text-right">
                      <div className={`text-[10px] font-black uppercase tracking-widest font-['Outfit'] px-3 py-1 rounded-full ${
                        item.status === 'disrupted' 
                          ? 'bg-brand-primary text-white shadow-lg shadow-brand-primary/20' 
                          : item.status === 'healthy' 
                            ? 'bg-emerald-50 text-emerald-600' 
                            : 'bg-brand-primary/10 text-brand-primary'
                      }`}>
                         {item.status.replace('_', ' ')}
                      </div>
                       <div className="text-[10px] font-bold text-gray-400 mt-2 font-['Outfit']">{t('pages.overview.orderDrop')}: <span className="text-gray-900 dark:text-gray-100">{Math.round((item.order_drop || 0) * 100)}%</span></div>
                   </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="enterprise-panel p-10">
           <div className="flex items-center justify-between mb-10">
                <h2 className="text-sm font-black uppercase tracking-[0.3em] text-gray-400 font-['Outfit']">{t('pages.overview.automationOutcome')}</h2>
              <ShieldCheck className="h-4 w-4 text-brand-primary opacity-30" />
           </div>
           
           <div className="grid grid-cols-2 gap-6">
              <div className="p-8 rounded-3xl bg-brand-soft border border-brand-primary/10 shadow-lg shadow-brand-primary/5">
                 <div className="text-[10px] font-black uppercase tracking-[0.2em] text-brand-dark mb-3 font-['Outfit']">Payouts Verified</div>
                 <div className="text-4xl font-black text-black font-['Outfit'] leading-none">₹{Math.round(totals.paid).toLocaleString()}</div>
                 <div className="mt-4 text-[9px] font-bold text-brand-primary uppercase tracking-[0.1em] font-['Outfit']">Parametric Integrity ✓</div>
              </div>
              <div className="p-8 rounded-3xl bg-gray-900 border border-gray-800 shadow-xl shadow-gray-900/10">
                 <div className="text-[10px] font-black uppercase tracking-[0.2em] text-gray-400 mb-3 font-['Outfit']">Signal Flow Rate</div>
                 <div className="text-4xl font-black text-white font-['Outfit'] leading-none">{disruptions.length}</div>
                 <div className="mt-4 text-[9px] font-bold text-emerald-400 uppercase tracking-[0.1em] font-['Outfit']">Live Kafka Feed</div>
              </div>
           </div>

           <div className="mt-12 pt-10 border-t border-gray-100 dark:border-gray-800">
              <div className="space-y-4">
                 <div className="flex justify-between items-center text-[10px] font-black uppercase tracking-widest font-['Outfit']">
                    <span className="text-gray-400">Health Coverage Stability</span>
                    <span className="text-gray-900 dark:text-white">{totals.coverage}%</span>
                 </div>
                 <div className="w-full h-2 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
                    <div className="h-full transition-all duration-1000 bg-brand-primary rounded-full shadow-[0_0_12px_rgba(236,72,153,0.4)]" style={{ width: `${totals.coverage}%` }}></div>
                 </div>
                 <p className="text-[10px] font-medium text-gray-400 leading-relaxed italic">
                    Platform integrity is maintaining a {totals.coverage}% healthy node threshold across all verified zones.
                 </p>
              </div>
           </div>
        </div>
      </div>
    </div>
  )
}

function MetricCard({ label, value, icon: Icon, trend, color }: { label: string; value: string | number; icon: any; trend: string; color: string }) {
  const isPositive = trend.startsWith('+') || trend === 'Stable' || trend === 'None'
  return (
    <div className="enterprise-panel p-8 flex flex-col justify-between min-h-[160px] group hover:border-brand-primary/30 transition-all transform hover:-translate-y-1">
      <div className="flex justify-between items-start">
        <div className="h-12 w-12 flex items-center justify-center rounded-2xl bg-gray-50 dark:bg-gray-900 border border-gray-100 dark:border-gray-800 group-hover:bg-brand-soft transition-colors">
          <Icon className="h-5 w-5 text-gray-500 group-hover:text-brand-primary transition-colors" />
        </div>
        <div className={`flex items-center gap-1.5 text-[9px] font-black uppercase tracking-widest px-3 py-1 bg-white dark:bg-gray-900 rounded-full border border-gray-100 dark:border-gray-800 ${color}`}>
           {trend}
           {trend.includes('%') && <ArrowUpRight className="h-3 w-3" />}
        </div>
      </div>
      <div>
        <div className="text-4xl font-black text-gray-900 dark:text-white mt-6 font-['Outfit'] leading-none tracking-tight">{value}</div>
        <div className="text-[10px] font-black uppercase tracking-[0.3em] text-gray-400 mt-2 font-['Outfit'] group-hover:text-brand-primary transition-colors">{label}</div>
      </div>
    </div>
  )
}
