package com.imaginai.indel.data.model

import com.google.gson.annotations.SerializedName

data class Earnings(
    @SerializedName("this_week_actual") val thisWeekActual: Double,
    @SerializedName("this_week_baseline") val thisWeekBaseline: Double,
    @SerializedName("today_earnings") val todayEarnings: Double = 0.0,
    @SerializedName("protected_income") val protectedIncome: Double,
    @SerializedName("insight") val insight: String? = null,
    val history: List<EarningRecord> = emptyList()
)

data class EarningRecord(
    val date: String,
    val amount: Double
)

data class Order(
    @SerializedName(value = "order_id", alternate = ["id"]) val orderId: String,
    @SerializedName("worker_id") val workerId: Int? = null,
    @SerializedName("zone_id") val zoneId: Int? = null,
    @SerializedName("from_city") val fromCity: String? = null,
    @SerializedName("to_city") val toCity: String? = null,
    @SerializedName("from_state") val fromState: String? = null,
    @SerializedName("to_state") val toState: String? = null,
    @SerializedName("pickup_area") val pickupArea: String,
    @SerializedName("drop_area") val dropArea: String,
    @SerializedName("distance_km") val distanceKm: Double,
    @SerializedName(value = "earning_inr", alternate = ["order_value"]) val earningInr: Double,
    @SerializedName("package_weight_kg") val packageWeightKg: Double = 0.0,
    @SerializedName("package_size") val packageSize: String? = null,
    @SerializedName("tip_inr") val tipInr: Double = 0.0,
    @SerializedName("vehicle_type") val vehicleType: String? = null,
    @SerializedName("vehicle_capacity") val vehicleCapacity: Int? = null,
    @SerializedName("allowed_zones") val allowedZones: String? = null,
    val status: String,
    @SerializedName(value = "assigned_at", alternate = ["created_at"]) val assignedAt: String,
    @SerializedName("customer_name") val customerName: String? = null,
    @SerializedName(value = "customer_phone", alternate = ["customer_contact_number"]) val customerPhone: String? = null,
    @SerializedName("address") val address: String? = null,
    @SerializedName(value = "payment_type", alternate = ["payment_method"]) val paymentType: String? = null,
    @SerializedName("zone_level") val zoneLevel: String? = null,
    @SerializedName("zone_name") val zoneName: String? = null,
    @SerializedName("source_node") val sourceNode: String? = null,
    @SerializedName("destination_node") val destinationNode: String? = null,
    @SerializedName("current_node") val currentNode: String? = null,
    @SerializedName("route") val route: String? = null
)

data class OrderListResponse(
    @SerializedName("orders") val orders: List<Order>
)

data class Notification(
    val id: String,
    val type: String,
    val title: String,
    val body: String,
    @SerializedName("created_at") val createdAt: String,
    val read: Boolean
)

data class NotificationListResponse(
    @SerializedName("notifications") val notifications: List<Notification>
)

data class VerifyCodeRequest(
    @SerializedName("code") val code: String
)

data class ZoneConfigResponse(
    @SerializedName("zone_id") val zoneId: String,
    @SerializedName("name") val name: String,
    @SerializedName("require_ip_validation") val requireIpValidation: Boolean
)

data class Zone(
    @SerializedName("zone_id") val zoneId: Int,
    @SerializedName("name") val name: String,
    @SerializedName("city") val city: String,
    @SerializedName("state") val state: String,
    @SerializedName("risk_rating") val riskRating: Double,
    @SerializedName("active_workers") val activeWorkers: Int,
    @SerializedName("areas") val areas: List<String> = emptyList()
)

data class ZoneListResponse(
    @SerializedName("zones") val zones: List<Zone>
)

