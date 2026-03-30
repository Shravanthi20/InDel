package com.imaginai.indel.ui.auth

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.repository.AuthRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class RegisterViewModel @Inject constructor(
    private val authRepository: AuthRepository
) : ViewModel() {

    private val _username = MutableStateFlow("")
    val username = _username.asStateFlow()

    private val _email = MutableStateFlow("")
    val email = _email.asStateFlow()

    private val _phone = MutableStateFlow("")
    val phone = _phone.asStateFlow()

    private val _password = MutableStateFlow("")
    val password = _password.asStateFlow()

    private val _confirmPassword = MutableStateFlow("")
    val confirmPassword = _confirmPassword.asStateFlow()

    private val _uiState = MutableStateFlow<RegisterUiState>(RegisterUiState.Idle)
    val uiState = _uiState.asStateFlow()

    fun onUsernameChanged(value: String) { _username.value = value }
    fun onEmailChanged(value: String) { _email.value = value }
    fun onPhoneChanged(value: String) { _phone.value = value }
    fun onPasswordChanged(value: String) { _password.value = value }
    fun onConfirmPasswordChanged(value: String) { _confirmPassword.value = value }

    private fun isValidEmail(email: String): Boolean {
        return email.contains("@") && email.substringAfter("@").contains(".")
    }

    private fun isValidPhone(phone: String): Boolean {
        return phone.length == 10 && phone.all { it.isDigit() }
    }

    fun register() {
        val emailVal = _email.value.trim()
        val phoneVal = _phone.value.trim()
        val usernameVal = _username.value.trim()

        if (usernameVal.isBlank() || emailVal.isBlank() || phoneVal.isBlank() || _password.value.isBlank()) {
            _uiState.value = RegisterUiState.Error("Please fill all fields")
            return
        }

        if (!isValidEmail(emailVal)) {
            _uiState.value = RegisterUiState.Error("Invalid email format (must contain @ and .)")
            return
        }

        if (!isValidPhone(phoneVal)) {
            _uiState.value = RegisterUiState.Error("Phone number must be exactly 10 digits")
            return
        }

        if (_password.value != _confirmPassword.value) {
            _uiState.value = RegisterUiState.Error("Passwords do not match")
            return
        }
        
        viewModelScope.launch {
            _uiState.value = RegisterUiState.Loading
            try {
                val response = authRepository.register(
                    usernameVal,
                    phoneVal,
                    emailVal,
                    _password.value
                )
                if (response.isSuccessful) {
                    _uiState.value = RegisterUiState.Success
                } else {
                    _uiState.value = RegisterUiState.Error("Registration failed")
                }
            } catch (e: Exception) {
                _uiState.value = RegisterUiState.Error(e.message ?: "Unknown error")
            }
        }
    }
}

sealed class RegisterUiState {
    object Idle : RegisterUiState()
    object Loading : RegisterUiState()
    object Success : RegisterUiState()
    data class Error(val message: String) : RegisterUiState()
}
