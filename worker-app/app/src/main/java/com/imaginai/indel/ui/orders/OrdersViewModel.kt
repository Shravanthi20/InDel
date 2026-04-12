package com.imaginai.indel.ui.orders

import android.util.Log
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.BatchOrderDto
import com.imaginai.indel.data.model.DeliveryBatchDto
import java.util.Locale
import com.imaginai.indel.data.repository.WorkerRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class OrdersViewModel @Inject constructor(
    private val workerRepository: WorkerRepository
) : ViewModel() {

    companion object {
        private const val TAG = "OrdersViewModel"
    }

    private val _uiState = MutableStateFlow<OrdersUiState>(OrdersUiState.Loading)
    val uiState = _uiState.asStateFlow()

    private val _isRefreshing = MutableStateFlow(false)
    val isRefreshing = _isRefreshing.asStateFlow()

    private val _selectedTab = MutableStateFlow(BatchLifecycleTab.AVAILABLE_NEAR)
    val selectedTab = _selectedTab.asStateFlow()

    private var cachedBatches: List<DeliveryBatch> = emptyList()

    init {
        loadOrders()
    }

    fun loadOrders() {
        viewModelScope.launch {
            _uiState.value = OrdersUiState.Loading
            fetchOrders()
        }
    }

    fun refresh() {
        viewModelScope.launch {
            _isRefreshing.value = true
            fetchOrders()
            delay(500)
            _isRefreshing.value = false
        }
    }

    private suspend fun fetchOrders() {
        try {
            val availableRes = workerRepository.getAvailableOrders()
            val workerOrdersRes = workerRepository.getAllOrders()
            val availableBatchesRes = workerRepository.getAvailableBatches()
            val assignedBatchesRes = workerRepository.getAssignedBatches()
            val deliveredBatchesRes = workerRepository.getDeliveredBatches()

            val availableOrders = if (availableRes.isSuccessful) {
                availableRes.body()?.orders ?: emptyList()
            } else {
                emptyList()
            }

            val workerOrders = if (workerOrdersRes.isSuccessful) {
                workerOrdersRes.body()?.orders ?: emptyList()
            } else {
                emptyList()
            }

            val batches = buildList {
                if (availableBatchesRes.isSuccessful) {
                    addAll((availableBatchesRes.body()?.batches ?: emptyList()).map { it.toUiBatch() })
                }
                if (assignedBatchesRes.isSuccessful) {
                    addAll((assignedBatchesRes.body()?.batches ?: emptyList()).map { it.toUiBatch() })
                }
                if (deliveredBatchesRes.isSuccessful) {
                    addAll((deliveredBatchesRes.body()?.batches ?: emptyList()).map { it.toUiBatch() })
                }
            }.distinctBy { it.batchId }

            cachedBatches = batches

            val availableNearBatches = batches.filter {
                val status = it.status.trim().lowercase(Locale.getDefault()).replace(" ", "_")
                status == "pending" || status == "assigned" || status == "accepted"
            }
            val pickedUpBatches = batches.filter {
                val status = it.status.trim().lowercase(Locale.getDefault()).replace(" ", "_")
                status == "picked_up"
            }
            val deliveryBatches = batches.filter {
                val status = it.status.trim().lowercase(Locale.getDefault()).replace(" ", "_")
                status == "delivered"
            }

            Log.d(
                TAG,
                "fetchOrders availableStatus=${availableRes.code()} allStatus=${workerOrdersRes.code()} " +
                    "batches availableNear=${availableNearBatches.size} picked=${pickedUpBatches.size} delivered=${deliveryBatches.size}"
            )

            if (availableRes.isSuccessful || workerOrdersRes.isSuccessful || availableBatchesRes.isSuccessful || assignedBatchesRes.isSuccessful || deliveredBatchesRes.isSuccessful) {
                _uiState.value = OrdersUiState.Success(
                    availableNearBatches = availableNearBatches,
                    pickedUpBatches = pickedUpBatches,
                    deliveryBatches = deliveryBatches,
                )
            } else {
                _uiState.value = OrdersUiState.Error(
                    "Failed to load orders (available=${availableRes.code()}, all=${workerOrdersRes.code()})"
                )
            }
        } catch (e: Exception) {
            Log.e(TAG, "fetchOrders exception", e)
            _uiState.value = OrdersUiState.Error(e.message ?: "Unknown error")
        }
    }

    fun acceptOrder(orderId: String) {
        viewModelScope.launch {
            val response = workerRepository.acceptOrder(orderId)
            if (response.isSuccessful) {
                // Refresh list to see the status change from 'assigned' to 'accepted'
                fetchOrders()
            }
        }
    }

    fun pickedUpOrder(orderId: String) {
        viewModelScope.launch {
            val response = workerRepository.pickedUpOrder(orderId)
            if (response.isSuccessful) {
                fetchOrders()
            }
        }
    }

    fun deliveredOrder(orderId: String, customerCode: String) {
        viewModelScope.launch {
            val response = workerRepository.deliveredOrder(orderId, customerCode)
            if (response.isSuccessful) {
                fetchOrders()
            }
        }
    }

    fun selectTab(tab: BatchLifecycleTab) {
        _selectedTab.value = tab
    }

    fun getBatchById(batchId: String): DeliveryBatch? = cachedBatches.firstOrNull { it.batchId == batchId }

    fun pickupCodeForBatch(batchId: String): String = codeFromBatchId(batchId, seed = 31)

    fun deliveryCodeForBatch(batchId: String): String = codeFromBatchId(batchId, seed = 37)

    fun isZoneASingleStop(batch: DeliveryBatch): Boolean {
        val from = batch.fromCity.trim().lowercase()
        val to = batch.toCity.trim().lowercase()
        return batch.zoneLevel.trim().equals("A", ignoreCase = true) && from.isNotBlank() && from == to
    }

    suspend fun acceptBatch(batch: DeliveryBatch, pickupCode: String): Boolean {
        val response = workerRepository.acceptBatch(batch.batchId, batch.orders.map { it.orderId }, pickupCode)
        if (response.isSuccessful) {
            fetchOrders()
            return true
        }
        return false
    }

    suspend fun deliverBatch(batch: DeliveryBatch, deliveryCode: String): BatchDeliveryResult {
        val response = workerRepository.deliverBatch(batch.batchId, deliveryCode)
        return if (response.isSuccessful) {
            fetchOrders()
            val body = response.body()
            BatchDeliveryResult(
                success = true,
                batchCompleted = body?.batchCompleted == true,
                remainingOrders = body?.remainingOrders,
                errorMessage = null,
            )
        } else {
            BatchDeliveryResult(
                success = false,
                batchCompleted = false,
                remainingOrders = null,
                errorMessage = "Unable to complete delivery right now.",
            )
        }
    }

    private fun codeFromBatchId(batchId: String, seed: Int): String {
        var value = seed
        batchId.trim().uppercase().forEach { char ->
            value = (value * if (seed == 31) 31 else 37 + char.code) % 9000
        }
        return (1000 + value).toString().padStart(4, '0')
    }

    private fun DeliveryBatchDto.toUiBatch(): DeliveryBatch {
        return DeliveryBatch(
            batchId = batchId,
            batchKey = batchKey,
            batchGroupKey = batchGroupKey,
            batchIndex = batchIndex,
            zoneLevel = zoneLevel,
            fromCity = fromCity,
            toCity = toCity,
            totalWeight = totalWeight,
            targetWeight = targetWeight,
            maxWeight = maxWeight,
            orderCount = orderCount,
            status = status,
            pickupCode = pickupCode,
            deliveryCode = deliveryCode,
            pickupTime = pickupTime,
            deliveryTime = deliveryTime,
            batchEarningInr = batchEarningInr,
            orders = orders.map { it.toUiOrder() },
        )
    }

    private fun BatchOrderDto.toUiOrder(): BatchOrder {
        return BatchOrder(
            orderId = orderId,
            deliveryAddress = deliveryAddress,
            contactName = contactName,
            contactPhone = contactPhone,
            weight = weight,
            pickupArea = pickupArea,
            dropArea = dropArea,
            deliveryCode = deliveryCode,
            status = status,
            pickupTime = pickupTime,
            deliveryTime = deliveryTime,
        )
    }
}

data class BatchDeliveryResult(
    val success: Boolean,
    val batchCompleted: Boolean,
    val remainingOrders: Int?,
    val errorMessage: String?,
)

enum class BatchLifecycleTab(val title: String) {
    AVAILABLE_NEAR("Available / Near"),
    PICKED_UP("Picked Up"),
    DELIVERY("Delivery"),
}

sealed class OrdersUiState {
    object Loading : OrdersUiState()
    data class Success(
        val availableNearBatches: List<DeliveryBatch>,
        val pickedUpBatches: List<DeliveryBatch>,
        val deliveryBatches: List<DeliveryBatch>,
    ) : OrdersUiState()
    data class Error(val message: String) : OrdersUiState()
}
