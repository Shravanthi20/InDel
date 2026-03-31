package com.imaginai.indel.data.model

import com.google.gson.annotations.SerializedName

data class EarningsSummary(
    @SerializedName("currency") val currency: String,
    @SerializedName("this_week_actual") val thisWeekActual: Int,
    @SerializedName("this_week_baseline") val thisWeekBaseline: Int,
    @SerializedName("protected_income") val protectedIncome: Int,
    @SerializedName("history") val history: List<EarningsHistoryItem>
)

data class EarningsHistoryItem(
    @SerializedName("week") val week: String,
    @SerializedName("actual") val actual: Int,
    @SerializedName("baseline") val baseline: Int
)
