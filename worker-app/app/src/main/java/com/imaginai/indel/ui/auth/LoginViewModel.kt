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
class LoginViewModel @Inject constructor(
    private val authRepository: AuthRepository
) : ViewModel() {

    private val _identifier = MutableStateFlow("")
    val identifier = _identifier.asStateFlow()

    private val _password = MutableStateFlow("")
    val password = _password.asStateFlow()

    private val _uiState = MutableStateFlow<LoginUiState>(LoginUiState.Idle)
    val uiState = _uiState.asStateFlow()

    fun onIdentifierChanged(value: String) { _identifier.value = value }
    fun onPasswordChanged(value: String) { _password.value = value }

    private fun isValidIdentifier(id: String): Boolean {
        // Check if it's an email: contains '@' and then a '.' somewhere after '@'
        val isEmail = id.contains("@") && id.substringAfter("@").contains(".")
        // Check if it's a 10-digit number
        val isPhone = id.length == 10 && id.all { it.isDigit() }
        
        return isEmail || isPhone
    }

    fun login() {
        val id = _identifier.value.trim()
        if (id.isBlank() || _password.value.isBlank()) {
            _uiState.value = LoginUiState.Error("Please enter all fields")
            return
        }

        if (!isValidIdentifier(id)) {
            _uiState.value = LoginUiState.Error("Enter a valid email or 10-digit phone number")
            return
        }

        viewModelScope.launch {
            _uiState.value = LoginUiState.Loading
            try {
                val response = authRepository.login(id, _password.value)
                if (response.isSuccessful) {
                    _uiState.value = LoginUiState.Success
                } else {
                    _uiState.value = LoginUiState.Error("Invalid credentials")
                }
            } catch (e: Exception) {
                _uiState.value = LoginUiState.Error(e.message ?: "Unknown error")
            }
        }
    }
}

sealed class LoginUiState {
    object Idle : LoginUiState()
    object Loading : LoginUiState()
    object Success : LoginUiState()
    data class Error(val message: String) : LoginUiState()
}
