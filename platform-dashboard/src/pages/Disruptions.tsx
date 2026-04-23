import { useEffect, useState, useRef } from 'react'
import { Activity, AlertTriangle, Radio, ShieldAlert, WifiOff, CloudRain, Zap, RefreshCw, Terminal, PlayCircle, Settings2 } from 'lucide-react'
import { getZoneHealth, getDisruptions, postTriggerDemo, postExternalSignal } from '../api/platform'
import { getZones } from '../api/zones'
import { useLocalization } from '../context/LocalizationContext'

interface ZoneHealth {
  zone_id: number
  status: string
  order_drop: number
  current_orders: number
  baseline_orders: number
  active_signals: Record<string, boolean>
}

interface Disruption {
  disruption_id: string
  zone_id: string
  type: string
  severity: string
  confidence: number
  status: string
  claims_generated: number
  claims_in_review: number
  payouts_processed: number
  payout_amount_total: number
  automation_status: string
  signals: { source: string; value: number }[]
  started_at: string
}

interface HistoryEvent {
  zone: number
  drop: number
  signals: string[]
  severity: string
  time: string
}

export default function Disruptions() {
  const { t } = useLocalization()
  const [healths, setHealths] = useState<ZoneHealth[]>([])
  const [zones, setZones] = useState<any[]>([])
  const [selectedZoneId, setSelectedZoneId] = useState<number | null>(null)
  const [disruptions, setDisruptions] = useState<Disruption[]>([])
  const [history, setHistory] = useState<HistoryEvent[]>([])
  const lastSignatureRef = useRef<string>("")
  const zonesInitializedRef = useRef(false)
  const [loadingAction, setLoadingAction] = useState(false)
  const [actionStatus, setActionStatus] = useState("")
  const [delaying, setDelaying] = useState<string | null>(null)
  const [fetchLatencyMs, setFetchLatencyMs] = useState<number | null>(null)
  const [showFlash, setShowFlash] = useState(false)

  const triggerFlash = () => {
    setShowFlash(true)
    setTimeout(() => setShowFlash(false), 2000)
  }

  const fetchData = async () => {
    const startedAt = performance.now()
    try {
      const [hRes, dRes, zRes] = await Promise.all([getZoneHealth(), getDisruptions(), getZones()])
      setHealths(hRes.data.data)
      setDisruptions(dRes.data.data)
      setZones(zRes.data?.zones ?? [])
      
      // Only auto-select first zone on initial load, then preserve user selection
      if (!zonesInitializedRef.current && zRes.data?.zones?.length > 0) {
        const firstZoneId = zRes.data.zones[0]?.zone_id
        if (firstZoneId) {
          setSelectedZoneId(firstZoneId)
          zonesInitializedRef.current = true
        }
      }
    } catch (e) {
      console.error('Failed to fetch platform status', e)
    } finally {
      setFetchLatencyMs(Math.round(performance.now() - startedAt))
    }
  }

  useEffect(() => {
    fetchData()
    const int = setInterval(fetchData, 1000)
    return () => clearInterval(int)
  }, [])

  useEffect(() => {
    healths.forEach((z) => {
      const activeSignalKeys = Object.entries(z.active_signals)
        .filter(([_, v]) => v === true)
        .map(([k]) => k)
        .sort();
      const signalsKey = activeSignalKeys.join("-");
      
      const dropPctNum = z.order_drop * 100;
      let severityBucket = "LOW";
      if (dropPctNum > 70) severityBucket = "HIGH";
      else if (dropPctNum > 40) severityBucket = "MEDIUM";
      
      const currentSignature = `${z.zone_id}-${signalsKey}`;
      
      if (z.status === 'disrupted' && activeSignalKeys.length > 0 && currentSignature !== lastSignatureRef.current) {
        const newEvent: HistoryEvent = {
          zone: z.zone_id,
          drop: Math.round(dropPctNum),
          signals: activeSignalKeys,
          severity: severityBucket,
          time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
        }
        
        setHistory(prev => [newEvent, ...prev].slice(0, 5));
        lastSignatureRef.current = currentSignature;
      }
    })
  }, [healths])

  const formatSignal = (s: string) => 
    s.replace("_", " ").toLowerCase().replace(/\b\w/g, c => c.toUpperCase())

  const formatZoneLabel = (zone: any) => {
    return String(zone?.zone_name || zone?.name || `Zone ${zone?.zone_id ?? ''}`).trim()
  }

  const handleExternalSignal = async (zone_id: number, type: string, active: boolean) => {
    if (loadingAction || delaying) return;
    setLoadingAction(true)
    setDelaying(type)
    
    setActionStatus(active ? `Injecting ${type.toUpperCase()} signal...` : "Clearing signal state...")
    await new Promise(r => setTimeout(r, 600))

    try {
      await postExternalSignal({
        zone_id,
        source: type,
        status: active ? "active" : "resolved"
      })
      if (active) triggerFlash()
      await fetchData()
    } finally {
      setDelaying(null)
      setActionStatus("")
      setLoadingAction(false)
    }
  }

  const handleOrderDrop = async (zone_id: number) => {
    if (loadingAction) return
    setLoadingAction(true)
    setActionStatus('Simulating demand collapse...')
    try {
      await postTriggerDemo({ zone_id, force_order_drop: true, external_signal: '' })
      triggerFlash()
      await new Promise((resolve) => setTimeout(resolve, 700))
      await fetchData()
    } finally {
      setActionStatus('')
      setLoadingAction(false)
    }
  }

  const handleReset = async (zone_id: number) => {
    setLoadingAction(true)
    setActionStatus("Resetting system state...")
    try {
      await postTriggerDemo({ zone_id, force_order_drop: false, external_signal: "" }) 
      setHistory([])
      lastSignatureRef.current = ""
      await new Promise(r => setTimeout(r, 600))
      await fetchData()
    } finally {
      setActionStatus("")
      setLoadingAction(false)
    }
  }

  return (
    <div className="space-y-10 font-['Outfit']">
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-2xl font-black tracking-tight text-slate-900 dark:text-white">{t('pages.disruptions.title')}</h1>
          <p className="mt-1 text-sm text-slate-500">{t('pages.disruptions.description')}</p>
        </div>
        <div className="flex gap-4">
           <div className="enterprise-panel px-4 py-2 flex flex-col items-center justify-center">
              <div className="text-[9px] font-black uppercase tracking-widest text-slate-400">Engine Latency</div>
              <div className="text-sm font-black text-emerald-600">{fetchLatencyMs === null ? '...' : `${fetchLatencyMs}ms`}</div>
           </div>
        </div>
      </div>

      <div className="grid grid-cols-12 gap-8">
        <div className="col-span-12 lg:col-span-8 space-y-8">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <StatsCard label={t('pages.analytics.disruptions')} value={disruptions.length} color="text-slate-900 dark:text-white" />
            <StatsCard label={t('pages.analytics.claimsGenerated')} value={disruptions.reduce((sum, item) => sum + item.claims_generated, 0)} color="text-brand-primary" />
            <StatsCard label={t('pages.workers.title')} value={disruptions.reduce((sum, item) => sum + item.payouts_processed, 0)} color="text-emerald-600" />
            <StatsCard label={t('pages.analytics.payoutAmount')} value={`Rs ${Math.round(disruptions.reduce((sum, item) => sum + item.payout_amount_total, 0))}`} color="text-brand-primary" />
          </div>

          <div className="enterprise-panel p-8">
            <h2 className="text-sm font-black text-slate-900 dark:text-white mb-8 flex items-center gap-2 uppercase tracking-widest">
              <Radio className="w-4 h-4 text-orange-500 animate-pulse" />
              Real-time Zone Telemetry
            </h2>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              {healths.map((z) => {
                const status = z.status.toUpperCase().replace('_', ' ')
                const drop = Math.round(z.order_drop * 100)
                const isDisrupted = z.status === 'disrupted'
                
                return (
                  <div key={z.zone_id} className={`p-8 rounded-3xl border-2 transition-all ${
                    isDisrupted 
                      ? 'bg-white dark:bg-gray-950 border-brand-primary shadow-xl shadow-brand-primary/10' 
                      : 'bg-gray-50/50 dark:bg-gray-900/30 border-gray-100 dark:border-gray-800'
                  }`}>
                    <div className="flex justify-between items-start mb-8">
                       <div className="flex items-center gap-3">
                          <div className={`h-2.5 w-2.5 rounded-full ${isDisrupted ? 'bg-brand-primary animate-pulse' : 'bg-emerald-500'}`}></div>
                          <span className="text-sm font-black tracking-tighter text-gray-900 dark:text-white uppercase italic">Zone {z.zone_id}</span>
                       </div>
                       <div className={`text-[10px] font-black px-4 py-1.5 rounded-full border-2 uppercase tracking-widest ${
                         isDisrupted ? 'bg-brand-primary text-white border-brand-primary' : 'bg-white dark:bg-gray-950 text-gray-400 border-gray-100 dark:border-gray-800'
                       }`}>
                          {status}
                       </div>
                    </div>

                    <div className="grid grid-cols-2 gap-8 mb-8">
                       <div>
                          <div className="text-3xl font-black text-gray-900 dark:text-white leading-none">{z.current_orders}</div>
                          <div className="text-[10px] text-gray-400 font-black uppercase tracking-[0.2em] mt-3">Live Volume</div>
                       </div>
                       <div>
                          <div className={`text-3xl font-black leading-none ${drop >= 30 ? 'text-brand-primary italic' : 'text-emerald-500'}`}>{drop}%</div>
                          <div className="text-[10px] text-gray-400 font-black uppercase tracking-[0.2em] mt-3">Delta Gap</div>
                       </div>
                    </div>

                    <div className="pt-8 border-t border-gray-100 dark:border-gray-800 flex flex-wrap gap-2">
                       {Object.entries(z.active_signals).filter(([_, v]) => v).map(([sig]) => (
                         <div key={sig} className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-brand-soft border border-brand-primary/10 text-brand-primary text-[10px] font-black uppercase tracking-widest">
                            <Zap className="h-3.5 w-3.5" /> {sig.split('_')[1] || sig}
                         </div>
                       ))}
                       {isDisrupted && (
                         <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-gray-900 text-white text-[10px] font-black uppercase tracking-widest">
                            <Activity className="h-3.5 w-3.5" /> VERIFYING LOSS
                         </div>
                       )}
                    </div>
                  </div>
                )
              })}
            </div>
          </div>

          <div className="enterprise-panel p-8">
            <h2 className="text-sm font-black text-slate-900 dark:text-white mb-8 flex items-center gap-2 uppercase tracking-widest">
              Outcome Timeline
            </h2>
            <div className="space-y-4">
              {history.length === 0 ? (
                <div className="text-center py-10 text-xs text-slate-400 italic">No events recorded in current cycle.</div>
              ) : history.map((item, i) => (
                <div key={i} className="flex items-center justify-between p-4 rounded-xl bg-slate-50 dark:bg-slate-800/40 border border-slate-100 dark:border-slate-800/60">
                   <div className="flex items-center gap-4">
                      <div className={`h-8 w-8 rounded flex items-center justify-center ${item.severity === 'HIGH' ? 'bg-brand-primary text-white' : 'bg-brand-primary/80 text-white'}`}>
                         <ShieldAlert className="h-4 w-4" />
                      </div>
                      <div>
                         <div className="text-xs font-bold text-slate-900 dark:text-white">Zone {item.zone} — Disruption Locked</div>
                         <div className="text-[10px] text-slate-500 mt-0.5 uppercase tracking-tighter">
                            Triggered by: {item.signals.map(s => s.split('_')[1] || s).join(' + ')}
                         </div>
                      </div>
                   </div>
                   <div className="text-right">
                      <div className="text-xs font-black text-rose-500">{item.drop}% DROP</div>
                      <div className="text-[9px] text-slate-400 font-bold uppercase tracking-widest mt-0.5">{item.time}</div>
                   </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="col-span-12 lg:col-span-4 space-y-8">
          <div className="enterprise-panel p-8 sticky top-6">
            <h2 className="text-sm font-black text-slate-900 dark:text-white mb-8 uppercase tracking-widest flex items-center gap-2">
               <Terminal className="h-4 w-4 text-orange-500" />
               Chaos Control Panel
            </h2>
            
            <div className="space-y-10">
              <section className="space-y-4">
                <div className="text-[10px] font-black text-slate-400 uppercase tracking-widest">Statistical Ingress</div>
                <div className="space-y-3">
                  <label className="block text-[10px] font-black uppercase tracking-widest text-slate-500">Zone</label>
                  <select
                    value={selectedZoneId ? String(selectedZoneId) : ''}
                    onChange={(e) => setSelectedZoneId(e.target.value ? Number(e.target.value) : null)}
                    className="w-full rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 px-3 py-3 text-sm text-slate-900 dark:text-white"
                  >
                    <option value="">Select Zone</option>
                    {zones.map((zone) => (
                      <option key={zone.zone_id} value={String(zone.zone_id)}>
                        {formatZoneLabel(zone)}
                      </option>
                    ))}
                  </select>
                </div>
                 <button
                  disabled={loadingAction || !selectedZoneId}
                  onClick={() => selectedZoneId && handleOrderDrop(selectedZoneId)}
                  className="w-full flex items-center justify-between p-4 rounded-xl border border-brand-primary/20 bg-brand-soft hover:bg-brand-primary/10 transition-all text-left group"
                >
                  <div className="max-w-[180px]">
                     <div className="text-[11px] font-black uppercase text-brand-primary">Collapse Demand</div>
                     <p className="text-[10px] text-slate-500 dark:text-slate-400 leading-tight mt-1">Force order drop in the selected zone to test engine anomaly detection.</p>
                  </div>
                  <PlayCircle className="h-5 w-5 text-brand-primary group-hover:scale-110 transition-transform" />
                </button>
              </section>

              <section className="space-y-4 pt-8 border-t border-slate-100 dark:border-slate-800">
                <div className="text-[10px] font-black text-slate-400 uppercase tracking-widest">External Signals</div>
                <div className="grid grid-cols-2 gap-3">
                  <SignalButton label="Rain" icon={CloudRain} active={delaying === 'weather_rain'} onClick={() => selectedZoneId && handleExternalSignal(selectedZoneId, 'weather_rain', true)} />
                  <SignalButton label="Curfew" icon={ShieldAlert} active={delaying === 'zone_curfew'} onClick={() => selectedZoneId && handleExternalSignal(selectedZoneId, 'zone_curfew', true)} />
                  <SignalButton label="AQI" icon={Radio} active={delaying === 'aqi_hazardous'} onClick={() => selectedZoneId && handleExternalSignal(selectedZoneId, 'aqi_hazardous', true)} />
                  <button 
                    disabled={loadingAction || !selectedZoneId}
                    onClick={() => selectedZoneId && handleExternalSignal(selectedZoneId, 'all_signals', false)}
                    className="h-14 flex items-center justify-center rounded-xl bg-slate-100 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700 text-slate-500 hover:text-slate-900 dark:hover:text-white transition-none">
                     <RefreshCw className="h-4 w-4" />
                  </button>
                </div>
              </section>

              <button
                disabled={loadingAction || !selectedZoneId}
                onClick={() => selectedZoneId && handleReset(selectedZoneId)}
                className="w-full py-4 rounded-xl border-2 border-slate-100 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-all flex items-center justify-center gap-2 group"
              >
                <RefreshCw className={`h-4 w-4 text-slate-400 ${loadingAction ? 'animate-spin' : ''}`} />
                <span className="text-[11px] font-black uppercase tracking-[0.2em] text-slate-600 dark:text-slate-300">System Reset</span>
              </button>

              {actionStatus && (
                <div className="px-4 py-2 rounded bg-slate-900 border border-slate-700 font-mono text-[9px] text-emerald-500 animate-pulse">
                  &gt; {actionStatus}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Chaos Flash Overlay */}
      {showFlash && (
        <div className="fixed inset-0 z-[100] pointer-events-none flex items-center justify-center">
          <div className="absolute inset-0 bg-brand-primary/10 animate-pulse transition-opacity duration-1000"></div>
          <div className="relative glass p-10 rounded-full border-4 border-brand-primary shadow-2xl animate-bounce">
             <Zap className="h-20 w-20 text-brand-primary fill-brand-primary" />
          </div>
        </div>
      )}
    </div>
  )
}

function StatsCard({ label, value, color }: { label: string, value: string | number, color: string }) {
  return (
    <div className="enterprise-panel p-6 flex flex-col justify-center">
       <div className="text-[9px] font-black uppercase tracking-[0.2em] text-slate-400 mb-1">{label}</div>
       <div className={`text-2xl font-black ${color}`}>{value}</div>
    </div>
  )
}

function SignalButton({ label, icon: Icon, active, onClick }: { label: string, icon: any, active: boolean, onClick: () => void }) {
  return (
    <button
      disabled={active}
      onClick={onClick}
      className={`h-14 flex flex-col items-center justify-center rounded-xl border transition-all ${
        active 
          ? 'bg-brand-primary border-brand-primary text-white animate-pulse' 
          : 'bg-white dark:bg-slate-900 border-slate-200 dark:border-slate-800 text-slate-400 hover:border-brand-primary/50 hover:text-brand-primary'
      }`}
    >
      <Icon className="h-4 w-4 mb-1" />
      <span className="text-[9px] font-black uppercase tracking-widest">{label}</span>
    </button>
  )
}
