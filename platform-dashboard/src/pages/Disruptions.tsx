import { useEffect, useState, useRef } from 'react'
import { Activity, AlertTriangle, Radio, ShieldAlert, WifiOff, CloudRain, Zap, RefreshCw } from 'lucide-react'
import { getZoneHealth, getDisruptions, postTriggerDemo, postExternalSignal } from '../api/platform'

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
  const [healths, setHealths] = useState<ZoneHealth[]>([])
  const [disruptions, setDisruptions] = useState<Disruption[]>([])
  const [history, setHistory] = useState<HistoryEvent[]>([])
  const lastSignatureRef = useRef<string>("")
  const [loadingAction, setLoadingAction] = useState(false)
  const [actionStatus, setActionStatus] = useState("")
  const [delaying, setDelaying] = useState<string | null>(null)

  const fetchData = async () => {
    try {
      const [hRes] = await Promise.all([getZoneHealth()])
      setHealths(hRes.data.data)
    } catch (e) {
      console.error('Failed to fetch platform status', e)
    }
  }

  useEffect(() => {
    fetchData()
    // Poll every 1 second for snappy demo feel
    const int = setInterval(fetchData, 1000)
    return () => clearInterval(int)
  }, [])

  // DISRUPTION "MEMORY" ENGINE (Fixed with Synchronous Ref to prevent race conditions during polling)
  useEffect(() => {
    healths.forEach((z) => {
      // 1. Precise Signal Extraction (True values only, sorted)
      const activeSignalKeys = Object.entries(z.active_signals)
        .filter(([_, v]) => v === true)
        .map(([k]) => k)
        .sort();
      const signalsKey = activeSignalKeys.join("-");
      
      // 2. Severity Bucketing (Logic-driven)
      const dropPctNum = z.order_drop * 100;
      let severityBucket = "LOW";
      if (dropPctNum > 70) severityBucket = "HIGH";
      else if (dropPctNum > 40) severityBucket = "MEDIUM";
      
      // 3. Construct Context-Only Signature (Omits severity for maximum stability)
      const currentSignature = `${z.zone_id}-${signalsKey}`;
      
      // 4. Detection: Capture ONLY on meaningful signal context transition
      // We use useRef for immediate synchronous comparison to prevent duplicates during fast polling
      if (z.status === 'disrupted' && activeSignalKeys.length > 0 && currentSignature !== lastSignatureRef.current) {
        const newEvent: HistoryEvent = {
          zone: z.zone_id,
          drop: Math.round(dropPctNum),
          signals: activeSignalKeys,
          severity: severityBucket,
          time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
        }
        
        setHistory(prev => [newEvent, ...prev].slice(0, 5));
        lastSignatureRef.current = currentSignature; // 🔥 INSTANT SYNCHRONOUS UPDATE
      }
    })
  }, [healths])

  // Formatting Helper: professional Title Case (e.g., zone_curfew -> Zone Curfew)
  const formatSignal = (s: string) => 
    s.replace("_", " ").toLowerCase().replace(/\b\w/g, c => c.toUpperCase())

  const handleExternalSignal = async (zone_id: number, type: string, active: boolean) => {
    if (loadingAction || delaying) return;
    setLoadingAction(true)
    setDelaying(type)
    
    if (active) {
      setActionStatus(`Processing ${type.replace('_', ' ').toUpperCase()} logic...`)
      await new Promise(r => setTimeout(r, 1000))
    } else {
      setActionStatus("Clearing registered signals...")
      await new Promise(r => setTimeout(r, 800))
    }

    try {
      await postExternalSignal({
        zone_id,
        source: type,
        status: active ? "active" : "resolved"
      })
      await fetchData()
    } finally {
      setDelaying(null)
      setActionStatus("")
      setLoadingAction(false)
    }
  }

  const handleReset = async (zone_id: number) => {
    setLoadingAction(true)
    setActionStatus("Resetting engine & syncing simulator...")
    try {
      await postTriggerDemo({ zone_id, force_order_drop: false, external_signal: "" }) 
      setHistory([])         // Clear History list
      lastSignatureRef.current = "" // 🔥 Clear signature memory synchronously
      await new Promise(r => setTimeout(r, 1000))
      await fetchData()
    } finally {
      setActionStatus("")
      setLoadingAction(false)
    }
  }

  return (
    <div className="p-8 bg-slate-950 min-h-screen text-slate-300 font-sans">
      <div className="mb-8 border-b border-white/10 pb-6 flex justify-between items-end">
        <div>
          <h1 className="text-4xl font-black text-white tracking-tight flex items-center gap-3">
            <Activity className="w-10 h-10 text-cyan-400" />
            Platform Disruption Engine
          </h1>
          <p className="text-slate-400 mt-2 text-lg font-medium">
            <span className="text-emerald-400/80">● Live Ingestion</span> — Detecting anomalies from streaming platform events.
          </p>
        </div>
        <div className="flex gap-4">
           <div className="px-4 py-2 bg-slate-900 border border-white/10 rounded-lg text-xs font-mono">
              <div className="text-slate-500 uppercase mb-1">Engine Latency</div>
              <div className="text-cyan-400 font-bold">~42ms</div>
           </div>
        </div>
      </div>

      <div className="grid grid-cols-12 gap-8">
        
        {/* TELEMETRY CARDS */}
        <div className="col-span-12 lg:col-span-8 space-y-6">
          <div className="bg-slate-900/50 border border-white/5 rounded-2xl p-6 shadow-2xl backdrop-blur-xl">
            <h2 className="text-xl font-bold text-white mb-4 flex items-center gap-2">
              <Radio className="w-5 h-5 text-emerald-400 animate-pulse" />
              Real-time Zone Telemetry
            </h2>
            
            <div className="mb-8 p-4 rounded-xl border border-white/5 bg-slate-900/40 text-xs flex gap-6 text-slate-400 leading-snug shadow-inner">
               <div className="max-w-[200px]">
                  <strong className="text-emerald-400 text-[10px] uppercase font-black tracking-widest block mb-1">HEALTHY</strong>
                  <span className="opacity-80">No disruption: baseline conditions stable.</span>
               </div>
               <div className="max-w-[200px] border-l border-white/5 pl-6">
                  <strong className="text-cyan-400 text-[10px] uppercase font-black tracking-widest block mb-1">MONITORING</strong>
                  <span className="opacity-80">External signal detected, no internal anomaly yet.</span>
               </div>
               <div className="max-w-[200px] border-l border-white/5 pl-6">
                  <strong className="text-amber-400 text-[10px] uppercase font-black tracking-widest block mb-1">ANOMALOUS</strong>
                  <span className="opacity-80">Demand drop detected, awaiting edge confirmation.</span>
               </div>
               <div className="max-w-[200px] border-l border-white/5 pl-6">
                  <strong className="text-rose-400 text-[10px] uppercase font-black tracking-widest block mb-1">DISRUPTED</strong>
                  <span className="text-white font-medium">Multi-signal validation complete → disruption confirmed.</span>
               </div>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {healths.map((z) => {
                const isDisrupted = z.status === 'disrupted'
                const isAnomolous = z.status === 'anomalous_demand'
                const clampedDrop = Math.max(0, Math.min(100, z.order_drop * 100))
                const isMonitoring = z.status === 'monitoring'
                
                return (
                  <div key={z.zone_id} className={`group relative overflow-hidden p-6 rounded-2xl border transition-all duration-700
                    ${isDisrupted ? 'bg-red-950/40 border-red-500/60 shadow-[0_0_30px_rgba(239,68,68,0.3)]' :
                      isAnomolous ? 'bg-amber-950/30 border-amber-500/50 shadow-[0_0_20px_rgba(245,158,11,0.1)]' : 
                      isMonitoring ? 'bg-cyan-950/30 border-cyan-500/50 shadow-[0_0_20px_rgba(6,182,212,0.1)]' :
                      'bg-slate-800/40 border-emerald-500/10 hover:border-emerald-500/30'}`}>
                    
                    {isDisrupted && <div className="absolute inset-0 bg-red-500/5 animate-pulse pointer-events-none" />}
                    
                    <div className="flex justify-between items-center mb-6">
                      <div className="flex flex-col gap-1">
                        <div className="flex items-center gap-3">
                           <div className={`w-2 h-2 rounded-full ${isDisrupted ? 'bg-red-500 shadow-[0_0_8px_#ef4444]' : isAnomolous ? 'bg-amber-500' : isMonitoring ? 'bg-cyan-500' : 'bg-emerald-500'} animate-ping`} />
                           <div className="font-mono text-sm font-bold text-white tracking-widest">ZONE {z.zone_id}</div>
                        </div>
                        <div className={`text-[10px] font-medium leading-tight mt-1 transition-all duration-500 ${isDisrupted ? 'text-red-400' : isAnomolous ? 'text-amber-400' : isMonitoring ? 'text-cyan-400' : 'text-slate-500'}`}>
                          {isDisrupted ? `Order volume dropped by ${clampedDrop.toFixed(0)}% + external signals detected → Disruption confirmed.` :
                           isAnomolous ? `Order volume dropped by ${clampedDrop.toFixed(0)}%. Awaiting external signal to confirm disruption.` :
                           isMonitoring ? `External signal received. Monitoring for volume impact...` :
                           'System stable. Order volume meeting expected baseline.'}
                        </div>
                      </div>
                      
                      {isDisrupted ? (
                        <div className="px-3 py-1 bg-red-500 text-white text-[10px] font-black rounded-full flex items-center gap-1 shadow-lg shadow-red-500/40 animate-pulse">
                          <ShieldAlert className="w-3 h-3" /> DISRUPTED
                        </div>
                      ) : isAnomolous ? (
                        <div className="px-3 py-1 bg-amber-500 text-black text-[10px] font-black rounded-full flex items-center gap-1">
                          <AlertTriangle className="w-3 h-3" /> ANOMALOUS
                        </div>
                      ) : isMonitoring ? (
                        <div className="px-3 py-1 bg-cyan-500 text-black text-[10px] font-black rounded-full flex items-center gap-1">
                          <Activity className="w-3 h-3" /> MONITORING
                        </div>
                      ) : (
                        <div className="px-3 py-1 bg-emerald-500/10 text-emerald-400 text-[10px] font-black rounded-full border border-emerald-500/20">
                          HEALTHY
                        </div>
                      )}
                    </div>
                    
                    <div className="grid grid-cols-2 gap-8 relative z-10">
                      <div>
                        <div className="text-3xl font-black text-white leading-none">
                           {z.current_orders}
                           <span className="text-xs text-slate-500 ml-1 font-medium italic">current / {z.baseline_orders.toFixed(0)} expected</span>
                        </div>
                        <div className="text-[10px] text-slate-500 mt-2 uppercase tracking-tighter font-bold">Live Volume (20s window)</div>
                      </div>
                      <div>
                        <div className={`text-3xl font-black leading-none ${clampedDrop >= 30 ? 'text-rose-400' : 'text-emerald-400'}`}>
                          {clampedDrop.toFixed(0)}%
                        </div>
                        <div className="text-[10px] text-slate-500 mt-2 uppercase tracking-tighter font-bold">Volume Drop %</div>
                      </div>
                    </div>

                    <div className="mt-6 pt-4 border-t border-white/5 space-y-2">
                       <div className="text-[10px] text-slate-600 font-bold uppercase tracking-widest">Active Inputs</div>
                       <div className="flex flex-wrap gap-2 min-h-[24px]">
                          <span className={`px-2 py-1 rounded text-[10px] font-bold border transition-colors ${clampedDrop >= 30 ? 'bg-amber-500/10 border-amber-500/20 text-amber-300' : 'bg-slate-800 border-white/5 text-slate-600'}`}>
                             DEMAND_DROP
                          </span>
                          {Object.keys(z.active_signals).map((sig) => (
                            <span key={sig} className={`px-2 py-1 text-[10px] font-bold rounded-sm flex items-center gap-1 border transition-colors ${isDisrupted ? 'bg-red-500/20 border-red-500/40 text-rose-300' : 'bg-slate-800 border-white/5 text-slate-400'}`}>
                               <Zap className="w-3 h-3" /> {formatSignal(sig)}
                            </span>
                          ))}
                       </div>
                    </div>

                    <div className="mt-5 pt-5 border-t border-white/5 bg-slate-900/40 -mx-6 -mb-6 p-6">
                      <div className="text-[10px] text-slate-500 font-bold uppercase tracking-widest mb-4">Event Logic Validation</div>
                      <div className="flex flex-col gap-3 text-[10px] uppercase font-black">
                         {/* Step 1 */}
                         <div className="flex items-center gap-3 text-emerald-400">
                           <span className="w-5 h-5 rounded-full bg-emerald-500/10 flex items-center justify-center border border-emerald-500/20">✓</span> 
                           <span>Order volume monitored</span>
                         </div>
                         {/* Step 2 */}
                         <div className={`flex items-center gap-3 transition-colors duration-300 ${clampedDrop >= 30 ? 'text-amber-400' : 'text-slate-600'}`}>
                           <span className={`w-5 h-5 rounded-full flex items-center justify-center border transition-all duration-300 ${clampedDrop >= 30 ? 'bg-amber-500/10 border-amber-500/30' : 'border-slate-700'}`}>
                               {clampedDrop >= 30 ? '✓' : ''}
                            </span> 
                           <span>{clampedDrop >= 30 ? 'Significant drop detected (>30%)' : 'Stable Platform Baseline'}</span>
                         </div>
                         {/* Step 3 */}
                         <div className={`flex items-center gap-3 transition-colors duration-300 ${isDisrupted ? 'text-rose-400' : (isMonitoring ? 'text-cyan-400' : 'text-slate-600')}`}>
                           <span className={`w-5 h-5 rounded-full flex items-center justify-center border transition-all duration-300 ${isDisrupted ? 'bg-rose-500/20 border-rose-500/40 shadow-[0_0_10px_rgba(244,63,94,0.2)]' : (isMonitoring ? 'bg-cyan-500/10 border-cyan-500/30' : 'border-slate-700')}`}>
                             {isDisrupted ? '!' : (isMonitoring ? '✓' : '')}
                           </span> 
                           <span>{isDisrupted ? 'Multi-signal validation complete' : (isMonitoring ? 'External signal received' : 'Awaiting external validation')}</span>
                         </div>
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>

          {/* CONFIRMED DISRUPTIONS TIMELINE */}
          <div className="bg-slate-900/50 border border-white/5 rounded-2xl p-6 shadow-2xl overflow-hidden">
            <h2 className="text-xl font-black text-white mb-6 uppercase tracking-tight flex items-center gap-2">
               Confirmed Disruptions Timeline
            </h2>
            <div className="space-y-4 max-h-[600px] overflow-y-auto pr-2 custom-scrollbar">
              {history.length === 0 ? (
                <div className="text-center py-12 border-2 border-dashed border-white/5 bg-slate-800/10 rounded-xl">
                   <p className="text-slate-600 italic font-medium">No disruptions recorded yet.</p>
                   <p className="text-[10px] text-slate-700 uppercase mt-2 tracking-widest font-black">Memory engine: Listening for events...</p>
                </div>
              ) : (
                history.map((item, i) => (
                  <div key={i} className="group p-5 bg-white/5 rounded-xl border border-white/5 hover:border-red-500/20 transition-all duration-500 animate-in fade-in slide-in-from-top-4">
                    <div className="flex justify-between items-start mb-3">
                      <div>
                        <div className="font-mono text-cyan-400 font-black text-sm uppercase tracking-widest">Zone {item.zone} — DISRUPTION CONFIRMED</div>
                        <div className="text-[9px] font-black text-slate-500 uppercase tracking-widest mt-1">
                           Signature: <span className="text-cyan-600">#{item.zone}-{item.signals.join('-').substr(0, 10)}</span>
                        </div>
                      </div>
                      <span className="px-2 py-0.5 bg-slate-800 text-white text-[9px] font-black rounded border border-white/10 uppercase tracking-widest">
                        {item.time}
                      </span>
                    </div>
                    
                    <div className="grid grid-cols-2 gap-3 mt-4">
                       <div className="p-3 bg-slate-950/40 rounded-lg border border-white/5">
                          <div className="text-[9px] text-slate-600 font-black uppercase mb-1">Impact Analysis</div>
                          <div className="text-sm font-black text-rose-400">Drop ({item.drop}%)</div>
                       </div>
                       <div className="p-3 bg-slate-950/40 rounded-lg border border-white/5 border-l-2 border-l-cyan-500/30">
                          <div className="text-[9px] text-slate-600 font-black uppercase mb-1">Contextual Triggers</div>
                          <div className="text-[10px] font-black text-cyan-500 uppercase truncate">
                            {item.signals.length > 0 ? item.signals.map(formatSignal).join(' + ') : 'Internal Anomaly'}
                          </div>
                       </div>
                    </div>

                    <div className="mt-4 flex items-center justify-between border-t border-white/5 pt-4">
                       <div className="flex items-center gap-2">
                         <div className={`w-2 h-2 rounded-full ${item.severity === 'HIGH' ? 'bg-red-500 shadow-[0_0_8px_red]' : item.severity === 'MEDIUM' ? 'bg-orange-500' : 'bg-yellow-500'} animate-pulse`} />
                         <span className={`text-[10px] font-black uppercase tracking-widest ${item.severity === 'HIGH' ? 'text-red-400' : item.severity === 'MEDIUM' ? 'text-orange-400' : 'text-yellow-400'}`}>
                           {item.severity} Severity Disruption
                         </span>
                       </div>
                       <span className="text-[8px] font-bold text-slate-700 uppercase tracking-widest">Event Captured</span>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>

        {/* CHAOS CONTROL PANEL */}
        <div className="col-span-12 lg:col-span-4 space-y-6">
          <div className="bg-slate-900 border-2 border-cyan-500/20 rounded-2xl p-8 shadow-[0_0_50px_rgba(6,182,212,0.1)] sticky top-6">
            <div className="flex items-center gap-3 mb-8">
               <div>
                  <h2 className="text-xl font-black text-white leading-tight">Chaos Control</h2>
                  <div className="text-xs text-cyan-400/60 font-bold uppercase tracking-widest">Simulation Gateway</div>
               </div>
            </div>
            
            <div className="space-y-8">
              <div className="space-y-4">
                <div className="text-[10px] font-black text-slate-500 uppercase tracking-widest flex items-center gap-2">
                   <Activity className="w-3 h-3" /> Step 1: Manage Order Flow
                </div>
                <div className="p-4 bg-slate-800/40 rounded-xl border border-white/5">
                   <p className="text-xs text-slate-400 leading-relaxed">
                      The Disruption Engine reacts automatically to the <strong>Statistical Drop</strong> in volume. 
                      It is currently listening directly to the platform's live Webhook stream.
                   </p>
                </div>
              </div>

              <div className="space-y-4 pt-4 border-t border-white/5">
                <div className="text-[10px] font-black text-slate-500 uppercase tracking-widest flex items-center gap-2">
                   <CloudRain className="w-3 h-3" /> Step 2: Inject External Signal
                </div>
                
                {actionStatus && (
                  <div className="px-3 py-2 rounded border border-cyan-500/30 bg-cyan-900/20 text-cyan-400 text-[10px] font-mono animate-pulse">
                    &gt; {actionStatus}
                  </div>
                )}
                
                <div className="grid grid-cols-2 gap-3 pb-2">
                  <button 
                    disabled={loadingAction || delaying === 'weather_rain'}
                    onClick={() => handleExternalSignal(1, 'weather_rain', true)}
                    className={`flex flex-col items-center justify-center p-3 rounded-xl border transition-all active:scale-95 group relative overflow-hidden
                      ${delaying === 'weather_rain' ? 'bg-blue-500/20 border-blue-500/60' : 'border-blue-500/30 bg-blue-500/10 hover:bg-blue-500/20'}`}>
                    {delaying === 'weather_rain' && <div className="absolute inset-0 bg-blue-500/10 animate-pulse" />}
                    <CloudRain className={`w-5 h-5 mb-1 transition-colors ${delaying === 'weather_rain' ? 'text-white' : 'text-blue-400'}`} />
                    <span className="text-[9px] font-black text-blue-300 uppercase truncate">
                      {delaying === 'weather_rain' ? 'Processing...' : 'Inject Rain'}
                    </span>
                  </button>
                  
                  <button 
                    disabled={loadingAction || delaying === 'aqi_hazardous'}
                    onClick={() => handleExternalSignal(1, 'aqi_hazardous', true)}
                    className={`flex flex-col items-center justify-center p-3 rounded-xl border transition-all active:scale-95 group relative overflow-hidden
                      ${delaying === 'aqi_hazardous' ? 'bg-yellow-500/20 border-yellow-500/60' : 'border-yellow-500/40 bg-yellow-500/10 hover:bg-yellow-500/20'}`}>
                    {delaying === 'aqi_hazardous' && <div className="absolute inset-0 bg-yellow-500/10 animate-pulse" />}
                    <CloudRain className={`w-5 h-5 mb-1 transition-colors ${delaying === 'aqi_hazardous' ? 'text-white' : 'text-yellow-400'}`} />
                    <span className="text-[9px] font-black text-yellow-300 uppercase truncate">
                      {delaying === 'aqi_hazardous' ? 'Processing...' : 'Inject AQI Spike'}
                    </span>
                  </button>

                  <button 
                    disabled={loadingAction || delaying === 'zone_curfew'}
                    onClick={() => handleExternalSignal(1, 'zone_curfew', true)}
                    className={`flex flex-col items-center justify-center p-3 rounded-xl border transition-all active:scale-95 group relative overflow-hidden
                      ${delaying === 'zone_curfew' ? 'bg-orange-500/20 border-orange-500/60' : 'border-orange-500/40 bg-orange-500/10 hover:bg-orange-500/20'}`}>
                    {delaying === 'zone_curfew' && <div className="absolute inset-0 bg-orange-500/10 animate-pulse" />}
                    <ShieldAlert className={`w-5 h-5 mb-1 transition-colors ${delaying === 'zone_curfew' ? 'text-white' : 'text-orange-400'}`} />
                    <span className="text-[9px] font-black text-orange-300 uppercase truncate">
                      {delaying === 'zone_curfew' ? 'Processing...' : 'Simulate Curfew'}
                    </span>
                  </button>

                  <button 
                    disabled={loadingAction}
                    onClick={() => handleExternalSignal(1, 'all_signals', false)}
                    className="flex flex-col items-center justify-center p-3 rounded-xl border border-slate-700 bg-slate-800 hover:bg-slate-700 transition-all active:scale-95 group">
                    <RefreshCw className="w-5 h-5 text-slate-500 mb-1" />
                    <span className="text-[9px] font-black text-slate-400 uppercase truncate">Clear Signals</span>
                  </button>
                </div>
              </div>

              <div className="pt-6 border-t border-white/5">
                 <button 
                    disabled={loadingAction}
                    onClick={() => handleReset(1)}
                    className="w-full py-4 rounded-xl bg-slate-800 border border-white/10 hover:bg-slate-700 transition flex items-center justify-center gap-2 group shadow-lg">
                    <RefreshCw className={`w-4 h-4 text-slate-400 ${loadingAction ? 'animate-spin' : ''}`} />
                    <span className="text-xs font-black text-white tracking-widest uppercase">Reset Engine State</span>
                 </button>
              </div>

            </div>
          </div>
        </div>

      </div>
    </div>
  )
}
