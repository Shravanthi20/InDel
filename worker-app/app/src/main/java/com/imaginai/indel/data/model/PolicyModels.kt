package com.imaginai.indel.data.model

import com.google.gson.annotations.SerializedName

data class Policy(
    @SerializedName("policy_id") val policyId: String,
    @SerializedName("status") val status: String,
    @SerializedName("weekly_premium_inr") val weeklyPremiumInr: Int,
    @SerializedName("coverage_ratio") val coverageRatio: Double,
    @SerializedName("zone") val zone: String,
    @SerializedName("next_due_date") val nextDueDate: String,
    @SerializedName("shap_breakdown") val shapBreakdown: List<ShapImpact>
)

data class ShapImpact(
    @SerializedName("feature") val feature: String,
    @SerializedName("impact") val impact: Double
)

data class PolicyResponse(
    @SerializedName("policy") val policy: Policy
)

data class PremiumResponse(
    @SerializedName("weekly_premium_inr") val weeklyPremiumInr: Int,
    @SerializedName("currency") val currency: String,
    @SerializedName("shap_breakdown") val shapBreakdown: List<ShapImpact>
)

data class EnrollResponse(
    @SerializedName("message") val message: String,
    @SerializedName("policy") val policy: PolicyStatus
)

data class PolicyStatus(
    @SerializedName("status") val status: String
)

data class PayPremiumRequest(
    @SerializedName("amount") val amount: Int? = null
)

data class PayPremiumResponse(
    @SerializedName("message") val message: String,
    @SerializedName("amount") val amount: Int,
    @SerializedName("currency") val currency: String,
    @SerializedName("payment_id") val paymentId: String,
    @SerializedName("checkout_id") val checkoutId: String? = null,
    @SerializedName("payment_status") val paymentStatus: String? = null,
    @SerializedName("checkout_mode") val checkoutMode: String? = null
)

data class SimpleMessageResponse(
    @SerializedName("message") val message: String,
    @SerializedName("policy") val policy: PolicyStatus? = null,
    @SerializedName("registered") val registered: Boolean? = null,
    @SerializedName("status") val status: String? = null
)
