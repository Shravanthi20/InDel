import axios from 'axios'
import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import {
  postAddBatches,
  generateClaimsForDisruption,
  getWorkers,
  getAssignedBatches,
  getAvailableBatches,
  getSimulationBatches,
  getDisruptions,
  getOrders,
  getZoneLevels,
  putAcceptBatch,
  putDeliverBatch,
  postExternalSignal,
  getZoneHealth,
  getZonePaths,
  getZones,
  postSimulateOrders,
  postTriggerDemo,
} from '../../api/platform'

export type EnvInputs = {
  temperature: number
  rain: number
  aqi: number
  traffic: number
}

export type PolicyInputs = {
  maxPayoutPerDay: number
  coverageRatio: number
}

export type ZoneLevel = 'ALL' | 'A' | 'B' | 'C'

export type ZoneRecord = {
  zone_id: number
  name: string
  city: string
  state: string
  risk_rating?: number
  active_workers?: number
  areas?: string[]
}

export type ZoneOption = {
  value: string
  label: string
  zoneId: number
}

export type ZonePathOption = ZoneOption & {
  city?: string
  zoneName?: string
  zoneState?: string
  fromCity?: string
  toCity?: string
  state?: string
  fromState?: string
  toState?: string
}

export type ZoneHealth = {
  zone_id: number
  status: string
  order_drop: number
  current_orders: number
  baseline_orders: number
  active_signals: Record<string, boolean>
}

export type BatchOrder = {
  orderId: string
  pickupArea?: string
  dropArea?: string
  deliveryAddress?: string
  deliveryCode?: string
  status?: string
  pickupTime?: string
  deliveryTime?: string
}

export type BatchRow = {
  batchId: string
  zoneLevel?: string
  fromCity?: string
  toCity?: string
  totalWeight?: number
  targetWeight?: number
  orderCount?: number
  status?: string
  pickupCode?: string
  deliveryCode?: string
  orders?: BatchOrder[]
}

export type SimulationResult = {
  status: 'Normal' | 'Reduced Operation' | 'Critical' | 'Severe'
  reason: string
  scopeLabel: string
  actionTaken: string
  riskScore: number
  aqiScore: number
  tempScore: number
  rainScore: number
  trafficScore: number
  payout: number
  cashOutflow: number
  notification: string
  reducedBatchSizeBy: number
  deliveryBufferMinutes: number
  zoneOperational: boolean
  affectedZoneCount: number
  source: 'god-mode-override' | 'api-mock'
}

export type EndpointStatus = {
  zoneHealth: 'ok' | 'failed' | 'pending'
  availableBatches: 'ok' | 'failed' | 'pending'
  assignedBatches: 'ok' | 'failed' | 'pending'
}

export type Notice = {
  tone: 'success' | 'error'
  message: string
}

export type BatchFlowCheckResult = {
  status: 'success' | 'error'
  checkedAt: string
  batchId?: string
  zoneLevel?: string
  pickupMessage?: string
  deliveryMessage?: string
  detail: string
}

export type IntegrationSelfTestCheck = {
  name: string
  status: 'pass' | 'fail' | 'skipped'
  detail: string
}

export type IntegrationSelfTestResult = {
  checkedAt: string
  checks: IntegrationSelfTestCheck[]
  passed: number
  failed: number
  skipped: number
}

export type DisruptionSignal = {
  sent: boolean
  sentAt: string
  scopeLabel: string
  triggerMode: 'all-factors-combined'
  zonesCount: number
  successfulRequests: number
  claimsCreated: number
  notificationsCreated: number
}

const DEFAULT_ENV_INPUTS: EnvInputs = {
  temperature: 33,
  rain: 3,
  aqi: 190,
  traffic: 58,
}

const DEFAULT_POLICY_INPUTS: PolicyInputs = {
  maxPayoutPerDay: 2000,
  coverageRatio: 0.8,
}

function clamp01(value: number) {
  return Math.max(0, Math.min(1, value))
}

function clampRange(value: number, min: number, max: number) {
  return Math.max(min, Math.min(max, value))
}

function round2(value: number) {
  return Math.round(value * 100) / 100
}

function normalizeText(value: string) {
  return value.trim().toLowerCase()
}

function isAuthDenied(error: unknown) {
  return axios.isAxiosError(error) && (error.response?.status === 401 || error.response?.status === 403)
}

function mapApiToInputs(zoneHealths: ZoneHealth[]): EnvInputs {
  if (!zoneHealths.length) {
    return DEFAULT_ENV_INPUTS
  }

  const primary = zoneHealths[0]
  const orderDrop = clamp01(primary.order_drop)
  const signalKeys = Object.keys(primary.active_signals || {})

  const hasWeather = signalKeys.some((sig) => sig.toLowerCase().includes('weather') || sig.toLowerCase().includes('rain'))
  const hasAqi = signalKeys.some((sig) => sig.toLowerCase().includes('aqi') || sig.toLowerCase().includes('air'))
  const hasTraffic = signalKeys.some((sig) => sig.toLowerCase().includes('traffic') || sig.toLowerCase().includes('road'))

  return {
    temperature: round2(clampRange(30 + orderDrop * 16, 20, 50)),
    rain: round2(clampRange((hasWeather ? 8 : 2) + orderDrop * 10, 0, 20)),
    aqi: Math.round(clampRange((hasAqi ? 230 : 165) + orderDrop * 120, 80, 350)),
    traffic: Math.round(clampRange((hasTraffic ? 70 : 52) + orderDrop * 35, 0, 100)),
  }
}

