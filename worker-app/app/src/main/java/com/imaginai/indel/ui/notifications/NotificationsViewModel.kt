package com.imaginai.indel.ui.notifications

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.Notification
import com.imaginai.indel.data.repository.WorkerRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class NotificationsViewModel @Inject constructor(
    private val workerRepository: WorkerRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow<NotificationsUiState>(NotificationsUiState.Loading)
    val uiState = _uiState.asStateFlow()

    init {
        loadNotifications()
    }

    fun loadNotifications() {
        viewModelScope.launch {
            val currentPreferences = (_uiState.value as? NotificationsUiState.Success)?.preferences ?: NotificationPreferencesState()
            _uiState.value = NotificationsUiState.Loading
            try {
                val response = workerRepository.getNotifications()
                _uiState.value = NotificationsUiState.Success(
                    notifications = response.body()?.notifications ?: emptyList(),
                    preferences = currentPreferences,
                )
            } catch (e: Exception) {
                _uiState.value = NotificationsUiState.Error(e.message ?: "Unknown error")
            }
        }
    }

    fun refreshNotifications() {
        loadNotifications()
    }

    fun togglePreference(key: String, enabled: Boolean) {
        val current = _uiState.value
        if (current !is NotificationsUiState.Success) return
        val preferences = when (key) {
            "delivery_updates" -> current.preferences.copy(deliveryUpdates = enabled)
            "payout_updates" -> current.preferences.copy(payoutUpdates = enabled)
            "policy_reminders" -> current.preferences.copy(policyReminders = enabled)
            "disruption_alerts" -> current.preferences.copy(disruptionAlerts = enabled)
            else -> current.preferences
        }
        _uiState.value = current.copy(preferences = preferences, feedback = null)
    }

    fun savePreferences() {
        val current = _uiState.value
        if (current !is NotificationsUiState.Success) return
        viewModelScope.launch {
            _uiState.value = current.copy(isSaving = true, feedback = null)
            try {
                workerRepository.updateNotificationPreferences(
                    mapOf(
                        "delivery_updates" to current.preferences.deliveryUpdates,
                        "payout_updates" to current.preferences.payoutUpdates,
                        "policy_reminders" to current.preferences.policyReminders,
                        "disruption_alerts" to current.preferences.disruptionAlerts,
                    )
                )
                _uiState.value = current.copy(isSaving = false, feedback = "Notification preferences saved.")
            } catch (e: Exception) {
                _uiState.value = current.copy(isSaving = false, feedback = e.message ?: "Failed to save preferences.")
            }
        }
    }
}

data class NotificationPreferencesState(
    val deliveryUpdates: Boolean = true,
    val payoutUpdates: Boolean = true,
    val policyReminders: Boolean = true,
    val disruptionAlerts: Boolean = true,
)

sealed class NotificationsUiState {
    object Loading : NotificationsUiState()
    data class Success(
        val notifications: List<Notification>,
        val preferences: NotificationPreferencesState,
        val isSaving: Boolean = false,
        val feedback: String? = null,
    ) : NotificationsUiState()
    data class Error(val message: String) : NotificationsUiState()
}