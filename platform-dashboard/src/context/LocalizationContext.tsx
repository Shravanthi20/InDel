import { createContext, useContext, useMemo, useState, type ReactNode } from 'react'

export type Language = 'en' | 'ta' | 'hi'

export type TranslationKey =
  | 'common.language'
  | 'lang.english'
  | 'lang.tamil'
  | 'lang.hindi'
  | 'sidebar.inventory'
  | 'sidebar.operations'
  | 'sidebar.overview'
  | 'sidebar.workers'
  | 'sidebar.zones'
  | 'sidebar.analytics'
  | 'sidebar.viewBatches'
  | 'sidebar.chaosEngine'
  | 'sidebar.reconciliation'
  | 'sidebar.backendConnected'
  | 'sidebar.backendOffline'
  | 'sidebar.connecting'
  | 'sidebar.loadingZoneInventory'
  | 'sidebar.zonesLoaded'
  | 'navbar.platform'
  | 'navbar.searchPlaceholder'
  // Overview page
  | 'pages.overview.title'
  | 'pages.overview.description'
  | 'pages.overview.activeWorkers'
  | 'pages.overview.trackedZones'
  | 'pages.overview.liveOrders'
  | 'pages.overview.disrupted'
  | 'pages.overview.zonePressure'
  | 'pages.overview.automationOutcome'
  | 'pages.overview.stable'
  | 'pages.overview.loading'
  | 'pages.overview.connecting'
  | 'pages.overview.critical'
  | 'pages.overview.none'
  | 'pages.overview.refreshed'
  | 'pages.overview.disabled'
  | 'pages.overview.orderDrop'
  | 'pages.overview.payoutAmount'
  | 'pages.overview.workerDelta'
  // Workers page
  | 'pages.workers.title'
  | 'pages.workers.description'
  | 'pages.workers.exportCSV'
  | 'pages.workers.searchPlaceholder'
  | 'pages.workers.filterStatus'
  | 'pages.workers.allStatus'
  | 'pages.workers.live'
  | 'pages.workers.offline'
  | 'pages.workers.headerWorkerID'
  | 'pages.workers.headerName'
  | 'pages.workers.headerPhone'
  | 'pages.workers.headerZone'
  | 'pages.workers.headerStatus'
  | 'pages.workers.unknownZone'
  | 'pages.workers.exportFileName'
  | 'pages.workers.headerWorker'
  | 'pages.workers.headerZoneAssignment'
  | 'pages.workers.headerPolicyStatus'
  | 'pages.workers.headerActivity'
  | 'pages.workers.headerActions'
  | 'pages.workers.noData'
  | 'pages.workers.activeCoverage'
  | 'pages.workers.inactive'
  | 'pages.workers.liveOnShift'
  | 'pages.workers.contact'
  | 'pages.workers.showingNodes'
  | 'pages.workers.prev'
  | 'pages.workers.next'
  | 'pages.workers.lastSeen'
  | 'pages.workers.neverSeen'
  | 'pages.workers.lastUpdated'
  | 'pages.workers.refresh'
  // Zones page
  | 'pages.zones.title'
  | 'pages.zones.selectLevel'
  | 'pages.zones.selectZone'
  | 'pages.zones.selectLevelFirst'
  | 'pages.zones.disruptionDropdown'
  | 'pages.zones.triggerDisruption'
  | 'pages.zones.close'
  | 'pages.zones.loading'
  | 'pages.zones.allZones'
  | 'pages.zones.levelA'
  | 'pages.zones.levelB'
  | 'pages.zones.levelC'
  | 'pages.zones.searchZone'
  | 'pages.zones.filterStatus'
  | 'pages.zones.statusAll'
  | 'pages.zones.statusHealthy'
  | 'pages.zones.statusDisrupted'
  | 'pages.zones.statusAnomalous'
  | 'pages.zones.zoneData'
  | 'pages.zones.healthy'
  | 'pages.zones.disrupted'
  | 'pages.zones.monitoring'
  | 'pages.zones.anomalous'
  // Analytics page
  | 'pages.analytics.title'
  | 'pages.analytics.avgOrderDrop'
  | 'pages.analytics.manualReview'
  | 'pages.analytics.activeDisruptions'
  | 'pages.analytics.timeFilter'
  | 'pages.analytics.allTime'
  | 'pages.analytics.weekly'
  | 'pages.analytics.realTime'
  | 'pages.analytics.selectedZone'
  | 'pages.analytics.forecastMetadata'
  | 'pages.analytics.retrainingCadence'
  | 'pages.analytics.scope'
  | 'pages.analytics.loadingForecast'
  | 'pages.analytics.forecastError'
  | 'pages.analytics.disruptions'
  | 'pages.analytics.disruptionId'
  | 'pages.analytics.startTime'
  | 'pages.analytics.claimsGenerated'
  | 'pages.analytics.claimsInReview'
  | 'pages.analytics.payoutAmount'
  | 'pages.analytics.selectZone'
  // Disruptions page
  | 'pages.disruptions.title'
  | 'pages.disruptions.description'
  // Common status labels
  | 'status.pending'
  | 'status.approved'
  | 'status.rejected'

