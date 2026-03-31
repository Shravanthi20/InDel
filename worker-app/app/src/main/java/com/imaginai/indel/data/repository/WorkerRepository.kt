package com.imaginai.indel.data.repository

import com.imaginai.indel.data.api.WorkerApiService
import com.imaginai.indel.data.model.*
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class WorkerRepository @Inject constructor(
    private val workerApiService: WorkerApiService
) {
    // Profile
    suspend fun onboard(name: String, zone: String, vehicleType: String, upiId: String) =
        workerApiService.onboard(OnboardRequest(name, zone, vehicleType, upiId))

    suspend fun getProfile() = workerApiService.getProfile()

    suspend fun updateProfile(name: String, zone: String, vehicleType: String, upiId: String) =
        workerApiService.updateProfile(OnboardRequest(name, zone, vehicleType, upiId))

    // Orders & Delivery
    suspend fun getAvailableOrders() = workerApiService.getAvailableOrders()

    suspend fun getAssignedOrders() = workerApiService.getAssignedOrders()

    suspend fun getAllOrders() = workerApiService.getAllOrders()

    suspend fun getOrderDetail(orderId: String) = workerApiService.getOrderDetail(orderId)

    suspend fun acceptOrder(orderId: String) = workerApiService.acceptOrder(orderId)

    suspend fun pickedUpOrder(orderId: String) = workerApiService.pickedUpOrder(orderId)

    suspend fun deliveredOrder(orderId: String, customerCode: String) = 
        workerApiService.deliveredOrder(orderId, customerCode)

    suspend fun sendCustomerCode(orderId: String) = workerApiService.sendCustomerCode(orderId)

    // Verification
    suspend fun sendVerificationCode() = workerApiService.sendVerificationCode()

    suspend fun verifyCode(request: VerifyCodeRequest) = workerApiService.verifyCode(request)

    suspend fun getZoneConfig() = workerApiService.getZoneConfig()

    // Session
    suspend fun getSession(sessionId: String) = workerApiService.getSession(sessionId)

    suspend fun getSessionDeliveries(sessionId: String) = workerApiService.getSessionDeliveries(sessionId)

    suspend fun getSessionFraudSignals(sessionId: String) = workerApiService.getSessionFraudSignals(sessionId)

    suspend fun endSession(sessionId: String) = workerApiService.endSession(sessionId)

    // Notifications
    suspend fun getNotifications() = workerApiService.getNotifications()

    suspend fun updateNotificationPreferences(prefs: Map<String, Boolean>) = 
        workerApiService.updateNotificationPreferences(prefs)

    suspend fun updateFcmToken(token: String) = 
        workerApiService.updateFcmToken(FcmTokenRequest(token))

    // Demo / Debug
    suspend fun triggerDisruption(disruptionType: String, zone: String) = 
        workerApiService.triggerDisruption(DisruptionRequest(disruptionType, zone))

    suspend fun assignOrders(count: Int) = workerApiService.assignOrders(CountRequest(count))

    suspend fun simulateDeliveries(count: Int) = workerApiService.simulateDeliveries(CountRequest(count))

    suspend fun settleEarnings() = workerApiService.settleEarnings()

    suspend fun resetZone() = workerApiService.resetZone()

    suspend fun resetDemo() = workerApiService.resetDemo()
}
