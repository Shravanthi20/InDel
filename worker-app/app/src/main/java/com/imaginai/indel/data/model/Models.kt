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
    @SerializedName("order_id") val orderId: String,
    @SerializedName("pickup_area") val pickupArea: String,
    @SerializedName("drop_area") val dropArea: String,
    @SerializedName("distance_km") val distanceKm: Double,
    @SerializedName(value = "earning_inr", alternate = ["order_value"]) val earningInr: Double,
    @SerializedName("tip_inr") val tipInr: Double = 0.0,
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
