import { useState, useEffect } from 'react'
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/layout/Sidebar'
import Overview from './pages/Overview'
import LossRatio from './pages/LossRatio'
import Claims from './pages/Claims'
import FraudQueue from './pages/FraudQueue'
import Forecast from './pages/Forecast'
import MaintenanceChecks from './pages/MaintenanceChecks'

export default function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)

  useEffect(() => {
    // Check if user is authenticated (JWT token in localStorage)
    const token = localStorage.getItem('token')
    setIsAuthenticated(!!token)
  }, [])

  if (!isAuthenticated) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <h1 className="text-3xl font-bold mb-4">InDel Insurer Portal</h1>
          <button 
            className="bg-blue-500 text-white px-6 py-2 rounded"
            onClick={() => setIsAuthenticated(true)}
          >
            Login
          </button>
        </div>
      </div>
    )
  }

  return (
    <Router>
      <Layout>
        <Routes>
          <Route path="/" element={<Overview />} />
          <Route path="/loss-ratio" element={<LossRatio />} />
          <Route path="/claims" element={<Claims />} />
          <Route path="/fraud-queue" element={<FraudQueue />} />
          <Route path="/forecast" element={<Forecast />} />
          <Route path="/maintenance-checks" element={<MaintenanceChecks />} />
          <Route path="*" element={<Navigate to="/" />} />
        </Routes>
      </Layout>
    </Router>
  )
}
