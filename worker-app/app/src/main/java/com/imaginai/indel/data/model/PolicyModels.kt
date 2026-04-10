package com.imaginai.indel.data.model

import com.google.gson.annotations.SerializedName

data class Policy(
    @SerializedName("policy_id") val policyId: String = "",
    @SerializedName("status") val status: String = "inactive",
    @SerializedName("plan_status") val planStatus: String? = null,
    @SerializedName("weekly_premium_inr") val weeklyPremiumInr: Int = 0,
    @SerializedName("coverage_ratio") val coverageRatio: Double = 0.0,
    @SerializedName("zone") val zone: String = "",
    @SerializedName("next_due_date") val nextDueDate: String = "--",
    @SerializedName("shap_breakdown") val shapBreakdown: List<ShapImpact> = emptyList(),
    @SerializedName("plan_id") val planId: String? = null,
    @SerializedName("plan_name") val planName: String? = null,
    @SerializedName("range_start") val rangeStart: Int? = null,
    @SerializedName("range_end") val rangeEnd: Int? = null,
    @SerializedName("selected_deliveries") val selectedDeliveries: Int? = null,
    @SerializedName("payment_status") val paymentStatus: String? = null,
    @SerializedName("days_since_last_payment") val daysSinceLastPayment: Int? = null,
    @SerializedName("next_payment_enabled") val nextPaymentEnabled: Boolean? = null,
    @SerializedName("coverage_status") val coverageStatus: String? = null,
    @SerializedName("late_fee_inr") val lateFeeInr: Int? = null,
    @SerializedName("required_payment_inr") val requiredPaymentInr: Int? = null,
    @SerializedName("last_payment_timestamp") val lastPaymentTimestamp: String? = null,
    @SerializedName("grace_days_remaining") val graceDaysRemaining: Int? = null,
    @SerializedName("billing_cycle_days") val billingCycleDays: Int? = null,
    @SerializedName("grace_period_days") val gracePeriodDays: Int? = null,
    @SerializedName("initial_payment_multiplier") val initialPaymentMultiplier: Int? = null,
    @SerializedName("plan_info") val planInfo: PlanInfo? = null
)

data class PlanInfo(
    @SerializedName("initial_payment_rule") val initialPaymentRule: String? = null,
    @SerializedName("weekly_cycle_days") val weeklyCycleDays: Int? = null,
    @SerializedName("grace_period_days") val gracePeriodDays: Int? = null,
    @SerializedName("late_fee_rule") val lateFeeRule: String? = null,
    @SerializedName("termination_rule") val terminationRule: String? = null
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
