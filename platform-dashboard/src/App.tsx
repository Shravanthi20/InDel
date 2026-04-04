import { BrowserRouter as Router, Navigate, Route, Routes } from 'react-router-dom'
import Sidebar from './components/layout/Sidebar'
import GodModeLayout from './pages/god-mode/GodModeLayout'
import TemperaturePage from './pages/god-mode/TemperaturePage'
import RainPage from './pages/god-mode/RainPage'
import AqiPage from './pages/god-mode/AqiPage'
import TrafficPage from './pages/god-mode/TrafficPage'
import ResultsPage from './pages/god-mode/ResultsPage'
import BatchesPage from './pages/god-mode/BatchesPage'
import { GodModeProvider } from './pages/god-mode/state'

export default function App() {
  return (
    <Router>
      <Sidebar>
        <Routes>
          <Route path="/" element={<Navigate to="/god-mode/temperature" replace />} />
          <Route
            path="/god-mode"
            element={(
              <GodModeProvider>
                <GodModeLayout />
              </GodModeProvider>
            )}
          >
            <Route index element={<Navigate to="/god-mode/temperature" replace />} />
            <Route path="temperature" element={<TemperaturePage />} />
            <Route path="rain" element={<RainPage />} />
            <Route path="aqi" element={<AqiPage />} />
            <Route path="traffic" element={<TrafficPage />} />
            <Route path="results" element={<ResultsPage />} />
            <Route path="batches" element={<BatchesPage />} />
          </Route>
        </Routes>
      </Sidebar>
    </Router>
  )
}
