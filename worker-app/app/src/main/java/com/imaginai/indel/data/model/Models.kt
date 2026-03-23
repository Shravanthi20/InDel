package com.imaginai.indel.data.model

import com.google.gson.annotations.SerializedName

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

data class Notification(
    val id: String,
    val type: String,
    val title: String,
    val body: String,
    @SerializedName("created_at") val createdAt: String,
    val read: Boolean
)
