package com.imaginai.indel.ui.orders

data class BatchOrder(
    val orderId: String,
    val deliveryAddress: String,
    val contactName: String,
    val contactPhone: String,
    val weight: Double,
    val pickupArea: String? = null,
    val dropArea: String? = null,
    val deliveryCode: String? = null,
    val status: String? = null,
    val pickupTime: String? = null,
    val deliveryTime: String? = null,
)

data class DeliveryBatch(
    val batchId: String,
    val batchKey: String? = null,
    val batchGroupKey: String? = null,
    val batchIndex: Int? = null,
    val zoneLevel: String,
    val fromCity: String,
    val toCity: String,
    val totalWeight: Double,
    val targetWeight: Double? = null,
    val maxWeight: Double? = null,
    val orderCount: Int,
    val status: String,
    val pickupCode: String? = null,
    val deliveryCode: String? = null,
    val pickupTime: String? = null,
    val deliveryTime: String? = null,
    val batchEarningInr: Double? = null,
    val orders: List<BatchOrder>,
)
