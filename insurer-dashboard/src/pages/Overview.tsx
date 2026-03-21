import { useEffect, useState } from 'react'
import { getOverview } from '../api/insurer'

export default function Overview() {
  const [stats, setStats] = useState(null)

  useEffect(() => {
    getOverview().then(res => setStats(res.data)).catch(console.error)
  }, [])

  if (!stats) return <div>Loading...</div>

  return (
    <div className="grid grid-cols-2 gap-6 p-6">
      <div className="bg-white rounded-lg p-6 shadow">
        <h3 className="text-lg font-semibold">Pool Health</h3>
        <p className="text-3xl font-bold mt-2">Excellent</p>
      </div>
      
      <div className="bg-white rounded-lg p-6 shadow">
        <h3 className="text-lg font-semibold">Loss Ratio</h3>
        <p className="text-3xl font-bold mt-2">45%</p>
      </div>
      
      <div className="bg-white rounded-lg p-6 shadow">
        <h3 className="text-lg font-semibold">Total Premiums</h3>
        <p className="text-3xl font-bold mt-2">₹2.5L</p>
      </div>
      
      <div className="bg-white rounded-lg p-6 shadow">
        <h3 className="text-lg font-semibold">Total Payouts</h3>
        <p className="text-3xl font-bold mt-2">₹1.1L</p>
      </div>
    </div>
  )
}
