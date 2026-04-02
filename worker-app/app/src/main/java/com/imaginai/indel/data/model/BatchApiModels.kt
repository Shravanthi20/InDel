package com.imaginai.indel.data.model

import com.google.gson.annotations.SerializedName

data class BatchOrderDto(
    @SerializedName("orderId") val orderId: String,
    @SerializedName("deliveryAddress") val deliveryAddress: String,
    @SerializedName("contactName") val contactName: String,
    @SerializedName("contactPhone") val contactPhone: String,
    @SerializedName("weight") val weight: Double,
)

data class DeliveryBatchDto(
    @SerializedName("batchId") val batchId: String,
    @SerializedName("zoneLevel") val zoneLevel: String,
    @SerializedName("fromCity") val fromCity: String,
    @SerializedName("toCity") val toCity: String,
    @SerializedName("totalWeight") val totalWeight: Double,
    @SerializedName("orderCount") val orderCount: Int,
    @SerializedName("status") val status: String,
    @SerializedName("orders") val orders: List<BatchOrderDto>,
)

data class BatchListResponse(
    @SerializedName("batches") val batches: List<DeliveryBatchDto>,
)