type TranslationTable = Record<Language, Record<TranslationKey, string>>

const enTranslations: Record<TranslationKey, string> = {
  'common.language': 'Language',
  'lang.english': 'English',
  'lang.tamil': 'Tamil',
  'lang.hindi': 'Hindi',
  'sidebar.inventory': 'Inventory',
  'sidebar.operations': 'Operations',
  'sidebar.overview': 'Overview',
  'sidebar.workers': 'Workers',
  'sidebar.zones': 'Zones',
  'sidebar.analytics': 'Analytics',
  'sidebar.viewBatches': 'Batch Browser',
  'sidebar.chaosEngine': 'Chaos Engine',
  'sidebar.reconciliation': 'Reconciliation',
  'sidebar.backendConnected': 'Backend Connected',
  'sidebar.backendOffline': 'Backend Offline',
  'sidebar.connecting': 'Connecting',
  'sidebar.loadingZoneInventory': 'Loading zone inventory...',
  'sidebar.zonesLoaded': 'zones loaded from backend',
  'navbar.platform': 'Platform',
  'navbar.searchPlaceholder': 'Search platform...',
  // Overview page
  'pages.overview.title': 'Platform Command',
  'pages.overview.description': 'Real-time telemetry and disruption automation across all covered regions.',
  'pages.overview.activeWorkers': 'Active Workers',
  'pages.overview.trackedZones': 'Tracked Zones',
  'pages.overview.liveOrders': 'Live Orders',
  'pages.overview.disrupted': 'Disrupted',
  'pages.overview.zonePressure': 'Zone Pressure Matrix',
  'pages.overview.automationOutcome': 'Automation Outcome',
  'pages.overview.stable': 'Stable',
  'pages.overview.loading': 'Loading',
  'pages.overview.connecting': 'Connecting to regional nodes...',
  'pages.overview.critical': 'Critical',
  'pages.overview.none': 'None',
  'pages.overview.refreshed': 'since refresh',
  'pages.overview.disabled': 'Disabled',
  'pages.overview.orderDrop': 'Drop',
  'pages.overview.payoutAmount': 'Total Payouts',
  'pages.overview.workerDelta': 'Worker Change',
  // Workers page
  'pages.workers.title': 'Worker Directory',
  'pages.workers.description': 'Managing global gig-worker identity and regional zone assignments.',
  'pages.workers.exportCSV': 'EXPORT CSV',
  'pages.workers.searchPlaceholder': 'Filter by name, ID or zone...',
  'pages.workers.filterStatus': 'Status Filter',
  'pages.workers.allStatus': 'All',
  'pages.workers.live': 'Live',
  'pages.workers.offline': 'Offline',
  'pages.workers.headerWorkerID': 'Worker ID',
  'pages.workers.headerName': 'Name',
  'pages.workers.headerPhone': 'Phone',
  'pages.workers.headerZone': 'Zone',
  'pages.workers.headerStatus': 'Status',
  'pages.workers.unknownZone': 'Unknown Zone',
  'pages.workers.exportFileName': 'workers_export_',
  'pages.workers.headerWorker': 'Worker',
  'pages.workers.headerZoneAssignment': 'Zone Assignment',
  'pages.workers.headerPolicyStatus': 'Policy Status',
  'pages.workers.headerActivity': 'Activity',
  'pages.workers.headerActions': 'Actions',
  'pages.workers.noData': 'No workers match the current criteria.',
  'pages.workers.activeCoverage': 'ACTIVE_COVERAGE',
  'pages.workers.inactive': 'INACTIVE',
  'pages.workers.liveOnShift': 'Live | On Shift',
  'pages.workers.contact': 'Contact',
  'pages.workers.showingNodes': 'Showing {filtered} of {total} nodes',
  'pages.workers.prev': 'PREV',
  'pages.workers.next': 'NEXT',
  'pages.workers.lastSeen': 'Last seen',
  'pages.workers.neverSeen': 'Never seen',
  'pages.workers.lastUpdated': 'Last updated',
  'pages.workers.refresh': 'REFRESH',
  // Zones page
  'pages.zones.title': 'Zone Operations & Disruption Control',
  'pages.zones.selectLevel': 'Select Level',
  'pages.zones.selectZone': 'Select Zone',
  'pages.zones.selectLevelFirst': 'Select Level First',
  'pages.zones.disruptionDropdown': 'Chaos Engine: Dynamic Zone & Disruption Selection',
  'pages.zones.triggerDisruption': 'Trigger Disruption',
  'pages.zones.close': 'Close',
  'pages.zones.loading': 'Loading...',
  'pages.zones.allZones': 'All Zones',
  'pages.zones.levelA': 'A',
  'pages.zones.levelB': 'B',
  'pages.zones.levelC': 'C',
  'pages.zones.searchZone': 'Search zone...',
  'pages.zones.filterStatus': 'Status Filter',
  'pages.zones.statusAll': 'All',
  'pages.zones.statusHealthy': 'Healthy',
  'pages.zones.statusDisrupted': 'Disrupted',
  'pages.zones.statusAnomalous': 'Anomalous',
  'pages.zones.zoneData': 'Zone Data',
  'pages.zones.healthy': 'healthy',
  'pages.zones.disrupted': 'disrupted',
  'pages.zones.monitoring': 'monitoring',
  'pages.zones.anomalous': 'anomalous_demand',
  // Analytics page
  'pages.analytics.title': 'Analytics & Disruption Intelligence',
  'pages.analytics.avgOrderDrop': 'Avg Order Drop',
  'pages.analytics.manualReview': 'Manual Review',
  'pages.analytics.activeDisruptions': 'Active Disruptions',
  'pages.analytics.timeFilter': 'Time Frame',
  'pages.analytics.allTime': 'All Time',
  'pages.analytics.weekly': 'Weekly',
  'pages.analytics.realTime': 'Real-time',
  'pages.analytics.selectedZone': 'Selected Zone',
  'pages.analytics.forecastMetadata': 'Forecast Metadata',
  'pages.analytics.retrainingCadence': 'Retraining Cadence',
  'pages.analytics.scope': 'Scope',
  'pages.analytics.loadingForecast': 'Loading forecast...',
  'pages.analytics.forecastError': 'Failed to load forecast',
  'pages.analytics.disruptions': 'Disruptions',
  'pages.analytics.disruptionId': 'Disruption ID',
  'pages.analytics.startTime': 'Start Time',
  'pages.analytics.claimsGenerated': 'Claims Generated',
  'pages.analytics.claimsInReview': 'Claims in Review',
  'pages.analytics.payoutAmount': 'Payout Amount',
  'pages.analytics.selectZone': 'SELECT ZONE',
  // Disruptions page
  'pages.disruptions.title': 'Disruptions',
  'pages.disruptions.description': 'Real-time disruption tracking and automation',
  // Common status labels
  'status.pending': 'Pending',
  'status.approved': 'Approved',
  'status.rejected': 'Rejected',
}

