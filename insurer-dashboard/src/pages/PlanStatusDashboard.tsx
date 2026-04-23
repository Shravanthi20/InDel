import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { endUserPlan, getPlanUsers, startUserPlan } from '../api/insurer'
import type { PlanUser } from '../types'
import { useLocalization } from '../context/LocalizationContext'
import { PageShell, Panel } from './OperationsShared'
import { 
  Users, 
  ShieldCheck, 
  ShieldAlert, 
  AlertTriangle, 
  RefreshCw,
  Zap,
  CheckCircle2,
  XCircle,
  Phone,
  MapPin,
  Info,
  TrendingDown,
  TrendingUp,
  Receipt
} from 'lucide-react'

const AuditTooltip = ({ explainability, basePremium }: { explainability?: Array<{ feature: string; impact: number }>, basePremium: number }) => {
  if (!explainability) return null
  return (
    <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-4 w-64 p-5 rounded-2xl bg-white/95 dark:bg-slate-900/95 backdrop-blur-xl border border-slate-200 dark:border-slate-800 shadow-2xl z-50 pointer-events-none opacity-0 group-hover:opacity-100 transition-all duration-300 transform scale-95 group-hover:scale-100 origin-bottom">
       <p className="text-[10px] font-black uppercase tracking-widest text-slate-400 mb-4 flex items-center gap-2">
         <Zap size={12} className="text-brand-primary" />
         Actuarial SHAP Audit
       </p>
       <div className="space-y-3">
         {explainability.map((item, idx) => (
           <div key={idx} className="flex items-center justify-between">
              <span className="text-[10px] font-bold text-slate-500 uppercase tracking-tight">{item.feature}</span>
              <span className={`text-[10px] font-black ${item.impact >= 0 ? 'text-rose-500' : 'text-emerald-500'}`}>
                {item.impact >= 0 ? '+' : ''}₹{item.impact.toFixed(2)}
              </span>
           </div>
         ))}
         <div className="pt-3 border-t border-slate-100 dark:border-slate-800 mt-3 flex items-center justify-between">
            <span className="text-[10px] font-black text-slate-900 dark:text-white uppercase">Final Premium</span>
            <span className="text-sm font-black text-brand-primary">₹{basePremium.toFixed(2)}</span>
         </div>
       </div>
       <div className="absolute -bottom-2 left-1/2 -translate-x-1/2 w-4 h-4 bg-white/95 dark:bg-slate-900/95 border-b border-r border-slate-200 dark:border-slate-800 rotate-45" />
    </div>
  )
}

