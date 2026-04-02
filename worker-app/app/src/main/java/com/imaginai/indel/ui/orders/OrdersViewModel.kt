package com.imaginai.indel.ui.orders

import android.util.Log
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.DeliveryBatchDto
import com.imaginai.indel.data.model.Order
import com.imaginai.indel.data.repository.WorkerRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import java.time.LocalDateTime
import java.time.format.DateTimeFormatter
import java.util.Locale
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

    private val acceptedBatchIds = mutableSetOf<String>()

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
            // First, get the worker's profile to extract zone_id
            val profileRes = workerRepository.getProfile()
            val worker = profileRes.body()?.worker
            val workerZoneId = worker?.zoneId

            // Updated based on Backend implementation:
            // 1. Available pool for everyone (from demo route)
            // 2. Assigned/Active orders for the logged-in worker
            // Both now filtered by worker's zone if available
            val availableRes = workerRepository.getAvailableBatches()
            val assignedRes = workerRepository.getAssignedBatches()

            val availableBatches = if (availableRes.isSuccessful) {
                availableRes.body()?.batches?.map { it.toUiModel() } ?: emptyList()
            } else {
                emptyList()
            }
            val assignedBatches = if (assignedRes.isSuccessful) {
                assignedRes.body()?.batches?.map { it.toUiModel() } ?: emptyList()
            } else {
                emptyList()
            }

            Log.d(
                TAG,
                "fetchOrders availableStatus=${availableRes.code()} assignedStatus=${assignedRes.code()} " +
                    "availableCount=${availableBatches.size} assignedCount=${assignedBatches.size} zoneId=$workerZoneId"
            )

            val diagnostics = buildString {
                append("profile=")
                append(profileRes.code())
                append(" available=")
                append(availableRes.code())
                append(" assigned=")
                append(assignedRes.code())
                append(" zoneId=")
                append(workerZoneId ?: "null")
                append(" counts=")
                append(availableBatches.size)
                append("/")
                append(assignedBatches.size)
            }

            // Success if at least one call works (resilient loading)
            if (availableRes.isSuccessful || assignedRes.isSuccessful) {
                _uiState.value = OrdersUiState.Success(
                    availableBatches = availableBatches,
                    assignedBatches = assignedBatches,
                    diagnostics = diagnostics
                )
            } else {
                _uiState.value = OrdersUiState.Error(
                    "Failed to load orders (profile=${profileRes.code()}, available=${availableRes.code()}, assigned=${assignedRes.code()})"
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

    fun acceptBatch(batchId: String) {
        acceptedBatchIds.add(batchId)
        val current = _uiState.value
        if (current is OrdersUiState.Success) {
            _uiState.value = current.copy(
                availableBatches = current.availableBatches.map { batch ->
                    if (batch.batchId == batchId) batch.copy(status = "Assigned") else batch
                }
            )
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

    private fun buildBatches(orders: List<Order>, isAssignedSection: Boolean): List<DeliveryBatch> {
        if (orders.isEmpty()) return emptyList()

        val grouped = orders.groupBy {
            Triple(
                normalizeLocation(it.fromCity ?: it.pickupArea),
                normalizeLocation(it.toCity ?: it.dropArea),
                inferZoneLevel(it)
            )
        }

        return grouped.entries.mapIndexed { index, entry ->
            val fromCity = entry.key.first
            val toCity = entry.key.second
            val zoneLevel = entry.key.third
            val batchOrders = entry.value.map { order ->
                BatchOrder(
                    orderId = order.orderId,
                    deliveryAddress = order.address ?: order.dropArea,
                    contactName = order.customerName ?: "Customer",
                    contactPhone = order.customerPhone ?: "N/A",
                    weight = normalizeWeight(order.packageWeightKg),
                )
            }
            val timestamp = LocalDateTime.now().plusSeconds(index.toLong()).format(DateTimeFormatter.ofPattern("yyyyMMddHHmmss"))
            val batchId = buildBatchId(
                zoneLevel = zoneLevel,
                fromCity = fromCity,
                toCity = toCity,
                fromState = entry.value.firstOrNull()?.fromState.orEmpty(),
                toState = entry.value.firstOrNull()?.toState.orEmpty(),
                timestamp = timestamp,
            )
            val status = when {
                acceptedBatchIds.contains(batchId) -> "Assigned"
                isAssignedSection -> "Out for Delivery"
                else -> "Pending"
            }

            DeliveryBatch(
                batchId = batchId,
                zoneLevel = zoneLevel,
                fromCity = fromCity,
                toCity = toCity,
                totalWeight = batchOrders.sumOf { it.weight },
                orderCount = batchOrders.size,
                status = status,
                orders = batchOrders,
            )
        }
    }

    private fun inferZoneLevel(order: Order): String {
        val from = normalizeLocation(order.fromCity ?: order.pickupArea)
        val to = normalizeLocation(order.toCity ?: order.dropArea)
        val fromState = normalizeState(order.fromState)
        val toState = normalizeState(order.toState)
        return when {
            from.equals(to, ignoreCase = true) -> "A"
            fromState.isNotBlank() && toState.isNotBlank() && fromState == toState -> "B"
            else -> "C"
        }
    }

    private fun buildBatchId(
        zoneLevel: String,
        fromCity: String,
        toCity: String,
        fromState: String,
        toState: String,
        timestamp: String,
    ): String {
        val zone = zoneLevel.uppercase(Locale.ROOT)
        val cityCode = when (zone) {
            "A" -> codePart(fromCity, 6)
            else -> codePart(fromCity, 3) + codePart(toCity, 3)
        }
        val stateCode = when (zone) {
            "C" -> codePart(fromState, 2) + codePart(toState, 2)
            else -> codePart(if (fromState.isNotBlank()) fromState else toState, 4)
        }
        return zone + cityCode + stateCode + timestamp
    }

    private fun codePart(value: String, length: Int): String {
        val cleaned = value.uppercase(Locale.ROOT).replace(" ", "")
        return cleaned.take(length).padEnd(length, 'X')
    }

    private fun normalizeState(state: String?): String = state.orEmpty().trim().lowercase(Locale.ROOT)

    private fun normalizeLocation(location: String): String {
        val trimmed = location.trim()
        return if (trimmed.isBlank()) "Unknown" else trimmed
    }

    private fun normalizeWeight(weight: Double): Double {
        return if (weight > 0) weight else 1.2
    }

    private fun DeliveryBatchDto.toUiModel(): DeliveryBatch {
        return DeliveryBatch(
            batchId = batchId,
            zoneLevel = zoneLevel,
            fromCity = fromCity,
            toCity = toCity,
            totalWeight = totalWeight,
            orderCount = orderCount,
            status = status,
            orders = orders.map {
                BatchOrder(
                    orderId = it.orderId,
                    deliveryAddress = it.deliveryAddress,
                    contactName = it.contactName,
                    contactPhone = it.contactPhone,
                    weight = it.weight,
                )
            },
        )
    }

    fun getBatchById(batchId: String): DeliveryBatch? {
        val state = _uiState.value
        if (state !is OrdersUiState.Success) return null
        return state.assignedBatches.firstOrNull { it.batchId == batchId }
            ?: state.availableBatches.firstOrNull { it.batchId == batchId }
    }
}

sealed class OrdersUiState {
    object Loading : OrdersUiState()
    data class Success(
        val availableBatches: List<DeliveryBatch>,
        val assignedBatches: List<DeliveryBatch>,
        val diagnostics: String
    ) : OrdersUiState()
    data class Error(val message: String) : OrdersUiState()
}
