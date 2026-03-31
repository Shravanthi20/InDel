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
class DeliveryCompletionViewModel @Inject constructor(
    private val workerRepository: WorkerRepository
) : ViewModel() {

    private val _order = MutableStateFlow<Order?>(null)
    val order = _order.asStateFlow()

    private val _uiState = MutableStateFlow<CompletionUiState>(CompletionUiState.Idle)
    val uiState = _uiState.asStateFlow()

    fun loadOrder(orderId: String) {
        viewModelScope.launch {
            _uiState.value = CompletionUiState.Loading
            try {
                val response = workerRepository.getOrderDetail(orderId)
                if (response.isSuccessful) {
                    _order.value = response.body()
                    _uiState.value = CompletionUiState.Idle
                } else {
                    _uiState.value = CompletionUiState.Error("Failed to load order details")
                }
            } catch (e: Exception) {
                _uiState.value = CompletionUiState.Error(e.message ?: "Unknown error")
            }
        }
    }

    fun completeDelivery(orderId: String, customerCode: String) {
        viewModelScope.launch {
            _uiState.value = CompletionUiState.Loading
            try {
                val response = workerRepository.deliveredOrder(orderId, customerCode)
                if (response.isSuccessful) {
                    _uiState.value = CompletionUiState.Success
                } else {
                    _uiState.value = CompletionUiState.Error("Invalid customer code or failed to complete delivery")
                }
            } catch (e: Exception) {
                _uiState.value = CompletionUiState.Error(e.message ?: "Unknown error")
            }
        }
    }
}

sealed class CompletionUiState {
    object Idle : CompletionUiState()
    object Loading : CompletionUiState()
    object Success : CompletionUiState()
    data class Error(val message: String) : CompletionUiState()
}