function parseZonePaths(level: ZoneLevel, body: any): ZonePathOption[] {
  const payload = body?.data ?? body
  const options: ZonePathOption[] = []

  if (level === 'A' && Array.isArray(payload?.cities)) {
    payload.cities.forEach((city: unknown) => {
      const cityName = typeof city === 'string' ? city.trim() : String(city ?? '').trim()
      if (!cityName) {
        return
      }
      options.push({
        value: cityName,
        label: cityName,
        zoneId: 0,
        city: cityName,
        zoneName: cityName,
      })
    })
  }

  const zones = payload?.zones ?? payload?.zonePairs ?? payload?.zone_pairs ?? payload?.cityPairs ?? payload?.city_pairs ?? []
  if (Array.isArray(zones)) {
    zones.forEach((zone: any) => {
      const zoneName = String(zone?.zone_name ?? zone?.zoneName ?? zone?.name ?? zone?.city ?? '').trim()
      const zoneState = String(zone?.zone_state ?? zone?.zoneState ?? zone?.state ?? '').trim()
      const zoneCity = String(zone?.city ?? '').trim()

      if (!zoneName && !zoneCity) {
        return
      }

      const label = zoneState && zoneName ? `${zoneState} - ${zoneName}` : (zoneName || zoneCity)
      options.push({
        value: label,
        label,
        zoneId: Number(zone?.zone_id ?? zone?.zoneId ?? 0),
        city: zoneCity || zoneName || undefined,
        zoneName: zoneName || undefined,
        zoneState: zoneState || undefined,
        fromCity: zoneName || undefined,
        toCity: zoneName || undefined,
        state: zoneState || undefined,
        fromState: zoneState || undefined,
        toState: zoneState || undefined,
      })
    })
  }

  if (Array.isArray(payload?.paths) && options.length === 0) {
    payload.paths.forEach((path: any) => {
      const label = String(path?.displayName ?? '').trim()
      if (!label) {
        return
      }
      options.push({
        value: label,
        label,
        zoneId: 0,
        city: String(path?.city ?? '').trim() || undefined,
        fromCity: String(path?.fromCity ?? '').trim() || undefined,
        toCity: String(path?.toCity ?? '').trim() || undefined,
      })
    })
  }

  return Array.from(new Map(options.map((option) => [option.value, option])).values()).slice(0, 10)
}

function zoneMatchesAnyField(zone: ZoneRecord, tokens: string[]) {
  const fields = [zone.name, zone.city, zone.state].map(normalizeText)
  return tokens.some((token) => token && fields.some((field) => field.includes(token)))
}

function resolveTargetZoneIds(zoneLevel: ZoneLevel, zoneName: string, zones: ZoneRecord[], zonePaths: ZonePathOption[]) {
  if (zones.length === 0) {
    return []
  }

  if (zoneLevel === 'ALL' || zoneName === 'ALL ZONES') {
    return zones.map((zone) => zone.zone_id)
  }

  const selectedPath = zonePaths.find((option) => option.value === zoneName)
  if (selectedPath) {
    const tokens = [
      selectedPath.zoneName,
      selectedPath.zoneState,
      selectedPath.city,
      selectedPath.fromCity,
      selectedPath.toCity,
      selectedPath.state,
      selectedPath.fromState,
      selectedPath.toState,
    ]
      .map((token) => normalizeText(token ?? ''))
      .filter(Boolean)

    const matchedZones = zones.filter((zone) => zoneMatchesAnyField(zone, tokens))
    if (matchedZones.length > 0) {
      return matchedZones.map((zone) => zone.zone_id)
    }
  }

  const fallbackTokens = [zoneName, zoneName.replace(/^zone\s+/i, '')].map(normalizeText).filter(Boolean)
  const matchedZones = zones.filter((zone) => zoneMatchesAnyField(zone, fallbackTokens))
  return matchedZones.map((zone) => zone.zone_id)
}

function computeResult(
  inputs: EnvInputs,
  policy: PolicyInputs,
  scopeLabel: string,
  affectedZoneCount: number,
  source: SimulationResult['source'],
): SimulationResult {
  const aqiScore = clamp01((inputs.aqi - 200) / 100)
  const tempScore = clamp01((inputs.temperature - 35) / 10)
  const rainScore = clamp01(inputs.rain / 15)
  const trafficScore = clamp01((inputs.traffic - 60) / 25)

  const riskScore = round2((0.3 * rainScore) + (0.25 * tempScore) + (0.25 * aqiScore) + (0.2 * trafficScore))
  const payout = riskScore < 0.3 ? 0 : round2(policy.maxPayoutPerDay * clamp01(policy.coverageRatio) * riskScore)
  const cashOutflow = payout

  const contributors = [
    { label: 'Rain intensity', value: rainScore * 0.3 },
    { label: 'Temperature', value: tempScore * 0.25 },
    { label: 'AQI', value: aqiScore * 0.25 },
    { label: 'Traffic congestion', value: trafficScore * 0.2 },
  ]

  const top = contributors.reduce((winner, current) => (current.value > winner.value ? current : winner), contributors[0])

  let status: SimulationResult['status'] = 'Normal'
  if (riskScore >= 0.75) {
    status = 'Severe'
  } else if (riskScore >= 0.5) {
    status = 'Critical'
  } else if (riskScore >= 0.3) {
    status = 'Reduced Operation'
  }

  const reducedBatchSizeBy = status === 'Severe' ? 50 : status === 'Critical' ? 35 : status === 'Reduced Operation' ? 15 : 0
  const deliveryBufferMinutes = status === 'Severe' ? 35 : status === 'Critical' ? 20 : status === 'Reduced Operation' ? 10 : 0

  return {
    status,
    reason: `${top.label} is the dominant disruption trigger`,
    scopeLabel,
    actionTaken: status === 'Severe' ? 'stopped' : status === 'Critical' ? 'reduced' : status === 'Reduced Operation' ? 'reduced' : 'monitored',
    riskScore,
    aqiScore: round2(aqiScore),
    tempScore: round2(tempScore),
    rainScore: round2(rainScore),
    trafficScore: round2(trafficScore),
    payout,
    cashOutflow,
    notification: riskScore >= 0.5
      ? `${scopeLabel}: disruption detected, action taken: ${status === 'Severe' ? 'stopped' : 'reduced'} and claims/payout updated.`
      : `${scopeLabel}: no active disruption, monitoring only.`,
    reducedBatchSizeBy,
    deliveryBufferMinutes,
    zoneOperational: status !== 'Severe',
    affectedZoneCount,
    source,
  }
}

