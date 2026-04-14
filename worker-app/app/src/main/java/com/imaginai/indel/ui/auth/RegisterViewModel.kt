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
                Log.d(TAG, "Fetching platform zones for registration level=$level")
                val response = workerRepository.getZonePaths(level.lowercase())
                if (response.isSuccessful) {
                    val body = response.body()
                    val paths = mutableListOf<ZonePath>()

                    body?.zones?.forEach { zone ->
                        val zoneLabel = buildString {
                            append(zone.zoneName)
                            zone.zoneState?.takeIf { it.isNotBlank() }?.let {
                                append(" ($it)")
                            }
                        }
                        paths.add(
                            ZonePath(
                                id = zone.zoneId?.toString(),
                                displayName = zoneLabel,
                                city = zone.city ?: zone.zoneName
                            )
                        )
                    }

                    body?.cities?.forEach { city ->
                        val cityLabel = buildString {
                            append(city.city)
                            city.state?.takeIf { it.isNotBlank() }?.let {
                                append(" ($it)")
                            }
                        }
                        paths.add(
                            ZonePath(
                                displayName = cityLabel,
                                city = city.city
                            )
                        )
                    }

                    body?.cityPairs?.forEach { pair ->
                        val pairState = pair.state
                            ?: listOfNotNull(pair.fromState, pair.toState).distinct().joinToString(" / ")
                        val pairLabel = buildString {
                            append(pair.from)
                            append(" to ")
                            append(pair.to)
                            pairState.takeIf { it.isNotBlank() }?.let {
                                append(" ($it)")
                            }
                        }
                        paths.add(
                            ZonePath(
                                displayName = pairLabel,
                                fromCity = pair.from,
                                toCity = pair.to
                            )
                        )
                    }

                    if (paths.isEmpty() && body?.paths != null) {
                        paths.addAll(body.paths)
                    }

                    _availablePaths.value = paths.distinctBy { it.displayName.orEmpty() }
                    _uiState.value = if (_availablePaths.value.isEmpty()) {
                        RegisterUiState.Error("No zones returned for selected level")
                    } else {
                        RegisterUiState.Idle
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
