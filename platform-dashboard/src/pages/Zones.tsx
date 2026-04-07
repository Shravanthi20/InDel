import { useEffect, useState, useMemo } from 'react'
import { getZoneHealth, getZones } from '../api/platform'
import { Activity, ShieldAlert, BarChart, MapPin, Wind, CloudRain, Zap, Search, Filter } from 'lucide-react'

export default function Zones() {
  const [zones, setZones] = useState<any[]>([])
  const [health, setHealth] = useState<any[]>([])
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<'all' | 'healthy' | 'disrupted' | 'anomalous'>('all')

  useEffect(() => {
    async function load() {
      const [zonesRes, healthRes] = await Promise.all([getZones(), getZoneHealth()])
      setZones(zonesRes.data.zones ?? [])
      setHealth(healthRes.data.data ?? [])
    }

    load().catch((error) => console.error('Failed to load zones', error))
    const timer = setInterval(() => load().catch(() => undefined), 5000)
    return () => clearInterval(timer)
  }, [])

  const filteredZones = useMemo(() => {
    return zones.filter(zone => {
      const zoneHealth = health.find((item) => item.zone_id === zone.zone_id)
      const status = zoneHealth?.status || 'healthy'
      
      const matchesSearch = 
        zone.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        zone.city.toLowerCase().includes(searchQuery.toLowerCase()) ||
        zone.state.toLowerCase().includes(searchQuery.toLowerCase())
      
      const matchesStatus = 
        statusFilter === 'all' || 
        (statusFilter === 'healthy' && status === 'healthy') || 
        (statusFilter === 'disrupted' && status === 'disrupted') ||
        (statusFilter === 'anomalous' && (status === 'anomalous_demand' || status === 'monitoring'))

      return matchesSearch && matchesStatus
    })
  }, [zones, health, searchQuery, statusFilter])

  return (
    <div className="space-y-10">
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-2xl font-black tracking-tight text-slate-900 dark:text-white">Zone Monitoring</h1>
          <p className="mt-1 text-sm text-slate-500">Live regional telemetry and risk assessment for automated insurance payout triggers.</p>
        </div>
        <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-50 dark:bg-emerald-500/10 border border-emerald-100 dark:border-emerald-500/20">
           <div className="h-1.5 w-1.5 rounded-full bg-emerald-500"></div>
           <span className="text-[9px] font-black uppercase tracking-widest text-emerald-600 dark:text-emerald-400">All Nodes Active</span>
        </div>
      </div>

      <div className="enterprise-panel p-4 flex items-center justify-between mb-8">
         <div className="relative group w-72">
            <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-slate-400">
              <Search className="h-3.5 w-3.5" />
            </div>
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Filter by name, city or state..."
              className="w-full rounded border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 py-1.5 pl-9 pr-3 text-[11px] text-slate-900 dark:text-white outline-none focus:border-orange-500 transition-none"
            />
         </div>
         <div className="flex items-center gap-2">
            <span className="text-[10px] font-black uppercase text-slate-400 mr-2">Health Filter:</span>
            {(['all', 'healthy', 'anomalous', 'disrupted'] as const).map((s) => (
              <button
                key={s}
                onClick={() => setStatusFilter(s)}
                className={`px-3 py-1.5 rounded border text-[10px] font-bold transition-none uppercase ${
                  statusFilter === s 
                    ? 'bg-orange-600 border-orange-600 text-white shadow-sm' 
                    : 'border-slate-200 dark:border-slate-700 text-slate-500 hover:text-slate-900 dark:hover:text-white'
                }`}
              >
                {s}
              </button>
            ))}
         </div>
      </div>

      <div className="grid gap-6 md:grid-cols-2 xl:grid-cols-3">
        {filteredZones.length === 0 ? (
          <div className="col-span-full py-20 text-center text-slate-400 text-xs italic border-2 border-dashed border-slate-100 dark:border-slate-800 rounded-2xl">
             No zones match the current monitoring parameters.
          </div>
        ) : filteredZones.map((zone) => {
          const zoneHealth = health.find((item) => item.zone_id === zone.zone_id)
          const isDisrupted = zoneHealth?.status === 'disrupted'
          const isAnomalous = zoneHealth?.status === 'anomalous_demand' || zoneHealth?.status === 'monitoring'
          const drop = Math.round((zoneHealth?.order_drop ?? 0) * 100)
          
          return (
            <div key={zone.zone_id} className={`enterprise-panel relative group overflow-hidden transition-all duration-300 ${isDisrupted ? 'border-rose-500/40 shadow-[0_0_20px_rgba(244,63,94,0.1)]' : isAnomalous ? 'border-amber-500/40 shadow-[0_0_15px_rgba(245,158,11,0.05)]' : ''}`}>
              <div className="p-8">
                <div className="flex items-start justify-between mb-8">
                  <div className="max-w-[180px]">
                    <div className="flex items-center gap-2">
                       <MapPin className="h-4 w-4 text-slate-400" />
                       <h2 className="text-sm font-black text-slate-900 dark:text-white uppercase tracking-tight truncate">{zone.name}, {zone.city}</h2>
                    </div>
                    <p className="text-[10px] text-slate-400 font-bold uppercase tracking-widest mt-1.5 truncate">{zone.state} • Risk {zone.risk_rating}</p>
                  </div>
                  <div className={`px-2 py-1 rounded-sm text-[9px] font-black uppercase tracking-widest border transition-all ${
                    isDisrupted 
                      ? 'bg-rose-500/10 border-rose-500/40 text-rose-500 shadow-[0_0_10px_rgba(244,63,94,0.2)]' 
                      : isAnomalous
                      ? 'bg-amber-500/10 border-amber-500/40 text-amber-600 dark:text-amber-400'
                      : 'bg-emerald-50/50 dark:bg-emerald-500/5 border-emerald-100 dark:border-emerald-500/20 text-emerald-600 dark:text-emerald-400'
                  }`}>
                    {zoneHealth?.status?.replace('_', ' ') ?? 'healthy'}
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-8">
                   <div>
                      <div className="text-2xl font-black text-slate-900 dark:text-white tracking-tight">{zone.active_workers}</div>
                      <div className="text-[9px] text-slate-400 font-bold uppercase tracking-widest mt-1">Live Workers</div>
                   </div>
                   <div>
                      <div className={`text-2xl font-black tracking-tight ${drop >= 30 ? 'text-rose-500' : 'text-emerald-500'}`}>{drop}%</div>
                      <div className="text-[9px] text-slate-400 font-bold uppercase tracking-widest mt-1">Volume Drop</div>
                   </div>
                </div>

                <div className="mt-8 pt-8 border-t border-slate-100 dark:border-slate-800">
                   <div className="flex items-center justify-between mb-4">
                      <span className="text-[10px] font-black uppercase tracking-widest text-slate-400">Regional Signals</span>
                      <Activity className={`h-3.5 w-3.5 transition-colors ${isDisrupted ? 'text-rose-500 animate-pulse' : 'text-slate-300'}`} />
                   </div>
                   <div className="flex flex-wrap gap-2">
                      <SignalBadge icon={Wind} label="AQI" active={zoneHealth?.active_signals?.aqi_hazardous} />
                      <SignalBadge icon={CloudRain} label="RAIN" active={zoneHealth?.active_signals?.weather_rain} />
                      <SignalBadge icon={ShieldAlert} label="CURFEW" active={zoneHealth?.active_signals?.zone_curfew} />
                      <SignalBadge icon={Zap} label="DEMAND" active={drop >= 30} />
                   </div>
                </div>
              </div>
              
              <div className="h-1 w-full bg-slate-100 dark:bg-slate-800">
                 <div className={`h-full transition-all duration-1000 ${isDisrupted ? 'bg-rose-500 w-full' : isAnomalous ? 'bg-amber-400 w-2/3' : 'bg-emerald-400 w-full opacity-30'}`}></div>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

function SignalBadge({ icon: Icon, label, active }: { icon: any, label: string, active?: boolean }) {
  return (
    <div className={`flex items-center gap-1.5 px-3 py-1.5 rounded border transition-all ${
      active 
        ? 'bg-orange-500/10 border-orange-500/40 text-orange-600 dark:text-orange-400' 
        : 'bg-slate-50 dark:bg-slate-800/40 border-slate-100 dark:border-slate-800 text-slate-400'
    }`}>
       <Icon className="h-3 w-3" />
       <span className="text-[9px] font-black tracking-widest">{label}</span>
    </div>
  )
}
