package com.imaginai.indel.ui.orders

data class BatchOrder(
    val orderId: String,
    val deliveryAddress: String,
    val contactName: String,
    val contactPhone: String,
    val weight: Double,
)

data class DeliveryBatch(
    val batchId: String,
    val zoneLevel: String,
    val fromCity: String,
    val toCity: String,
    val totalWeight: Double,
    val orderCount: Int,
    val status: String,
    val orders: List<BatchOrder>,
)
