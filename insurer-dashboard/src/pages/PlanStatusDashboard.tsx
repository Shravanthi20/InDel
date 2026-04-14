import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { PieChart, Pie, Cell, BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'
import { endUserPlan, getPlanUsers, startUserPlan } from '../api/insurer'
import type { PlanUser } from '../types'

const chartTypes = [
  { label: 'Pie: In Plan vs Not', value: 'pie' },
  { label: 'Bar: Users by Zone', value: 'bar' },
  { label: 'Table: User Details', value: 'table' },
]

const COLORS = ['#0088FE', '#FF8042']

export default function PlanStatusDashboard() {
  const [users, setUsers] = useState<PlanUser[]>([])
  const [chartType, setChartType] = useState('pie')
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [pendingUserId, setPendingUserId] = useState<number | null>(null)

  const loadUsers = useCallback(async (showSkeleton = false) => {
    if (showSkeleton) {
      setLoading(true)
    } else {
      setRefreshing(true)
    }

    try {
      setError(null)
      const result = await getPlanUsers<PlanUser>()
      setUsers(result)
    } catch (err: any) {
      setError(err?.response?.data?.error?.message || err?.message || 'Failed to load plan status users')
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }, [])

  useEffect(() => {
    void loadUsers(true)
  }, [loadUsers])

  const handleTogglePlan = useCallback(async (user: PlanUser) => {
    setPendingUserId(user.id)
    setError(null)

    try {
      const updatedUser = user.status === 'active'
        ? await endUserPlan<PlanUser>(user.id)
        : await startUserPlan<PlanUser>(user.id)

      setUsers((current) => current.map((entry) => (
        entry.id === updatedUser.id ? updatedUser : entry
      )))
    } catch (err: any) {
      setError(err?.response?.data?.error?.message || err?.message || 'Failed to update plan status')
    } finally {
      setPendingUserId(null)
    }
  }, [])

  const pieData = useMemo(() => ([
    { name: 'In Plan', value: users.filter((user) => user.status === 'active').length },
    { name: 'Not in Plan', value: users.filter((user) => user.status !== 'active').length },
  ]), [users])

  const barData = useMemo(() => Object.values(
    users.reduce((acc, user) => {
      const key = user.zone || 'Unknown'
      if (!acc[key]) {
        acc[key] = { zone: key, inPlan: 0, notInPlan: 0 }
      }
      if (user.status === 'active') {
        acc[key].inPlan += 1
      } else {
        acc[key].notInPlan += 1
      }
      return acc
    }, {} as Record<string, { zone: string; inPlan: number; notInPlan: number }>)
  ), [users])

  return (
    <div style={{ padding: 24 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 16, marginBottom: 16, flexWrap: 'wrap' }}>
        <div>
          <h2 style={{ fontWeight: 700, fontSize: 22, marginBottom: 8 }}>Plan Status Analytics</h2>
          <p style={{ color: '#64748b', margin: 0 }}>
            Start or end a worker plan from the same dataset used by the charts below.
          </p>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <label>
            Chart Type:{' '}
            <select value={chartType} onChange={(e) => setChartType(e.target.value)}>
              {chartTypes.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </label>
          <button
            type="button"
            onClick={() => void loadUsers(false)}
            disabled={loading || refreshing}
            style={{
              border: '1px solid #cbd5e1',
              background: '#fff',
              borderRadius: 8,
              padding: '8px 12px',
              cursor: loading || refreshing ? 'not-allowed' : 'pointer',
            }}
          >
            {refreshing ? 'Refreshing...' : 'Refresh'}
          </button>
        </div>
      </div>

      {error && (
        <div style={{ color: '#b91c1c', marginBottom: 16, fontWeight: 600 }}>
          Error: {error}
        </div>
      )}

      <div style={{ marginTop: 24 }}>
        {loading ? (
          <div style={{ marginTop: 32 }}>Loading...</div>
        ) : users.length === 0 ? (
          <div style={{ marginTop: 32, color: '#888' }}>No user data available.</div>
        ) : (
          <>
            {chartType === 'pie' && (
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie
                    data={pieData}
                    dataKey="value"
                    nameKey="name"
                    cx="50%"
                    cy="50%"
                    outerRadius={80}
                    fill="#8884d8"
                    label
                  >
                    {pieData.map((entry, index) => (
                      <Cell key={`${entry.name}-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            )}

            {chartType === 'bar' && (
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={barData} margin={{ top: 20, right: 30, left: 20, bottom: 5 }}>
                  <XAxis dataKey="zone" />
                  <YAxis />
                  <Tooltip />
                  <Bar dataKey="inPlan" stackId="a" fill="#0088FE" name="In Plan" />
                  <Bar dataKey="notInPlan" stackId="a" fill="#FF8042" name="Not in Plan" />
                </BarChart>
              </ResponsiveContainer>
            )}

            <div style={{ marginTop: 16 }}>
              {(chartType === 'table' || chartType === 'pie' || chartType === 'bar') && (
                <div style={{ maxHeight: 420, overflow: 'auto', marginTop: 16 }}>
                  <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                    <thead>
                      <tr>
                        <th style={{ borderBottom: '1px solid #eee', padding: 8, textAlign: 'left' }}>User ID</th>
                        <th style={{ borderBottom: '1px solid #eee', padding: 8, textAlign: 'left' }}>Name</th>
                        <th style={{ borderBottom: '1px solid #eee', padding: 8, textAlign: 'left' }}>Phone</th>
                        <th style={{ borderBottom: '1px solid #eee', padding: 8, textAlign: 'left' }}>Zone</th>
                        <th style={{ borderBottom: '1px solid #eee', padding: 8, textAlign: 'left' }}>Status</th>
                        <th style={{ borderBottom: '1px solid #eee', padding: 8, textAlign: 'left' }}>Plan</th>
                        <th style={{ borderBottom: '1px solid #eee', padding: 8, textAlign: 'left' }}>Action</th>
                      </tr>
                    </thead>
                    <tbody>
                      {users.map((user) => {
                        const isActive = user.status === 'active'
                        const isPending = pendingUserId === user.id

                        return (
                          <tr key={user.id}>
                            <td style={{ padding: 8 }}>{user.id}</td>
                            <td style={{ padding: 8 }}>{user.name || 'Unknown'}</td>
                            <td style={{ padding: 8 }}>{user.phone || '-'}</td>
                            <td style={{ padding: 8 }}>{user.zone || 'Unknown'}</td>
                            <td style={{ padding: 8, fontWeight: 600, color: isActive ? '#0369a1' : '#b45309' }}>
                              {isActive ? 'In Plan' : 'Not in Plan'}
                            </td>
                            <td style={{ padding: 8 }}>{user.plan_id || '-'}</td>
                            <td style={{ padding: 8 }}>
                              <button
                                type="button"
                                onClick={() => void handleTogglePlan(user)}
                                disabled={isPending}
                                style={{
                                  border: 'none',
                                  borderRadius: 8,
                                  padding: '8px 12px',
                                  color: '#fff',
                                  background: isActive ? '#dc2626' : '#0284c7',
                                  cursor: isPending ? 'not-allowed' : 'pointer',
                                  opacity: isPending ? 0.7 : 1,
                                }}
                              >
                                {isPending ? 'Saving...' : isActive ? 'End Plan' : 'Start Plan'}
                              </button>
                            </td>
                          </tr>
                        )
                      })}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  )
}
