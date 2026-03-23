package com.imaginai.indel.ui.auth

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.repository.WorkerRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class OnboardingViewModel @Inject constructor(
    private val workerRepository: WorkerRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<OnboardingUiState>(OnboardingUiState.Idle)
    val uiState = _uiState.asStateFlow()

    private val _name = MutableStateFlow("")
    val name = _name.asStateFlow()

    private val _zone = MutableStateFlow("")
    val zone = _zone.asStateFlow()

    private val _vehicleType = MutableStateFlow("")
    val vehicleType = _vehicleType.asStateFlow()

    private val _upiId = MutableStateFlow("")
    val upiId = _upiId.asStateFlow()

    fun onNameChanged(value: String) { _name.value = value }
    fun onZoneChanged(value: String) { _zone.value = value }
    fun onVehicleTypeChanged(value: String) { _vehicleType.value = value }
    fun onUpiIdChanged(value: String) { _upiId.value = value }

    fun submitOnboarding() {
        viewModelScope.launch {
            _uiState.value = OnboardingUiState.Loading
            try {
                val response = workerRepository.onboard(
                    name = _name.value,
                    zone = _zone.value,
                    vehicleType = _vehicleType.value,
                    upiId = _upiId.value
                )
                if (response.isSuccessful) {
                    _uiState.value = OnboardingUiState.Success
                } else {
                    _uiState.value = OnboardingUiState.Error("Failed to submit onboarding")
                }
            } catch (e: Exception) {
                _uiState.value = OnboardingUiState.Error(e.message ?: "Unknown error")
            }
        }
    }
}

sealed class OnboardingUiState {
    object Idle : OnboardingUiState()
    object Loading : OnboardingUiState()
    object Success : OnboardingUiState()
    data class Error(val message: String) : OnboardingUiState()
}
