export interface Claim {
  id: number
  worker_id: number
  claim_amount: number
  status: string
  created_at: string
}

export interface Zone {
  zone_id: number
  name: string
  city: string
  state: string
  risk_rating: number
  active_workers: number
  areas: string[]
}

export interface Worker {
  id: number
  name: string
  zone_id: number
  total_earnings: number
}

export interface PlanUser {
  id: number
  name: string
  phone: string
  zone: string
  status: 'active' | 'inactive' | string
  policy_id?: number
  plan_id?: string
  started_at?: string
  updated_at?: string
}

// --- Zone Path API Types ---
export interface ZoneACitiesResponse {
  cities: string[];
}

export interface ZoneBCityPair {
  from: string;
  to: string;
  state: string;
}
export interface ZoneBCityPairsResponse {
  city_pairs: ZoneBCityPair[];
}

export interface ZoneCCityPair {
  from: string;
  to: string;
  from_state: string;
  to_state: string;
}
export interface ZoneCCityPairsResponse {
  city_pairs: ZoneCCityPair[];
}
