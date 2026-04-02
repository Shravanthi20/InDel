import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import Sidebar from './components/layout/Sidebar'
import Overview from './pages/Overview'
import Workers from './pages/Workers'
import Zones from './pages/Zones'
import Analytics from './pages/Analytics'
import Disruptions from './pages/Disruptions'

export default function App() {
  return (
    <Router>
      <Sidebar>
        <Routes>
          <Route path="/" element={<Overview />} />
          <Route path="/workers" element={<Workers />} />
          <Route path="/zones" element={<Zones />} />
          <Route path="/analytics" element={<Analytics />} />
          <Route path="/disruptions" element={<Disruptions />} />
        </Routes>
      </Sidebar>
    </Router>
  )
}