function deriveDisruptionSignal(inputs: EnvInputs) {
  const temperatureRisk = clamp01((inputs.temperature - 35) / 10)
  const rainRisk = clamp01(inputs.rain / 15)
  const aqiRisk = clamp01((inputs.aqi - 200) / 100)
  const trafficRisk = clamp01((inputs.traffic - 60) / 25)

  const weights = [
    { key: 'temperature' as const, score: temperatureRisk, signal: 'temperature_extreme' },
    { key: 'rain' as const, score: rainRisk, signal: 'weather_rain' },
    { key: 'aqi' as const, score: aqiRisk, signal: 'aqi_hazardous' },
    { key: 'traffic' as const, score: trafficRisk, signal: 'traffic_congestion' },
  ]

  const top = weights.reduce((winner, current) => (current.score > winner.score ? current : winner), weights[0])
  const compositeScore = round2((temperatureRisk * 0.3) + (rainRisk * 0.25) + (aqiRisk * 0.25) + (trafficRisk * 0.2))

  return {
    signal: top.signal,
    dominantFactor: top.key,
    compositeScore,
    details: {
      temperatureRisk: round2(temperatureRisk),
      rainRisk: round2(rainRisk),
      aqiRisk: round2(aqiRisk),
      trafficRisk: round2(trafficRisk),
    },
  }
}

export function formatCurrency(value: number) {
  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency: 'INR',
    maximumFractionDigits: 0,
  }).format(Math.max(0, value))
}

export function pickupCodeFromBatchId(batchId: string) {
  let seed = 0
  for (const char of batchId.trim().toUpperCase()) {
    seed = (seed * 31 + char.charCodeAt(0)) % 9000
  }
  return String(1000 + seed).padStart(4, '0')
}

export function deliveryCodeFromBatchId(batchId: string) {
  let seed = 7
  for (const char of batchId.trim().toUpperCase()) {
    seed = (seed * 37 + char.charCodeAt(0)) % 9000
  }
  return String(1000 + seed).padStart(4, '0')
}

export function deliveryCodeFromOrderId(orderId: string) {
  let seed = 11
  for (const char of orderId.trim().toUpperCase()) {
    seed = (seed * 41 + char.charCodeAt(0)) % 9000
  }
  return String(1000 + seed).padStart(4, '0')
}

function normalizeBatchWorkflowStatus(status?: string) {
  return (status || '').trim().toLowerCase().replace(/\s+/g, '_')
}

function applyWorkerBatchSnapshot(
  availableBatchRows: BatchRow[],
  assignedBatchRows: BatchRow[],
  setAvailableBatches: (batches: BatchRow[]) => void,
  setAssignedBatches: (batches: BatchRow[]) => void,
) {
  setAvailableBatches(availableBatchRows)
  setAssignedBatches(assignedBatchRows)
}

type GodModeContextValue = {
  godModeEnabled: boolean
  setGodModeEnabled: (enabled: boolean) => void
  zoneLevel: ZoneLevel
  setZoneLevel: (value: ZoneLevel) => void
  zoneName: string
  setZoneName: (value: string) => void
  zoneNameOptions: ZoneOption[]
  zoneNameLoading: boolean
  affectedZoneIds: number[]
  scopeLabel: string
  apiInputs: EnvInputs
  manualInputs: EnvInputs
  policyInputs: PolicyInputs
  setManualInput: (key: keyof EnvInputs, value: number) => void
  setPolicyInput: (key: keyof PolicyInputs, value: number) => void
  loading: boolean
  running: boolean
  addingDisruption: boolean
  error: string
  result: SimulationResult
  runSimulation: () => void
  addDisruption: () => Promise<void>
  generatingBatches: boolean
  generateBatches: () => void
  lastDisruptionSignal: DisruptionSignal | null
  notice: Notice | null
  previewResult: SimulationResult
  clearNotice: () => void
  zones: ZoneRecord[]
  availableBatches: BatchRow[]
  assignedBatches: BatchRow[]
  batches: BatchRow[]
  showCodes: boolean
  setShowCodes: (value: boolean) => void
  endpointStatus: EndpointStatus
  checkingBatchFlow: boolean
  runBatchFlowCheck: () => Promise<void>
  lastBatchFlowCheck: BatchFlowCheckResult | null
  checkingIntegrationSelfTest: boolean
  runIntegrationSelfTest: () => Promise<void>
  integrationSelfTestResult: IntegrationSelfTestResult | null
}

