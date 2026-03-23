package com.imaginai.indel.data.repository

import com.imaginai.indel.data.api.WorkerApiService
import com.imaginai.indel.data.model.PayPremiumRequest
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class PolicyRepository @Inject constructor(
    private val workerApiService: WorkerApiService
) {
    suspend fun getPolicy() = workerApiService.getPolicy()
    suspend fun getPremium() = workerApiService.getPremium()
    suspend fun enrollPolicy() = workerApiService.enrollPolicy()
    suspend fun payPremium(amount: Int?) = workerApiService.payPremium(PayPremiumRequest(amount))
    suspend fun pausePolicy() = workerApiService.pausePolicy()
    suspend fun cancelPolicy() = workerApiService.cancelPolicy()
}
