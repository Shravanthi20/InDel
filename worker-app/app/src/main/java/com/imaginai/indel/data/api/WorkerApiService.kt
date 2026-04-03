package com.imaginai.indel.data.api

import com.imaginai.indel.data.model.*
import retrofit2.Response
import retrofit2.http.*

interface WorkerApiService {
    // Auth
    @POST("api/v1/auth/register")
    suspend fun register(@Body request: RegisterRequest): Response<AuthResponse>

    @POST("api/v1/auth/login")
    suspend fun login(@Body request: LoginRequest): Response<AuthResponse>

    @POST("api/v1/auth/otp/send")
    suspend fun sendOtp(@Body request: OtpSendRequest): Response<OtpSendResponse>

    @POST("api/v1/auth/otp/verify")
    suspend fun verifyOtp(@Body request: OtpVerifyRequest): Response<OtpVerifyResponse>

    // Profile & Onboarding
    @POST("api/v1/worker/onboard")
    suspend fun onboard(@Body request: OnboardRequest): Response<OnboardResponse>

    @GET("api/v1/worker/profile")
    suspend fun getProfile(): Response<WorkerProfileResponse>

    @PUT("api/v1/worker/profile")
    suspend fun updateProfile(@Body request: OnboardRequest): Response<WorkerProfileResponse>

    // Orders & Delivery
    @GET("api/v1/demo/orders/available")
    suspend fun getAvailableOrders(): Response<OrderListResponse>

    @GET("api/v1/worker/orders/assigned")
    suspend fun getAssignedOrders(): Response<OrderListResponse>

    @GET("api/v1/worker/orders")
    suspend fun getAllOrders(): Response<OrderListResponse>

    @GET("api/v1/worker/orders/{order_id}")
    suspend fun getOrderDetail(@Path("order_id") orderId: String): Response<Order>

    @PUT("api/v1/worker/orders/{order_id}/accept")
    suspend fun acceptOrder(@Path("order_id") orderId: String): Response<SimpleMessageResponse>

    @PUT("api/v1/worker/orders/{order_id}/picked-up")
    suspend fun pickedUpOrder(@Path("order_id") orderId: String): Response<SimpleMessageResponse>

    @PUT("api/v1/worker/orders/{order_id}/delivered")
    suspend fun deliveredOrder(
        @Path("order_id") orderId: String,
        @Query("customer_code") customerCode: String
    ): Response<SimpleMessageResponse>

    @POST("api/v1/worker/orders/{order_id}/code/send")
    suspend fun sendCustomerCode(@Path("order_id") orderId: String): Response<SimpleMessageResponse>

    // Verification
    @POST("api/v1/worker/fetch-verification/send-code")
    suspend fun sendVerificationCode(): Response<SimpleMessageResponse>

    @POST("api/v1/worker/fetch-verification/verify")
    suspend fun verifyCode(@Body request: VerifyCodeRequest): Response<SimpleMessageResponse>

    @GET("api/v1/worker/zone-config")
    suspend fun getZoneConfig(): Response<ZoneConfigResponse>

    // Session Tracking
    @GET("api/v1/worker/session/{session_id}")
    suspend fun getSession(@Path("session_id") sessionId: String): Response<SessionResponse>

    @GET("api/v1/worker/session/{session_id}/deliveries")
    suspend fun getSessionDeliveries(@Path("session_id") sessionId: String): Response<OrderListResponse>

    @GET("api/v1/worker/session/{session_id}/fraud-signals")
    suspend fun getSessionFraudSignals(@Path("session_id") sessionId: String): Response<FraudSignalResponse>

    @PUT("api/v1/worker/session/{session_id}/end")
    suspend fun endSession(@Path("session_id") sessionId: String): Response<SimpleMessageResponse>

    // Earnings
    @GET("api/v1/worker/earnings")
    suspend fun getEarnings(): Response<EarningsSummary>

    @GET("api/v1/worker/earnings/history")
    suspend fun getEarningsHistory(): Response<EarningHistoryResponse>

    @GET("api/v1/worker/earnings/baseline")
    suspend fun getBaseline(): Response<BaselineResponse>

    // Policy
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

    // Claims & Wallet
    @GET("api/v1/worker/claims")
    suspend fun getClaims(): Response<ClaimsResponse>

    @GET("api/v1/worker/claims/{claim_id}")
    suspend fun getClaimDetail(@Path("claim_id") claimId: String): Response<Claim>

    @GET("api/v1/worker/wallet")
    suspend fun getWallet(): Response<WalletResponse>

    @GET("api/v1/worker/payouts")
    suspend fun getPayouts(@Query("limit") limit: Int = 10): Response<PayoutListResponse>

    // Notifications
    @GET("api/v1/worker/notifications")
    suspend fun getNotifications(): Response<NotificationListResponse>

    @PUT("api/v1/worker/notifications/preferences")
    suspend fun updateNotificationPreferences(@Body preferences: Map<String, Boolean>): Response<SimpleMessageResponse>

    @POST("api/v1/worker/notifications/fcm-token")
    suspend fun updateFcmToken(@Body request: FcmTokenRequest): Response<SimpleMessageResponse>

    // Demo Tools (Debug only)
    @POST("api/v1/demo/trigger-disruption")
    suspend fun triggerDisruption(@Body request: DisruptionRequest): Response<SimpleMessageResponse>

    @POST("api/v1/demo/assign-orders")
    suspend fun assignOrders(@Body request: CountRequest): Response<SimpleMessageResponse>

    @POST("api/v1/demo/simulate-deliveries")
    suspend fun simulateDeliveries(@Body request: CountRequest): Response<SimpleMessageResponse>

    @POST("api/v1/demo/settle-earnings")
    suspend fun settleEarnings(): Response<SimpleMessageResponse>

    @POST("api/v1/demo/reset-zone")
    suspend fun resetZone(): Response<SimpleMessageResponse>

    @POST("api/v1/demo/reset")
    suspend fun resetDemo(): Response<SimpleMessageResponse>
}
