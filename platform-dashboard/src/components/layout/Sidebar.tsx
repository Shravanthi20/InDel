import { Link } from 'react-router-dom'

export default function Sidebar({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen bg-gray-100">
      <aside className="w-64 bg-slate-900 text-white p-6">
        <h1 className="text-2xl font-bold mb-8">InDel Admin</h1>
        <nav className="space-y-4">
          <Link to="/" className="block hover:bg-slate-800 p-2 rounded">Overview</Link>
          <Link to="/workers" className="block hover:bg-slate-800 p-2 rounded">Workers</Link>
          <Link to="/zones" className="block hover:bg-slate-800 p-2 rounded">Zones</Link>
          <Link to="/analytics" className="block hover:bg-slate-800 p-2 rounded">Analytics</Link>
        </nav>
      </aside>
      <main className="flex-1">
        {children}
      </main>
    </div>
  )
}