const taTranslations: Record<TranslationKey, string> = {
  ...enTranslations,
  'common.language': 'மொழி',
  'lang.english': 'ஆங்கிலம்',
  'lang.tamil': 'தமிழ்',
  'lang.hindi': 'இந்தி',
  'sidebar.inventory': 'சரக்கு',
  'sidebar.operations': 'செயல்பாடுகள்',
  'sidebar.overview': 'மேலோட்டம்',
  'sidebar.workers': 'பணியாளர்கள்',
  'sidebar.zones': 'மண்டலங்கள்',
  'sidebar.analytics': 'பகுப்பாய்வு',
  'sidebar.viewBatches': 'தொகுதிகளை காண்க',
  'sidebar.chaosEngine': 'கோளாறு இயந்திரம்',
  'sidebar.reconciliation': 'ஒப்புமை',
  'sidebar.backendConnected': 'பின்னணி இணைந்துள்ளது',
  'sidebar.backendOffline': 'பின்னணி ஆஃப்லைன்',
  'sidebar.connecting': 'இணைக்கப்படுகிறது',
  'sidebar.loadingZoneInventory': 'மண்டல பட்டியல் ஏற்றப்படுகிறது...',
  'sidebar.zonesLoaded': 'மண்டலங்கள் பின்னணியில் இருந்து ஏற்றப்பட்டன',
  'navbar.platform': 'தளம்',
  'navbar.searchPlaceholder': 'தளத்தில் தேடுக...',
  'pages.workers.title': 'பணியாளர் அடைவுப்பட்டி',
  'pages.workers.description': 'உலகளாவிய கிக் பணியாளர் அடையாளம் மற்றும் மண்டல ஒதுக்கீடுகளை நிர்வகிக்கிறது.',
  'pages.workers.exportCSV': 'CSV ஏற்றுமதி',
  'pages.workers.searchPlaceholder': 'பெயர், ஐடி அல்லது மண்டலத்தின் மூலம் வடிகட்டு...',
  'pages.workers.filterStatus': 'நிலை வடிகட்டி',
  'pages.workers.allStatus': 'அனைத்தும்',
  'pages.workers.live': 'நேரலை',
  'pages.workers.offline': 'ஆஃப்லைன்',
  'pages.workers.headerWorkerID': 'பணியாளர் ஐடி',
  'pages.workers.headerName': 'பெயர்',
  'pages.workers.headerPhone': 'தொலைபேசி',
  'pages.workers.headerZone': 'மண்டலம்',
  'pages.workers.headerStatus': 'நிலை',
  'pages.workers.unknownZone': 'அறியப்படாத மண்டலம்',
  'pages.workers.headerWorker': 'பணியாளர்',
  'pages.workers.headerZoneAssignment': 'மண்டல ஒதுக்கீடு',
  'pages.workers.headerPolicyStatus': 'பாலிசி நிலை',
  'pages.workers.headerActivity': 'செயற்பாடு',
  'pages.workers.headerActions': 'செயல்கள்',
  'pages.workers.noData': 'தற்போதைய அளவுகோல்களுக்கு பொருந்தும் பணியாளர்கள் இல்லை.',
  'pages.workers.activeCoverage': 'செயலில்_காப்பு',
  'pages.workers.inactive': 'செயலற்றது',
  'pages.workers.liveOnShift': 'நேரலை | பணியில்',
  'pages.workers.contact': 'தொடர்பு',
  'pages.workers.showingNodes': '{total} இல் {filtered} பதிவுகள் காட்டப்படுகிறது',
  'pages.workers.prev': 'முந்தையது',
  'pages.workers.next': 'அடுத்தது',
  'pages.workers.lastSeen': 'கடைசியாக பார்த்தது',
  'pages.workers.neverSeen': 'இதுவரை பார்த்ததில்லை',
  'pages.workers.lastUpdated': 'கடைசியாக புதுப்பிக்கப்பட்டது',
  'pages.workers.refresh': 'புதுப்பி',
  'pages.analytics.selectZone': 'மண்டலத்தை தேர்ந்தெடுக்கவும்',
}

