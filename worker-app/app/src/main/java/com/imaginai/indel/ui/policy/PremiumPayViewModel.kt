package com.imaginai.indel.ui.policy

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.repository.PolicyRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class PremiumPayViewModel @Inject constructor(
    private val policyRepository: PolicyRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<PayUiState>(PayUiState.Idle)
    val uiState = _uiState.asStateFlow()

    private val _amount = MutableStateFlow("")
    val amount = _amount.asStateFlow()

    fun onAmountChanged(value: String) { _amount.value = value }

    fun pay() {
        viewModelScope.launch {
            _uiState.value = PayUiState.Loading
            try {
                val response = policyRepository.payPremium(_amount.value.toIntOrNull())
                if (response.isSuccessful) {
                    val body = response.body()
                    val summary = buildString {
                        append(body?.message ?: "Payment Successful")
                        if (body?.paymentId != null) {
                            append(" | ")
                            append(body.paymentId)
                        }
                        if (body?.paymentStatus != null) {
                            append(" | ")
                            append(body.paymentStatus)
                        }
                    }
                    _uiState.value = PayUiState.Success(summary)
                } else {
                    _uiState.value = PayUiState.Error("Payment failed")
                }
            } catch (e: Exception) {
                _uiState.value = PayUiState.Error(e.message ?: "Unknown error")
            }
        }
    }

    fun reset() {
        viewModelScope.launch {
            delay(50)
            _uiState.value = PayUiState.Idle
        }
    }
}

sealed class PayUiState {
    object Idle : PayUiState()
    object Loading : PayUiState()
    data class Success(val message: String) : PayUiState()
    data class Error(val message: String) : PayUiState()
}
