import { useEffect, useState, useMemo } from 'react'
import { getWorkers } from '../api/platform'
import { Search, Filter, MoreVertical, CreditCard, MapPin, Download } from 'lucide-react'

export default function Workers() {
  const [workers, setWorkers] = useState<any[]>([])
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<'all' | 'live' | 'offline'>('all')

  useEffect(() => {
    async function load() {
      const response = await getWorkers()
      setWorkers(response.data.workers ?? [])
    }

    load().catch((error) => console.error('Failed to load workers', error))
  }, [])

  const filteredWorkers = useMemo(() => {
    return workers.filter(worker => {
      const matchesSearch = 
        worker.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        worker.worker_id.toString().includes(searchQuery) ||
        worker.zone.toLowerCase().includes(searchQuery.toLowerCase())
      
      const isLive = worker.zone !== 'Unknown Zone'
      const matchesStatus = 
        statusFilter === 'all' || 
        (statusFilter === 'live' && isLive) || 
        (statusFilter === 'offline' && !isLive)

      return matchesSearch && matchesStatus
    })
  }, [workers, searchQuery, statusFilter])

  const exportToCSV = () => {
    const headers = ['Worker ID', 'Name', 'Phone', 'Zone', 'Status']
    const rows = filteredWorkers.map(w => [
      w.worker_id,
      w.name,
      w.phone,
      w.zone,
      w.zone !== 'Unknown Zone' ? 'Live' : 'Offline'
    ])
    
    const csvContent = [headers, ...rows].map(e => e.join(",")).join("\n")
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement("a")
    const url = URL.createObjectURL(blob)
    link.setAttribute("href", url)
    link.setAttribute("download", `workers_export_${new Date().toISOString().split('T')[0]}.csv`)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  return (
    <div className="space-y-10">
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-2xl font-black tracking-tight text-slate-900 dark:text-white">Worker Directory</h1>
          <p className="mt-1 text-sm text-slate-500">Managing global gig-worker identity and regional zone assignments.</p>
        </div>
        <div className="flex gap-3">
           <button 
             onClick={exportToCSV}
             className="flex items-center gap-2 h-9 px-4 rounded-md border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-xs font-bold font-['Outfit'] hover:bg-slate-50 dark:hover:bg-slate-800 transition-none">
              <Download className="h-3.5 w-3.5" />
              EXPORT CSV
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
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Filter by name, ID or zone..."
                className="w-full rounded border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 py-1.5 pl-9 pr-3 text-[11px] text-slate-900 dark:text-white outline-none focus:border-orange-500 transition-none"
              />
           </div>
           <div className="flex items-center gap-2">
              <span className="text-[10px] font-black uppercase text-slate-400 mr-2">Status Filter:</span>
              {(['all', 'live', 'offline'] as const).map((s) => (
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
              {filteredWorkers.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-8 py-20 text-center text-slate-400 text-xs italic">
                    No workers match the current criteria.
                  </td>
                </tr>
              ) : filteredWorkers.map((worker) => {
                const isLive = worker.zone !== 'Unknown Zone'
                return (
                  <tr key={worker.worker_id} className="hover:bg-slate-50 dark:hover:bg-slate-800/30">
                    <td className="px-8 py-5">
                      <div className="flex items-center gap-3">
                         <div className="h-9 w-9 rounded bg-slate-100 dark:bg-slate-800 flex items-center justify-center text-[10px] font-black text-slate-500 uppercase">
                            {worker.name.split(' ').map((n: string) => n[0]).join('')}
                         </div>
                         <div>
                            <div className="text-xs font-bold text-slate-900 dark:text-white">{worker.name}</div>
                            <div className="text-[10px] text-slate-400 font-medium">#{String(worker.worker_id).padStart(6, '0')}</div>
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
                       <div className={`flex items-center gap-2 px-2 py-1 rounded w-fit border ${
                         isLive 
                           ? 'bg-emerald-50 dark:bg-emerald-500/10 border-emerald-100 dark:border-emerald-500/20' 
                           : 'bg-slate-100 dark:bg-slate-800 border-slate-200 dark:border-slate-700 opacity-60'
                       }`}>
                          <div className={`h-1 w-1 rounded-full ${isLive ? 'bg-emerald-500' : 'bg-slate-400'}`}></div>
                          <span className={`text-[9px] font-black uppercase tracking-tighter ${isLive ? 'text-emerald-600 dark:text-emerald-400' : 'text-slate-500'}`}>
                            {isLive ? 'ACTIVE_COVERAGE' : 'INACTIVE'}
                          </span>
                       </div>
                    </td>
                    <td className="px-8 py-5">
                       <div className="text-xs font-medium text-slate-600 dark:text-slate-300">
                          {isLive ? 'Live • On Shift' : 'Offline'}
                       </div>
                       <div className="text-[10px] text-slate-400 mt-0.5">Contact: {worker.phone}</div>
                    </td>
                    <td className="px-8 py-5">
                       <div className="flex items-center gap-3">
                          <button className="p-1 text-slate-400 hover:text-orange-500 transition-colors">
                             <CreditCard className="h-4 w-4" />
                          </button>
                          <button className="p-1 text-slate-400 hover:text-slate-900 dark:hover:text-white transition-colors">
                             <MoreVertical className="h-4 w-4" />
                          </button>
                       </div>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
        
        <div className="p-6 border-t border-slate-100 dark:border-slate-800 bg-slate-50/30 dark:bg-slate-800/10 flex items-center justify-between">
           <div className="text-[10px] font-bold text-slate-400 uppercase tracking-widest">
             Showing {filteredWorkers.length} of {workers.length} nodes
           </div>
           <div className="flex gap-2">
              <button disabled className="px-3 py-1 rounded border border-slate-200 dark:border-slate-800 text-[10px] font-bold text-slate-300 dark:text-slate-700 cursor-not-allowed">PREV</button>
              <button disabled className="px-3 py-1 rounded border border-slate-200 dark:border-slate-800 text-[10px] font-bold text-slate-300 dark:text-slate-700 cursor-not-allowed">NEXT</button>
           </div>
        </div>
      </div>
    </div>
  )
}
