package com.imaginai.indel.data.api

import com.imaginai.indel.data.model.*
import retrofit2.Response
import retrofit2.http.*

interface WorkerApiService {
    @POST("api/v1/worker/onboard")
    suspend fun onboard(@Body request: OnboardRequest): Response<OnboardResponse>

    @GET("api/v1/worker/profile")
    suspend fun getProfile(): Response<WorkerProfileResponse>

    @GET("api/v1/worker/policy")
    suspend fun getPolicy(): Response<PolicyResponse>

    @GET("api/v1/worker/policy/premium")
    suspend fun getPremium(): Response<PremiumResponse>

    @POST("api/v1/worker/policy/enroll")
    suspend fun enrollPolicy(): Response<EnrollResponse>

    @POST("api/v1/worker/policy/premium/pay")
    suspend fun payPremium(@Body request: PayPremiumRequest): Response<PayPremiumResponse>

    @PUT("api/v1/worker/policy/pause")
    suspend fun pausePolicy(): Response<SimpleMessageResponse>

    @PUT("api/v1/worker/policy/cancel")
    suspend fun cancelPolicy(): Response<SimpleMessageResponse>

    @GET("api/v1/worker/earnings")
    suspend fun getEarnings(): Response<EarningsSummary>

    @GET("api/v1/worker/earnings/history")
    suspend fun getEarningsHistory(): Response<EarningsHistoryResponse>

    @GET("api/v1/worker/earnings/baseline")
    suspend fun getBaseline(): Response<BaselineResponse>

    @GET("api/v1/worker/claims")
    suspend fun getClaims(): Response<ClaimsResponse>

    @GET("api/v1/worker/claims/{claim_id}")
    suspend fun getClaimDetail(@Path("claim_id") claimId: String): Response<Claim>

    @GET("api/v1/worker/wallet")
    suspend fun getWallet(): Response<WalletResponse>

    @GET("api/v1/worker/payouts")
    suspend fun getPayouts(@Query("limit") limit: Int = 10): Response<PayoutsResponse>
}
