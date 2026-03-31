package com.imaginai.indel.ui.auth

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.repository.WorkerRepository
import com.imaginai.indel.ui.shared.isVehicleAllowedForZoneLevel
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

    private val _name = MutableStateFlow("")
    val name = _name.asStateFlow()

    private val _zoneLevel = MutableStateFlow("")
    val zoneLevel = _zoneLevel.asStateFlow()

    private val _zoneName = MutableStateFlow("")
    val zoneName = _zoneName.asStateFlow()

    private val _vehicleType = MutableStateFlow("")
    val vehicleType = _vehicleType.asStateFlow()

    private val _upiId = MutableStateFlow("")
    val upiId = _upiId.asStateFlow()

    fun onNameChanged(value: String) { _name.value = value }
    
    fun onZoneLevelChanged(value: String) {
        _zoneLevel.value = value
        _zoneName.value = "" // Reset zone name when level changes
        if (!isVehicleAllowedForZoneLevel(_zoneLevel.value, _vehicleType.value)) {
            _vehicleType.value = ""
        }
    }

    fun onZoneNameChanged(value: String) {
        _zoneName.value = value
    }

    fun onVehicleTypeChanged(value: String) { _vehicleType.value = value }
    fun onUpiIdChanged(value: String) { _upiId.value = value }

    fun submitOnboarding() {
        val upi = _upiId.value.trim()
        if (_name.value.isBlank() || _zoneLevel.value.isBlank() || _zoneName.value.isBlank() || 
            _vehicleType.value.isBlank() || upi.isBlank()) {
            _uiState.value = OnboardingUiState.Error("Please fill all fields")
            return
        }

        if (!isValidUpiId(upi)) {
            _uiState.value = OnboardingUiState.Error("Invalid UPI ID format (username@bankid)")
            return
        }

        if (!isVehicleAllowedForZoneLevel(_zoneLevel.value, _vehicleType.value)) {
            _uiState.value = OnboardingUiState.Error("Selected vehicle is not allowed for this zone level")
            return
        }

        viewModelScope.launch {
            _uiState.value = OnboardingUiState.Loading
            try {
                val response = workerRepository.onboard(
                    name = _name.value,
                    zoneLevel = _zoneLevel.value,
                    zoneName = _zoneName.value,
                    vehicleType = _vehicleType.value,
                    upiId = upi
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
