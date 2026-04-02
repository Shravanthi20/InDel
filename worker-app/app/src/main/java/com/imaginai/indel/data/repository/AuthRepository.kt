package com.imaginai.indel.data.repository

import com.imaginai.indel.data.api.AuthApiService
import com.imaginai.indel.data.local.PreferencesDataStore
import com.imaginai.indel.data.model.*
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AuthRepository @Inject constructor(
    private val authApiService: AuthApiService,
    private val preferencesDataStore: PreferencesDataStore
) {
    suspend fun register(
        username: String,
        phone: String,
        email: String,
        password: String,
        zoneLevel: String? = null,
        zoneName: String? = null
    ) = authApiService.register(
        RegisterRequest(username, phone, email, password, zoneLevel, zoneName)
    ).also { response ->
        if (response.isSuccessful) {
            response.body()?.let {
                preferencesDataStore.saveAuthToken(it.token)
                preferencesDataStore.saveWorkerId(it.workerId)
            }
        }
    }

    suspend fun login(identifier: String, password: String) =
        authApiService.login(LoginRequest(phone = identifier, password = password)).also { response ->
            if (response.isSuccessful) {
                response.body()?.let {
                    preferencesDataStore.saveAuthToken(it.token)
                    preferencesDataStore.saveWorkerId(it.workerId)
                }
            }
        }

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
