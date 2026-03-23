package com.imaginai.indel.ui.policy

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.Policy
import com.imaginai.indel.data.model.PremiumResponse
import com.imaginai.indel.data.repository.PolicyRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class PolicyViewModel @Inject constructor(
    private val policyRepository: PolicyRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<PolicyUiState>(PolicyUiState.Loading)
    val uiState = _uiState.asStateFlow()

    init {
        loadPolicy()
    }

    fun loadPolicy() {
        viewModelScope.launch {
            _uiState.value = PolicyUiState.Loading
            try {
                val policyRes = policyRepository.getPolicy()
                val premiumRes = policyRepository.getPremium()

                if (policyRes.isSuccessful && premiumRes.isSuccessful) {
                    _uiState.value = PolicyUiState.Success(
                        policy = policyRes.body()!!.policy,
                        premium = premiumRes.body()!!
                    )
                } else {
                    _uiState.value = PolicyUiState.Error("Failed to load policy")
                }
            } catch (e: Exception) {
                _uiState.value = PolicyUiState.Error(e.message ?: "Unknown error")
            }
        }
    }

    fun enroll() {
        viewModelScope.launch {
            policyRepository.enrollPolicy()
            loadPolicy()
        }
    }

    fun pause() {
        viewModelScope.launch {
            policyRepository.pausePolicy()
            loadPolicy()
        }
    }

    fun cancel() {
        viewModelScope.launch {
            policyRepository.cancelPolicy()
            loadPolicy()
        }
    }
}

sealed class PolicyUiState {
    object Loading : PolicyUiState()
    data class Success(val policy: Policy, val premium: PremiumResponse) : PolicyUiState()
    data class Error(val message: String) : PolicyUiState()
}
