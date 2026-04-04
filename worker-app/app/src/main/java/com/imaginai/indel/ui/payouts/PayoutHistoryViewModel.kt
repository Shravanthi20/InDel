package com.imaginai.indel.ui.payouts

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.PayoutRecord
import com.imaginai.indel.data.repository.WorkerRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class PayoutHistoryViewModel @Inject constructor(
    private val workerRepository: WorkerRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow<PayoutHistoryUiState>(PayoutHistoryUiState.Loading)
    val uiState = _uiState.asStateFlow()

    init {
        loadPayouts()
    }

    fun loadPayouts() {
        viewModelScope.launch {
            _uiState.value = PayoutHistoryUiState.Loading
            try {
                val response = workerRepository.getPayouts(limit = 20)
                val payouts = response.body()?.payouts ?: emptyList()
                _uiState.value = PayoutHistoryUiState.Success(
                    payouts = payouts,
                    latestAmount = payouts.firstOrNull()?.amount?.toInt() ?: 0,
                    latestProcessedAt = payouts.firstOrNull()?.processedAt,
                )
            } catch (e: Exception) {
                _uiState.value = PayoutHistoryUiState.Error(e.message ?: "Unknown error")
            }
        }
    }
}

sealed class PayoutHistoryUiState {
    object Loading : PayoutHistoryUiState()
    data class Success(
        val payouts: List<PayoutRecord>,
        val latestAmount: Int,
        val latestProcessedAt: String?,
    ) : PayoutHistoryUiState()
    data class Error(val message: String) : PayoutHistoryUiState()
}