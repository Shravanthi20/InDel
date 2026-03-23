package com.imaginai.indel.data.model

import com.google.gson.annotations.SerializedName

data class OtpSendRequest(
    @SerializedName("phone") val phone: String
)

data class OtpSendResponse(
    @SerializedName("message") val message: String,
    @SerializedName("otp_for_testing") val otpForTesting: String?,
    @SerializedName("phone") val phone: String,
    @SerializedName("expires_in_seconds") val expiresInSeconds: Int
)

data class OtpVerifyRequest(
    @SerializedName("phone") val phone: String,
    @SerializedName("otp") val otp: String
)

data class OtpVerifyResponse(
    @SerializedName("message") val message: String,
    @SerializedName("token") val token: String,
    @SerializedName("token_type") val tokenType: String,
    @SerializedName("worker_id") val workerId: String
)