const hiTranslations: Record<TranslationKey, string> = {
  ...enTranslations,
  'common.language': 'भाषा',
  'lang.english': 'अंग्रेजी',
  'lang.tamil': 'तमिल',
  'lang.hindi': 'हिंदी',
  'sidebar.inventory': 'इन्वेंटरी',
  'sidebar.operations': 'ऑपरेशंस',
  'sidebar.overview': 'ओवरव्यू',
  'sidebar.workers': 'वर्कर्स',
  'sidebar.zones': 'ज़ोन्स',
  'sidebar.analytics': 'एनालिटिक्स',
  'sidebar.viewBatches': 'बैच देखें',
  'sidebar.chaosEngine': 'कैओस इंजन',
  'sidebar.reconciliation': 'रिकंसिलिएशन',
  'sidebar.backendConnected': 'बैकएंड कनेक्टेड',
  'sidebar.backendOffline': 'बैकएंड ऑफलाइन',
  'sidebar.connecting': 'कनेक्ट हो रहा है',
  'sidebar.loadingZoneInventory': 'ज़ोन इन्वेंटरी लोड हो रही है...',
  'sidebar.zonesLoaded': 'ज़ोन बैकएंड से लोड हुए',
  'navbar.platform': 'प्लेटफॉर्म',
  'navbar.searchPlaceholder': 'प्लेटफॉर्म खोजें...',
  'pages.workers.title': 'वर्कर डायरेक्टरी',
  'pages.workers.description': 'ग्लोबल गिग-वर्कर पहचान और क्षेत्रीय ज़ोन असाइनमेंट का प्रबंधन।',
  'pages.workers.exportCSV': 'CSV एक्सपोर्ट',
  'pages.workers.searchPlaceholder': 'नाम, आईडी या ज़ोन से फ़िल्टर करें...',
  'pages.workers.filterStatus': 'स्थिति फ़िल्टर',
  'pages.workers.allStatus': 'सभी',
  'pages.workers.live': 'लाइव',
  'pages.workers.offline': 'ऑफलाइन',
  'pages.workers.headerWorkerID': 'वर्कर आईडी',
  'pages.workers.headerName': 'नाम',
  'pages.workers.headerPhone': 'फ़ोन',
  'pages.workers.headerZone': 'ज़ोन',
  'pages.workers.headerStatus': 'स्थिति',
  'pages.workers.unknownZone': 'अज्ञात ज़ोन',
  'pages.workers.headerWorker': 'वर्कर',
  'pages.workers.headerZoneAssignment': 'ज़ोन असाइनमेंट',
  'pages.workers.headerPolicyStatus': 'पॉलिसी स्थिति',
  'pages.workers.headerActivity': 'गतिविधि',
  'pages.workers.headerActions': 'कार्रवाइयां',
  'pages.workers.noData': 'वर्तमान मानदंड से कोई वर्कर मेल नहीं खाता।',
  'pages.workers.activeCoverage': 'ACTIVE_COVERAGE',
  'pages.workers.inactive': 'INACTIVE',
  'pages.workers.liveOnShift': 'लाइव | ड्यूटी पर',
  'pages.workers.contact': 'संपर्क',
  'pages.workers.showingNodes': '{total} में से {filtered} नोड्स दिखाए जा रहे हैं',
  'pages.workers.prev': 'पिछला',
  'pages.workers.next': 'अगला',
  'pages.workers.lastSeen': 'अंतिम बार देखा गया',
  'pages.workers.neverSeen': 'कभी नहीं देखा गया',
  'pages.workers.lastUpdated': 'अंतिम बार अपडेट किया गया',
  'pages.workers.refresh': 'ताज़ा करें',
  'pages.analytics.selectZone': 'ज़ोन चुनें',
}

