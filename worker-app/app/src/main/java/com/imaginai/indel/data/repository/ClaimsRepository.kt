package com.imaginai.indel.data.repository

import com.imaginai.indel.data.api.WorkerApiService
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ClaimsRepository @Inject constructor(
    private val workerApiService: WorkerApiService
) {
    suspend fun getClaims() = workerApiService.getClaims()
    suspend fun getClaimDetail(claimId: String) = workerApiService.getClaimDetail(claimId)
    suspend fun getWallet() = workerApiService.getWallet()
    suspend fun getPayouts(limit: Int = 10) = workerApiService.getPayouts(limit)
}
