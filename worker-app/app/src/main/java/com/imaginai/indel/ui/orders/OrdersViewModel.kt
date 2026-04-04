package com.imaginai.indel.ui.orders

import android.util.Log
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.DeliveryBatchDto
import com.imaginai.indel.data.model.Order
import com.imaginai.indel.data.model.WorkerProfile
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

    data class BatchDeliveryResult(
        val success: Boolean,
        val batchCompleted: Boolean,
        val remainingOrders: Int? = null,
    )

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
            val deliveredRes = workerRepository.getDeliveredBatches()

            val availableRaw = if (availableRes.isSuccessful) {
                availableRes.body()?.batches?.map { it.toUiModel() } ?: emptyList()
            } else {
                emptyList()
            }
            val assignedRaw = if (assignedRes.isSuccessful) {
                assignedRes.body()?.batches?.map { it.toUiModel() } ?: emptyList()
            } else {
                emptyList()
            }
            val deliveredRaw = if (deliveredRes.isSuccessful) {
                deliveredRes.body()?.batches?.map { it.toUiModel() } ?: emptyList()
            } else {
                emptyList()
            }

            val workerCity = resolveWorkerCity(worker)
            val workerOwned = dedupeByBatchId(assignedRaw + deliveredRaw)
            val assignedBatches = workerOwned.filter { normalizeBatchStatus(it.status) == "Assigned" }
            val pickedUpBatches = workerOwned.filter { normalizeBatchStatus(it.status) == "Picked Up" }
            val deliveredBatches = workerOwned.filter { normalizeBatchStatus(it.status) == "Delivered" }

            val usedOwnedIds = (assignedBatches + pickedUpBatches + deliveredBatches).map { it.batchId }.toHashSet()
            val availableBatches = dedupeByBatchId(availableRaw)
                .filter { normalizeBatchStatus(it.status) == "Assigned" }
                .filter { it.batchId !in usedOwnedIds }
                .filter { isReachableFromWorkerLocation(it, workerCity) }

            Log.d(
                TAG,
                "fetchOrders availableStatus=${availableRes.code()} assignedStatus=${assignedRes.code()} deliveredStatus=${deliveredRes.code()} " +
                    "availableCount=${availableBatches.size} assignedCount=${assignedBatches.size} pickedUpCount=${pickedUpBatches.size} deliveredCount=${deliveredBatches.size} zoneId=$workerZoneId workerCity=$workerCity"
            )

            val diagnostics = buildString {
                append("profile=")
                append(profileRes.code())
                append(" available=")
                append(availableRes.code())
                append(" assigned=")
                append(assignedRes.code())
                append(" delivered=")
                append(deliveredRes.code())
                append(" zoneId=")
                append(workerZoneId ?: "null")
                append(" counts=")
                append(availableBatches.size)
                append("/")
                append(assignedBatches.size)
                append("/")
                append(pickedUpBatches.size)
                append("/")
                append(deliveredBatches.size)
            }

            // Success if at least one call works (resilient loading)
            if (availableRes.isSuccessful || assignedRes.isSuccessful) {
                _uiState.value = OrdersUiState.Success(
                    availableBatches = availableBatches,
                    assignedBatches = assignedBatches,
                    pickedUpBatches = pickedUpBatches,
                    deliveredBatches = deliveredBatches,
                    diagnostics = diagnostics
                )
            } else {
                _uiState.value = OrdersUiState.Error(
                    "Failed to load orders (profile=${profileRes.code()}, available=${availableRes.code()}, assigned=${assignedRes.code()}, delivered=${deliveredRes.code()})"
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

    suspend fun acceptBatch(batch: DeliveryBatch, pickupCode: String): Boolean {
        return try {
            val response = workerRepository.acceptBatch(batch.batchId, batch.orders.map { it.orderId }, pickupCode)
            if (response.isSuccessful) {
                fetchOrders()
                true
            } else {
                Log.w(TAG, "acceptBatch failed code=${response.code()} batchId=${batch.batchId}")
                false
            }
        } catch (e: Exception) {
            Log.e(TAG, "acceptBatch exception", e)
            false
        }
    }

    suspend fun deliverBatch(batch: DeliveryBatch, deliveryCode: String): BatchDeliveryResult {
        return try {
            val response = workerRepository.deliverBatch(batch.batchId, deliveryCode)
            if (response.isSuccessful) {
                val responseBody = response.body()
                fetchOrders()
                BatchDeliveryResult(
                    success = true,
                    batchCompleted = responseBody?.batchCompleted ?: true,
                    remainingOrders = responseBody?.remainingOrders,
                )
            } else {
                Log.w(TAG, "deliverBatch failed code=${response.code()} batchId=${batch.batchId}")
                BatchDeliveryResult(success = false, batchCompleted = false)
            }
        } catch (e: Exception) {
            Log.e(TAG, "deliverBatch exception", e)
            BatchDeliveryResult(success = false, batchCompleted = false)
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
                isAssignedSection -> "Picked Up"
                else -> "Assigned"
            }

            DeliveryBatch(
                batchId = batchId,
                zoneLevel = zoneLevel,
                fromCity = fromCity,
                toCity = toCity,
                totalWeight = batchOrders.sumOf { it.weight },
                orderCount = batchOrders.size,
                status = status,
                pickupCode = pickupCodeForBatch(batchId),
                deliveryCode = deliveryCodeForBatch(batchId),
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
        val fallback = if (weight > 0) weight else 1.2
        return fallback.coerceIn(0.05, 5.0)
    }

    private fun normalizeBatchStatus(status: String): String {
        return when (status.trim().lowercase(Locale.ROOT)) {
            "assigned", "accepted" -> "Assigned"
            "picked_up", "picked up" -> "Picked Up"
            "delivered" -> "Delivered"
            else -> status.replace("_", " ").replaceFirstChar { it.uppercase() }
        }
    }

    private fun dedupeByBatchId(batches: List<DeliveryBatch>): List<DeliveryBatch> {
        val score: (String) -> Int = { status ->
            when (normalizeBatchStatus(status)) {
                "Delivered" -> 3
                "Picked Up" -> 2
                "Assigned" -> 1
                else -> 0
            }
        }

        val byId = mutableMapOf<String, DeliveryBatch>()
        batches.forEach { batch ->
            val existing = byId[batch.batchId]
            if (existing == null || score(batch.status) >= score(existing.status)) {
                byId[batch.batchId] = batch.copy(status = normalizeBatchStatus(batch.status))
            }
        }

        return byId.values.toList()
    }

    private fun resolveWorkerCity(worker: WorkerProfile?): String {
        if (worker == null) return ""

        val fromCity = worker.fromCity?.trim().orEmpty()
        val city = worker.city?.trim().orEmpty()
        val zoneName = worker.zoneName.trim()
        val zone = worker.zone?.trim().orEmpty()

        if (fromCity.isNotBlank()) return fromCity
        if (city.isNotBlank()) return city
        if (zoneName.isNotBlank()) return zoneName
        if (zone.isNotBlank()) return zone.split(",").first().trim()
        return ""
    }

    private fun isReachableFromWorkerLocation(batch: DeliveryBatch, workerCityRaw: String): Boolean {
        val workerCity = workerCityRaw.trim()
        if (workerCity.isBlank()) return false

        val fromCity = batch.fromCity.trim()
        val toCity = batch.toCity.trim()
        val zoneLevel = batch.zoneLevel.trim().uppercase(Locale.ROOT)

        return when (zoneLevel) {
            "A" -> fromCity.equals(workerCity, ignoreCase = true) && toCity.equals(workerCity, ignoreCase = true)
            "B", "C" -> fromCity.equals(workerCity, ignoreCase = true)
            else -> false
        }
    }

    fun pickupCodeForBatch(batchId: String): String {
        val normalized = batchId.trim().uppercase(Locale.ROOT)
        var seed = 0
        normalized.forEach { ch ->
            seed = (seed * 31 + ch.code) % 9000
        }
        return String.format(Locale.ROOT, "%04d", 1000 + seed)
    }

    fun deliveryCodeForBatch(batchId: String): String {
        val normalized = batchId.trim().uppercase(Locale.ROOT)
        var seed = 7
        normalized.forEach { ch ->
            seed = (seed * 37 + ch.code) % 9000
        }
        return String.format(Locale.ROOT, "%04d", 1000 + seed)
    }

    fun deliveryCodeForOrder(orderId: String): String {
        val normalized = orderId.trim().uppercase(Locale.ROOT)
        var seed = 11
        normalized.forEach { ch ->
            seed = (seed * 41 + ch.code) % 9000
        }
        return String.format(Locale.ROOT, "%04d", 1000 + seed)
    }

    fun isZoneASingleStop(batch: DeliveryBatch): Boolean {
        if (!batch.zoneLevel.equals("A", ignoreCase = true)) return false
        return batch.fromCity.trim().equals(batch.toCity.trim(), ignoreCase = true)
    }

    private fun DeliveryBatchDto.toUiModel(): DeliveryBatch {
        return DeliveryBatch(
            batchId = batchId,
            zoneLevel = zoneLevel,
            fromCity = fromCity,
            toCity = toCity,
            totalWeight = totalWeight,
            orderCount = orderCount,
            status = normalizeBatchStatus(status),
            pickupCode = pickupCode,
            deliveryCode = deliveryCode,
            pickupTime = pickupTime,
            deliveryTime = deliveryTime,
            batchEarningInr = batchEarningInr,
            orders = orders.map {
                val zoneASingleStop = zoneLevel.equals("A", ignoreCase = true) && fromCity.trim().equals(toCity.trim(), ignoreCase = true)
                BatchOrder(
                    orderId = it.orderId,
                    deliveryAddress = it.deliveryAddress,
                    contactName = it.contactName,
                    contactPhone = it.contactPhone,
                    weight = it.weight,
                    pickupArea = it.pickupArea,
                    dropArea = it.dropArea,
                    deliveryCode = if (zoneASingleStop) {
                        it.deliveryCode ?: deliveryCodeForOrder(it.orderId)
                    } else {
                        null
                    },
                    status = it.status,
                    pickupTime = it.pickupTime,
                    deliveryTime = it.deliveryTime,
                )
            },
        )
    }

    fun getBatchById(batchId: String): DeliveryBatch? {
        val state = _uiState.value
        if (state !is OrdersUiState.Success) return null
        return state.assignedBatches.firstOrNull { it.batchId == batchId }
            ?: state.pickedUpBatches.firstOrNull { it.batchId == batchId }
            ?: state.availableBatches.firstOrNull { it.batchId == batchId }
            ?: state.deliveredBatches.firstOrNull { it.batchId == batchId }
    }
}

sealed class OrdersUiState {
    object Loading : OrdersUiState()
    data class Success(
        val availableBatches: List<DeliveryBatch>,
        val assignedBatches: List<DeliveryBatch>,
        val pickedUpBatches: List<DeliveryBatch>,
        val deliveredBatches: List<DeliveryBatch>,
        val diagnostics: String
    ) : OrdersUiState()
    data class Error(val message: String) : OrdersUiState()
}
