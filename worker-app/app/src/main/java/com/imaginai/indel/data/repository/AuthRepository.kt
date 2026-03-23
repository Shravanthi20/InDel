package com.imaginai.indel.data.repository

import com.imaginai.indel.data.api.AuthApiService
import com.imaginai.indel.data.local.PreferencesDataStore
import com.imaginai.indel.data.model.OtpSendRequest
import com.imaginai.indel.data.model.OtpVerifyRequest
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AuthRepository @Inject constructor(
    private val authApiService: AuthApiService,
    private val preferencesDataStore: PreferencesDataStore
) {
    suspend fun sendOtp(phone: String) = authApiService.sendOtp(OtpSendRequest(phone))

    suspend fun verifyOtp(phone: String, otp: String) = 
        authApiService.verifyOtp(OtpVerifyRequest(phone, otp)).also { response ->
            if (response.isSuccessful) {
                response.body()?.let {
                    preferencesDataStore.saveAuthToken(it.token)
                    preferencesDataStore.saveWorkerId(it.workerId)
                }
            }
        }
    
    val authToken = preferencesDataStore.authToken
    val workerId = preferencesDataStore.workerId
    
    suspend fun logout() = preferencesDataStore.clearAll()
}
