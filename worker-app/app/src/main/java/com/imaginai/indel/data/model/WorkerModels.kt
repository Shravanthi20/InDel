package com.imaginai.indel.data.model

import com.google.gson.annotations.SerializedName

data class WorkerProfile(
    @SerializedName("worker_id") val workerId: String,
    @SerializedName("name") val name: String,
    @SerializedName("phone") val phone: String,
    @SerializedName("zone") val zone: String,
    @SerializedName("vehicle_type") val vehicleType: String,
    @SerializedName("upi_id") val upiId: String,
    @SerializedName("coverage_status") val coverageStatus: String,
    @SerializedName("enrolled") val enrolled: Boolean
)

data class OnboardRequest(
    @SerializedName("name") val name: String,
    @SerializedName("zone") val zone: String,
    @SerializedName("vehicle_type") val vehicleType: String,
    @SerializedName("upi_id") val upiId: String
)

data class OnboardResponse(
    @SerializedName("message") val message: String,
    @SerializedName("worker") val worker: WorkerProfile
)

data class WorkerProfileResponse(
    @SerializedName("worker") val worker: WorkerProfile
)
