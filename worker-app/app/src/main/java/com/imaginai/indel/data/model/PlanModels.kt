package com.imaginai.indel.data.model

import com.google.gson.annotations.SerializedName

data class DeliveryPlan(
    @SerializedName("plan_id") val planId: String,
    @SerializedName("plan_name") val planName: String,
    @SerializedName("range_start") val rangeStart: Int,
    @SerializedName("range_end") val rangeEnd: Int,
    @SerializedName("weekly_premium_inr") val weeklyPremiumInr: Int,
    @SerializedName("weekly_premium_min_inr") val weeklyPremiumMinInr: Int? = null,
    @SerializedName("weekly_premium_max_inr") val weeklyPremiumMaxInr: Int? = null,
    @SerializedName("coverage_ratio") val coverageRatio: Double,
    @SerializedName("max_payout_inr") val maxPayoutInr: Int,
    @SerializedName("description") val description: String = ""
)

data class PlanListResponse(
    @SerializedName("plans") val plans: List<DeliveryPlan>
)

data class PlanSelectionRequest(
    @SerializedName("plan_id") val planId: String,
    @SerializedName("expected_deliveries") val expectedDeliveries: Int? = null,
    @SerializedName("payment_amount_inr") val paymentAmountInr: Int,
    @SerializedName("payment_confirmed") val paymentConfirmed: Boolean = true
)

data class PlanSelectionResponse(
    @SerializedName("message") val message: String,
    @SerializedName("plan") val plan: DeliveryPlan,
    @SerializedName("policy") val policy: PlanSelectionPolicy? = null
)

data class PlanSelectionPolicy(
    @SerializedName("policy_id") val policyId: String? = null,
    @SerializedName("status") val status: String? = null,
    @SerializedName("weekly_premium_inr") val weeklyPremiumInr: Int? = null,
    @SerializedName("coverage_ratio") val coverageRatio: Double? = null,
    @SerializedName("payment_amount_inr") val paymentAmountInr: Int? = null,
    @SerializedName("payment_status") val paymentStatus: String? = null
)
