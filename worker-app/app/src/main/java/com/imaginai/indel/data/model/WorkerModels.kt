package com.imaginai.indel.data.model

import com.google.gson.annotations.SerializedName

data class WorkerProfile(
    @SerializedName("worker_id") val workerId: String,
    @SerializedName("name") val name: String,
    @SerializedName("phone") val phone: String,
    @SerializedName("zone") val zone: String? = null,
    @SerializedName("zone_level") val zoneLevel: String,
    @SerializedName("zone_name") val zoneName: String,
    @SerializedName("area") val area: String? = null,
    @SerializedName("zone_id") val zoneId: Int? = null,
    @SerializedName("city") val city: String? = null,
    @SerializedName("from_city") val fromCity: String? = null,
    @SerializedName("to_city") val toCity: String? = null,
    @SerializedName("vehicle_type") val vehicleType: String,
    @SerializedName("vehicle_name") val vehicleName: String? = null,
    @SerializedName("upi_id") val upiId: String,
    @SerializedName("coverage_status") val coverageStatus: String,
    @SerializedName("enrolled") val enrolled: Boolean
)

data class OnboardRequest(
    @SerializedName("name") val name: String,
    @SerializedName("zone_level") val zoneLevel: String? = null,
    @SerializedName("zone_name") val zoneName: String? = null,
    @SerializedName("area") val area: String? = null,
    @SerializedName("zone_id") val zoneId: Int? = null,
    @SerializedName("city") val city: String? = null,
    @SerializedName("from_city") val fromCity: String? = null,
    @SerializedName("to_city") val toCity: String? = null,
    @SerializedName("vehicle_type") val vehicleType: String,
    @SerializedName("vehicle_name") val vehicleName: String? = null,
    @SerializedName("upi_id") val upiId: String
)

data class OnboardResponse(
    @SerializedName("message") val message: String,
    @SerializedName("worker") val worker: WorkerProfile
)

data class WorkerProfileResponse(
    @SerializedName("worker") val worker: WorkerProfile
)
