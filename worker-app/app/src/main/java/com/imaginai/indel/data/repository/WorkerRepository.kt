package com.imaginai.indel.data.repository

import com.imaginai.indel.data.api.WorkerApiService
import com.imaginai.indel.data.api.PlatformApiService
import com.imaginai.indel.data.local.dao.WorkerDao
import com.imaginai.indel.data.local.entity.WorkerEntity
import com.imaginai.indel.data.model.*
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class WorkerRepository @Inject constructor(
    private val workerApiService: WorkerApiService,
    private val platformApiService: PlatformApiService,
    private val workerDao: WorkerDao
) {
    // Local Offline Profile Flow
    fun getProfileFlow(): Flow<WorkerProfile?> {
        return workerDao.getWorkerProfileFlow().map { it.firstOrNull()?.toNetworkModel() }
    }
    suspend fun onboard(
        name: String,
        zoneLevel: String? = null,
        zoneName: String? = null,
        zoneId: Int? = null,
        city: String? = null,
        fromCity: String? = null,
        toCity: String? = null,
        vehicleType: String,
        vehicleName: String? = null,
        upiId: String
    ) = workerApiService.onboard(
        OnboardRequest(
            name = name,
            zoneLevel = zoneLevel,
            zoneName = zoneName,
            zoneId = zoneId,
            city = city,
            fromCity = fromCity,
            toCity = toCity,
            vehicleType = vehicleType,
            vehicleName = vehicleName,
            upiId = upiId
        )
    )

    suspend fun getProfile(): retrofit2.Response<WorkerProfileResponse> {
        val response = workerApiService.getProfile()
        if (response.isSuccessful) {
            response.body()?.worker?.let { worker ->
                workerDao.insertWorkerProfile(WorkerEntity.fromNetworkModel(worker))
            }
        }
        return response
    }

    suspend fun clearLocalProfile() {
        workerDao.clearProfile()
    }

    suspend fun updateOnlineStatus(online: Boolean) =
        workerApiService.updateOnlineStatus(OnlineStatusRequest(online))

    suspend fun heartbeat() = workerApiService.heartbeat()

    suspend fun getPolicy() = workerApiService.getPolicy()

    suspend fun updateProfile(
        name: String,
        zoneLevel: String,
        zoneName: String,
        zoneId: Int? = null,
        city: String? = null,
        fromCity: String? = null,
        toCity: String? = null,
        vehicleType: String,
        upiId: String
    ) = workerApiService.updateProfile(
        OnboardRequest(
            name = name,
            zoneLevel = zoneLevel,
            zoneName = zoneName,
            zoneId = zoneId,
            city = city,
            fromCity = fromCity,
            toCity = toCity,
            vehicleType = vehicleType,
            vehicleName = null,
            upiId = upiId
        )
    )

    // Zones
    suspend fun getZones() = platformApiService.getZones()

    suspend fun getZoneLevels() = platformApiService.getZoneLevels()
    
    suspend fun getZonePaths(type: String) = platformApiService.getZonePaths(type)

    // Orders & Delivery
    suspend fun getAvailableOrders(path: String? = null) = workerApiService.getAvailableOrders(path)

    suspend fun getAssignedOrders(path: String? = null) = workerApiService.getAssignedOrders(path)

    suspend fun getAllOrders(path: String? = null) = workerApiService.getAllOrders(path)

    suspend fun getAvailableBatches(limit: Int = 100) = workerApiService.getAvailableBatches(limit)

    suspend fun getAssignedBatches() = workerApiService.getAssignedBatches()

    suspend fun getDeliveredBatches() = workerApiService.getDeliveredBatches()

    suspend fun acceptBatch(batchId: String, orderIds: List<String>, pickupCode: String) =
        workerApiService.acceptBatch(batchId, BatchAcceptRequest(orderIds, pickupCode))

    suspend fun deliverBatch(batchId: String, deliveryCode: String) =
        workerApiService.deliverBatch(batchId, BatchDeliverRequest(deliveryCode))

    suspend fun getOrderDetail(orderId: String) = workerApiService.getOrderDetail(orderId)

    suspend fun acceptOrder(orderId: String) = workerApiService.acceptOrder(orderId)

    suspend fun pickedUpOrder(orderId: String) = workerApiService.pickedUpOrder(orderId)

    suspend fun deliveredOrder(orderId: String, customerCode: String) = 
        workerApiService.deliveredOrder(orderId, customerCode)

    suspend fun sendCustomerCode(orderId: String) = workerApiService.sendCustomerCode(orderId)

    // Verification
    suspend fun sendVerificationCode() = workerApiService.sendVerificationCode()

    suspend fun verifyCode(request: VerifyCodeRequest) = workerApiService.verifyCode(request)

    // Session
    suspend fun getSession(sessionId: String) = workerApiService.getSession(sessionId)

    suspend fun getSessionDeliveries(sessionId: String) = workerApiService.getSessionDeliveries(sessionId)

    suspend fun getSessionFraudSignals(sessionId: String) = workerApiService.getSessionFraudSignals(sessionId)

    suspend fun endSession(sessionId: String) = workerApiService.endSession(sessionId)

    // Earnings
    suspend fun getPayouts(limit: Int = 10) = workerApiService.getPayouts(limit)

    // Notifications
    suspend fun getNotifications() = workerApiService.getNotifications()

    suspend fun updateNotificationPreferences(prefs: Map<String, Boolean>) = 
        workerApiService.updateNotificationPreferences(prefs)

    suspend fun updateFcmToken(token: String) = 
        workerApiService.updateFcmToken(FcmTokenRequest(token))

    // Plans
    suspend fun getPlans() = workerApiService.getPlans()

    suspend fun selectPlan(planId: String, expectedDeliveries: Int?, paymentAmountInr: Int) =
        workerApiService.selectPlan(
            PlanSelectionRequest(
                planId = planId,
                expectedDeliveries = expectedDeliveries,
                paymentAmountInr = paymentAmountInr,
                paymentConfirmed = true,
            )
        )

    suspend fun skipPlan() = workerApiService.skipPlan()

    // Demo / Debug
    suspend fun triggerDisruption(disruptionType: String, zoneLevel: String, zoneName: String) = 
        workerApiService.triggerDisruption(DisruptionRequest(disruptionType, zoneLevel, zoneName))

    suspend fun assignOrders(count: Int) = workerApiService.assignOrders(CountRequest(count))

    suspend fun simulateDeliveries(count: Int) = workerApiService.simulateDeliveries(CountRequest(count))

    suspend fun settleEarnings() = workerApiService.settleEarnings()

    suspend fun resetZone() = workerApiService.resetZone()

    suspend fun resetDemo() = workerApiService.resetDemo()
}
