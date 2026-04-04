import { Outlet } from 'react-router-dom'
import { useGodMode } from './state'

export default function GodModeLayout() {
  const {
    godModeEnabled,
    setGodModeEnabled,
    zoneLevel,
    setZoneLevel,
    zoneName,
    setZoneName,
    zoneNameOptions,
    zoneNameLoading,
    scopeLabel,
    affectedZoneIds,
    running,
    loading,
    runSimulation,
    generatingBatches,
    generateBatches,
    notice,
    clearNotice,
  } = useGodMode()

  return (
    <div className="min-h-screen bg-slate-50 p-5 text-slate-900 lg:p-8">
      <div className="mx-auto max-w-[1560px] space-y-6">
        <section className="rounded-[2rem] border border-slate-200 bg-white p-6 shadow-sm">
          <div className="flex flex-col gap-5 xl:flex-row xl:items-center xl:justify-between">
            <div>
              <p className="text-[11px] uppercase tracking-[0.35em] text-sky-700">God Mode Simulation</p>
              <h1 className="mt-2 text-4xl font-black tracking-tight text-slate-900">Factor pages with shared backend state</h1>
              <p className="mt-2 max-w-3xl text-sm leading-7 text-slate-600">
                Each disruption factor now lives in its own page. The run action and backend endpoint data are shared across all tabs.
              </p>
            </div>

            <div className="grid gap-3 rounded-[1.5rem] border border-slate-200 bg-slate-50 p-4 sm:grid-cols-2 xl:min-w-[640px] xl:grid-cols-[repeat(3,minmax(0,1fr))_auto]">
              <label className="grid gap-2 text-xs font-semibold uppercase tracking-[0.22em] text-slate-500">
                Zone Level
                <select
                  value={zoneLevel}
                  onChange={(event) => setZoneLevel(event.target.value as 'ALL' | 'A' | 'B' | 'C')}
                  className="rounded-2xl border border-slate-200 bg-white px-3 py-3 text-sm font-semibold text-slate-900 outline-none transition focus:border-sky-300"
                >
                  <option value="ALL">ALL ZONES</option>
                  <option value="A">Zone A</option>
                  <option value="B">Zone B</option>
                  <option value="C">Zone C</option>
                </select>
              </label>

              <label className="grid gap-2 text-xs font-semibold uppercase tracking-[0.22em] text-slate-500">
                Zone Name
                <select
                  value={zoneName}
                  onChange={(event) => setZoneName(event.target.value)}
                  className="rounded-2xl border border-slate-200 bg-white px-3 py-3 text-sm font-semibold text-slate-900 outline-none transition focus:border-sky-300"
                  disabled={zoneLevel === 'ALL' || zoneNameLoading || zoneNameOptions.length <= 1}
                >
                  {zoneNameOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </label>

              <div className="flex flex-col gap-2">
                <span className="text-xs font-semibold uppercase tracking-[0.22em] text-slate-500">Scope</span>
                <div className="rounded-2xl border border-slate-200 bg-white px-3 py-3 text-sm font-semibold text-slate-800">
                  {scopeLabel}
                </div>
                <div className="text-[11px] uppercase tracking-[0.22em] text-slate-500">Affected zones: {affectedZoneIds.length}</div>
              </div>

              <div className="flex flex-wrap items-end gap-3 xl:justify-end">
                <button
                  type="button"
                  onClick={() => setGodModeEnabled(!godModeEnabled)}
                  className={`rounded-full border px-5 py-2 text-sm font-semibold uppercase tracking-[0.2em] transition ${godModeEnabled
                    ? 'border-rose-300 bg-rose-50 text-rose-700 hover:bg-rose-100'
                    : 'border-sky-300 bg-sky-50 text-sky-700 hover:bg-sky-100'}`}
                >
                  {godModeEnabled ? 'Disable God Mode' : 'Enable God Mode'}
                </button>

                <button
                  type="button"
                  onClick={runSimulation}
                  disabled={running || loading}
                  className="rounded-full border border-slate-300 bg-white px-5 py-2 text-sm font-semibold uppercase tracking-[0.2em] text-slate-800 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {running ? 'Running...' : 'Run Simulation'}
                </button>
                <button
                  type="button"
                  onClick={generateBatches}
                  disabled={generatingBatches}
                  className="rounded-full border border-slate-300 bg-white px-5 py-2 text-sm font-semibold uppercase tracking-[0.2em] text-slate-800 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {generatingBatches ? 'Generating...' : 'Generate Batches'}
                </button>
              </div>
            </div>
          </div>

          {notice ? (
            <div
              className={`mt-4 flex items-center justify-between gap-3 rounded-2xl border px-4 py-3 text-sm font-medium ${notice.tone === 'success'
                ? 'border-emerald-200 bg-emerald-50 text-emerald-800'
                : 'border-rose-200 bg-rose-50 text-rose-800'}`}
            >
              <span>{notice.message}</span>
              <button
                type="button"
                onClick={clearNotice}
                className="rounded-full border border-current px-3 py-1 text-xs font-semibold uppercase tracking-[0.2em] opacity-80 transition hover:opacity-100"
              >
                Dismiss
              </button>
            </div>
          ) : null}
        </section>

        <Outlet />
      </div>
    </div>
  )
}
