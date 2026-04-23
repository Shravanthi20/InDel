import { useEffect, useState } from 'react'
import { getMLForecast, getZones } from '../api/insurer'
import { PageShell, Panel } from './OperationsShared'
import { 
  BrainCircuit, 
  Info, 
  RefreshCw,
  AlertTriangle,
  Zap,
  ShieldCheck,
  TrendingUp
} from 'lucide-react'
import { 
  AreaChart, 
  Area, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer, 
  ReferenceArea,
  Dot
} from 'recharts'
import { useLocalization } from '../context/LocalizationContext'

interface ForecastPoint {
  date: string
  disruption_probability: number
}

interface ChartDataPoint {
  name: string
  fullDate: string
  prob: number
  rawProb: number
}

export default function Forecast() {
  const { t } = useLocalization()
  const [zones, setZones] = useState<any[]>([])
  const [forecast, setForecast] = useState<ForecastPoint[]>([])
  const [forecastMeta, setForecastMeta] = useState<{ retraining_cadence?: string; scope?: string } | null>(null)
  const [forecastLoading, setForecastLoading] = useState(true)
  const [forecastError, setForecastError] = useState(false)
  const [selectedZone, setSelectedZone] = useState<number | null>(null)

  useEffect(() => {
    async function load() {
      try {
        const zonesRes = await getZones()
        const nextZones = zonesRes.data?.zones ?? []
        setZones(nextZones)
        setSelectedZone((current) => current ?? nextZones[0]?.zone_id ?? null)
      } catch (error) {
        console.error('Failed to load zones', error)
      }
    }
    load()
  }, [])

  useEffect(() => {
    async function loadForecast() {
      if (selectedZone === null) {
        setForecast([])
        setForecastMeta(null)
        setForecastLoading(false)
        return
      }

      setForecastLoading(true)
      setForecastError(false)
      try {
        const res = await getMLForecast(selectedZone)
        setForecast(res.data.forecast ?? [])
        setForecastMeta({
          retraining_cadence: res.data.retraining_cadence,
          scope: res.data.scope,
        })
      } catch {
        setForecastError(true)
      } finally {
        setForecastLoading(false)
      }
    }

    loadForecast()
  }, [selectedZone])

  const chartData: ChartDataPoint[] = forecast.map(p => {
    const d = new Date(p.date)
    return {
      name: d.toLocaleDateString('en', { weekday: 'short' }),
      fullDate: d.toLocaleDateString('en', { month: 'short', day: 'numeric' }),
      prob: Math.round(p.disruption_probability * 100),
      rawProb: p.disruption_probability
    }
  })

  // Custom Tooltip for a premium feel
  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      const data = payload[0].payload
      const isHigh = data.prob >= 35
      const isMed = data.prob >= 20 && !isHigh
      
      return (
        <div className="bg-white/90 dark:bg-slate-900/90 backdrop-blur-xl border border-slate-200 dark:border-slate-800 p-4 rounded-2xl shadow-2xl">
          <p className="text-[10px] font-black uppercase tracking-widest text-slate-400 mb-2">
            {data.name} · {data.fullDate}
          </p>
          <div className="flex items-center gap-3">
            <div className={`h-10 w-10 rounded-xl flex items-center justify-center ${
              isHigh ? 'bg-rose-500/10 text-rose-500' : isMed ? 'bg-amber-500/10 text-amber-500' : 'bg-emerald-500/10 text-emerald-500'
            }`}>
              {isHigh ? <AlertTriangle size={20} /> : isMed ? <Zap size={20} /> : <ShieldCheck size={20} />}
            </div>
            <div>
              <p className="text-2xl font-black tracking-tighter text-slate-900 dark:text-white leading-none">
                {data.prob}%
              </p>
              <p className={`text-[9px] font-black uppercase tracking-widest mt-1 ${
                isHigh ? 'text-rose-500' : isMed ? 'text-amber-500' : 'text-emerald-500'
              }`}>
                {isHigh ? 'Critical Risk' : isMed ? 'Incr. Pressure' : 'Stable Ops'}
              </p>
            </div>
          </div>
        </div>
      )
    }
    return null
  }

  return (
    <PageShell
      eyebrow={t('pages.forecast.eyebrow')}
      title={t('pages.forecast.title')}
      description={t('pages.forecast.description')}
    >
      <Panel title={t('pages.forecast.upcomingRisk')}>
        <div className="flex flex-wrap items-center justify-between gap-6 mb-10">
          <div className="flex items-center gap-5">
            <div className="h-14 w-14 rounded-2xl bg-brand-soft/50 dark:bg-brand-primary/5 flex items-center justify-center border border-brand-primary/10 shadow-inner">
              <TrendingUp className="h-7 w-7 text-brand-primary" strokeWidth={3} />
            </div>
            <div>
              <p className="text-[10px] text-slate-400 uppercase tracking-[0.25em] font-black font-['Outfit']">{t('pages.analytics.selectedZone')}</p>
              <select
                id="forecast-zone-select"
                value={selectedZone || ''}
                onChange={(e) => setSelectedZone(Number(e.target.value))}
                className="mt-1 bg-transparent text-2xl font-black text-slate-900 dark:text-white outline-none cursor-pointer border-b-2 border-transparent hover:border-brand-primary focus:border-brand-primary transition-all pr-12 appearance-none font-['Outfit']"
                style={{ backgroundImage: 'url("data:image/svg+xml,%3Csvg xmlns=\'http://www.w3.org/2000/svg\' fill=\'none\' viewBox=\'0 0 24 24\' stroke=\'%23EC4899\' stroke-width=\'3\'%3E%3Cpath stroke-linecap=\'round\' stroke-linejoin=\'round\' d=\'M19.5 8.25l-7.5 7.5-7.5-7.5\' /%3E%3C/svg%3E")', backgroundRepeat: 'no-repeat', backgroundPosition: 'right center', backgroundSize: '1.2rem' }}
              >
                {zones.map((zone) => (
                  <option key={zone.zone_id} value={zone.zone_id} className="dark:bg-gray-900">
                    {zone.name || `Zone ${zone.zone_id}`}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className="flex items-center gap-8">
            <div className="hidden lg:block text-right">
              <p className="text-[10px] text-slate-400 uppercase tracking-[0.2em] font-black mb-1">{t('pages.analytics.retrainingCadence')}</p>
              <div className="flex items-center gap-2 justify-end">
                <div className="h-2 w-2 rounded-full bg-emerald-500 animate-pulse" />
                <p className="text-xs font-black text-slate-600 dark:text-slate-300">{forecastMeta?.retraining_cadence || 'Weekly — Mon 02:00 UTC'}</p>
              </div>
            </div>
            <button 
              onClick={() => setSelectedZone(selectedZone)}
              className="group h-12 w-12 rounded-2xl bg-white dark:bg-slate-950 border border-slate-100 dark:border-slate-800 flex items-center justify-center hover:border-brand-primary hover:shadow-2xl transition-all duration-300"
            >
              <RefreshCw className={`h-5 w-5 text-brand-primary transition-transform duration-700 ${forecastLoading ? 'animate-spin' : 'group-hover:rotate-180'}`} />
            </button>
          </div>
        </div>

        <div className="mb-10 flex items-center gap-4 p-6 rounded-2xl border border-amber-100/50 dark:border-amber-900/20 bg-amber-50/20 dark:bg-amber-900/5 backdrop-blur-sm">
          <div className="h-8 w-8 rounded-lg bg-amber-500/10 flex items-center justify-center text-amber-600 shrink-0">
            <Info size={16} />
          </div>
          <p className="text-xs text-amber-800/80 dark:text-amber-300/60 leading-relaxed font-bold tracking-tight">
            Prophet Intelligence Engine: Disruption signals detected via multi-seasonal trend analysis. Shaded regions indicate historical risk thresholds.
          </p>
        </div>

        <div className="h-[450px] w-full relative">
          {forecastError ? (
            <div className="absolute inset-0 flex flex-col items-center justify-center rounded-3xl border-2 border-dashed border-slate-100 dark:border-slate-800">
              <AlertTriangle className="h-12 w-12 text-rose-500/30 mb-4" />
              <p className="text-slate-400 font-black uppercase tracking-widest">{t('pages.analytics.forecastError')}</p>
            </div>
          ) : forecastLoading ? (
            <div className="absolute inset-0 flex flex-col items-center justify-center">
              <RefreshCw className="h-10 w-10 text-brand-primary animate-spin mb-4" />
              <p className="text-[10px] text-slate-400 font-black uppercase tracking-[0.3em]">{t('pages.analytics.loadingForecast')}</p>
            </div>
          ) : (
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart
                data={chartData}
                margin={{ top: 20, right: 30, left: 0, bottom: 20 }}
              >
                <defs>
                  <linearGradient id="forecastGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#EC4899" stopOpacity={0.4} />
                    <stop offset="95%" stopColor="#EC4899" stopOpacity={0} />
                  </linearGradient>
                </defs>
                
                <CartesianGrid 
                  strokeDasharray="8 8" 
                  vertical={false} 
                  stroke="rgba(148, 163, 184, 0.1)" 
                />
                
                <XAxis 
                  dataKey="name" 
                  axisLine={false}
                  tickLine={false}
                  tick={{ fill: '#94a3b8', fontSize: 10, fontWeight: 900 }}
                  dy={15}
                />
                
                <YAxis 
                  domain={[0, 100]} 
                  axisLine={false}
                  tickLine={false}
                  tick={{ fill: '#94a3b8', fontSize: 10, fontWeight: 700 }}
                  dx={-10}
                />
                
                <Tooltip 
                  content={<CustomTooltip />} 
                  cursor={{ stroke: '#EC4899', strokeWidth: 1, strokeDasharray: '4 4' }}
                />

                {/* Highlight Danger Zone (>35%) */}
                <ReferenceArea 
                  y1={35} 
                  y2={100} 
                  fill="rgba(244, 63, 94, 0.03)" 
                  stroke="none"
                />

                <Area
                  type="monotone"
                  dataKey="prob"
                  stroke="#EC4899"
                  strokeWidth={5}
                  fillOpacity={1}
                  fill="url(#forecastGradient)"
                  animationDuration={2000}
                  activeDot={{ 
                    r: 8, 
                    fill: '#EC4899', 
                    stroke: '#fff', 
                    strokeWidth: 4, 
                    className: 'shadow-2xl' 
                  }}
                />
              </AreaChart>
            </ResponsiveContainer>
          )}
        </div>

        <div className="mt-12 flex flex-wrap gap-8 items-center justify-center py-8 border-t border-slate-50 dark:border-slate-800">
          <LegendItem icon={<div className="h-3 w-3 rounded-full bg-rose-500 shadow-lg shadow-rose-500/30" />} label="Critical Zone (>35%)" />
          <LegendItem icon={<div className="h-3 w-3 rounded-full bg-slate-200 dark:bg-slate-700" />} label="Stable Threshold" />
          <div className="h-4 w-px bg-slate-100 dark:bg-slate-800 mx-4" />
          <p className="text-[10px] font-black text-slate-400 tracking-[0.2em] uppercase">
            Model: Facebook Prophet (Weekly)
          </p>
        </div>
      </Panel>
    </PageShell>
  )
}

function LegendItem({ icon, label }: { icon: React.ReactNode; label: string }) {
  return (
    <div className="flex items-center gap-3">
      {icon}
      <span className="text-[10px] font-black uppercase tracking-[0.2em] text-slate-500 font-['Outfit']">{label}</span>
    </div>
  )
}
