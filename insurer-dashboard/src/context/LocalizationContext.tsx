import { createContext, useContext, useMemo, useState, type ReactNode } from 'react'

export type Language = 'en' | 'ta' | 'hi'

export type TranslationKey =
  | 'common.language'
  | 'lang.english'
  | 'lang.tamil'
  | 'lang.hindi'
  | 'auth.secureTerminal'
  | 'auth.initializeSession'
  | 'auth.startSession'
  | 'auth.jwt'
  | 'auth.statusReady'
  | 'auth.copyright'
  | 'sidebar.dashboard'
  | 'sidebar.analysis'
  | 'sidebar.claims'
  | 'sidebar.overview'
  | 'sidebar.planStatus'
  | 'sidebar.lossRatio'
  | 'sidebar.forecast'
  | 'sidebar.claimsMenu'
  | 'sidebar.fraudQueue'
  | 'sidebar.networkConnected'
  | 'sidebar.node'
  | 'navbar.insurer'
  | 'navbar.overview'
  | 'navbar.searchPlaceholder'
  | 'route.register'
  // Overview page
  | 'pages.overview.eyebrow'
  | 'pages.overview.title'
  | 'pages.overview.description'
  | 'pages.overview.dynamicControls'
  | 'pages.overview.controlsSubtitle'
  | 'pages.overview.zoneLevel'
  | 'pages.overview.zoneSearch'
  | 'pages.overview.searchPlaceholder'
  | 'pages.overview.refresh'
  | 'pages.overview.premiumPool'
  | 'pages.overview.subscribedPlan'
  | 'pages.overview.claimsHappened'
  | 'pages.overview.moneyExchange'
  | 'pages.overview.enterpriseTrends'
  | 'pages.overview.claimsDistribution'
  | 'pages.overview.pending'
  | 'pages.overview.approved'
  | 'pages.overview.flagged'
  | 'pages.overview.netFlow'
  | 'pages.overview.zoneBreakdown'
  | 'pages.overview.weekPremiums'
  | 'pages.overview.weekPayouts'
  | 'pages.overview.poolReserve'
  | 'pages.overview.allZones'
  | 'pages.overview.levelA'
  | 'pages.overview.levelB'
  | 'pages.overview.levelC'
  // Claims page
  | 'pages.claims.eyebrow'
  | 'pages.claims.title'
  | 'pages.claims.description'
  | 'pages.claims.activeStream'
  | 'pages.claims.activeStreamSubtitle'
  | 'pages.claims.headerID'
  | 'pages.claims.headerWorker'
  | 'pages.claims.headerZone'
  | 'pages.claims.headerValuation'
  | 'pages.claims.headerStatus'
  | 'pages.claims.headerSecurity'
  | 'pages.claims.noData'
  | 'pages.claims.safe'
  | 'pages.claims.flagged'
  // Forecast page
  | 'pages.forecast.eyebrow'
  | 'pages.forecast.title'
  | 'pages.forecast.description'
  | 'pages.forecast.upcomingRisk'
  | 'pages.forecast.disruptionProbability'
  | 'pages.forecast.noData'
  // Loss Ratio page
  | 'pages.lossRatio.eyebrow'
  | 'pages.lossRatio.title'
  | 'pages.lossRatio.description'
  | 'pages.lossRatio.zoneMetrics'
  | 'pages.lossRatio.zoneMetricsSubtitle'
  | 'pages.lossRatio.dataGrid'
  | 'pages.lossRatio.summaryInsights'
  | 'pages.lossRatio.summaryInsightsSubtitle'
  | 'pages.lossRatio.exposureAlert'
  | 'pages.lossRatio.exposureAlertDesc'
  | 'pages.lossRatio.growthOpportunity'
  | 'pages.lossRatio.growthOpportunityDesc'
  | 'pages.lossRatio.headerZone'
  | 'pages.lossRatio.headerPremiums'
  | 'pages.lossRatio.headerClaims'
  | 'pages.lossRatio.headerRatio'
  // Fraud Queue page
  | 'pages.fraudQueue.eyebrow'
  | 'pages.fraudQueue.title'
  | 'pages.fraudQueue.description'
  | 'pages.fraudQueue.securityFlags'
  | 'pages.fraudQueue.securityFlagsSubtitle'
  | 'pages.fraudQueue.headerID'
  | 'pages.fraudQueue.headerSignal'
  | 'pages.fraudQueue.headerAnomalyScore'
  | 'pages.fraudQueue.headerVerification'
  | 'pages.fraudQueue.noData'
  | 'pages.planStatus.title'
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

