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

    init {
        fetchPolicyPremium()
    }

    private fun fetchPolicyPremium() {
        viewModelScope.launch {
            try {
                val response = policyRepository.getPolicy()
                if (response.isSuccessful) {
                    val policy = response.body()?.policy
                    val payable = policy?.requiredPaymentInr ?: policy?.weeklyPremiumInr ?: 0
                    _amount.value = payable.toString()
                }
            } catch (e: Exception) {
                // Fallback or ignore for init
            }
        }
    }

    fun setLoading(isLoading: Boolean) {
        if (isLoading) {
            _uiState.value = PayUiState.Loading
        } else {
            _uiState.value = PayUiState.Idle
        }
    }

    fun setPaymentError(error: String) {
        _uiState.value = PayUiState.Error(error)
    }

    fun recordPaymentSuccess(paymentId: String?) {
        viewModelScope.launch {
            _uiState.value = PayUiState.Loading  
            try {
                // We call the backend to record that Razorpay succeeded
                val response = policyRepository.payPremium(_amount.value.toIntOrNull())
                if (response.isSuccessful) {
                    val summary = "Payment Successful via Razorpay | ID: $paymentId"
                    _uiState.value = PayUiState.Success(summary)
                } else {
                    _uiState.value = PayUiState.Error("Failed to sync backend")
                }
            } catch (e: Exception) {
                _uiState.value = PayUiState.Error(e.message ?: "Unknown backend error")
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
