package com.imaginai.indel.ui.delivery

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.*
import com.imaginai.indel.data.repository.EarningsRepository
import com.imaginai.indel.data.repository.PolicyRepository
import com.imaginai.indel.data.repository.WorkerRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class LandingViewModel @Inject constructor(
    private val workerRepository: WorkerRepository,
    private val earningsRepository: EarningsRepository,
    private val policyRepository: PolicyRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<LandingUiState>(LandingUiState.Loading)
    val uiState = _uiState.asStateFlow()

    init {
        loadLandingData()
    }

    fun loadLandingData() {
        viewModelScope.launch {
            _uiState.value = LandingUiState.Loading
            try {
                val profileRes = workerRepository.getProfile()
                val earningsRes = earningsRepository.getEarnings()
                
                if (profileRes.isSuccessful && earningsRes.isSuccessful) {
                    _uiState.value = LandingUiState.Success(
                        worker = profileRes.body()!!.worker,
                        earningsToday = earningsRes.body()!!.thisWeekActual.toDouble() // Using this for demo
                    )
                } else {
                    _uiState.value = LandingUiState.Error("Failed to load landing data")
                }
            } catch (e: Exception) {
                _uiState.value = LandingUiState.Error(e.message ?: "Unknown error")
            }
        }
    }
}

sealed class LandingUiState {
    object Loading : LandingUiState()
    data class Success(val worker: WorkerProfile, val earningsToday: Double) : LandingUiState()
    data class Error(val message: String) : LandingUiState()
}
