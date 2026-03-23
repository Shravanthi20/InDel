package com.imaginai.indel.data.model

import com.google.gson.annotations.SerializedName

data class Claim(
    @SerializedName("claim_id") val claimId: String,
    @SerializedName("status") val status: String,
    @SerializedName("zone") val zone: String,
    @SerializedName("disruption_type") val disruptionType: String,
    @SerializedName("income_loss") val incomeLoss: Int,
    @SerializedName("payout_amount") val payoutAmount: Int,
    @SerializedName("fraud_verdict") val fraudVerdict: String,
    @SerializedName("disruption_window") val disruptionWindow: DisruptionWindow? = null,
    @SerializedName("created_at") val createdAt: String? = null
)

data class DisruptionWindow(
    @SerializedName("start") val start: String,
    @SerializedName("end") val end: String
)

data class ClaimsResponse(
    @SerializedName("claims") val claims: List<Claim>
)

data class WalletResponse(
    @SerializedName("currency") val currency: String,
    @SerializedName("available_balance") val availableBalance: Int,
    @SerializedName("last_payout_amount") val lastPayoutAmount: Int,
    @SerializedName("last_payout_at") val lastPayoutAt: String
)

data class Payout(
    @SerializedName("payout_id") val payoutId: String,
    @SerializedName("claim_id") val claimId: String,
    @SerializedName("amount") val amount: Int,
    @SerializedName("method") val method: String,
    @SerializedName("status") val status: String,
    @SerializedName("processed_at") val processedAt: String
)

data class PayoutsResponse(
    @SerializedName("payouts") val payouts: List<Payout>
)
