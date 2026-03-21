export interface Claim {
  id: number
  worker_id: number
  claim_amount: number
  status: string
  created_at: string
}

export interface Zone {
  id: number
  name: string
  city: string
  risk_rating: number
}

export interface Worker {
  id: number
  name: string
  zone_id: number
  total_earnings: number
}
