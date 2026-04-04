import { BrowserRouter as Router, Navigate, Route, Routes } from 'react-router-dom'
import Sidebar from './components/layout/Sidebar'
import GodModeLayout from './pages/god-mode/GodModeLayout'
import { GodModeProvider } from './pages/god-mode/state'

export default function App() {
  return (
    <Router>
      <Sidebar>
        <Routes>
          <Route path="/" element={<Navigate to="/god-mode" replace />} />
          <Route
            path="/god-mode"
            element={(
              <GodModeProvider>
                <GodModeLayout />
              </GodModeProvider>
            )}
          />
        </Routes>
      </Sidebar>
    </Router>
  )
}
