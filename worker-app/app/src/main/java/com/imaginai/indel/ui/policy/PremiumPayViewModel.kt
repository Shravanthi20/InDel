package com.imaginai.indel.ui.policy

import android.util.Log
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

    companion object {
        private const val TAG = "PremiumPayVM"
    }

    private val _uiState = MutableStateFlow<PayUiState>(PayUiState.Idle)
    val uiState = _uiState.asStateFlow()

    private val _amount = MutableStateFlow("")
    val amount = _amount.asStateFlow()

    private val _basePremium = MutableStateFlow(0)
    val basePremium = _basePremium.asStateFlow()

    private val _lateFee = MutableStateFlow(0)
    val lateFee = _lateFee.asStateFlow()

    private val _paymentEnabled = MutableStateFlow(false)
    val paymentEnabled = _paymentEnabled.asStateFlow()

    private val _paymentHint = MutableStateFlow<String?>(null)
    val paymentHint = _paymentHint.asStateFlow()

    init {
        fetchMlThenPolicy()
    }

    private fun fetchMlThenPolicy() {
        viewModelScope.launch {
            try {
                var mlPremium: Int? = null
                val mlResp = policyRepository.getPremiumQuote()
                if (mlResp.isSuccessful) {
                    mlPremium = mlResp.body()?.weeklyPremiumInr
                    Log.d(TAG, "[ML] Premium=Rs $mlPremium")
                } else {
                    Log.w(TAG, "[ML] Failed, falling back to policy stored value")
                }

                val policyResp = policyRepository.getPolicy()
                val policy = if (policyResp.isSuccessful) policyResp.body()?.policy else null

                val base = mlPremium
                    ?: policy?.requiredPaymentInr
                    ?: policy?.weeklyPremiumInr
                    ?: 35

                val late = policy?.lateFeeInr ?: 0
                val total = base + late

                _basePremium.value = base
                _lateFee.value = late
                _amount.value = total.toString()
                _paymentEnabled.value = policy?.nextPaymentEnabled == true
                _paymentHint.value = when {
                    policy == null -> "Unable to confirm payment eligibility right now."
                    policy.coverageStatus.equals("NeedsActivation", ignoreCase = true) ->
                        "Start the plan from plan selection before paying the first premium."
                    policy.coverageStatus.equals("Deactivated", ignoreCase = true) ->
                        "This plan is deactivated. Re-enroll before making a payment."
                    policy.nextPaymentEnabled == true -> null
                    policy.daysSinceLastPayment != null && policy.billingCycleDays != null ->
                        "Next premium unlocks after ${policy.billingCycleDays} days. Current cycle day: ${policy.daysSinceLastPayment}."
                    else -> "Premium payment is not available for this billing cycle yet."
                }
            } catch (e: Exception) {
                Log.e(TAG, "Error fetching premium", e)
                _paymentEnabled.value = false
                _paymentHint.value = "Unable to load payment status."
            }
        }
    }

    fun setLoading(isLoading: Boolean) {
        _uiState.value = if (isLoading) PayUiState.Loading else PayUiState.Idle
    }

    fun setPaymentError(error: String) {
        _uiState.value = PayUiState.Error(error)
    }

    fun recordPaymentSuccess(paymentId: String?) {
        viewModelScope.launch {
            _uiState.value = PayUiState.Loading
            try {
                val totalAmount = _amount.value.toIntOrNull() ?: 0
                val response = policyRepository.payPremium(totalAmount)
                if (response.isSuccessful) {
                    val paidAmount = response.body()?.amount ?: totalAmount
                    val summary = "Payment of Rs $paidAmount recorded and coverage renewed."
                    Log.d(TAG, "[Payment] Success: $summary, Razorpay ID=$paymentId")
                    _uiState.value = PayUiState.Success(summary)
                } else {
                    val err = response.errorBody()?.string().orEmpty()
                    Log.w(TAG, "[Payment] Backend sync failed: $err")
                    _uiState.value = PayUiState.Error("Payment completed in Razorpay but backend sync failed. Please contact support before retrying.")
                }
            } catch (e: Exception) {
                Log.e(TAG, "[Payment] Exception", e)
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