data class CityPair(
    @SerializedName("from") val from: String,
    @SerializedName("to") val to: String,
    @SerializedName("state") val state: String? = null,
    @SerializedName("from_state") val fromState: String? = null,
    @SerializedName("to_state") val toState: String? = null,
    @SerializedName("distance_km") val distanceKm: Double = 0.0,
    @SerializedName("from_lat") val fromLat: Double? = null,
    @SerializedName("from_lon") val fromLon: Double? = null,
    @SerializedName("to_lat") val toLat: Double? = null,
    @SerializedName("to_lon") val toLon: Double? = null
)

data class ZonePath(
    @SerializedName("id") val id: String? = null,
    @SerializedName("display_name") val displayName: String? = null,
    @SerializedName("city") val city: String? = null,
    @SerializedName("from_city") val fromCity: String? = null,
    @SerializedName("to_city") val toCity: String? = null
)

// Zone A cities from backend: each entry has city, state, lat, lon
data class ZonePathCity(
    @SerializedName("city") val city: String,
    @SerializedName("state") val state: String? = null,
    @SerializedName("lat") val lat: Double? = null,
    @SerializedName("lon") val lon: Double? = null
)

// Zone B/C entries from backend: zone_name, zone_state, city, level
data class ZonePathEntry(
    @SerializedName("zone_id") val zoneId: Int? = null,
    @SerializedName("zone_name") val zoneName: String,
    @SerializedName("zone_state") val zoneState: String? = null,
    @SerializedName("city") val city: String? = null,
    @SerializedName("level") val level: String? = null
)

// Updated to match actual backend response shape
data class ZonePathResponse(
    @SerializedName("cities") val cities: List<ZonePathCity>? = null,  // Zone A: [{city,state,lat,lon}]
    @SerializedName("zones") val zones: List<ZonePathEntry>? = null,   // Zone B/C: [{zone_name,zone_state,...}]
    @SerializedName("city_pairs") val cityPairs: List<CityPair>? = null,
    @SerializedName("paths") val paths: List<ZonePath>? = null
)

// Zone level options (A/B/C) from /platform/zone-levels
data class ZoneLevelOption(
    @SerializedName("level") val level: String,
    @SerializedName("label") val label: String,
    @SerializedName("description") val description: String? = null
)

data class ZoneLevelResponse(
    @SerializedName("levels") val levels: List<ZoneLevelOption>
)

data class SessionResponse(
    @SerializedName("session_id") val sessionId: String,
    @SerializedName("start_time") val startTime: String,
    @SerializedName("end_time") val endTime: String?,
    @SerializedName("status") val status: String,
    @SerializedName("deliveries_completed") val deliveriesCompleted: Int,
    @SerializedName("earnings_in_session") val earningsInSession: Double
)

data class FraudSignalResponse(
    @SerializedName("signals") val signals: List<FraudSignal>
)

data class FraudSignal(
    @SerializedName("type") val type: String,
    @SerializedName("severity") val severity: String,
    @SerializedName("timestamp") val timestamp: String
)

data class EarningHistoryResponse(
    @SerializedName("history") val history: List<EarningRecord>
)

data class BaselineResponse(
    @SerializedName("baseline") val baseline: Double,
    @SerializedName("currency") val currency: String
)

data class PayoutListResponse(
    @SerializedName("payouts") val payouts: List<PayoutRecord>
)

data class PayoutRecord(
    @SerializedName("payout_id") val payoutId: String,
    @SerializedName("claim_id") val claimId: String?,
    @SerializedName("amount") val amount: Double,
    @SerializedName("method") val method: String,
    @SerializedName("status") val status: String,
    @SerializedName("processed_at") val processedAt: String
)

data class FcmTokenRequest(
    @SerializedName("fcm_token") val fcmToken: String
)

data class DisruptionRequest(
    @SerializedName("disruption_type") val disruptionType: String,
    @SerializedName("zone_level") val zoneLevel: String,
    @SerializedName("zone_name") val zoneName: String
)

data class CountRequest(
    @SerializedName("count") val count: Int
)
