package com.imaginai.indel.ui.auth

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.repository.AuthRepository
import com.imaginai.indel.data.repository.WorkerRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class OtpViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    private val workerRepository: WorkerRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<OtpUiState>(OtpUiState.Idle)
    val uiState = _uiState.asStateFlow()

    private val _phone = MutableStateFlow("")
    val phone = _phone.asStateFlow()

    private val _otp = MutableStateFlow("")
    val otp = _otp.asStateFlow()

    fun onPhoneChanged(newPhone: String) {
        _phone.value = newPhone
    }

    fun onOtpChanged(newOtp: String) {
        _otp.value = newOtp
    }

    fun sendOtp() {
        if (_phone.value.isBlank()) {
             _uiState.value = OtpUiState.Error("Please enter phone number")
             return
        }
        viewModelScope.launch {
            _uiState.value = OtpUiState.Loading
            try {
                val response = authRepository.sendOtp(_phone.value)
                if (response.isSuccessful) {
                    val body = response.body()
                    _uiState.value = OtpUiState.OtpSent(body?.otpForTesting)
                } else {
                    _uiState.value = OtpUiState.Error("Failed to send OTP")
                }
            } catch (e: Exception) {
                _uiState.value = OtpUiState.Error(e.message ?: "Unknown error")
            }
        }
    }

    fun verifyOtp() {
        if (_otp.value.isBlank()) {
            _uiState.value = OtpUiState.Error("Please enter OTP")
            return
        }
        viewModelScope.launch {
            _uiState.value = OtpUiState.Loading
            try {
                val response = authRepository.verifyOtp(_phone.value, _otp.value)
                if (response.isSuccessful) {
                    checkWorkerProfile()
                } else {
                    _uiState.value = OtpUiState.Error("Invalid OTP")
                }
            } catch (e: Exception) {
                _uiState.value = OtpUiState.Error(e.message ?: "Unknown error")
            }
        }
    }

    private suspend fun checkWorkerProfile() {
        try {
            val response = workerRepository.getProfile()
            // Updated to handle WorkerProfileResponse which wraps the worker object
            if (response.isSuccessful && response.body()?.worker?.name?.isNotEmpty() == true) {
                _uiState.value = OtpUiState.Success(hasProfile = true)
            } else {
                _uiState.value = OtpUiState.Success(hasProfile = false)
            }
        } catch (e: Exception) {
            // If profile check fails, assume onboarding is needed
            _uiState.value = OtpUiState.Success(hasProfile = false)
        }
    }
}

sealed class OtpUiState {
    object Idle : OtpUiState()
    object Loading : OtpUiState()
    data class OtpSent(val testOtp: String?) : OtpUiState()
    data class Success(val hasProfile: Boolean) : OtpUiState()
    data class Error(val message: String) : OtpUiState()
}
