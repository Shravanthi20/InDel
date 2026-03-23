package com.imaginai.indel.data.repository

import com.imaginai.indel.data.api.WorkerApiService
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class EarningsRepository @Inject constructor(
    private val workerApiService: WorkerApiService
) {
    suspend fun getEarnings() = workerApiService.getEarnings()
    suspend fun getEarningsHistory() = workerApiService.getEarningsHistory()
    suspend fun getBaseline() = workerApiService.getBaseline()
}
