package com.imaginai.indel.ui.auth

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import android.util.Log
import com.imaginai.indel.data.model.ZonePath
import com.imaginai.indel.data.repository.AuthRepository
import com.imaginai.indel.data.repository.WorkerRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class RegisterViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    private val workerRepository: WorkerRepository
) : ViewModel() {
    companion object {
        private const val TAG = "RegisterViewModel"
    }

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

    private val _zoneLevel = MutableStateFlow("")
    val zoneLevel = _zoneLevel.asStateFlow()

    private val _zoneName = MutableStateFlow("")
    val zoneName = _zoneName.asStateFlow()

    private val _availablePaths = MutableStateFlow<List<ZonePath>>(emptyList())
    val availablePaths = _availablePaths.asStateFlow()

    private val _uiState = MutableStateFlow<RegisterUiState>(RegisterUiState.Idle)
    val uiState = _uiState.asStateFlow()

    fun onUsernameChanged(value: String) { _username.value = value }
    fun onEmailChanged(value: String) { _email.value = value }
    fun onPhoneChanged(value: String) { _phone.value = value }
    fun onPasswordChanged(value: String) { _password.value = value }
    fun onConfirmPasswordChanged(value: String) { _confirmPassword.value = value }

    fun onZoneLevelChanged(value: String) {
        _zoneLevel.value = value
        _zoneName.value = ""
        _availablePaths.value = emptyList()
        if (value.isNotBlank()) {
            fetchZonePaths(value)
        }
    }

    private fun fetchZonePaths(level: String) {
        viewModelScope.launch {
            try {
                Log.d(TAG, "Fetching zone paths for level=$level")
                val response = workerRepository.getZonePaths(level.lowercase())
                Log.d(TAG, "Zone path response: ${response.code()}")
                
                if (response.isSuccessful) {
                    val body = response.body()
                    val paths = mutableListOf<ZonePath>()
                    
                    body?.cities?.let { cities ->
                        Log.d(TAG, "Processing ${cities.size} cities")
                        cities.forEach { city ->
                            val displayName = city.city + (city.state?.let { " ($it)" } ?: "")
                            paths.add(ZonePath(displayName = displayName, city = city.city))
                        }
                    }
                    
                    body?.cityPairs?.let { pairs ->
                        Log.d(TAG, "Processing ${pairs.size} city pairs")
                        pairs.forEach { pair ->
                            paths.add(ZonePath(
                                displayName = "${pair.from} to ${pair.to}" + (pair.state?.let { " (${it})" } ?: ""),
                                fromCity = pair.from,
                                toCity = pair.to
                            ))
                        }
                    }

                    if (paths.size > 10) {
                        paths.subList(10, paths.size).clear()
                    }
                    
                    // Fallback to existing paths if any
                    if (paths.isEmpty() && body?.paths != null) {
                        Log.d(TAG, "Using fallback paths")
                        paths.addAll(body.paths)
                        if (paths.size > 10) {
                            paths.subList(10, paths.size).clear()
                        }
                    }
                    
                    Log.d(TAG, "Final paths count: ${paths.size}")
                    _availablePaths.value = paths
                    if (paths.isEmpty()) {
                        _uiState.value = RegisterUiState.Error("No zone paths returned for selected level")
                    } else {
                        _uiState.value = RegisterUiState.Idle
                    }
                } else {
                    val errorText = response.errorBody()?.string() ?: "Zone path fetch failed"
                    Log.e(TAG, "fetchZonePaths failed status=${response.code()} body=$errorText")
                    _uiState.value = RegisterUiState.Error("Unable to load zones (${response.code()})")
                }
            } catch (e: Exception) {
                Log.e(TAG, "fetchZonePaths exception", e)
                e.printStackTrace()
                _uiState.value = RegisterUiState.Error(e.message ?: "Unable to load zones")
            }
        }
    }

    fun onZonePathSelected(path: ZonePath) {
        _zoneName.value = path.displayName ?: ""
    }

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
        val zoneLevelVal = _zoneLevel.value
        val zoneNameVal = _zoneName.value

        if (usernameVal.isBlank() || emailVal.isBlank() || phoneVal.isBlank() || 
            _password.value.isBlank() || zoneLevelVal.isBlank() || zoneNameVal.isBlank()) {
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
                    _password.value,
                    zoneLevelVal,
                    zoneNameVal
                )
                if (response.isSuccessful) {
                    _uiState.value = RegisterUiState.Success
                } else {
                    val errorMsg = response.errorBody()?.string() ?: "Registration failed"
                    _uiState.value = RegisterUiState.Error(errorMsg)
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
