import { useEffect, useState } from 'react'
import { getWorkers } from '../api/platform'
import { Search, Filter, MoreVertical, CreditCard, MapPin } from 'lucide-react'

export default function Workers() {
  const [workers, setWorkers] = useState<any[]>([])

  useEffect(() => {
    async function load() {
      const response = await getWorkers()
      setWorkers(response.data.workers ?? [])
    }

    load().catch((error) => console.error('Failed to load workers', error))
  }, [])

  return (
    <div className="space-y-10">
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-2xl font-black tracking-tight text-slate-900 dark:text-white">Worker Directory</h1>
          <p className="mt-1 text-sm text-slate-500">Managing global gig-worker identity and regional zone assignments.</p>
        </div>
        <div className="flex gap-3">
           <button className="h-9 px-4 rounded-md border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-xs font-bold font-['Outfit'] hover:bg-slate-50 dark:hover:bg-slate-800 transition-none">
              EXPORT CSV
           </button>
           <button className="h-9 px-4 rounded-md bg-orange-600 text-white text-xs font-bold font-['Outfit'] hover:bg-orange-700 transition-none">
              ADD WORKER
           </button>
        </div>
      </div>

      <div className="enterprise-panel overflow-hidden">
        <div className="border-b border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/20 p-4 flex items-center justify-between">
           <div className="relative group w-72">
              <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-slate-400">
                <Search className="h-3.5 w-3.5" />
              </div>
              <input
                type="text"
                placeholder="Filter by name, ID or zone..."
                className="w-full rounded border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 py-1.5 pl-9 pr-3 text-[11px] text-slate-900 dark:text-white outline-none focus:border-orange-500 transition-none"
              />
           </div>
           <button className="flex items-center gap-2 px-3 py-1.5 rounded border border-slate-200 dark:border-slate-700 text-[11px] font-bold text-slate-500 hover:text-slate-900 dark:hover:text-white transition-none">
              <Filter className="h-3 w-3" />
              ADVANCED FILTER
           </button>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full border-collapse text-left">
            <thead>
              <tr className="border-b border-slate-100 dark:border-slate-800">
                <th className="px-8 py-4 text-[10px] font-black uppercase tracking-widest text-slate-400">Worker</th>
                <th className="px-8 py-4 text-[10px] font-black uppercase tracking-widest text-slate-400">Zone Assignment</th>
                <th className="px-8 py-4 text-[10px] font-black uppercase tracking-widest text-slate-400">Policy Status</th>
                <th className="px-8 py-4 text-[10px] font-black uppercase tracking-widest text-slate-400">Activity</th>
                <th className="px-8 py-4 text-[10px] font-black uppercase tracking-widest text-slate-400">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
              {workers.map((worker) => (
                <tr key={worker.worker_id} className="hover:bg-slate-50 dark:hover:bg-slate-800/30">
                  <td className="px-8 py-5">
                    <div className="flex items-center gap-3">
                       <div className="h-9 w-9 rounded bg-slate-100 dark:bg-slate-800 flex items-center justify-center text-[10px] font-black text-slate-500">
                          {worker.name.split(' ').map((n: string) => n[0]).join('')}
                       </div>
                       <div>
                          <div className="text-xs font-bold text-slate-900 dark:text-white">{worker.name}</div>
                          <div className="text-[10px] text-slate-400 font-medium">#{worker.worker_id.substr(0, 12)}</div>
                       </div>
                    </div>
                  </td>
                  <td className="px-8 py-5">
                     <div className="flex items-center gap-1.5">
                        <MapPin className="h-3 w-3 text-slate-400" />
                        <span className="text-xs font-medium text-slate-600 dark:text-slate-300">{worker.zone}</span>
                     </div>
                  </td>
                  <td className="px-8 py-5">
                     <div className="flex items-center gap-2 px-2 py-1 rounded bg-emerald-50 dark:bg-emerald-500/10 border border-emerald-100 dark:border-emerald-500/20 w-fit">
                        <div className="h-1 w-1 rounded-full bg-emerald-500"></div>
                        <span className="text-[9px] font-black uppercase tracking-tighter text-emerald-600 dark:text-emerald-400">ACTIVE_COVERAGE</span>
                     </div>
                  </td>
                  <td className="px-8 py-5">
                     <div className="text-xs font-medium text-slate-600 dark:text-slate-300">Live • On Shift</div>
                     <div className="text-[10px] text-slate-400 mt-0.5">Contact: {worker.phone}</div>
                  </td>
                  <td className="px-8 py-5">
                     <div className="flex items-center gap-3">
                        <button className="p-1 text-slate-400 hover:text-orange-500">
                           <CreditCard className="h-4 w-4" />
                        </button>
                        <button className="p-1 text-slate-400 hover:text-slate-900 dark:hover:text-white">
                           <MoreVertical className="h-4 w-4" />
                        </button>
                     </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        
        <div className="p-6 border-t border-slate-100 dark:border-slate-800 bg-slate-50/30 dark:bg-slate-800/10 flex items-center justify-between">
           <div className="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Showing {workers.length} active nodes</div>
           <div className="flex gap-2">
              <button className="px-3 py-1 rounded border border-slate-200 dark:border-slate-800 text-[10px] font-bold text-slate-400">PREV</button>
              <button className="px-3 py-1 rounded border border-slate-200 dark:border-slate-800 text-[10px] font-bold text-slate-400">NEXT</button>
           </div>
        </div>
      </div>
    </div>
  )
}
