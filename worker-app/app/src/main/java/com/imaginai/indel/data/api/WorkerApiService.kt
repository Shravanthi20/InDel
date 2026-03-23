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
    suspend fun getPremium(): Response<SimpleMessageResponse>

    @POST("api/v1/worker/policy/enroll")
    suspend fun enrollPolicy(): Response<SimpleMessageResponse>

    @POST("api/v1/worker/policy/premium/pay")
    suspend fun payPremium(@Body request: PayPremiumRequest): Response<SimpleMessageResponse>

    @PUT("api/v1/worker/policy/pause")
    suspend fun pausePolicy(): Response<SimpleMessageResponse>

    @PUT("api/v1/worker/policy/cancel")
    suspend fun cancelPolicy(): Response<SimpleMessageResponse>

    @GET("api/v1/worker/earnings")
    suspend fun getEarnings(): Response<EarningsSummary>

    @GET("api/v1/worker/earnings/history")
    suspend fun getEarningsHistory(): Response<List<EarningRecord>>

    @GET("api/v1/worker/earnings/baseline")
    suspend fun getBaseline(): Response<SimpleMessageResponse>

    @GET("api/v1/worker/claims")
    suspend fun getClaims(): Response<ClaimsResponse>

    @GET("api/v1/worker/claims/{claim_id}")
    suspend fun getClaimDetail(@Path("claim_id") claimId: String): Response<Claim>

    @GET("api/v1/worker/wallet")
    suspend fun getWallet(): Response<WalletResponse>

    @GET("api/v1/worker/payouts")
    suspend fun getPayouts(@Query("limit") limit: Int = 10): Response<List<SimpleMessageResponse>>

    @GET("api/v1/worker/orders/assigned")
    suspend fun getAssignedOrders(): Response<List<Order>>

    @GET("api/v1/worker/orders")
    suspend fun getAllOrders(): Response<List<Order>>

    @PUT("api/v1/worker/orders/{order_id}/accept")
    suspend fun acceptOrder(@Path("order_id") orderId: String): Response<SimpleMessageResponse>

    @PUT("api/v1/worker/orders/{order_id}/picked-up")
    suspend fun pickedUpOrder(@Path("order_id") orderId: String): Response<SimpleMessageResponse>

    @PUT("api/v1/worker/orders/{order_id}/delivered")
    suspend fun deliveredOrder(@Path("order_id") orderId: String): Response<SimpleMessageResponse>

    @GET("api/v1/worker/notifications")
    suspend fun getNotifications(): Response<List<Notification>>
}
