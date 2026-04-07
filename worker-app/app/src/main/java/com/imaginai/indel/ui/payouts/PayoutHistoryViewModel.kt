package com.imaginai.indel.ui.payouts

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.PayoutRecord
import com.imaginai.indel.data.repository.ClaimsRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class PayoutHistoryViewModel @Inject constructor(
    private val claimsRepository: ClaimsRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<PayoutHistoryUiState>(PayoutHistoryUiState.Loading)
    val uiState = _uiState.asStateFlow()

    private val _isRefreshing = MutableStateFlow(false)
    val isRefreshing = _isRefreshing.asStateFlow()

    init {
        loadPayouts()
    }

    fun loadPayouts() {
        viewModelScope.launch {
            _uiState.value = PayoutHistoryUiState.Loading
            fetchPayouts()
        }
    }

    fun refresh() {
        viewModelScope.launch {
            _isRefreshing.value = true
            fetchPayouts()
            delay(400)
            _isRefreshing.value = false
        }
    }

    private suspend fun fetchPayouts() {
        try {
            val response = claimsRepository.getPayouts()
            if (response.isSuccessful) {
                _uiState.value = PayoutHistoryUiState.Success(response.body()?.payouts ?: emptyList())
            } else {
                _uiState.value = PayoutHistoryUiState.Error("Failed to load payout history")
            }
        } catch (e: Exception) {
            _uiState.value = PayoutHistoryUiState.Error(e.message ?: "Unknown error")
        }
    }
}

sealed class PayoutHistoryUiState {
    object Loading : PayoutHistoryUiState()
    data class Success(val payouts: List<PayoutRecord>) : PayoutHistoryUiState()
    data class Error(val message: String) : PayoutHistoryUiState()
}