const GodModeContext = createContext<GodModeContextValue | null>(null)

export function GodModeProvider({ children }: { children: ReactNode }) {
  const [godModeEnabled, setGodModeEnabled] = useState(false)
  const [zoneLevel, setZoneLevel] = useState<ZoneLevel>('ALL')
  const [zoneName, setZoneName] = useState('ALL ZONES')
  const [zonePaths, setZonePaths] = useState<ZonePathOption[]>([])
  const [zoneNameLoading, setZoneNameLoading] = useState(false)
  const [apiInputs, setApiInputs] = useState<EnvInputs>(DEFAULT_ENV_INPUTS)
  const [manualInputs, setManualInputs] = useState<EnvInputs>(DEFAULT_ENV_INPUTS)
  const [policyInputs, setPolicyInputs] = useState<PolicyInputs>(DEFAULT_POLICY_INPUTS)
  const [zones, setZones] = useState<ZoneRecord[]>([])
  const [availableBatches, setAvailableBatches] = useState<BatchRow[]>([])
  const [assignedBatches, setAssignedBatches] = useState<BatchRow[]>([])
  const [showCodes, setShowCodes] = useState(true)
  const [loading, setLoading] = useState(true)
  const [running, setRunning] = useState(false)
  const [addingDisruption, setAddingDisruption] = useState(false)
  const [generatingBatches, setGeneratingBatches] = useState(false)
  const [checkingBatchFlow, setCheckingBatchFlow] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState<Notice | null>(null)
  const [lastBatchFlowCheck, setLastBatchFlowCheck] = useState<BatchFlowCheckResult | null>(null)
  const [checkingIntegrationSelfTest, setCheckingIntegrationSelfTest] = useState(false)
  const [integrationSelfTestResult, setIntegrationSelfTestResult] = useState<IntegrationSelfTestResult | null>(null)
  const [lastDisruptionSignal, setLastDisruptionSignal] = useState<DisruptionSignal | null>(null)
  const [endpointStatus, setEndpointStatus] = useState<EndpointStatus>({
    zoneHealth: 'pending',
    availableBatches: 'pending',
    assignedBatches: 'pending',
  })
  const [committedResult, setCommittedResult] = useState<SimulationResult>(
    computeResult(DEFAULT_ENV_INPUTS, DEFAULT_POLICY_INPUTS, 'ALL ZONES', 0, 'api-mock'),
  )

  const zoneNameOptions = useMemo(() => {
    if (zoneLevel === 'ALL') {
      return [{ value: 'ALL ZONES', label: 'ALL ZONES', zoneId: 0 }]
    }

    if (zoneNameLoading) {
      return [{ value: '', label: 'Loading zone names...', zoneId: 0 }]
    }

    if (zonePaths.length === 0) {
      return [{ value: '', label: 'No zone names available', zoneId: 0 }]
    }

    return [{ value: '', label: 'Select Zone Name', zoneId: 0 }, ...zonePaths]
  }, [zoneLevel, zoneNameLoading, zonePaths])

  const affectedZoneIds = useMemo(() => zones.map((zone) => zone.zone_id), [zones])

  const scopeLabel = useMemo(() => 'ALL ZONES', [])

  const batches = useMemo(() => {
    const merged = [...availableBatches, ...assignedBatches]
    const byId = new Map<string, BatchRow>()
    merged.forEach((batch) => byId.set(batch.batchId, batch))
    return Array.from(byId.values())
  }, [availableBatches, assignedBatches])

  useEffect(() => {
    let mounted = true

    async function loadData() {
      setLoading(true)
      setError('')
      setEndpointStatus({
        zoneHealth: 'pending',
        availableBatches: 'pending',
        assignedBatches: 'pending',
      })

      const [zonesResponse, zoneHealthResponse, availableResponse, assignedResponse] = await Promise.allSettled([
        getZones(),
        getZoneHealth(),
        getAvailableBatches(),
        getAssignedBatches(),
      ])

      if (!mounted) {
        return
      }

      setEndpointStatus({
        zoneHealth: zoneHealthResponse.status === 'fulfilled' ? 'ok' : 'failed',
        availableBatches: availableResponse.status === 'fulfilled' ? 'ok' : 'failed',
        assignedBatches: assignedResponse.status === 'fulfilled' ? 'ok' : 'failed',
      })

      if (zonesResponse.status === 'fulfilled') {
        setZones((zonesResponse.value.data?.zones || []) as ZoneRecord[])
      }

      if (zoneHealthResponse.status === 'fulfilled') {
        const zoneHealths = (zoneHealthResponse.value.data?.data || []) as ZoneHealth[]
        const mapped = mapApiToInputs(zoneHealths)
        setApiInputs(mapped)
        setManualInputs(mapped)
        setCommittedResult(computeResult(mapped, policyInputs, 'ALL ZONES', 0, 'api-mock'))
      }

      if (availableResponse.status === 'fulfilled' || assignedResponse.status === 'fulfilled') {
        const availableRows = availableResponse.status === 'fulfilled'
          ? ((availableResponse.value.data?.batches || []) as BatchRow[])
          : []
        const assignedRows = assignedResponse.status === 'fulfilled'
          ? ((assignedResponse.value.data?.batches || []) as BatchRow[])
          : []
        applyWorkerBatchSnapshot(availableRows, assignedRows, setAvailableBatches, setAssignedBatches)
      }

      if (zoneHealthResponse.status === 'rejected' && availableResponse.status === 'rejected' && assignedResponse.status === 'rejected' && zonesResponse.status === 'rejected') {
        setError('Unable to load God Mode context')
      }

      setLoading(false)
    }

    loadData()

    return () => {
      mounted = false
    }
  }, [])

  useEffect(() => {
    let mounted = true

    async function loadZonePathsForLevel(level: ZoneLevel) {
      if (level === 'ALL') {
        setZonePaths([])
        setZoneNameLoading(false)
        return
      }

      setZoneNameLoading(true)
      try {
        const response = await getZonePaths(level.toLowerCase() as 'a' | 'b' | 'c')
        if (!mounted) {
          return
        }

        setZonePaths(parseZonePaths(level, response.data))
      } catch {
        if (mounted) {
          setZonePaths([])
        }
      } finally {
        if (mounted) {
          setZoneNameLoading(false)
        }
      }
    }

    loadZonePathsForLevel(zoneLevel)

    return () => {
      mounted = false
    }
  }, [zoneLevel])

  useEffect(() => {
    if (!godModeEnabled) {
      setCommittedResult(computeResult(apiInputs, policyInputs, scopeLabel, affectedZoneIds.length, 'api-mock'))
    }
  }, [apiInputs, affectedZoneIds.length, godModeEnabled, policyInputs, scopeLabel])

  useEffect(() => {
    if (!notice) {
      return
    }

    const timeoutId = setTimeout(() => {
      setNotice(null)
    }, 4500)

    return () => {
      clearTimeout(timeoutId)
    }
  }, [notice])

  const livePreview = useMemo(() => {
    const source: SimulationResult['source'] = godModeEnabled ? 'god-mode-override' : 'api-mock'
    return computeResult(godModeEnabled ? manualInputs : apiInputs, policyInputs, scopeLabel, affectedZoneIds.length, source)
  }, [affectedZoneIds.length, apiInputs, godModeEnabled, manualInputs, policyInputs, scopeLabel])

  const setManualInput = (key: keyof EnvInputs, value: number) => {
    setManualInputs((current) => ({ ...current, [key]: value }))
  }

  const setPolicyInput = (key: keyof PolicyInputs, value: number) => {
    setPolicyInputs((current) => ({ ...current, [key]: value }))
  }

  const handleZoneLevelChange = (value: ZoneLevel) => {
    setZoneLevel(value)
    if (value === 'ALL') {
      setZoneName('ALL ZONES')
      return
    }

    setZoneName('')
  }

  const refreshBatchContext = async () => {
    const [zonesResponse, availableResponse, assignedResponse] = await Promise.allSettled([
      getZones(),
      getAvailableBatches(),
      getAssignedBatches(),
    ])

    if (zonesResponse.status === 'fulfilled') {
      setZones((zonesResponse.value.data?.zones || []) as ZoneRecord[])
    }
    setEndpointStatus((current) => ({
      ...current,
      availableBatches: availableResponse.status === 'fulfilled' ? 'ok' : 'failed',
      assignedBatches: assignedResponse.status === 'fulfilled' ? 'ok' : 'failed',
    }))
    if (availableResponse.status === 'fulfilled' || assignedResponse.status === 'fulfilled') {
      const availableRows = availableResponse.status === 'fulfilled'
        ? ((availableResponse.value.data?.batches || []) as BatchRow[])
        : []
      const assignedRows = assignedResponse.status === 'fulfilled'
        ? ((assignedResponse.value.data?.batches || []) as BatchRow[])
        : []
      applyWorkerBatchSnapshot(availableRows, assignedRows, setAvailableBatches, setAssignedBatches)
    }
  }

  const generateBatches = async () => {
    if (generatingBatches) {
      return
    }

    setGeneratingBatches(true)
    setError('')
    setNotice(null)

    const selectedPath = zonePaths.find((option) => option.value === zoneName)
    const fromCity = selectedPath?.zoneName ?? selectedPath?.fromCity ?? selectedPath?.city ?? zones[0]?.city ?? 'Tambaram'
    const toCity = selectedPath?.zoneName ?? selectedPath?.toCity ?? selectedPath?.city ?? zones[0]?.name ?? 'Velachery'
    const fromState = selectedPath?.zoneState ?? selectedPath?.fromState ?? selectedPath?.state ?? zones[0]?.state ?? ''
    const toState = selectedPath?.zoneState ?? selectedPath?.toState ?? selectedPath?.state ?? zones[0]?.state ?? ''
    const targetZoneId = affectedZoneIds[0] ?? zones[0]?.zone_id ?? 1

    try {
      let successMessage = 'Generate Fake Orders completed.'

      if (zoneLevel === 'ALL') {
        const response = await postSimulateOrders({ count: 6 })
        const createdOrders = Number(response.data?.data?.count ?? response.data?.count ?? 0)
        successMessage = createdOrders > 0
          ? `Generated ${createdOrders} fake orders across all zones.`
          : 'Generated fake orders across all zones.'
      } else {
        const response = await postAddBatches({
          count: 6,
          zone_id: targetZoneId,
          zone_level: zoneLevel,
          from_city: fromCity,
          to_city: toCity,
          from_state: fromState || undefined,
          to_state: toState || undefined,
        })

        const createdOrders = Number(response.data?.created_orders ?? 0)
        const estimatedBatches = Number(response.data?.estimated_batches ?? 0)
        successMessage = createdOrders > 0
          ? `Generated ${createdOrders} fake orders and formed ${estimatedBatches || 'new'} optimized batches.`
          : 'Generated fake orders for the selected zone.'
      }

      await refreshBatchContext()
      setNotice({ tone: 'success', message: successMessage })
    } catch (batchErr) {
      const message = batchErr instanceof Error ? batchErr.message : 'Unable to generate batches'
      setError(message)
      setNotice({ tone: 'error', message })
    } finally {
      setGeneratingBatches(false)
    }
  }

  const runSimulation = async () => {
    setRunning(true)
    setError('')

    try {
      const preview = livePreview
      const targetZoneIds = zones.map((zone) => zone.zone_id)

      if (targetZoneIds.length === 0) {
        setError('No zones available for simulation.')
        setCommittedResult(preview)
        return
      }

      const signal = preview.riskScore >= 0.75
        ? 'system_failure'
        : preview.aqiScore >= preview.rainScore && preview.aqiScore >= preview.trafficScore
          ? 'aqi_hazardous'
          : preview.rainScore >= preview.trafficScore
            ? 'weather_rain'
            : 'traffic_congestion'

      await Promise.allSettled(
        targetZoneIds.map((zoneId) => postTriggerDemo({
          zone_id: zoneId,
          force_order_drop: preview.riskScore >= 0.3,
          external_signal: signal,
        })),
      )

      const disruptionsResponse = await getDisruptions()
      const disruptionRows = (disruptionsResponse.data?.data || []) as Array<{ disruption_id: string; zone_id: string; type: string; severity: string }>

      const claimRequests = targetZoneIds.map((zoneId) => {
        const latest = disruptionRows.find((row) => {
          const parsedZoneId = Number(String(row.zone_id).replace(/\D+/g, ''))
          return parsedZoneId === zoneId
        })
        return latest ? generateClaimsForDisruption(Number(String(latest.disruption_id).replace(/\D+/g, ''))) : Promise.resolve(null)
      })

      await Promise.allSettled(claimRequests)

      const [zonesResponse, zoneHealthResponse, availableResponse, assignedResponse] = await Promise.allSettled([
        getZones(),
        getZoneHealth(),
        getAvailableBatches(),
        getAssignedBatches(),
      ])

      if (zonesResponse.status === 'fulfilled') {
        setZones((zonesResponse.value.data?.zones || []) as ZoneRecord[])
      }
      if (zoneHealthResponse.status === 'fulfilled') {
        const zoneHealths = (zoneHealthResponse.value.data?.data || []) as ZoneHealth[]
        const mapped = mapApiToInputs(zoneHealths)
        setApiInputs(mapped)
        setManualInputs(mapped)
      }
      if (availableResponse.status === 'fulfilled' || assignedResponse.status === 'fulfilled') {
        const availableRows = availableResponse.status === 'fulfilled'
          ? ((availableResponse.value.data?.batches || []) as BatchRow[])
          : []
        const assignedRows = assignedResponse.status === 'fulfilled'
          ? ((assignedResponse.value.data?.batches || []) as BatchRow[])
          : []
        applyWorkerBatchSnapshot(availableRows, assignedRows, setAvailableBatches, setAssignedBatches)
      }

      setCommittedResult(preview)
    } catch (runErr) {
      setError(runErr instanceof Error ? runErr.message : 'Unable to run simulation')
      setCommittedResult(livePreview)
    } finally {
      setRunning(false)
    }
  }

  const runBatchFlowCheck = async () => {
    if (checkingBatchFlow) {
      return
    }

    setCheckingBatchFlow(true)
    setError('')
    setNotice(null)

    try {
      const availableResponse = await getAvailableBatches()
      const assignedResponse = await getAssignedBatches()
      let freshAvailableBatches = (availableResponse.data?.batches || []) as BatchRow[]
      const freshAssignedBatches = (assignedResponse.data?.batches || []) as BatchRow[]
      applyWorkerBatchSnapshot(freshAvailableBatches, freshAssignedBatches, setAvailableBatches, setAssignedBatches)
      let freshAvailable = freshAvailableBatches.filter((batch) => normalizeBatchWorkflowStatus(batch.status) === 'assigned')

      let candidate = freshAvailable.find((batch) => (batch.orders?.length || 0) > 0)
      if (!candidate) {
        // Try to self-heal: seed orders and batches, then retry once.
        await generateBatches()
        const retryAvailableResponse = await getAvailableBatches()
        const retryAssignedResponse = await getAssignedBatches()
        freshAvailableBatches = (retryAvailableResponse.data?.batches || []) as BatchRow[]
        const retryAssignedBatches = (retryAssignedResponse.data?.batches || []) as BatchRow[]
        applyWorkerBatchSnapshot(freshAvailableBatches, retryAssignedBatches, setAvailableBatches, setAssignedBatches)
        freshAvailable = freshAvailableBatches.filter((batch) => normalizeBatchWorkflowStatus(batch.status) === 'assigned')
        candidate = freshAvailable.find((batch) => (batch.orders?.length || 0) > 0)
      }

      if (!candidate) {
        const detail = 'No available batch found for flow check even after auto-seeding.'
        setNotice({ tone: 'error', message: detail })
        setLastBatchFlowCheck({
          status: 'error',
          checkedAt: new Date().toISOString(),
          detail,
        })
        return
      }

      const batchId = candidate.batchId
      const orderIds = (candidate.orders || []).map((order) => order.orderId).filter(Boolean)
      if (orderIds.length === 0) {
        const detail = `Batch ${batchId} has no orders to validate.`
        setNotice({ tone: 'error', message: detail })
        setLastBatchFlowCheck({
          status: 'error',
          checkedAt: new Date().toISOString(),
          batchId,
          zoneLevel: candidate.zoneLevel,
          detail,
        })
        return
      }

      const pickupCode = candidate.pickupCode || pickupCodeFromBatchId(batchId)
      const pickupRes = await putAcceptBatch(batchId, { orderIds, pickupCode })

      const isZoneA = (candidate.zoneLevel || '').trim().toUpperCase() === 'A'
      const deliveryCode = isZoneA
        ? (candidate.orders || []).find((order) => order.orderId === orderIds[0])?.deliveryCode || deliveryCodeFromOrderId(orderIds[0])
        : candidate.deliveryCode || deliveryCodeFromBatchId(batchId)
      const deliverRes = await putDeliverBatch(batchId, { deliveryCode })

      await refreshBatchContext()

      const pickupMessage = String(pickupRes.data?.message || 'batch_picked_up')
      const deliverMessage = String(deliverRes.data?.message || 'batch_delivered')
      setNotice({
        tone: 'success',
        message: `Batch flow check passed for ${batchId}: ${pickupMessage} -> ${deliverMessage}.`,
      })
      setLastBatchFlowCheck({
        status: 'success',
        checkedAt: new Date().toISOString(),
        batchId,
        zoneLevel: candidate.zoneLevel,
        pickupMessage,
        deliveryMessage: deliverMessage,
        detail: `Batch flow check passed for ${batchId}.`,
      })
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Batch flow check failed'
      setError(message)
      setNotice({ tone: 'error', message })
      setLastBatchFlowCheck({
        status: 'error',
        checkedAt: new Date().toISOString(),
        detail: message,
      })
    } finally {
      setCheckingBatchFlow(false)
    }
  }

  const runIntegrationSelfTest = async () => {
    if (checkingIntegrationSelfTest) {
      return
    }

    setCheckingIntegrationSelfTest(true)
    setError('')
    setNotice(null)

    const checks: IntegrationSelfTestCheck[] = []

    const runCheck = async (name: string, action: () => Promise<unknown>) => {
      try {
        await action()
        checks.push({ name, status: 'pass', detail: '200 OK' })
      } catch (err) {
        let detail = err instanceof Error ? err.message : 'request failed'
        if (axios.isAxiosError(err) && err.response) {
          detail = `${err.response.status} ${err.response.statusText || ''}`.trim()
        }
        checks.push({ name, status: 'fail', detail })
      }
    }

    try {
      await runCheck('GET /api/v1/platform/workers', async () => { await getWorkers() })
      await runCheck('GET /api/v1/platform/zones', async () => { await getZones() })
      await runCheck('GET /api/v1/platform/zone-levels', async () => { await getZoneLevels() })
      await runCheck('GET /api/v1/platform/zone-paths?type=a', async () => { await getZonePaths('a') })
      await runCheck('GET /api/v1/platform/zones/health', async () => { await getZoneHealth() })
      await runCheck('GET /api/v1/platform/disruptions', async () => { await getDisruptions() })
      await runCheck('GET /api/v1/worker/orders', async () => { await getOrders() })
      await runCheck('GET /api/v1/worker/batches', async () => { await getAvailableBatches() })
      await runCheck('GET /api/v1/worker/batches/assigned', async () => { await getAssignedBatches() })
      await runCheck('GET /api/v1/demo/batches', async () => { await getSimulationBatches() })
      await runCheck('POST /api/v1/demo/simulate-orders', async () => { await postSimulateOrders({ count: 1 }) })
      await runCheck('POST /api/v1/platform/demo/add-batches', async () => {
        await postAddBatches({
          count: 1,
          zone_id: zones[0]?.zone_id || 1,
          zone_level: 'B',
          from_city: 'Chennai',
          to_city: 'Bangalore',
          from_state: 'Tamil Nadu',
          to_state: 'Karnataka',
        })
      })
      await runCheck('POST /api/v1/platform/demo/trigger-disruption', async () => {
        await postTriggerDemo({
          zone_id: zones[0]?.zone_id || 1,
          force_order_drop: false,
          external_signal: 'traffic_congestion',
        })
      })
      await runCheck('POST /api/v1/platform/webhooks/external-signal', async () => {
        await postExternalSignal({
          zone_id: zones[0]?.zone_id || 1,
          source: 'platform_dashboard_self_test',
          status: 'active',
        })
      })

      try {
        const disruptions = await getDisruptions()
        const disruptionRows = disruptions.data?.data || []
        const disruptionId = Number(disruptionRows[0]?.disruption_id || 0)
        if (Number.isFinite(disruptionId) && disruptionId > 0) {
          await runCheck(`POST /api/v1/internal/claims/generate-for-disruption/${disruptionId}`, async () => {
            await generateClaimsForDisruption(disruptionId)
          })
        } else {
          checks.push({
            name: 'POST /api/v1/internal/claims/generate-for-disruption/:id',
            status: 'skipped',
            detail: 'No disruption id available',
          })
        }
      } catch (err) {
        let detail = err instanceof Error ? err.message : 'unable to read disruptions'
        if (axios.isAxiosError(err) && err.response) {
          detail = `${err.response.status} ${err.response.statusText || ''}`.trim()
        }
        checks.push({
          name: 'POST /api/v1/internal/claims/generate-for-disruption/:id',
          status: 'fail',
          detail,
        })
      }

      const passed = checks.filter((c) => c.status === 'pass').length
      const failed = checks.filter((c) => c.status === 'fail').length
      const skipped = checks.filter((c) => c.status === 'skipped').length

      setIntegrationSelfTestResult({
        checkedAt: new Date().toISOString(),
        checks,
        passed,
        failed,
        skipped,
      })

      await refreshBatchContext()

      if (failed === 0) {
        setNotice({ tone: 'success', message: `Integration self-test passed (${passed} checks, ${skipped} skipped).` })
      } else {
        setNotice({ tone: 'error', message: `Integration self-test found ${failed} failed checks.` })
      }
    } finally {
      setCheckingIntegrationSelfTest(false)
    }
  }

  const addDisruption = async () => {
  if (addingDisruption) {
    return
  }

  setAddingDisruption(true)
  setError('')

  try {
    const sourceInputs = godModeEnabled ? manualInputs : apiInputs
    const targetZoneIds = zones.map((zone) => zone.zone_id)

    if (targetZoneIds.length === 0) {
    setError('No zones available to add disruption.')
    return
    }

    const signalPlan = deriveDisruptionSignal(sourceInputs)
    const results = await Promise.allSettled(
    targetZoneIds.map((zoneId) => postTriggerDemo({
      zone_id: zoneId,
      force_order_drop: true,
      external_signal: 'composite_all_factors',
      generate_claims: true,
      aqi: sourceInputs.aqi,
      rain: sourceInputs.rain,
      traffic: sourceInputs.traffic,
      max_payout_inr: policyInputs.maxPayoutPerDay,
      max_payout_per_day: policyInputs.maxPayoutPerDay,
      coverage_ratio: policyInputs.coverageRatio,
      temperature: sourceInputs.temperature,
    })),
    )

    let totalClaims = 0
    let totalNotifications = 0
    let successfulRequests = 0
    results.forEach((entry) => {
    if (entry.status !== 'fulfilled') {
      return
    }
    successfulRequests += 1
    const payload = entry.value.data?.data
    totalClaims += Number(payload?.claims_generated || 0)
    totalNotifications += Number(payload?.notifications_created || 0)
    })

    const [zoneHealthResponse, availableResponse, assignedResponse] = await Promise.allSettled([
    getZoneHealth(),
    getAvailableBatches(),
    getAssignedBatches(),
    ])

    if (zoneHealthResponse.status === 'fulfilled') {
    const zoneHealths = (zoneHealthResponse.value.data?.data || []) as ZoneHealth[]
    const mapped = mapApiToInputs(zoneHealths)
    setApiInputs(mapped)
    setManualInputs(mapped)
    }
    if (availableResponse.status === 'fulfilled' || assignedResponse.status === 'fulfilled') {
    const availableRows = availableResponse.status === 'fulfilled'
      ? ((availableResponse.value.data?.batches || []) as BatchRow[])
      : []
    const assignedRows = assignedResponse.status === 'fulfilled'
      ? ((assignedResponse.value.data?.batches || []) as BatchRow[])
      : []
    applyWorkerBatchSnapshot(availableRows, assignedRows, setAvailableBatches, setAssignedBatches)
    }

    setCommittedResult(computeResult(sourceInputs, policyInputs, scopeLabel, targetZoneIds.length, 'god-mode-override'))
    setLastDisruptionSignal({
    sent: successfulRequests > 0,
    sentAt: new Date().toISOString(),
    scopeLabel,
    triggerMode: 'all-factors-combined',
    zonesCount: targetZoneIds.length,
    successfulRequests,
    claimsCreated: totalClaims,
    notificationsCreated: totalNotifications,
    })
    setNotice({
    tone: 'success',
    message: `Combined disruption sent for ${scopeLabel} (score ${signalPlan.compositeScore.toFixed(2)}). Claims created: ${totalClaims}, notifications created: ${totalNotifications}.`,
    })
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Unable to add disruption'
    setError(message)
    setNotice({ tone: 'error', message })
  } finally {
    setAddingDisruption(false)
  }
  }

  return (
    <GodModeContext.Provider
      value={{
        godModeEnabled,
        setGodModeEnabled,
        zoneLevel,
        setZoneLevel: handleZoneLevelChange,
        zoneName,
        setZoneName,
        zoneNameOptions,
        zoneNameLoading,
        affectedZoneIds,
        scopeLabel,
        apiInputs,
        manualInputs,
        policyInputs,
        setManualInput,
        setPolicyInput,
        loading,
        running,
        addingDisruption,
        generatingBatches,
        lastDisruptionSignal,
        error,
        notice,
        result: committedResult,
        previewResult: livePreview,
        runSimulation,
        addDisruption,
        generateBatches,
        clearNotice: () => setNotice(null),
        zones,
        availableBatches,
        assignedBatches,
        batches,
        showCodes,
        setShowCodes,
        endpointStatus,
        checkingBatchFlow,
        runBatchFlowCheck,
        lastBatchFlowCheck,
        checkingIntegrationSelfTest,
        runIntegrationSelfTest,
        integrationSelfTestResult,
      }}
    >
      {children}
    </GodModeContext.Provider>
  )
}

export function useGodMode() {
  const context = useContext(GodModeContext)
  if (!context) {
    throw new Error('useGodMode must be used inside GodModeProvider')
  }
  return context
}













