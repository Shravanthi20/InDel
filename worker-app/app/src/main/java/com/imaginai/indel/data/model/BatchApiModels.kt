package com.imaginai.indel.data.model

import com.google.gson.annotations.SerializedName

data class BatchOrderDto(
    @SerializedName("orderId") val orderId: String,
    @SerializedName("deliveryAddress") val deliveryAddress: String,
    @SerializedName("contactName") val contactName: String,
    @SerializedName("contactPhone") val contactPhone: String,
    @SerializedName("weight") val weight: Double,
    @SerializedName("pickupArea") val pickupArea: String? = null,
    @SerializedName("dropArea") val dropArea: String? = null,
    @SerializedName("deliveryCode") val deliveryCode: String? = null,
    @SerializedName("status") val status: String? = null,
    @SerializedName("pickupTime") val pickupTime: String? = null,
    @SerializedName("deliveryTime") val deliveryTime: String? = null,
)

data class DeliveryBatchDto(
    @SerializedName("batchId") val batchId: String,
    @SerializedName("batchKey") val batchKey: String? = null,
    @SerializedName("batchGroupKey") val batchGroupKey: String? = null,
    @SerializedName("batchIndex") val batchIndex: Int? = null,
    @SerializedName("zoneLevel") val zoneLevel: String,
    @SerializedName("fromCity") val fromCity: String,
    @SerializedName("toCity") val toCity: String,
    @SerializedName("totalWeight") val totalWeight: Double,
    @SerializedName("targetWeight") val targetWeight: Double? = null,
    @SerializedName("maxWeight") val maxWeight: Double? = null,
    @SerializedName("orderCount") val orderCount: Int,
    @SerializedName("status") val status: String,
    @SerializedName("pickupCode") val pickupCode: String? = null,
    @SerializedName("deliveryCode") val deliveryCode: String? = null,
    @SerializedName("pickupTime") val pickupTime: String? = null,
    @SerializedName("deliveryTime") val deliveryTime: String? = null,
    @SerializedName("batchEarningInr") val batchEarningInr: Double? = null,
    @SerializedName("orders") val orders: List<BatchOrderDto>,
)

data class BatchListResponse(
    @SerializedName("batches") val batches: List<DeliveryBatchDto>,
)

data class BatchAcceptRequest(
    @SerializedName("orderIds") val orderIds: List<String>,
    @SerializedName("pickupCode") val pickupCode: String,
)

data class BatchAcceptResponse(
    @SerializedName("message") val message: String,
    @SerializedName("batchId") val batchId: String,
    @SerializedName("acceptedOrderIds") val acceptedOrderIds: List<String>,
)

data class BatchDeliverRequest(
    @SerializedName("deliveryCode") val deliveryCode: String,
)

data class BatchDeliverResponse(
    @SerializedName("message") val message: String,
    @SerializedName("batchId") val batchId: String,
    @SerializedName("batchCompleted") val batchCompleted: Boolean? = null,
    @SerializedName("remainingOrders") val remainingOrders: Int? = null,
    @SerializedName("orderId") val orderId: String? = null,
)