type TranslationTable = Record<Language, Record<TranslationKey, string>>

const translations: TranslationTable = {
  en: {
    'common.language': 'Language',
    'lang.english': 'English',
    'lang.tamil': 'Tamil',
    'lang.hindi': 'Hindi',
    'auth.secureTerminal': 'Secure Terminal',
    'auth.initializeSession': 'Initialize insurer session to core services.',
    'auth.startSession': 'Start Session',
    'auth.jwt': 'Auth: JWT-256',
    'auth.statusReady': 'Status: Ready',
    'auth.copyright': '© 2026 InDel Technologies.',
    'sidebar.dashboard': 'Dashboard',
    'sidebar.analysis': 'Analysis',
    'sidebar.claims': 'Claims',
    'sidebar.overview': 'Overview',
    'sidebar.planStatus': 'Plan Status',
    'sidebar.lossRatio': 'Loss Ratio',
    'sidebar.forecast': 'Forecast',
    'sidebar.claimsMenu': 'Claims',
    'sidebar.fraudQueue': 'Fraud Queue',
    'sidebar.networkConnected': 'Network Connected',
    'sidebar.node': 'Node',
    'navbar.insurer': 'Insurer',
    'navbar.overview': 'Overview',
    'navbar.searchPlaceholder': 'Search console...',
    'route.register': 'Register',
    // Overview page
    'pages.overview.eyebrow': 'Console',
    'pages.overview.title': 'Global Portfolio Operations',
    'pages.overview.description': 'Track real-time worker coverage, enterprise claims pressure, and reserve posture across the ecosystem.',
    'pages.overview.dynamicControls': 'Dynamic Controls',
    'pages.overview.controlsSubtitle': 'Slice by zone level/name and refresh after each scenario change.',
    'pages.overview.zoneLevel': 'Zone Level',
    'pages.overview.zoneSearch': 'Zone / City Search',
    'pages.overview.searchPlaceholder': 'Filter by zone or city',
    'pages.overview.refresh': 'Refresh',
    'pages.overview.premiumPool': 'Premium Pool',
    'pages.overview.subscribedPlan': 'Subscribed On Plan',
    'pages.overview.claimsHappened': 'Claims Happened',
    'pages.overview.moneyExchange': 'Overall Money Exchange',
    'pages.overview.enterpriseTrends': 'Enterprise Trends',
    'pages.overview.claimsDistribution': 'Claims Distribution',
    'pages.overview.pending': 'Pending',
    'pages.overview.approved': 'Approved',
    'pages.overview.flagged': 'Flagged',
    'pages.overview.netFlow': 'Net Flow',
    'pages.overview.zoneBreakdown': 'Zone Breakdown',
    'pages.overview.weekPremiums': 'Week Premiums',
    'pages.overview.weekPayouts': 'Week Payouts',
    'pages.overview.poolReserve': 'Pool Reserve',
    'pages.overview.allZones': 'ALL',
    'pages.overview.levelA': 'A',
    'pages.overview.levelB': 'B',
    'pages.overview.levelC': 'C',
    // Claims page
    'pages.claims.eyebrow': 'Pipeline',
    'pages.claims.title': 'Global Claims Stream',
    'pages.claims.description': 'Inspect real-time payout requests, fraud scores, and settlement status across the ecosystem.',
    'pages.claims.activeStream': 'Active Stream',
    'pages.claims.activeStreamSubtitle': 'Showing the most recent 20 claim events.',
    'pages.claims.headerID': 'ID',
    'pages.claims.headerWorker': 'Worker',
    'pages.claims.headerZone': 'Zone',
    'pages.claims.headerValuation': 'Valuation',
    'pages.claims.headerStatus': 'Status',
    'pages.claims.headerSecurity': 'Security',
    'pages.claims.noData': 'Awaiting real-time claim signals.',
    'pages.claims.safe': 'safe',
    'pages.claims.flagged': 'flagged',
    // Forecast page
    'pages.forecast.eyebrow': 'Forecast',
    'pages.forecast.title': '7-Day Forecast',
    'pages.forecast.description': 'Surface near-term disruption probability so reserve planning can happen before claims arrive.',
    'pages.forecast.upcomingRisk': 'Upcoming Disruption Risk',
    'pages.forecast.disruptionProbability': 'Disruption probability',
    'pages.forecast.noData': 'No forecast outputs available yet.',
    // Loss Ratio page
    'pages.lossRatio.eyebrow': 'Analysis',
    'pages.lossRatio.title': 'Loss Ratio Distribution',
    'pages.lossRatio.description': 'Deep dive into zone performance and risk concentration across the active insurer book.',
    'pages.lossRatio.zoneMetrics': 'Zone Metrics',
    'pages.lossRatio.zoneMetricsSubtitle': 'Variance across active operational zones.',
    'pages.lossRatio.dataGrid': 'Data Grid',
    'pages.lossRatio.summaryInsights': 'Summary Insights',
    'pages.lossRatio.summaryInsightsSubtitle': 'Critical risk focuses.',
    'pages.lossRatio.exposureAlert': 'Exposure Alert',
    'pages.lossRatio.exposureAlertDesc': 'High variance detected in industrial zones. Suggested adjustment for ratios > 80%.',
    'pages.lossRatio.growthOpportunity': 'Growth Opportunity',
    'pages.lossRatio.growthOpportunityDesc': 'Zone scaling successful where loss ratio remains below 15% threshold.',
    'pages.lossRatio.headerZone': 'Zone',
    'pages.lossRatio.headerPremiums': 'Premiums',
    'pages.lossRatio.headerClaims': 'Claims',
    'pages.lossRatio.headerRatio': 'Ratio',
    // Fraud Queue page
    'pages.fraudQueue.eyebrow': 'Security',
    'pages.fraudQueue.title': 'Fraud Analysis Queue',
    'pages.fraudQueue.description': 'Review high-risk claims routed for manual verification by the ML scoring engine.',
    'pages.fraudQueue.securityFlags': 'Security Flags',
    'pages.fraudQueue.securityFlagsSubtitle': 'Sorted by anomaly priority for manual review.',
    'pages.fraudQueue.headerID': 'ID',
    'pages.fraudQueue.headerSignal': 'Signal',
    'pages.fraudQueue.headerAnomalyScore': 'Anomaly Score',
    'pages.fraudQueue.headerVerification': 'Verification',
    'pages.fraudQueue.noData': 'Security clearance complete. No flags.',
    'pages.planStatus.title': 'Plan Status Dashboard',
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
  },
  ta: {
    'common.language': 'மொழி',
    'lang.english': 'ஆங்கிலம்',
    'lang.tamil': 'தமிழ்',
    'lang.hindi': 'இந்தி',
    'auth.secureTerminal': 'பாதுகாப்பான டெர்மினல்',
    'auth.initializeSession': 'கோர் சேவைகளுக்கு காப்பீட்டாளர் அமர்வை தொடங்கவும்.',
    'auth.startSession': 'அமர்வு தொடங்கு',
    'auth.jwt': 'அங்கீகாரம்: JWT-256',
    'auth.statusReady': 'நிலை: தயார்',
    'auth.copyright': '© 2026 InDel டெக்னாலஜீஸ்.',
    'sidebar.dashboard': 'டாஷ்போர்டு',
    'sidebar.analysis': 'பகுப்பாய்வு',
    'sidebar.claims': 'கோரிக்கைகள்',
    'sidebar.overview': 'மேலோட்டம்',
    'sidebar.planStatus': 'திட்ட நிலை',
    'sidebar.lossRatio': 'இழப்பு விகிதம்',
    'sidebar.forecast': 'முன்கணிப்பு',
    'sidebar.claimsMenu': 'கோரிக்கைகள்',
    'sidebar.fraudQueue': 'மோசடி வரிசை',
    'sidebar.networkConnected': 'வலைப்பிணைப்பு இணைந்துள்ளது',
    'sidebar.node': 'நோடு',
    'navbar.insurer': 'காப்பீட்டாளர்',
    'navbar.overview': 'மேலோட்டம்',
    'navbar.searchPlaceholder': 'கணினி பலகையை தேடவும்...',
    'route.register': 'பதிவு',
    // Overview page
    'pages.overview.eyebrow': 'கன்சோல்',
    'pages.overview.title': 'உலகளாவிய போர்ட்ஃபோலியோ செயல்பாடுகள்',
    'pages.overview.description': 'நிகழ்நேர தொழிலாளர் கவரேஜ், நிறுவன கோரிக்கை அழுத்தம் மற்றும் ரிசர்வ் நிலையை சாரணி முழுவதும் கண்காணிக்கவும்.',
    'pages.overview.dynamicControls': 'இயக்கீய கட்டுப்பாட்டுக்கள்',
    'pages.overview.controlsSubtitle': 'மண்டல நிலை/பெயரால் வெட்டி ஒவ்வொரு சிற்றளவ மாற்றத்திற்குப் பிறகு புதுப்பிக்கவும்.',
    'pages.overview.zoneLevel': 'மண்டல நிலை',
    'pages.overview.zoneSearch': 'மண்டல / நகர தேடல்',
    'pages.overview.searchPlaceholder': 'மண்டல அல்லது நகரத்தால் வடிகட்டவும்',
    'pages.overview.refresh': 'புதுப்பிக்க',
    'pages.overview.premiumPool': 'பிரீமியம் குளம்',
    'pages.overview.subscribedPlan': 'திட்டத்தில் பதிவுசெய்யப்பட்டது',
    'pages.overview.claimsHappened': 'கோரிக்கைகள் நிகழ்ந்தன',
    'pages.overview.moneyExchange': 'ஒட்டுமொத்த பணப் பரிமாற்றம்',
    'pages.overview.enterpriseTrends': 'நிறுவன போக்குகள்',
    'pages.overview.claimsDistribution': 'கோரிக்கைகள் விநியோகம்',
    'pages.overview.pending': 'நிலுவையில் உள்ள',
    'pages.overview.approved': 'அங்கீகৃत',
    'pages.overview.flagged': 'குறிக்கப்பட்ட',
    'pages.overview.netFlow': 'நிகர ஓட்டம்',
    'pages.overview.zoneBreakdown': 'மண்டல பிரிப்பு',
    'pages.overview.weekPremiums': 'வாரப் பிரீமியம்',
    'pages.overview.weekPayouts': 'வாரப் பொகிப்பனவுகள்',
    'pages.overview.poolReserve': 'குளம் ரிசர்வ்',
    'pages.overview.allZones': 'அனைத்து',
    'pages.overview.levelA': 'A',
    'pages.overview.levelB': 'B',
    'pages.overview.levelC': 'C',
    // Claims page
    'pages.claims.eyebrow': 'குழாய்',
    'pages.claims.title': 'உலகளாவிய கோரிக்கைகள் ஸ்ட்ரீம்',
    'pages.claims.description': 'நிகழ்நேர கொடுப்பனவு கோரிக்கைகள், மோசடி மதிப்பெண்கள் மற்றும் தீடுமையிற்குப் பிறகு சாரணி முழுவதும் பகிர்வு நிலையை ஆராயவும்.',
    'pages.claims.activeStream': 'செயலில் ஸ்ட்ரீம்',
    'pages.claims.activeStreamSubtitle': 'மிகச் சமீபத்திய 20 கோரிக்கை நிகழ்வுகளைக் காட்டுதல்.',
    'pages.claims.headerID': 'ID',
    'pages.claims.headerWorker': 'தொழிலாளர்',
    'pages.claims.headerZone': 'மண்டலம்',
    'pages.claims.headerValuation': 'மதிப்பீடு',
    'pages.claims.headerStatus': 'நிலை',
    'pages.claims.headerSecurity': 'பாதுகாப்பு',
    'pages.claims.noData': 'நிகழ்நேர கோரிக்கை சிக்னல்கள் காத்திருந்தது.',
    'pages.claims.safe': 'பாதுகாப்பான',
    'pages.claims.flagged': 'குறிக்கப்பட்ட',
    // Forecast page
    'pages.forecast.eyebrow': 'முன்கணிப்பு',
    'pages.forecast.title': '7-நாள் முன்கணிப்பு',
    'pages.forecast.description': 'கோரிக்கைகள் வருவதற்கு முன்பு ரிசர்வ் திட்டமிடல் நடக்க வேண்டியதாக, அருகிலுள்ள இடையூறு நிகழ்தகவை மேல்நோக்கமாக வைக்கவும்.',
    'pages.forecast.upcomingRisk': 'விரக்ததீ இடையூறு ஆபத்து',
    'pages.forecast.disruptionProbability': 'இடையூறு நிகழ்தகவு',
    'pages.forecast.noData': 'முன்கணிப்பு வெளியீடுகள் இதுவரை கிடைக்கவில்லை.',
    // Loss Ratio page
    'pages.lossRatio.eyebrow': 'பகுப்பாய்வு',
    'pages.lossRatio.title': 'இழப்பு விகிதம் விநியோகம்',
    'pages.lossRatio.description': 'செயலில் இன்ற்ற் புத்தகத்தில் மண்டல செயல்திறன் மற்றும் ஆபத்து செறிவு பற்றிய ஆழமான பகுப்பாய்வு.',
    'pages.lossRatio.zoneMetrics': 'மண்டல மெட்ரிக்குகள்',
    'pages.lossRatio.zoneMetricsSubtitle': 'செயலில் இயக்க மண்டலங்கள் முழுவதும் மாறுபாடு.',
    'pages.lossRatio.dataGrid': 'தரவு கட்டம்',
    'pages.lossRatio.summaryInsights': 'சারைஞ்சய பார்வைக்கு',
    'pages.lossRatio.summaryInsightsSubtitle': 'விமர்சனமான ஆபத்து கவனங்கள்.',
    'pages.lossRatio.exposureAlert': 'வெளிப்பாடு எச்சரிக்கை',
    'pages.lossRatio.exposureAlertDesc': 'தொழிற்சாலை மண்டலங்களில் அதிக மாறுபாடு கண்டறியப்பட்டது. 80% ஐ விட அதிகமான விகிதங்களுக்குப் பரிந்துரைக்கப்பட்ட சரிசெய்தல்.',
    'pages.lossRatio.growthOpportunity': 'வளர்ச்சி வாய்ப்பு',
    'pages.lossRatio.growthOpportunityDesc': 'இழப்பு விகிதம் 15% எல்லை கீழே இருக்கும் மண்டல அளவிடல் வெற்றிகரமாகவும்.',
    'pages.lossRatio.headerZone': 'மண்டலம்',
    'pages.lossRatio.headerPremiums': 'பிரீமியம்',
    'pages.lossRatio.headerClaims': 'கோரிக்கைகள்',
    'pages.lossRatio.headerRatio': 'விகிதம்',
    // Fraud Queue page
    'pages.fraudQueue.eyebrow': 'பாதுகாப்பு',
    'pages.fraudQueue.title': 'மோசடி பகுப்பாய்வு வரிசை',
    'pages.fraudQueue.description': 'ML மதிப்பெண் இயந்திரத்தால் கையாளுவதற்கு வழிசெலுத்தப்பட்ட உচ்চ ঝுக்கு கோரிக்கைகளை மீண்டும் பார்க்கவும்.',
    'pages.fraudQueue.securityFlags': 'பாதுகாப்பு கொொொொொொொ',
    'pages.fraudQueue.securityFlagsSubtitle': 'கைகுறிப்பு மதிப்பீடு முன்னுரிமைக்கு வருகிற கையாளுவதற்கு வரிசையாக்கப்பட்டது.',
    'pages.fraudQueue.headerID': 'ID',
    'pages.fraudQueue.headerSignal': 'சிக்னல்',
    'pages.fraudQueue.headerAnomalyScore': 'வேற்றுமை மதிப்பெண்',
    'pages.fraudQueue.headerVerification': 'சரிபார்ப்பு',
    'pages.fraudQueue.noData': 'பாதுகாப்பு நிষ்கலুஷ்ணி முடிந்தது. சிறப்பு ஐயமற.',
    'pages.planStatus.title': 'திட்ட நிலை டாஷ்போர்டு',
    // Analytics page
    'pages.analytics.title': 'பகுப்பாய்வு மற்றும் இடையூறு நுண்ணறிவு',
    'pages.analytics.avgOrderDrop': 'சராசரி ஆர்டர் சரிவு',
    'pages.analytics.manualReview': 'கைமுறை மதிப்பாய்வு',
    'pages.analytics.activeDisruptions': 'செயலில் உள்ள இடையூறுகள்',
    'pages.analytics.timeFilter': 'நேர இடைவெளி',
    'pages.analytics.allTime': 'எல்லா நேரமும்',
    'pages.analytics.weekly': 'வாராந்திரம்',
    'pages.analytics.realTime': 'நிகழ்நேரம்',
    'pages.analytics.selectedZone': 'தேர்ந்தெடுக்கப்பட்ட மண்டலம்',
    'pages.analytics.forecastMetadata': 'முன்கணிப்பு மெட்டாடேட்டா',
    'pages.analytics.retrainingCadence': 'மறுபயிற்சி இடைவெளி',
    'pages.analytics.scope': 'நோக்கம்',
    'pages.analytics.loadingForecast': 'முன்கணிப்பு ஏற்றப்படுகிறது...',
    'pages.analytics.forecastError': 'முன்கணிப்பை ஏற்றுவதில் தோல்வி',
    'pages.analytics.disruptions': 'இடையூறுகள்',
    'pages.analytics.disruptionId': 'இடையூறு ஐடி',
    'pages.analytics.startTime': 'தொடக்க நேரம்',
    'pages.analytics.claimsGenerated': 'உருவாக்கப்பட்ட கோரிக்கைகள்',
    'pages.analytics.claimsInReview': 'மதிப்பாய்வில் உள்ள கோரிக்கைகள்',
    'pages.analytics.payoutAmount': 'கொடுப்பனவுத் தொகை',
  },
  hi: {
    'common.language': 'भाषा',
    'lang.english': 'अंग्रेजी',
    'lang.tamil': 'तमिल',
    'lang.hindi': 'हिंदी',
    'auth.secureTerminal': 'सुरक्षित टर्मिनल',
    'auth.initializeSession': 'कोर सेवाओं के लिए बीमाकर्ता सत्र शुरू करें।',
    'auth.startSession': 'सत्र शुरू करें',
    'auth.jwt': 'प्रमाणीकरण: JWT-256',
    'auth.statusReady': 'स्थिति: तैयार',
    'auth.copyright': '© 2026 InDel टेक्नोलॉजीज।',
    'sidebar.dashboard': 'डैशबोर्ड',
    'sidebar.analysis': 'विश्लेषण',
    'sidebar.claims': 'दावे',
    'sidebar.overview': 'ओवरव्यू',
    'sidebar.planStatus': 'प्लान स्थिति',
    'sidebar.lossRatio': 'लॉस रेशियो',
    'sidebar.forecast': 'पूर्वानुमान',
    'sidebar.claimsMenu': 'दावे',
    'sidebar.fraudQueue': 'धोखाधड़ी कतार',
    'sidebar.networkConnected': 'नेटवर्क कनेक्टेड',
    'sidebar.node': 'नोड',
    'navbar.insurer': 'बीमाकर्ता',
    'navbar.overview': 'ओवरव्यू',
    'navbar.searchPlaceholder': 'कंसोल खोजें...',
    'route.register': 'पंजीकरण',
    // Overview page
    'pages.overview.eyebrow': 'कंसोल',
    'pages.overview.title': 'वैश्विक पोर्टफोलियो संचालन',
    'pages.overview.description': 'पूरे इकोसिस्टम में वास्तविक समय में कार्यकर्ता कवरेज, एंटरप्राइज दावों के दबाव, और रिज़र्व स्थिति को ट्रैक करें।',
    'pages.overview.dynamicControls': 'गतिशील नियंत्रण',
    'pages.overview.controlsSubtitle': 'ज़ोन स्तर/नाम द्वारा स्लाइस करें और प्रत्येक परिदृश्य परिवर्तन के बाद रीफ्रेश करें।',
    'pages.overview.zoneLevel': 'ज़ोन स्तर',
    'pages.overview.zoneSearch': 'ज़ोन / शहर खोज',
    'pages.overview.searchPlaceholder': 'ज़ोन या शहर द्वारा फ़िल्टर करें',
    'pages.overview.refresh': 'रीफ्रेश',
    'pages.overview.premiumPool': 'प्रीमियम पूल',
    'pages.overview.subscribedPlan': 'योजना पर सदस्य',
    'pages.overview.claimsHappened': 'दावे हुए',
    'pages.overview.moneyExchange': 'कुल मुद्रा विनिमय',
    'pages.overview.enterpriseTrends': 'एंटरप्राइज ट्रेंड्स',
    'pages.overview.claimsDistribution': 'दावे वितरण',
    'pages.overview.pending': 'लंबित',
    'pages.overview.approved': 'स्वीकृत',
    'pages.overview.flagged': 'फ़्लैग किए गए',
    'pages.overview.netFlow': 'नेट फ्लो',
    'pages.overview.zoneBreakdown': 'ज़ोन ब्रेकडाउन',
    'pages.overview.weekPremiums': 'सप्ताह प्रीमियम',
    'pages.overview.weekPayouts': 'सप्ताह भुगतान',
    'pages.overview.poolReserve': 'पूल रिज़र्व',
    'pages.overview.allZones': 'सभी',
    'pages.overview.levelA': 'A',
    'pages.overview.levelB': 'B',
    'pages.overview.levelC': 'C',
    // Claims page
    'pages.claims.eyebrow': 'पाइपलाइन',
    'pages.claims.title': 'वैश्विक दावे स्ट्रीम',
    'pages.claims.description': 'पूरे इकोसिस्टम में वास्तविक समय भुगतान अनुरोध, धोखाधड़ी स्कोर, और निपटान स्थिति की जांच करें।',
    'pages.claims.activeStream': 'सक्रिय स्ट्रीम',
    'pages.claims.activeStreamSubtitle': 'सबसे हाल की 20 दावे घटनाएं दिखा रहे हैं।',
    'pages.claims.headerID': 'ID',
    'pages.claims.headerWorker': 'कार्यकर्ता',
    'pages.claims.headerZone': 'ज़ोन',
    'pages.claims.headerValuation': 'मूल्यांकन',
    'pages.claims.headerStatus': 'स्थिति',
    'pages.claims.headerSecurity': 'सुरक्षा',
    'pages.claims.noData': 'वास्तविक समय दावे संकेतों की प्रतीक्षा में।',
    'pages.claims.safe': 'सुरक्षित',
    'pages.claims.flagged': 'फ़्लैग किए गए',
    // Forecast page
    'pages.forecast.eyebrow': 'पूर्वानुमान',
    'pages.forecast.title': '7-दिन पूर्वानुमान',
    'pages.forecast.description': 'निकट-अवधि के व्यवधान संभावना को सतह पर लाएं ताकि दावे आने से पहले रिज़र्व योजना हो सके।',
    'pages.forecast.upcomingRisk': 'आने वाले व्यवधान जोखिम',
    'pages.forecast.disruptionProbability': 'व्यवधान संभावना',
    'pages.forecast.noData': 'अभी तक कोई पूर्वानुमान आउटपुट उपलब्ध नहीं है।',
    // Loss Ratio page
    'pages.lossRatio.eyebrow': 'विश्लेषण',
    'pages.lossRatio.title': 'हानि अनुपात वितरण',
    'pages.lossRatio.description': 'सक्रिय बीमाकर्ता बुक में ज़ोन प्रदर्शन और जोखिम सांद्रता में गहन विश्लेषण।',
    'pages.lossRatio.zoneMetrics': 'ज़ोन मेट्रिक्स',
    'pages.lossRatio.zoneMetricsSubtitle': 'सक्रिय परिचालन क्षेत्रों में भिन्नता।',
    'pages.lossRatio.dataGrid': 'डेटा ग्रिड',
    'pages.lossRatio.summaryInsights': 'सारांश अंतर्दृष्टि',
    'pages.lossRatio.summaryInsightsSubtitle': 'महत्वपूर्ण जोखिम फोकस।',
    'pages.lossRatio.exposureAlert': 'एक्सपोजर अलर्ट',
    'pages.lossRatio.exposureAlertDesc': 'औद्योगिक क्षेत्रों में उच्च भिन्नता का पता चला। 80% से अधिक अनुपात के लिए सुझाया गया समायोजन।',
    'pages.lossRatio.growthOpportunity': 'विकास अवसर',
    'pages.lossRatio.growthOpportunityDesc': 'जहां हानि अनुपात 15% थ्रेशोल्ड से नीचे रहता है वहां ज़ोन स्केलिंग सफल।',
    'pages.lossRatio.headerZone': 'ज़ोन',
    'pages.lossRatio.headerPremiums': 'प्रीमियम',
    'pages.lossRatio.headerClaims': 'दावे',
    'pages.lossRatio.headerRatio': 'अनुपात',
    // Fraud Queue page
    'pages.fraudQueue.eyebrow': 'सुरक्षा',
    'pages.fraudQueue.title': 'धोखाधड़ी विश्लेषण कतार',
    'pages.fraudQueue.description': 'ML स्कोरिंग इंजन द्वारा मैनुअल सत्यापन के लिए रूट की गई उच्च-जोखिम वाली दावों की समीक्षा करें।',
    'pages.fraudQueue.securityFlags': 'सुरक्षा झंडे',
    'pages.fraudQueue.securityFlagsSubtitle': 'मैनुअल समीक्षा के लिए विसंगति प्राथमिकता द्वारा क्रमबद्ध।',
    'pages.fraudQueue.headerID': 'ID',
    'pages.fraudQueue.headerSignal': 'सिग्नल',
    'pages.fraudQueue.headerAnomalyScore': 'विसंगति स्कोर',
    'pages.fraudQueue.headerVerification': 'सत्यापन',
    'pages.fraudQueue.noData': 'सुरक्षा मंजूरी पूर्ण। कोई झंडे नहीं।',
    'pages.planStatus.title': 'प्लान स्थिति डैशबोर्ड',
    // Analytics page
    'pages.analytics.title': 'एनालिटिक्स और व्यवधान इंटेलिजेंस',
    'pages.analytics.avgOrderDrop': 'औसत ऑर्डर ड्रॉप',
    'pages.analytics.manualReview': 'मैनुअल समीक्षा',
    'pages.analytics.activeDisruptions': 'सक्रिय व्यवधान',
    'pages.analytics.timeFilter': 'समय सीमा',
    'pages.analytics.allTime': 'पूरा समय',
    'pages.analytics.weekly': 'साप्ताहिक',
    'pages.analytics.realTime': 'वास्तविक समय',
    'pages.analytics.selectedZone': 'चयनित ज़ोन',
    'pages.analytics.forecastMetadata': 'पूर्वानुमान मेटाडेटा',
    'pages.analytics.retrainingCadence': 'रीट्रेनिंग कैडेंस',
    'pages.analytics.scope': 'दायरा',
    'pages.analytics.loadingForecast': 'पूर्वानुमान लोड हो रहा है...',
    'pages.analytics.forecastError': 'पूर्वानुमान लोड करने में विफल',
    'pages.analytics.disruptions': 'व्यवधान',
    'pages.analytics.disruptionId': 'व्यवधान आईडी',
    'pages.analytics.startTime': 'प्रारंभ समय',
    'pages.analytics.claimsGenerated': 'दावे उत्पन्न हुए',
    'pages.analytics.claimsInReview': 'समीक्षाधीन दावे',
    'pages.analytics.payoutAmount': 'भुगतान राशि',
  },
}

type LocalizationContextValue = {
  language: Language
  setLanguage: (language: Language) => void
  t: (key: TranslationKey) => string
}

const STORAGE_KEY = 'indel_insurer_language'

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
