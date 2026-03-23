package com.imaginai.indel.data.repository

import com.imaginai.indel.data.api.WorkerApiService
import com.imaginai.indel.data.model.OnboardRequest
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class WorkerRepository @Inject constructor(
    private val workerApiService: WorkerApiService
) {
    suspend fun onboard(name: String, zone: String, vehicleType: String, upiId: String) =
        workerApiService.onboard(OnboardRequest(name, zone, vehicleType, upiId))

    suspend fun getProfile() = workerApiService.getProfile()

    suspend fun getAssignedOrders() = workerApiService.getAssignedOrders()

    suspend fun getAllOrders() = workerApiService.getAllOrders()

    suspend fun acceptOrder(orderId: String) = workerApiService.acceptOrder(orderId)

    suspend fun pickedUpOrder(orderId: String) = workerApiService.pickedUpOrder(orderId)

    suspend fun deliveredOrder(orderId: String) = workerApiService.deliveredOrder(orderId)

    suspend fun getNotifications() = workerApiService.getNotifications()
}
