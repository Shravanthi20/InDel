package com.imaginai.indel.ui.policy

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.Policy
import com.imaginai.indel.data.repository.PolicyRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.delay
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

    private val _isRefreshing = MutableStateFlow(false)
    val isRefreshing = _isRefreshing.asStateFlow()

    init {
        loadPolicy()
        startAutoRefresh()
    }

    fun loadPolicy() {
        viewModelScope.launch {
            _uiState.value = PolicyUiState.Loading
            fetchPolicy()
        }
    }

    fun refresh() {
        viewModelScope.launch {
            _isRefreshing.value = true
            fetchPolicy()
            delay(500)
            _isRefreshing.value = false
        }
    }

    private suspend fun fetchPolicy() {
        try {
            val policyRes = policyRepository.getPolicy()
            if (policyRes.isSuccessful) {
                val policy = policyRes.body()?.policy
                _uiState.value = PolicyUiState.Success(
                    policy = policy ?: Policy()
                )
            } else {
                _uiState.value = PolicyUiState.Error("Failed to load policy")
            }
        } catch (e: Exception) {
            _uiState.value = PolicyUiState.Error(e.message ?: "Unknown error")
        }
    }

    fun enroll() {
        viewModelScope.launch {
            _uiState.value = PolicyUiState.Loading
            policyRepository.enrollPolicy()
            fetchPolicy()
        }
    }

    fun pause() {
        viewModelScope.launch {
            _uiState.value = PolicyUiState.Loading
            policyRepository.pausePolicy()
            fetchPolicy()
        }
    }

    fun cancel() {
        viewModelScope.launch {
            _uiState.value = PolicyUiState.Loading
            policyRepository.cancelPolicy()
            fetchPolicy()
        }
    }

    private fun startAutoRefresh() {
        viewModelScope.launch {
            while (true) {
                delay(15000)
                fetchPolicy()
            }
        }
    }
}

sealed class PolicyUiState {
    object Loading : PolicyUiState()
    data class Success(val policy: Policy) : PolicyUiState()
    data class Error(val message: String) : PolicyUiState()
}
