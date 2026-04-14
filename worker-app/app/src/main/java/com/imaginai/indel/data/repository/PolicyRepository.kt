package com.imaginai.indel.data.repository

import com.imaginai.indel.data.api.WorkerApiService
import com.imaginai.indel.data.local.PreferencesDataStore
import com.imaginai.indel.data.model.DisruptionPayoutRequest
import com.imaginai.indel.data.model.DisruptionPayoutResponse
import com.imaginai.indel.data.model.PayPremiumRequest
import com.imaginai.indel.data.model.PayPremiumResponse
import com.imaginai.indel.data.model.Policy
import com.imaginai.indel.data.model.PremiumResponse
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.firstOrNull
import retrofit2.Response
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class PolicyRepository @Inject constructor(
    private val workerApiService: WorkerApiService,
    private val preferencesDataStore: PreferencesDataStore
) {
    // ── Cache ──────────────────────────────────────────────────────────────

    fun getCachedPolicy(): Flow<Policy?> = preferencesDataStore.getPolicyCache()

    suspend fun savePolicyCache(policy: Policy) = preferencesDataStore.savePolicyCache(policy)

    suspend fun getPolicyFromCache(): Policy? = getCachedPolicy().firstOrNull()

    // ── Backend Fetches ────────────────────────────────────────────────────

    /** Fetch from backend and update the local cache. Returns null on failure. */
    suspend fun fetchAndCachePolicy(): Policy? {
        val response = workerApiService.getPolicy()
        if (response.isSuccessful) {
            val policy = response.body()?.policy
            if (policy != null) savePolicyCache(policy)
            return policy
        }
        return null
    }

    suspend fun getPolicyAfterPayment(): Policy? = fetchAndCachePolicy()

    suspend fun getPolicy() = workerApiService.getPolicy()

    /**
     * Fetches the latest ML-computed premium.
     * ML is PRIMARY — this always calls the backend, never uses cache alone.
     */
    suspend fun getPremiumQuote(): Response<PremiumResponse> = workerApiService.getPremium()

    // ── Lifecycle Actions ──────────────────────────────────────────────────

    suspend fun enrollPolicy() = workerApiService.enrollPolicy()

    /**
     * Pay a premium with a specific amount.
     * On success, refreshes the policy cache so UI reflects the new cycle state.
     */
    suspend fun payPremium(amount: Int): Response<PayPremiumResponse> {
        val response = workerApiService.payPremium(PayPremiumRequest(amount))
        if (response.isSuccessful) {
            // Refresh cache so next screen load reflects the updated payment cycle
            fetchAndCachePolicy()
        }
        return response
    }

    /** Legacy shim for old callers that pass nullable amount. */
    suspend fun payPremiumNullable(amount: Int?) = workerApiService.payPremium(PayPremiumRequest(amount))

    suspend fun pausePolicy() = workerApiService.pausePolicy()

    suspend fun cancelPolicy(): Boolean {
        val response = workerApiService.cancelPolicy()
        if (response.isSuccessful) {
            fetchAndCachePolicy()
        }
        return response.isSuccessful
    }

    // ── Disruption Payout ──────────────────────────────────────────────────

    /**
     * Triggers a disruption payout for the worker.
     * Back-end gates on: active plan, zone match, risk_score < threshold.
     */
    suspend fun triggerDisruptionPayout(
        disruptionType: String,
        zoneLevel: String,
        zoneName: String,
        disruptionHours: Double = 4.0
    ): Response<DisruptionPayoutResponse> {
        return workerApiService.triggerDisruptionPayout(
            DisruptionPayoutRequest(
                disruptionType = disruptionType,
                zoneLevel = zoneLevel,
                zoneName = zoneName,
                disruptionHours = disruptionHours
            )
        )
    }
}
