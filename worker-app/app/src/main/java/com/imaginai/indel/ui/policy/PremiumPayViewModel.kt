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

/**
 * PremiumPayViewModel — used by PremiumPayScreen (standalone pay screen).
 *
 * On init:
 *   1. Fetches ML premium first (primary source)
 *   2. Falls back to policy's required_payment_inr / weekly_premium_inr
 *   3. Adds any late fee shown in policy
 *
 * After Razorpay returns:
 *   Records payment with backend (payPremium), then refreshes policy cache.
 */
@HiltViewModel
class PremiumPayViewModel @Inject constructor(
    private val policyRepository: PolicyRepository
) : ViewModel() {

    companion object {
        private const val TAG = "PremiumPayVM"
    }

    private val _uiState = MutableStateFlow<PayUiState>(PayUiState.Idle)
    val uiState = _uiState.asStateFlow()

    /** The displayable total amount (base + late fee) */
    private val _amount = MutableStateFlow("")
    val amount = _amount.asStateFlow()

    /** Base weekly premium from ML */
    private val _basePremium = MutableStateFlow(0)
    val basePremium = _basePremium.asStateFlow()

    /** Late fee from backend (₹1/day in grace period) */
    private val _lateFee = MutableStateFlow(0)
    val lateFee = _lateFee.asStateFlow()

    init {
        fetchMlThenPolicy()
    }

    private fun fetchMlThenPolicy() {
        viewModelScope.launch {
            try {
                // 1. ML is primary — fetch fresh premium quote
                var mlPremium: Int? = null
                val mlResp = policyRepository.getPremiumQuote()
                if (mlResp.isSuccessful) {
                    mlPremium = mlResp.body()?.weeklyPremiumInr
                    Log.d(TAG, "[ML] Premium=₹$mlPremium")
                } else {
                    Log.w(TAG, "[ML] Failed, falling back to policy stored value")
                }

                // 2. Get policy for late fee & stored premium fallback
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

                Log.d(TAG, "[Premium] base=₹$base late=₹$late total=₹$total")
            } catch (e: Exception) {
                Log.e(TAG, "Error fetching premium", e)
                // Keep empty — UI will show 0 / disabled
            }
        }
    }

    fun setLoading(isLoading: Boolean) {
        _uiState.value = if (isLoading) PayUiState.Loading else PayUiState.Idle
    }

    fun setPaymentError(error: String) {
        _uiState.value = PayUiState.Error(error)
    }

    /**
     * Called after Razorpay reports a successful payment.
     * Records payment with backend using the computed total amount.
     * Backend enforces no-duplicate via payment state.
     */
    fun recordPaymentSuccess(paymentId: String?) {
        viewModelScope.launch {
            _uiState.value = PayUiState.Loading
            try {
                val totalAmount = _amount.value.toIntOrNull() ?: 0
                val response = policyRepository.payPremium(totalAmount)
                if (response.isSuccessful) {
                    val paidAmount = response.body()?.amount ?: totalAmount
                    val summary = "Payment of ₹$paidAmount recorded — coverage cycle renewed"
                    Log.d(TAG, "[Payment] Success: $summary, Razorpay ID=$paymentId")
                    _uiState.value = PayUiState.Success(summary)
                } else {
                    val err = response.errorBody()?.string().orEmpty()
                    Log.w(TAG, "[Payment] Backend sync failed: $err")
                    _uiState.value = PayUiState.Error("Payment recorded by gateway but sync failed. Please refresh.")
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