const translations: TranslationTable = {
  en: enTranslations,
  ta: taTranslations,
  hi: hiTranslations,
}

type LocalizationContextValue = {
  language: Language
  setLanguage: (language: Language) => void
  t: (key: TranslationKey) => string
}

const STORAGE_KEY = 'indel_platform_language'

const LocalizationContext = createContext<LocalizationContextValue | undefined>(undefined)

function resolveInitialLanguage(): Language {
  const stored = localStorage.getItem(STORAGE_KEY)
  if (stored === 'en' || stored === 'ta' || stored === 'hi') {
    return stored
  }
  return 'en'
}

export function LocalizationProvider({ children }: { children: ReactNode }) {
  const [language, setLanguageState] = useState<Language>(resolveInitialLanguage)

  const setLanguage = (nextLanguage: Language) => {
    localStorage.setItem(STORAGE_KEY, nextLanguage)
    setLanguageState(nextLanguage)
  }

  const value = useMemo<LocalizationContextValue>(
    () => ({
      language,
      setLanguage,
      t: (key: TranslationKey) => translations[language][key] ?? translations.en[key] ?? key,
    }),
    [language],
  )

  return <LocalizationContext.Provider value={value}>{children}</LocalizationContext.Provider>
}

export function useLocalization() {
  const context = useContext(LocalizationContext)
  if (!context) {
    throw new Error('useLocalization must be used within LocalizationProvider')
  }
  return context
}
