package com.imaginai.indel.ui.orders

import android.util.Log
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.Order
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
            // Updated based on Backend implementation:
            // 1. Available pool for everyone (from demo route)
            val availableRes = workerRepository.getAvailableOrders()
            // 2. Assigned/Active orders for the logged-in worker
            val assignedRes = workerRepository.getAssignedOrders()

            val availableOrders = if (availableRes.isSuccessful) {
                availableRes.body()?.orders ?: emptyList()
            } else {
                emptyList()
            }
            val assignedOrders = if (assignedRes.isSuccessful) {
                assignedRes.body()?.orders ?: emptyList()
            } else {
                emptyList()
            }

            Log.d(
                TAG,
                "fetchOrders availableStatus=${availableRes.code()} assignedStatus=${assignedRes.code()} " +
                    "availableCount=${availableOrders.size} assignedCount=${assignedOrders.size}"
            )

            // Success if at least one call works (resilient loading)
            if (availableRes.isSuccessful || assignedRes.isSuccessful) {
                _uiState.value = OrdersUiState.Success(
                    availableOrders = availableOrders,
                    assignedOrders = assignedOrders
                )
            } else {
                _uiState.value = OrdersUiState.Error(
                    "Failed to load orders (available=${availableRes.code()}, assigned=${assignedRes.code()})"
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
}

sealed class OrdersUiState {
    object Loading : OrdersUiState()
    data class Success(
        val availableOrders: List<Order>,
        val assignedOrders: List<Order>
    ) : OrdersUiState()
    data class Error(val message: String) : OrdersUiState()
}
