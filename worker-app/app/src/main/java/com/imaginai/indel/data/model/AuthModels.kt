package com.imaginai.indel.data.model

import com.google.gson.annotations.SerializedName

data class LoginRequest(
    @SerializedName("phone") val phone: String? = null,
    @SerializedName("email") val email: String? = null,
    @SerializedName("password") val password: String
)

data class RegisterRequest(
    @SerializedName("username") val username: String,
    @SerializedName("phone") val phone: String,
    @SerializedName("email") val email: String,
    @SerializedName("password") val password: String,
    @SerializedName("zone_level") val zoneLevel: String? = null,
    @SerializedName("zone_name") val zoneName: String? = null
)

data class AuthResponse(
    @SerializedName("token") val token: String,
    @SerializedName("token_type") val tokenType: String,
    @SerializedName("worker_id") val workerId: String
)

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
