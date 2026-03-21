import { Link } from 'react-router-dom'

export default function Sidebar({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen bg-gray-100">
      <aside className="w-64 bg-slate-900 text-white p-6">
        <h1 className="text-2xl font-bold mb-8">InDel</h1>
        <nav className="space-y-4">
          <Link to="/" className="block hover:bg-slate-800 p-2 rounded">Overview</Link>
          <Link to="/loss-ratio" className="block hover:bg-slate-800 p-2 rounded">Loss Ratio</Link>
          <Link to="/claims" className="block hover:bg-slate-800 p-2 rounded">Claims</Link>
          <Link to="/fraud-queue" className="block hover:bg-slate-800 p-2 rounded">Fraud Queue</Link>
          <Link to="/forecast" className="block hover:bg-slate-800 p-2 rounded">Forecast</Link>
          <Link to="/maintenance-checks" className="block hover:bg-slate-800 p-2 rounded">Maintenance</Link>
        </nav>
      </aside>
      <main className="flex-1">
        {children}
      </main>
    </div>
  )
}
