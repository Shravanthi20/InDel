package com.imaginai.indel.ui.policy

import android.util.Log
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.Policy
import com.imaginai.indel.data.model.ShapImpact
import com.imaginai.indel.data.repository.PolicyRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

/**
 * PolicyViewModel — full lifecycle state machine.
 *
 * State machine:
 *   INACTIVE / NeedsActivation  →  pay 2× weekly_premium (SelectPlan or re-enroll)
 *   ACTIVE (Locked)             →  7-day billing cycle running — payment button disabled
 *   ACTIVE (Eligible)           →  payment window open (may include late_fee)
 *   ACTIVE (Grace)              →  inside 2-day grace period, late_fee accumulating
 *   INACTIVE / Deactivated      →  missed grace — restart requires 2×
 *   INACTIVE (manual stop)      →  user cancelled via "Stop Plan"
 */
@HiltViewModel
class PolicyViewModel @Inject constructor(
    private val policyRepository: PolicyRepository
) : ViewModel() {

    companion object {
        private const val TAG = "PolicyViewModel"
    }

    private val _uiState = MutableStateFlow<PolicyUiState>(PolicyUiState.Loading)
    val uiState = _uiState.asStateFlow()

    private val _isRefreshing = MutableStateFlow(false)
    val isRefreshing = _isRefreshing.asStateFlow()

    // Premium ML state shown in payment flow
    private val _mlPremium = MutableStateFlow<MlPremiumState>(MlPremiumState.Idle)
    val mlPremium = _mlPremium.asStateFlow()

    // In-flight action feedback
    private val _actionError = MutableStateFlow<String?>(null)
    val actionError = _actionError.asStateFlow()

    init {
        loadPolicy()
    }

    // ── Load / Refresh ─────────────────────────────────────────────────────

    fun loadPolicy() {
        viewModelScope.launch {
            _uiState.value = PolicyUiState.Loading
            // Cache first for snappy UI
            val cached = policyRepository.getPolicyFromCache()
            if (cached != null) {
                _uiState.value = PolicyUiState.Success(cached)
            }
            // Always freshen from backend (ML re-prices on each GET /policy)
            val fresh = policyRepository.fetchAndCachePolicy()
            if (fresh != null) {
                _uiState.value = PolicyUiState.Success(fresh)
            } else if (cached == null) {
                _uiState.value = PolicyUiState.Error("Failed to load policy")
            }
        }
    }

    fun refresh() {
        viewModelScope.launch {
            _isRefreshing.value = true
            val policy = policyRepository.fetchAndCachePolicy()
            if (policy != null) {
                _uiState.value = PolicyUiState.Success(policy)
            }
            _isRefreshing.value = false
        }
    }

    // ── ML Premium Fetch ───────────────────────────────────────────────────

    /**
     * Called before the payment flow.
     * ML is primary — always hits the backend for a fresh quote.
     * Falls back to policy.weeklyPremiumInr if ML is unavailable.
     */
    fun fetchMlPremium() {
        viewModelScope.launch {
            _mlPremium.value = MlPremiumState.Loading
            try {
                val resp = policyRepository.getPremiumQuote()
                if (resp.isSuccessful) {
                    val body = resp.body()!!
                    _mlPremium.value = MlPremiumState.Ready(
                        weeklyPremium = body.weeklyPremiumInr,
                        riskScore = body.riskScore,
                        pricingSource = body.pricingSource ?: "ml",
                        modelVersion = body.modelVersion,
                        shapBreakdown = body.shapBreakdown
                    )
                    Log.d(TAG, "[ML] Premium=${body.weeklyPremiumInr} source=${body.pricingSource}")
                } else {
                    val fallback = currentPolicyWeeklyPremium()
                    _mlPremium.value = MlPremiumState.Fallback(fallback)
                    Log.w(TAG, "[ML] Failed — using fallback ₹$fallback")
                }
            } catch (e: Exception) {
                val fallback = currentPolicyWeeklyPremium()
                _mlPremium.value = MlPremiumState.Fallback(fallback)
                Log.e(TAG, "[ML] Exception — fallback ₹$fallback", e)
            }
        }
    }

    private fun currentPolicyWeeklyPremium(): Int {
        return (uiState.value as? PolicyUiState.Success)?.policy?.weeklyPremiumInr ?: 35
    }

    // ── Start Plan (Activation) ────────────────────────────────────────────

    /**
     * Start a plan for the first time or after deactivation.
     * First payment = 2 × weekly_premium (enforced on backend via SelectPlan).
     * This calls /policy/enroll then freshens state.
     */
    fun startPlanWithPayment(policy: Policy) {
        viewModelScope.launch {
            _uiState.value = PolicyUiState.Loading
            _actionError.value = null
            try {
                // Fetch latest ML premium first (ML is primary)
                val premiumResp = policyRepository.getPremiumQuote()
                val mlAmount = if (premiumResp.isSuccessful) {
                    premiumResp.body()?.weeklyPremiumInr ?: policy.weeklyPremiumInr
                } else {
                    policy.weeklyPremiumInr
                }

                // First payment = 2× weekly premium
                val activationAmount = (policy.initialPaymentMultiplier ?: 2) * mlAmount

                Log.d(TAG, "[Start Plan] ML premium=₹$mlAmount, activation=₹$activationAmount")

                val payResp = policyRepository.payPremium(activationAmount)
                if (payResp.isSuccessful) {
                    // Backend also enrolls on payment; just freshen
                    val updated = policyRepository.fetchAndCachePolicy()
                    if (updated != null) {
                        _uiState.value = PolicyUiState.Success(updated)
                    } else {
                        loadPolicy()
                    }
                } else {
                    val errBody = payResp.errorBody()?.string().orEmpty()
                    Log.w(TAG, "[Start Plan] Payment failed: $errBody")

                    // Try enroll then retry payment
                    policyRepository.enrollPolicy()
                    val retryResp = policyRepository.payPremium(activationAmount)
                    if (retryResp.isSuccessful) {
                        val updated = policyRepository.fetchAndCachePolicy()
                        if (updated != null) {
                            _uiState.value = PolicyUiState.Success(updated)
                            return@launch
                        }
                    }
                    _uiState.value = PolicyUiState.Error("Activation payment failed. Please try again.")
                }
            } catch (e: Exception) {
                Log.e(TAG, "[Start Plan] Exception", e)
                _uiState.value = PolicyUiState.Error("Failed to start plan: ${e.localizedMessage}")
            }
        }
    }

    // ── Weekly Premium Payment ─────────────────────────────────────────────

    /**
     * Pay the weekly premium.
     * 1. Fetch latest ML premium
     * 2. Compute total = base_premium + late_fee_from_policy
     * 3. POST /policy/premium/pay with that amount
     * 4. On success: refresh policy (button auto-disables via nextPaymentEnabled=false)
     *
     * Button should only be enabled when nextPaymentEnabled == true.
     */
    fun payWeeklyPremium() {
        viewModelScope.launch {
            val currentPolicy = (uiState.value as? PolicyUiState.Success)?.policy
            if (currentPolicy == null) {
                _actionError.value = "Policy not loaded"
                return@launch
            }

            _uiState.value = PolicyUiState.Loading
            _actionError.value = null

            try {
                // ML is primary — recalculate before every payment
                val premiumResp = policyRepository.getPremiumQuote()
                val basePremium = if (premiumResp.isSuccessful) {
                    premiumResp.body()?.weeklyPremiumInr ?: currentPolicy.weeklyPremiumInr
                } else {
                    currentPolicy.weeklyPremiumInr
                }

                val lateFee = currentPolicy.lateFeeInr ?: 0
                val totalAmount = basePremium + lateFee

                Log.d(TAG, "[Pay Weekly] ML base=₹$basePremium + late=₹$lateFee = ₹$totalAmount")

                val payResp = policyRepository.payPremium(totalAmount)
                if (payResp.isSuccessful) {
                    // Re-fetch with cache update — button will auto-disable via nextPaymentEnabled
                    val updated = policyRepository.fetchAndCachePolicy()
                    _uiState.value = if (updated != null) {
                        PolicyUiState.Success(updated)
                    } else {
                        PolicyUiState.PaymentSuccess(basePremium, lateFee, totalAmount)
                    }
                } else {
                    val err = payResp.errorBody()?.string().orEmpty()
                    Log.w(TAG, "[Pay Weekly] Failed: $err")
                    val updated = policyRepository.fetchAndCachePolicy()
                    _uiState.value = if (updated != null) {
                        PolicyUiState.Success(updated)
                    } else {
                        PolicyUiState.Error("Payment failed. Please check your plan status.")
                    }
                }
            } catch (e: Exception) {
                Log.e(TAG, "[Pay Weekly] Exception", e)
                _uiState.value = PolicyUiState.Error("Payment error: ${e.localizedMessage}")
            }
        }
    }

    // ── Stop Plan ──────────────────────────────────────────────────────────

    /**
     * Stop (cancel) the active plan immediately.
     * Shows confirmation dialog before calling this.
     */
    fun stopPlan() {
        viewModelScope.launch {
            _uiState.value = PolicyUiState.Loading
            _actionError.value = null
            try {
                val ok = policyRepository.cancelPolicy()
                val updated = policyRepository.fetchAndCachePolicy()
                _uiState.value = if (updated != null) {
                    PolicyUiState.Success(updated)
                } else if (ok) {
                    PolicyUiState.PlanStopped
                } else {
                    PolicyUiState.Error("Failed to stop plan")
                }
            } catch (e: Exception) {
                Log.e(TAG, "[Stop Plan] Exception", e)
                _uiState.value = PolicyUiState.Error("Failed to stop plan: ${e.localizedMessage}")
            }
        }
    }

    // ── Legacy / Compat ────────────────────────────────────────────────────

    fun enroll() {
        viewModelScope.launch {
            _uiState.value = PolicyUiState.Loading
            policyRepository.enrollPolicy()
            val policy = policyRepository.fetchAndCachePolicy()
            _uiState.value = if (policy != null) PolicyUiState.Success(policy)
            else PolicyUiState.Error("Failed to load policy after enroll")
        }
    }

    fun pause() {
        viewModelScope.launch {
            _uiState.value = PolicyUiState.Loading
            policyRepository.pausePolicy()
            val policy = policyRepository.fetchAndCachePolicy()
            _uiState.value = if (policy != null) PolicyUiState.Success(policy)
            else PolicyUiState.Error("Failed to sync after pause")
        }
    }

    /** Alias kept for existing callsites that navigate to Stop Plan confirmation. */
    fun cancel() = stopPlan()

    fun clearActionError() {
        _actionError.value = null
    }
}

// ── UI States ──────────────────────────────────────────────────────────────

sealed class PolicyUiState {
    object Loading : PolicyUiState()
    data class Success(val policy: Policy) : PolicyUiState()
    data class PaymentSuccess(
        val basePremium: Int,
        val lateFee: Int,
        val totalPaid: Int
    ) : PolicyUiState()
    object PlanStopped : PolicyUiState()
    data class Error(val message: String) : PolicyUiState()
}

sealed class MlPremiumState {
    object Idle : MlPremiumState()
    object Loading : MlPremiumState()
    data class Ready(
        val weeklyPremium: Int,
        val riskScore: Double?,
        val pricingSource: String,
        val modelVersion: String?,
        val shapBreakdown: List<ShapImpact>
    ) : MlPremiumState()
    data class Fallback(val weeklyPremium: Int) : MlPremiumState()
}
