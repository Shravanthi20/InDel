package com.imaginai.indel.ui.plan

import android.util.Log
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.DeliveryPlan
import com.imaginai.indel.data.model.Policy
import com.imaginai.indel.data.repository.WorkerRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import org.json.JSONObject
import javax.inject.Inject

@HiltViewModel
class PlanSelectionViewModel @Inject constructor(
    private val workerRepository: WorkerRepository
) : ViewModel() {

    companion object {
        private const val TAG = "PlanSelection"
    }

    private val _uiState = MutableStateFlow<PlanUiState>(PlanUiState.Loading)
    val uiState = _uiState.asStateFlow()

    private val _selectedPlan = MutableStateFlow<DeliveryPlan?>(null)
    val selectedPlan = _selectedPlan.asStateFlow()

    private val _selectedExpectedDeliveries = MutableStateFlow<Int?>(null)
    val selectedExpectedDeliveries = _selectedExpectedDeliveries.asStateFlow()

    private val _isPaymentRequired = MutableStateFlow(false)
    val isPaymentRequired = _isPaymentRequired.asStateFlow()

    private val _currentPolicy = MutableStateFlow<Policy?>(null)
    val currentPolicy = _currentPolicy.asStateFlow()

    private var cachedPlans: List<DeliveryPlan> = emptyList()

    init {
        loadPlans()
        loadCurrentPolicy()
    }

    fun loadPlans() {
        viewModelScope.launch {
            _uiState.value = PlanUiState.Loading
            try {
                val response = workerRepository.getPlans()
                if (response.isSuccessful) {
                    val plans = response.body()?.plans ?: emptyList()
                    cachedPlans = plans
                    Log.d(TAG, "Loaded ${plans.size} plans")
                    _uiState.value = PlanUiState.Success(plans)
                } else {
                    _uiState.value = PlanUiState.Error("Failed to load plans")
                }
            } catch (e: Exception) {
                Log.e(TAG, "Error loading plans", e)
                _uiState.value = PlanUiState.Error(e.message ?: "Unknown error")
            }
        }
    }

    fun loadCurrentPolicy() {
        viewModelScope.launch {
            try {
                val response = workerRepository.getPolicy()
                if (response.isSuccessful) {
                    _currentPolicy.value = response.body()?.policy
                }
            } catch (e: Exception) {
                Log.e(TAG, "Error loading current policy", e)
            }
        }
    }

    fun selectPlan(plan: DeliveryPlan) {
        _selectedPlan.value = plan
        _selectedExpectedDeliveries.value = plan.rangeStart
        _isPaymentRequired.value = true
    }

    fun selectExpectedDeliveries(deliveries: Int) {
        val plan = _selectedPlan.value ?: return
        if (deliveries in plan.rangeStart..plan.rangeEnd) {
            _selectedExpectedDeliveries.value = deliveries
        }
    }

    fun calculatePremium(plan: DeliveryPlan, deliveries: Int?): Int {
        val minPremium = plan.weeklyPremiumMinInr ?: plan.weeklyPremiumInr
        val maxPremium = plan.weeklyPremiumMaxInr ?: plan.weeklyPremiumInr
        val selectedDeliveries = deliveries ?: plan.rangeStart
        val span = (plan.rangeEnd - plan.rangeStart).coerceAtLeast(1)
        val progress = (selectedDeliveries - plan.rangeStart).coerceAtLeast(0).coerceAtMost(span)
        return minPremium + ((maxPremium - minPremium) * progress / span)
    }

    fun calculateUpgradeFee(plan: DeliveryPlan): Int {
        val currentPlanId = currentPolicy.value?.planId?.trim().orEmpty()
        if (currentPlanId.isEmpty() || currentPlanId == plan.planId) return 0

        val currentPlan = cachedPlans.firstOrNull { it.planId == currentPlanId } ?: return 0
        return if (plan.maxPayoutInr > currentPlan.maxPayoutInr) 5 else 0
    }

    fun confirmSelection() {
        val plan = _selectedPlan.value ?: return
        val deliveries = _selectedExpectedDeliveries.value ?: plan.rangeStart
        viewModelScope.launch {
            try {
                val premium = calculatePremium(plan, deliveries)
                val upgradeFee = calculateUpgradeFee(plan)
                val totalPayment = premium + upgradeFee
                var response = workerRepository.selectPlan(
                    planId = plan.planId,
                    expectedDeliveries = deliveries,
                    paymentAmountInr = totalPayment,
                )

                // Backend may require a higher first-activation amount.
                // If provided, retry once with required_amount_inr.
                if (!response.isSuccessful) {
                    val raw = response.errorBody()?.string().orEmpty()
                    val parsed = parsePlanSelectionError(raw)
                    if (parsed.errorCode == "insufficient_payment_amount" && parsed.requiredAmountInr != null && parsed.requiredAmountInr > totalPayment) {
                        response = workerRepository.selectPlan(
                            planId = plan.planId,
                            expectedDeliveries = deliveries,
                            paymentAmountInr = parsed.requiredAmountInr,
                        )
                    }
                }

                if (response.isSuccessful) {
                    val selectedFromApi = response.body()?.plan
                    val apiDescription: String? = selectedFromApi?.description
                    val selectedPlan = selectedFromApi?.copy(
                        weeklyPremiumInr = premium,
                        weeklyPremiumMinInr = plan.weeklyPremiumMinInr,
                        weeklyPremiumMaxInr = plan.weeklyPremiumMaxInr,
                        description = if (apiDescription.isNullOrBlank()) plan.description else apiDescription,
                    ) ?: plan.copy(weeklyPremiumInr = premium)

                    _selectedPlan.value = selectedPlan
                    _selectedExpectedDeliveries.value = deliveries
                    _isPaymentRequired.value = false
                    _uiState.value = PlanUiState.SelectionComplete(cachedPlans, selectedPlan)
                    Log.d(TAG, "Plan ${plan.planId} selected with premium Rs.$premium, fee Rs.$upgradeFee and deliveries $deliveries")
                } else {
                    val raw = response.errorBody()?.string().orEmpty()
                    val parsed = parsePlanSelectionError(raw)
                    val msg = parsed.message
                        ?: parsed.errorCode
                        ?: "Failed to confirm plan selection"
                    _uiState.value = PlanUiState.Error(msg)
                }
            } catch (e: Exception) {
                Log.e(TAG, "Error confirming plan selection", e)
                _uiState.value = PlanUiState.Error(e.message ?: "Failed to confirm plan selection")
            }
        }
    }

    private data class PlanSelectionError(
        val errorCode: String? = null,
        val message: String? = null,
        val requiredAmountInr: Int? = null,
    )

    private fun parsePlanSelectionError(raw: String): PlanSelectionError {
        if (raw.isBlank()) return PlanSelectionError()
        return try {
            val json = JSONObject(raw)
            PlanSelectionError(
                errorCode = json.optString("error").ifBlank { null },
                message = json.optString("message").ifBlank { null },
                requiredAmountInr = if (json.has("required_amount_inr")) json.optInt("required_amount_inr") else null,
            )
        } catch (_: Exception) {
            PlanSelectionError(message = raw)
        }
    }

    fun clearSelection() {
        _selectedPlan.value = null
        _selectedExpectedDeliveries.value = null
        _isPaymentRequired.value = false
    }

    fun skipPlan() {
        viewModelScope.launch {
            // NOTE: API call to backend disabled - plan skip stays local only
            // Previously called workerRepository.skipPlan() here
            _uiState.value = PlanUiState.Skipped(cachedPlans)
            Log.d(TAG, "Plan skipped (local state only)")
        }
    }
}

sealed class PlanUiState {
    object Loading : PlanUiState()
    data class Success(val plans: List<DeliveryPlan>) : PlanUiState()
    data class SelectionComplete(val plans: List<DeliveryPlan>, val selectedPlan: DeliveryPlan) : PlanUiState()
    data class Skipped(val plans: List<DeliveryPlan>) : PlanUiState()
    data class Error(val message: String) : PlanUiState()
}
