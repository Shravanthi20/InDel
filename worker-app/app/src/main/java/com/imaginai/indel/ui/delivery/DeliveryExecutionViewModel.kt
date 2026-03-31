package com.imaginai.indel.ui.delivery

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.Order
import com.imaginai.indel.data.repository.WorkerRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class DeliveryExecutionViewModel @Inject constructor(
    private val workerRepository: WorkerRepository
) : ViewModel() {

    private val _order = MutableStateFlow<Order?>(null)
    val order = _order.asStateFlow()

    private val _uiState = MutableStateFlow<ExecutionUiState>(ExecutionUiState.Idle)
    val uiState = _uiState.asStateFlow()

    fun loadOrder(orderId: String) {
        viewModelScope.launch {
            _uiState.value = ExecutionUiState.Loading
            try {
                val response = workerRepository.getOrderDetail(orderId)
                if (response.isSuccessful) {
                    _order.value = response.body()
                    _uiState.value = ExecutionUiState.Idle
                } else {
                    _uiState.value = ExecutionUiState.Error("Failed to load order details")
                }
            } catch (e: Exception) {
                _uiState.value = ExecutionUiState.Error(e.message ?: "Unknown error")
            }
        }
    }

    fun startPickup() {
        val currentOrder = _order.value ?: return
        viewModelScope.launch {
            _uiState.value = ExecutionUiState.Loading
            try {
                val response = workerRepository.pickedUpOrder(currentOrder.orderId)
                if (response.isSuccessful) {
                    loadOrder(currentOrder.orderId)
                } else {
                    _uiState.value = ExecutionUiState.Error("Failed to mark as picked up")
                }
            } catch (e: Exception) {
                _uiState.value = ExecutionUiState.Error(e.message ?: "Unknown error")
            }
        }
    }
}

sealed class ExecutionUiState {
    object Idle : ExecutionUiState()
    object Loading : ExecutionUiState()
    object Success : ExecutionUiState()
    data class Error(val message: String) : ExecutionUiState()
}
