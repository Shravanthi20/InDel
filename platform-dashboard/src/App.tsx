import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { ThemeProvider } from './context/ThemeContext'
import Sidebar from './components/layout/Sidebar'
import Overview from './pages/Overview'
import Workers from './pages/Workers'
import Zones from './pages/Zones'
import Analytics from './pages/Analytics'
import Disruptions from './pages/Disruptions'
import WeeklyCycleOps from './pages/WeeklyCycleOps'
import PayoutOps from './pages/PayoutOps'
import ReconciliationOps from './pages/ReconciliationOps'
import SyntheticDataOps from './pages/SyntheticDataOps'
import GodModeLayout from './pages/god-mode/GodModeLayout'
import { GodModeProvider } from './pages/god-mode/state'
import BatchSimulationPage from './pages/god-mode/BatchSimulationPage'

export default function App() {
  return (
    <ThemeProvider>
      <Router
        future={{
          v7_startTransition: true,
          v7_relativeSplatPath: true,
        }}
      >
        <Sidebar>
          <Routes>
            <Route path="/" element={<Overview />} />
            <Route path="/workers" element={<Workers />} />
            <Route path="/zones" element={<Zones />} />
            <Route path="/analytics" element={<Analytics />} />
            <Route path="/weekly-cycle" element={<WeeklyCycleOps />} />
            <Route path="/payout-ops" element={<PayoutOps />} />
            <Route path="/reconciliation" element={<ReconciliationOps />} />
            <Route path="/disruptions" element={<Disruptions />} />
            <Route
              path="/batches"
              element={(
                <GodModeProvider>
                  <GodModeLayout />
                </GodModeProvider>
              )}
            />
            <Route
              path="/synthetic-data"
              element={(
                <GodModeProvider>
                  <SyntheticDataOps />
                </GodModeProvider>
              )}
            />
            <Route
              path="/god-mode"
              element={(
                <GodModeProvider>
                  <GodModeLayout />
                </GodModeProvider>
              )}
            />
            <Route
              path="/god-mode/batch-simulation"
              element={(
                <GodModeProvider>
                  <BatchSimulationPage />
                </GodModeProvider>
              )}
            />
          </Routes>
        </Sidebar>
      </Router>
    </ThemeProvider>
  )
}
