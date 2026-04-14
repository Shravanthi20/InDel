import { useState, useEffect } from 'react'
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import { ThemeProvider } from './context/ThemeContext'
import PlanStatusDashboard from './pages/PlanStatusDashboard'
import Layout from './components/layout/Sidebar'
import Overview from './pages/Overview'
import LossRatio from './pages/LossRatio'
import Claims from './pages/Claims'
import FraudQueue from './pages/FraudQueue'
import Forecast from './pages/Forecast'
import Register from './pages/Register'

function AppContent() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)

  useEffect(() => {
    const token = localStorage.getItem('token')
    setIsAuthenticated(!!token)
  }, [])

  if (!isAuthenticated) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-slate-50 dark:bg-slate-950 font-['Outfit']">
        <div className="w-full max-w-[400px] p-10 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-2xl">
          <div className="mb-8 text-center">
            <div className="inline-flex h-12 w-12 items-center justify-center rounded bg-[var(--brand-primary)] mb-6 font-black text-white italic text-xl">
              ID
            </div>
            <h1 className="text-2xl font-black text-slate-900 dark:text-white tracking-tight">Secure Terminal</h1>
            <p className="mt-2 text-sm text-slate-500 dark:text-slate-400">Initialize insurer session to core services.</p>
          </div>
          
          <div className="space-y-4">
            <button 
              className="w-full rounded bg-[var(--brand-primary)] p-4 text-sm font-bold text-white transition-none hover:bg-[var(--brand-primary-deep)] active:bg-[var(--brand-primary-deep)] shadow-md"
              onClick={() => {
                localStorage.setItem('token', 'demo-insurer-token')
                setIsAuthenticated(true)
              }}
            >
              Start Session
            </button>
            <div className="flex items-center justify-between px-1">
               <span className="text-[10px] font-black uppercase tracking-widest text-slate-400">Auth: JWT-256</span>
               <span className="text-[10px] font-black uppercase tracking-widest text-slate-400">Status: Ready</span>
            </div>
          </div>
          
          <p className="mt-10 text-center text-[9px] font-black uppercase tracking-[0.2em] text-slate-400 dark:text-slate-600">
             &copy; 2026 InDel Technologies.
          </p>
        </div>
      </div>
    )
  }

  return (
    <Router
      future={{
        v7_startTransition: true,
        v7_relativeSplatPath: true,
      }}
    >
      <Layout>
        <Routes>
          <Route path="/" element={<Overview />} />
          <Route path="/plan-status" element={<PlanStatusDashboard />} />
          <Route path="/loss-ratio" element={<LossRatio />} />
          <Route path="/claims" element={<Claims />} />
          <Route path="/fraud-queue" element={<FraudQueue />} />
          <Route path="/forecast" element={<Forecast />} />
          <Route path="/register" element={<Register />} />
          <Route path="*" element={<Navigate to="/" />} />
        </Routes>
      </Layout>
    </Router>
  )
}

export default function App() {
  return (
    <ThemeProvider>
      <AppContent />
    </ThemeProvider>
  )
}
