package com.imaginai.indel.ui.policy

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.repository.PolicyRepository
import dagger.hilt.android.lifecycle.HiltViewModel
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
                    _uiState.value = PayUiState.Success(response.body()?.message ?: "Payment Successful")
                } else {
                    _uiState.value = PayUiState.Error("Payment failed")
                }
            } catch (e: Exception) {
                _uiState.value = PayUiState.Error(e.message ?: "Unknown error")
            }
        }
    }
}

sealed class PayUiState {
    object Idle : PayUiState()
    object Loading : PayUiState()
    data class Success(val message: String) : PayUiState()
    data class Error(val message: String) : PayUiState()
}
