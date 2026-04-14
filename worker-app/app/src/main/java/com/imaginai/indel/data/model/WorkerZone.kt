package com.imaginai.indel.data.model

data class WorkerZone(
    val zoneLevel: String,
    val zoneName: String,
    val zoneId: Int,
    val city: String,
    val fromCity: String? = null,
    val toCity: String? = null
)