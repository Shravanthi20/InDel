package com.imaginai.indel.ui.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.EarningsSummary
import com.imaginai.indel.data.model.Policy
import com.imaginai.indel.data.model.WorkerProfile
import com.imaginai.indel.data.repository.EarningsRepository
import com.imaginai.indel.data.repository.PolicyRepository
import com.imaginai.indel.data.repository.WorkerRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class HomeViewModel @Inject constructor(
    private val workerRepository: WorkerRepository,
    private val policyRepository: PolicyRepository,
    private val earningsRepository: EarningsRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<HomeUiState>(HomeUiState.Loading)
    val uiState = _uiState.asStateFlow()

    init {
        loadDashboard()
    }

    fun loadDashboard() {
        viewModelScope.launch {
            _uiState.value = HomeUiState.Loading
            try {
                val profileRes = workerRepository.getProfile()
                val policyRes = policyRepository.getPolicy()
                val earningsRes = earningsRepository.getEarnings()

                if (profileRes.isSuccessful && policyRes.isSuccessful && earningsRes.isSuccessful) {
                    _uiState.value = HomeUiState.Success(
                        worker = profileRes.body()?.worker!!,
                        policy = policyRes.body()?.policy!!,
                        earnings = earningsRes.body()!!
                    )
                } else {
                    _uiState.value = HomeUiState.Error("Failed to load dashboard data")
                }
            } catch (e: Exception) {
                _uiState.value = HomeUiState.Error(e.message ?: "Unknown error")
            }
        }
    }
}

sealed class HomeUiState {
    object Loading : HomeUiState()
    data class Success(
        val worker: WorkerProfile,
        val policy: Policy,
        val earnings: EarningsSummary
    ) : HomeUiState()
    data class Error(val message: String) : HomeUiState()
}
