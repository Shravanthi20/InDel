package com.imaginai.indel.ui.delivery

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.repository.WorkerRepository
import com.imaginai.indel.data.model.VerifyCodeRequest
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class FetchVerificationViewModel @Inject constructor(
    private val workerRepository: WorkerRepository
) : ViewModel() {

    private val _code = MutableStateFlow("")
    val code = _code.asStateFlow()

    private val _uiState = MutableStateFlow<VerificationUiState>(VerificationUiState.Idle)
    val uiState = _uiState.asStateFlow()

    fun onCodeChanged(value: String) { _code.value = value }

    fun sendCode() {
        viewModelScope.launch {
            _uiState.value = VerificationUiState.Loading
            try {
                val response = workerRepository.sendVerificationCode()
                if (response.isSuccessful) {
                    _uiState.value = VerificationUiState.CodeSent
                } else {
                    _uiState.value = VerificationUiState.Error("Failed to send code")
                }
            } catch (e: Exception) {
                _uiState.value = VerificationUiState.Error(e.message ?: "Unknown error")
            }
        }
    }

    fun verifyCode() {
        viewModelScope.launch {
            _uiState.value = VerificationUiState.Loading
            try {
                val response = workerRepository.verifyCode(VerifyCodeRequest(_code.value))
                if (response.isSuccessful) {
                    _uiState.value = VerificationUiState.Success
                } else {
                    _uiState.value = VerificationUiState.Error("Invalid code")
                }
            } catch (e: Exception) {
                _uiState.value = VerificationUiState.Error(e.message ?: "Unknown error")
            }
        }
    }
}

sealed class VerificationUiState {
    object Idle : VerificationUiState()
    object Loading : VerificationUiState()
    object CodeSent : VerificationUiState()
    object Success : VerificationUiState()
    data class Error(val message: String) : VerificationUiState()
}
