package com.imaginai.indel.data.model

import com.google.gson.annotations.SerializedName

data class OtpSendRequest(val phone: String)
data class OtpSendResponse(
    val message: String,
    @SerializedName("otp_for_testing") val otpForTesting: String?,
    @SerializedName("expires_in_seconds") val expiresInSeconds: Int
)

data class OtpVerifyRequest(val phone: String, val otp: String)
data class OtpVerifyResponse(
    val token: String,
    @SerializedName("token_type") val tokenType: String,
    @SerializedName("worker_id") val workerId: String
)

data class WorkerProfile(
    val id: String,
    val name: String,
    val zone: String,
    @SerializedName("vehicle_type") val vehicleType: String,
    @SerializedName("upi_id") val upiId: String,
    @SerializedName("coverage_status") val coverageStatus: String
)

data class OnboardRequest(
    val name: String,
    val zone: String,
    @SerializedName("vehicle_type") val vehicleType: String,
    @SerializedName("upi_id") val upiId: String
)

data class Policy(
    val status: String,
    @SerializedName("weekly_premium_inr") val weeklyPremiumInr: Double,
    @SerializedName("coverage_ratio") val coverageRatio: Double,
    @SerializedName("next_due_date") val nextDueDate: String,
    @SerializedName("shap_breakdown") val shapBreakdown: List<ShapFeature> = emptyList()
)

data class ShapFeature(
    val feature: String,
    val impact: Double
)

data class Earnings(
    @SerializedName("this_week_actual") val thisWeekActual: Double,
    @SerializedName("this_week_baseline") val thisWeekBaseline: Double,
    @SerializedName("protected_income") val protectedIncome: Double,
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
    @SerializedName("earning_inr") val earningInr: Double,
    val status: String,
    @SerializedName("assigned_at") val assignedAt: String
)

data class Claim(
    @SerializedName("id") val id: String,
    @SerializedName("disruption_type") val disruptionType: String,
    val zone: String,
    @SerializedName("income_loss") val incomeLoss: Double,
    @SerializedName("payout_amount") val payoutAmount: Double,
    val status: String,
    @SerializedName("created_at") val createdAt: String,
    @SerializedName("disruption_window") val disruptionWindow: DisruptionWindow? = null,
    @SerializedName("fraud_verdict") val fraudVerdict: String? = null
)

data class DisruptionWindow(
    val start: String,
    val end: String
)

data class Notification(
    val id: String,
    val type: String,
    val title: String,
    val body: String,
    @SerializedName("created_at") val createdAt: String,
    val read: Boolean
)
