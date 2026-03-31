package com.imaginai.indel.ui.profile

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
class ProfileEditViewModel @Inject constructor(
    private val workerRepository: WorkerRepository
) : ViewModel() {

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

    private val _uiState = MutableStateFlow<ProfileEditUiState>(ProfileEditUiState.Loading)
    val uiState = _uiState.asStateFlow()

    init {
        loadProfile()
    }

    fun onNameChanged(value: String) {
        _name.value = value
    }

    fun onZoneLevelChanged(value: String) {
        _zoneLevel.value = value
        _zoneName.value = ""
        if (!isVehicleAllowedForZoneLevel(_zoneLevel.value, _vehicleType.value)) {
            _vehicleType.value = ""
        }
    }

    fun onZoneNameChanged(value: String) {
        _zoneName.value = value
    }

    fun onVehicleTypeChanged(value: String) {
        _vehicleType.value = value
    }

    fun onUpiIdChanged(value: String) {
        _upiId.value = value
    }

    fun loadProfile() {
        viewModelScope.launch {
            _uiState.value = ProfileEditUiState.Loading
            try {
                val response = workerRepository.getProfile()
                if (response.isSuccessful) {
                    val worker = response.body()?.worker
                    if (worker != null) {
                        _name.value = worker.name
                        _zoneLevel.value = worker.zoneLevel
                        _zoneName.value = worker.zoneName
                        _vehicleType.value = worker.vehicleType
                        _upiId.value = worker.upiId
                        _uiState.value = ProfileEditUiState.Idle
                    } else {
                        _uiState.value = ProfileEditUiState.Error("Profile not found")
                    }
                } else {
                    _uiState.value = ProfileEditUiState.Error("Failed to load profile")
                }
            } catch (e: Exception) {
                _uiState.value = ProfileEditUiState.Error(e.message ?: "Unknown error")
            }
        }
    }

    fun saveProfile() {
        val upi = _upiId.value.trim()
        if (_name.value.isBlank() || _zoneLevel.value.isBlank() || _zoneName.value.isBlank() || 
            _vehicleType.value.isBlank() || upi.isBlank()) {
            _uiState.value = ProfileEditUiState.Error("All fields are required")
            return
        }

        if (!isValidUpiId(upi)) {
            _uiState.value = ProfileEditUiState.Error("Invalid UPI ID format (username@bankid)")
            return
        }

        if (!isVehicleAllowedForZoneLevel(_zoneLevel.value, _vehicleType.value)) {
            _uiState.value = ProfileEditUiState.Error("Selected vehicle is not allowed for this zone level")
            return
        }

        viewModelScope.launch {
            _uiState.value = ProfileEditUiState.Saving
            try {
                val response = workerRepository.updateProfile(
                    name = _name.value,
                    zoneLevel = _zoneLevel.value,
                    zoneName = _zoneName.value,
                    vehicleType = _vehicleType.value,
                    upiId = upi
                )
                if (response.isSuccessful) {
                    _uiState.value = ProfileEditUiState.Saved
                } else {
                    _uiState.value = ProfileEditUiState.Error("Failed to save profile")
                }
            } catch (e: Exception) {
                _uiState.value = ProfileEditUiState.Error(e.message ?: "Unknown error")
            }
        }
    }
}

sealed class ProfileEditUiState {
    object Loading : ProfileEditUiState()
    object Idle : ProfileEditUiState()
    object Saving : ProfileEditUiState()
    object Saved : ProfileEditUiState()
    data class Error(val message: String) : ProfileEditUiState()
}