export default function PlanStatusDashboard() {
  const { t } = useLocalization()
  const [users, setUsers] = useState<PlanUser[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [pendingUserId, setPendingUserId] = useState<number | null>(null)

  const loadUsers = useCallback(async (showSkeleton = false) => {
    if (showSkeleton) {
      setLoading(true)
    } else {
      setRefreshing(true)
    }

    try {
      setError(null)
      const result = await getPlanUsers<PlanUser>()
      setUsers(result)
    } catch (err: any) {
      setError(err?.response?.data?.error?.message || err?.message || 'Failed to load plan status users')
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }, [])

  useEffect(() => {
    void loadUsers(true)
  }, [loadUsers])

  const handleTogglePlan = useCallback(async (user: PlanUser) => {
    setPendingUserId(user.id)
    setError(null)

    try {
      const updatedUser = user.status === 'active'
        ? await endUserPlan<PlanUser>(user.id)
        : await startUserPlan<PlanUser>(user.id)

      setUsers((current) => current.map((entry) => (
        entry.id === updatedUser.id ? updatedUser : entry
      )))
    } catch (err: any) {
      setError(err?.response?.data?.error?.message || err?.message || 'Failed to update plan status')
    } finally {
      setPendingUserId(null)
    }
  }, [])

  const stats = useMemo(() => {
    const total = users.length
    const protectedCount = users.filter(u => u.status === 'active').length
    const unprotectedCount = total - protectedCount
    const percentage = total > 0 ? Math.round((protectedCount / total) * 100) : 0
    
    return {
      total,
      protectedCount,
      unprotectedCount,
      percentage
    }
  }, [users])

  const sortedUsers = useMemo(() => {
    return [...users].sort((a, b) => {
      // Unprotected first
      if (a.status !== 'active' && b.status === 'active') return -1
      if (a.status === 'active' && b.status !== 'active') return 1
      return 0
    })
  }, [users])

  return (
    <PageShell
      eyebrow="Coverage Operations"
      title="Plan Status Intelligence"
      description="Real-time monitoring of worker insurance coverage and proactive enrollment management."
    >
      {error && (
        <div className="mb-6 flex items-center gap-3 p-4 rounded-2xl bg-rose-50 border border-rose-100 text-rose-600 text-xs font-black uppercase tracking-widest shadow-soft-xl">
          <AlertTriangle size={16} />
          {error}
        </div>
      )}

      <div className="grid gap-8">
        {/* Intelligence Header */}
        <Panel title="System Coverage Snapshot">
          <div className="grid gap-12 lg:grid-cols-[1fr_1.5fr]">
            <div className="space-y-8">
              <div className="flex items-center gap-8">
                <div className="h-16 w-16 rounded-2xl bg-brand-primary/10 flex items-center justify-center text-brand-primary shadow-[0_20px_50px_rgba(236,72,153,0.15)]">
                  <Receipt size={32} strokeWidth={2.5} />
                </div>
                <div>
                  <p className="text-4xl font-black text-slate-900 dark:text-white tracking-tighter leading-none">Actuarial Stream</p>
                  <div className="flex items-center gap-3 mt-2">
                    <div className="flex items-center gap-1.5 px-2 py-0.5 rounded-lg bg-emerald-50 text-emerald-600 border border-emerald-100">
                      <ShieldCheck size={12} strokeWidth={3} />
                      <span className="text-[10px] font-black uppercase">{stats.protectedCount} Active</span>
                    </div>
                    <div className="flex items-center gap-1.5 px-2 py-0.5 rounded-lg bg-slate-900 text-white">
                      <Users size={12} strokeWidth={3} />
                      <span className="text-[10px] font-black uppercase">{stats.total} Total</span>
                    </div>
                  </div>
                </div>
              </div>

              {stats.unprotectedCount > 0 && (
                <div className="flex items-start gap-4 p-5 rounded-2xl bg-amber-50 border border-amber-100 animate-in fade-in slide-in-from-left-4 duration-500">
                  <AlertTriangle className="h-6 w-6 text-amber-600 shrink-0" />
                  <div>
                    <p className="text-xs font-black text-amber-900 uppercase tracking-tight">Vulnerability Identified</p>
                    <p className="text-xs text-amber-800 leading-relaxed font-bold mt-1">
                      {stats.unprotectedCount} worker{stats.unprotectedCount > 1 ? 's are' : ' is'} currently unprotected. Actuarial floor suggests immediate enrollment to stabilize risk.
                    </p>
                  </div>
                </div>
              )}
            </div>

            <div className="flex flex-col justify-center space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="flex items-center gap-2">
                     <div className="h-2 w-2 rounded-full bg-brand-primary animate-pulse" />
                     <p className="text-sm font-black text-slate-900 dark:text-white uppercase tracking-widest">Real-time Coverage</p>
                  </div>
                </div>
                <p className="text-4xl font-black text-brand-primary tracking-tighter">{stats.percentage}%</p>
              </div>
              
              <div className="h-4 w-full bg-slate-100 dark:bg-slate-800 rounded-full overflow-hidden shadow-inner flex relative">
                <div 
                  className="h-full bg-gradient-to-r from-brand-primary to-rose-500 transition-all duration-1000 ease-out flex items-center justify-end px-2"
                  style={{ width: `${stats.percentage}%` }}
                >
                  <div className="h-1.5 w-1.5 rounded-full bg-white animate-ping" />
                </div>
              </div>
              <div className="flex justify-between text-[10px] font-black uppercase tracking-[0.2em] text-slate-400">
                <span>Vulnerable</span>
                <span>Fully Protected</span>
              </div>
            </div>
          </div>
        </Panel>

        {/* Actionable Table */}
        <Panel 
           title="Worker Risk & Enrollment Audit" 
           subtitle="Surname-based identity verification with full SHAP actuarial audit trails."
        >
          <div className="flex items-center justify-between mb-8">
            <div className="flex items-center gap-4">
               <div className="flex items-center gap-2">
                  <div className="h-2 w-2 rounded-full bg-brand-primary animate-ping" />
                  <p className="text-[10px] font-black uppercase tracking-widest text-slate-400">Actuarial Sync Active</p>
               </div>
               <div className="h-4 w-[1px] bg-slate-200 dark:bg-slate-800" />
               <p className="text-[10px] font-black uppercase tracking-widest text-slate-500 italic">
                 Algorithm: Actuarial-XGB-v1
               </p>
            </div>
            <button
               onClick={() => void loadUsers(false)}
               disabled={loading || refreshing}
               className="flex items-center gap-2 px-6 py-2 rounded-xl border border-slate-100 dark:border-slate-800 text-[10px] font-black uppercase tracking-widest hover:bg-slate-50 dark:hover:bg-slate-950 transition-all shadow-soft-xl"
            >
              <RefreshCw className={`h-3 w-3 ${refreshing ? 'animate-spin' : ''}`} />
              {refreshing ? 'Refreshing Pricing...' : 'Force Price Sync'}
            </button>
          </div>

          <div className="overflow-x-auto -mx-10 px-10">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="border-b border-slate-50 dark:border-slate-800">
                  <th className="pb-5 px-4 font-black uppercase tracking-[0.2em] text-[10px] text-slate-400 font-['Outfit']">Identity / Locality</th>
                  <th className="pb-5 px-4 font-black uppercase tracking-[0.2em] text-[10px] text-slate-400 font-['Outfit'] text-center">Coverage</th>
                  <th className="pb-5 px-4 font-black uppercase tracking-[0.2em] text-[10px] text-slate-400 font-['Outfit'] text-right">Weekly Premium</th>
                  <th className="pb-5 px-4 font-black uppercase tracking-[0.2em] text-[10px] text-slate-400 font-['Outfit'] text-right">Max Payout</th>
                  <th className="pb-5 px-4 font-black uppercase tracking-[0.2em] text-[10px] text-slate-400 font-['Outfit'] text-right">Enrollment Fate</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-50 dark:divide-slate-800/50">
                {sortedUsers.map((user, index) => {
                  const isActive = user.status === 'active'
                  const isPending = pendingUserId === user.id

                  return (
                    <tr 
                      key={`${user.id}-${index}`} 
                      className={`group transition-all duration-300 ${!isActive ? 'bg-brand-soft/30 dark:bg-brand-primary/5' : 'hover:bg-slate-50/50 dark:hover:bg-slate-900/30'}`}
                    >
                      <td className="py-6 px-4">
                        <div className="flex items-center gap-4">
                           <div className={`h-11 w-11 rounded-xl flex items-center justify-center font-black text-sm border shadow-sm ${
                             isActive ? 'bg-white dark:bg-slate-950 border-slate-100 dark:border-slate-800 text-slate-400' : 'bg-brand-primary text-white border-brand-primary shadow-brand-primary/20'
                           }`}>
                             {user.name?.[0] || 'U'}
                           </div>
                           <div>
                              <p className="font-black text-slate-900 dark:text-white uppercase tracking-tight text-xs">{user.name || 'Unknown Entity'}</p>
                              <div className="flex items-center gap-2 mt-1 text-slate-400">
                                 <MapPin size={10} className="text-brand-primary" />
                                 <p className="text-[10px] font-black uppercase tracking-widest">{user.zone || 'Global'}</p>
                              </div>
                           </div>
                        </div>
                      </td>
                      <td className="py-6 px-4 text-center">
                        <div className={`inline-flex items-center gap-2 px-4 py-1.5 rounded-full border transition-all ${
                          isActive 
                            ? 'bg-emerald-50 dark:bg-emerald-500/10 border-emerald-100 dark:border-emerald-500/20 text-emerald-600'
                            : 'bg-rose-50 dark:bg-rose-500/10 border-rose-100 dark:border-rose-500/20 text-rose-600'
                        }`}>
                          {isActive ? <CheckCircle2 size={12} strokeWidth={3} /> : <ShieldAlert size={12} strokeWidth={3} />}
                          <span className="text-[10px] font-black uppercase tracking-widest">
                            {isActive ? 'Protected' : 'Volatile'}
                          </span>
                        </div>
                      </td>
                      <td className="py-6 px-4 text-right align-middle">
                        <div className="relative inline-block group">
                           <div className="flex flex-col items-end">
                              <div className="text-sm font-black text-slate-900 dark:text-white flex items-center gap-2">
                                ₹{user.weekly_premium?.toFixed(2) ?? '0.00'}
                                <div className="h-1.5 w-1.5 rounded-full bg-brand-primary animate-pulse" />
                              </div>
                              <p className="text-[9px] font-black uppercase tracking-widest text-slate-400 flex items-center gap-1 hover:text-brand-primary transition-colors cursor-help">
                                SHAP Audit <Info size={9} />
                              </p>
                           </div>
                           <AuditTooltip explainability={user.explainability} basePremium={user.weekly_premium ?? 0} />
                        </div>
                      </td>
                      <td className="py-6 px-4 text-right">
                        <p className="text-sm font-black text-slate-600 dark:text-slate-400">₹{Math.round(user.max_payout ?? 0).toLocaleString('en-IN')}</p>
                        <p className="text-[9px] font-black uppercase tracking-widest text-slate-400">Max Benefit</p>
                      </td>
                      <td className="py-6 px-4 text-right">
                        <button
                          onClick={() => void handleTogglePlan(user)}
                          disabled={isPending}
                          className={`relative overflow-hidden group/btn px-6 py-2.5 rounded-xl font-black text-[10px] uppercase tracking-widest transition-all duration-300 transform active:scale-95 ${
                            isActive 
                              ? 'bg-slate-100 dark:bg-slate-800 text-slate-500 hover:bg-slate-200 dark:hover:bg-slate-700'
                              : 'bg-brand-primary text-white shadow-lg shadow-brand-primary/20 hover:scale-[1.05] hover:shadow-brand-primary/40'
                          } ${isPending ? 'opacity-50 cursor-wait' : 'cursor-pointer'}`}
                        >
                          <span className="relative z-10">
                            {isPending ? 'Syncing...' : isActive ? 'Verify & End' : 'Start Plan Now'}
                          </span>
                          {!isActive && (
                            <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 to-transparent -translate-x-full group-hover/btn:animate-[shimmer_1.5s_infinite]" />
                          )}
                        </button>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </Panel>
      </div>
    </PageShell>
  )
}
