package com.imaginai.indel.ui.auth

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.repository.WorkerRepository
import com.imaginai.indel.ui.shared.isValidUpiId
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


    private val _vehicleType = MutableStateFlow("")
    val vehicleType = _vehicleType.asStateFlow()

    private val _vehicleName = MutableStateFlow("")
    val vehicleName = _vehicleName.asStateFlow()



    fun onVehicleTypeChanged(value: String) { _vehicleType.value = value }
    fun onVehicleNameChanged(value: String) { _vehicleName.value = value }

    fun submitOnboarding() {
        val vehicleName = _vehicleName.value.trim()
        if (_vehicleType.value.isBlank() || vehicleName.isBlank()) {
            _uiState.value = OnboardingUiState.Error("Please fill all fields")
            return
        }
        viewModelScope.launch {
            _uiState.value = OnboardingUiState.Loading
            try {
                val response = workerRepository.onboard(
                    vehicleType = _vehicleType.value,
                    vehicleName = vehicleName
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
