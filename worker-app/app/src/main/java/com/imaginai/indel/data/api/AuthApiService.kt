package com.imaginai.indel.data.api

import com.imaginai.indel.data.model.OtpSendRequest
import com.imaginai.indel.data.model.OtpSendResponse
import com.imaginai.indel.data.model.OtpVerifyRequest
import com.imaginai.indel.data.model.OtpVerifyResponse
import retrofit2.Response
import retrofit2.http.Body
import retrofit2.http.POST

interface AuthApiService {
    @POST("api/v1/auth/otp/send")
    suspend fun sendOtp(@Body request: OtpSendRequest): Response<OtpSendResponse>

    @POST("api/v1/auth/otp/verify")
    suspend fun verifyOtp(@Body request: OtpVerifyRequest): Response<OtpVerifyResponse>
}
