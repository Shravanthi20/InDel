import axios from 'axios'
import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import {
  generateClaimsForDisruption,
  getAssignedBatches,
  getAvailableBatches,
  getDisruptions,
  getZoneHealth,
  getZonePaths,
  getZones,
  postIngestDemoOrder,
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
      })
    })
  }

  const pairs = payload?.cityPairs ?? payload?.city_pairs ?? []
  if (Array.isArray(pairs)) {
    pairs.forEach((pair: any) => {
      const fromCity = String(pair?.from ?? pair?.fromCity ?? '').trim()
      const toCity = String(pair?.to ?? pair?.toCity ?? '').trim()
      const state = String(pair?.state ?? '').trim()
      const fromState = String(pair?.fromState ?? pair?.from_state ?? '').trim()
      const toState = String(pair?.toState ?? pair?.to_state ?? '').trim()

      if (!fromCity && !toCity) {
        return
      }

      const label = level === 'C'
        ? `${fromCity} (${fromState || ''}) to ${toCity} (${toState || ''})`.replace(/\s+/g, ' ').replace(/\(\s*\)/g, '').trim()
        : `${fromCity} to ${toCity}${state ? ` (${state})` : ''}`

      options.push({
        value: label,
        label,
        zoneId: 0,
        fromCity,
        toCity,
        state,
        fromState,
        toState,
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

function deriveZonePathsFromZones(level: ZoneLevel, zones: ZoneRecord[]): ZonePathOption[] {
  if (zones.length === 0 || level === 'ALL') {
    return []
  }

  if (level === 'A') {
    const cityOptions: ZonePathOption[] = zones
      .map((zone) => zone.city?.trim())
      .filter((city): city is string => Boolean(city))
      .map((city) => ({ value: city, label: city, zoneId: 0, city }))
    return Array.from(new Map(cityOptions.map((option) => [option.value, option])).values()).slice(0, 10)
  }

  const pairs: ZonePathOption[] = []
  for (let i = 0; i < zones.length; i += 1) {
    for (let j = 0; j < zones.length; j += 1) {
      if (i === j) {
        continue
      }

      const from = zones[i]
      const to = zones[j]
      const fromCity = from.city?.trim() ?? ''
      const toCity = to.city?.trim() ?? ''
      const fromState = from.state?.trim() ?? ''
      const toState = to.state?.trim() ?? ''
      if (!fromCity || !toCity) {
        continue
      }

      const sameState = fromState !== '' && toState !== '' && normalizeText(fromState) === normalizeText(toState)
      if (level === 'B' && !sameState) {
        continue
      }
      if (level === 'C' && sameState) {
        continue
      }

      const label = level === 'C'
        ? `${fromCity} (${fromState}) to ${toCity} (${toState})`
        : `${fromCity} to ${toCity}${fromState ? ` (${fromState})` : ''}`

      pairs.push({
        value: label,
        label,
        zoneId: 0,
        fromCity,
        toCity,
        state: fromState || undefined,
        fromState: fromState || undefined,
        toState: toState || undefined,
      })
    }
  }

  return Array.from(new Map(pairs.map((option) => [option.value, option])).values()).slice(0, 10)
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
    seed = (seed * 17 + char.charCodeAt(0)) % 9000
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
  error: string
  result: SimulationResult
  runSimulation: () => void
  generatingBatches: boolean
  generateBatches: () => void
  notice: Notice | null
  clearNotice: () => void
  batches: BatchRow[]
  showCodes: boolean
  setShowCodes: (value: boolean) => void
  endpointStatus: EndpointStatus
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
  const [showCodes, setShowCodes] = useState(false)
  const [loading, setLoading] = useState(true)
  const [running, setRunning] = useState(false)
  const [generatingBatches, setGeneratingBatches] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState<Notice | null>(null)
  const [endpointStatus, setEndpointStatus] = useState<EndpointStatus>({
    zoneHealth: 'pending',
    availableBatches: 'pending',
    assignedBatches: 'pending',
  })
  const [committedResult, setCommittedResult] = useState<SimulationResult>(
    computeResult(DEFAULT_ENV_INPUTS, DEFAULT_POLICY_INPUTS, 'ALL ZONES', 0, 'api-mock'),
  )

  const derivedZonePaths = useMemo(() => deriveZonePathsFromZones(zoneLevel, zones), [zoneLevel, zones])
  const effectiveZonePaths = useMemo(
    () => (zonePaths.length > 0 ? zonePaths : derivedZonePaths),
    [zonePaths, derivedZonePaths],
  )

  const zoneNameOptions = useMemo(() => {
    if (zoneLevel === 'ALL') {
      return [{ value: 'ALL ZONES', label: 'ALL ZONES', zoneId: 0 }]
    }

    if (zoneNameLoading) {
      return [{ value: '', label: 'Loading zone names...', zoneId: 0 }]
    }

    if (effectiveZonePaths.length === 0) {
      return [{ value: '', label: 'No zone names available', zoneId: 0 }]
    }

    return [{ value: '', label: 'Select Zone Name', zoneId: 0 }, ...effectiveZonePaths]
  }, [zoneLevel, zoneNameLoading, effectiveZonePaths])

  const affectedZoneIds = useMemo(
    () => resolveTargetZoneIds(zoneLevel, zoneName, zones, effectiveZonePaths),
    [zoneLevel, zoneName, zones, effectiveZonePaths],
  )

  const scopeLabel = useMemo(() => {
    if (zoneLevel === 'ALL' || zoneName === 'ALL ZONES') {
      return 'ALL ZONES'
    }

    const selected = zoneNameOptions.find((option) => option.value === zoneName)
    return selected ? `${zoneLevel} / ${selected.label}` : `${zoneLevel}`
  }, [zoneLevel, zoneName, zoneNameOptions])

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
        assignedBatches: assignedResponse.status === 'fulfilled' ? 'ok' : (isAuthDenied(assignedResponse.reason) ? 'pending' : 'failed'),
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

      if (availableResponse.status === 'fulfilled') {
        setAvailableBatches(availableResponse.value.data?.batches || [])
      }

      if (assignedResponse.status === 'fulfilled') {
        setAssignedBatches(assignedResponse.value.data?.batches || [])
      } else if (isAuthDenied(assignedResponse.reason)) {
        setAssignedBatches([])
      }

      if (zoneHealthResponse.status === 'rejected' && availableResponse.status === 'rejected' && zonesResponse.status === 'rejected') {
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
    if (availableResponse.status === 'fulfilled') {
      setAvailableBatches(availableResponse.value.data?.batches || [])
    }
    if (assignedResponse.status === 'fulfilled') {
      setAssignedBatches(assignedResponse.value.data?.batches || [])
    } else if (isAuthDenied(assignedResponse.reason)) {
      setAssignedBatches([])
    }
  }

  const generateBatches = async () => {
    if (generatingBatches) {
      return
    }

    setGeneratingBatches(true)
    setError('')
    setNotice(null)

    const selectedPath = effectiveZonePaths.find((option) => option.value === zoneName)
    const targetZoneId = affectedZoneIds[0] ?? zones[0]?.zone_id ?? 1
    const fromCity = selectedPath?.fromCity ?? selectedPath?.city ?? zones[0]?.city ?? 'Tambaram'
    const toCity = selectedPath?.toCity ?? selectedPath?.city ?? zones[0]?.name ?? 'Velachery'
    const fromState = selectedPath?.fromState ?? selectedPath?.state ?? zones[0]?.state ?? ''
    const toState = selectedPath?.toState ?? selectedPath?.state ?? zones[0]?.state ?? ''

    try {
      let successMessage = 'Batches generated successfully.'

      try {
        await postSimulateOrders({ count: 6 })
        successMessage = 'Batch generation triggered: 6 demo orders simulated.'
      } catch (simulateErr) {
        if (!isAuthDenied(simulateErr)) {
          throw simulateErr
        }

        const ordersToCreate = 6
        const seedRequests = Array.from({ length: ordersToCreate }, (_, idx) => {
          const orderId = `god_${Date.now()}_${idx + 1}`
          return postIngestDemoOrder({
            order_id: orderId,
            customer_name: `God Mode Customer ${idx + 1}`,
            customer_id: `god-cust-${idx + 1}`,
            customer_contact_number: `90000000${(10 + idx).toString().padStart(2, '0')}`,
            address: `${toCity} Address ${idx + 1}`,
            payment_method: 'cod',
            order_value: 220 + idx * 15,
            payment_amount: 220 + idx * 15,
            package_size: idx % 3 === 0 ? 'small' : idx % 3 === 1 ? 'medium' : 'large',
            package_weight_kg: idx % 3 === 0 ? 0.8 : idx % 3 === 1 ? 2.4 : 4.2,
            zone_id: targetZoneId,
            from_city: fromCity,
            to_city: toCity,
            from_state: fromState,
            to_state: toState,
            pickup_area: fromCity,
            drop_area: toCity,
            distance_km: 6 + idx,
            tip_inr: 10,
            delivery_fee_inr: 35,
            status: 'assigned',
            source: 'god-mode-batch-generator',
          })
        })

        const ingestResults = await Promise.allSettled(seedRequests)
        const ingestedCount = ingestResults.filter((result) => result.status === 'fulfilled').length

        if (ingestedCount === 0) {
          throw new Error('Unable to ingest demo orders for batch generation')
        }

        successMessage = `Batch generation completed using fallback ingest: ${ingestedCount} orders created.`
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
      const targetZoneIds = affectedZoneIds.length > 0 ? affectedZoneIds : zones.map((zone) => zone.zone_id)

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
      if (availableResponse.status === 'fulfilled') {
        setAvailableBatches(availableResponse.value.data?.batches || [])
      }
      if (assignedResponse.status === 'fulfilled') {
        setAssignedBatches(assignedResponse.value.data?.batches || [])
      } else if (isAuthDenied(assignedResponse.reason)) {
        setAssignedBatches([])
      }

      setCommittedResult(preview)
    } catch (runErr) {
      setError(runErr instanceof Error ? runErr.message : 'Unable to run simulation')
      setCommittedResult(livePreview)
    } finally {
      setRunning(false)
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
        generatingBatches,
        error,
        notice,
        result: committedResult,
        runSimulation,
        generateBatches,
        clearNotice: () => setNotice(null),
        batches,
        showCodes,
        setShowCodes,
        endpointStatus,
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
