package com.imaginai.indel.ui.notifications

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.Notification
import com.imaginai.indel.data.repository.WorkerRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class NotificationsViewModel @Inject constructor(
    private val workerRepository: WorkerRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<NotificationsUiState>(NotificationsUiState.Loading)
    val uiState = _uiState.asStateFlow()

    private val _isRefreshing = MutableStateFlow(false)
    val isRefreshing = _isRefreshing.asStateFlow()

    init {
        loadNotifications()
        startAutoRefresh()
    }

    fun loadNotifications() {
        viewModelScope.launch {
            _uiState.value = NotificationsUiState.Loading
            fetchNotifications()
        }
    }

    fun refresh() {
        viewModelScope.launch {
            _isRefreshing.value = true
            fetchNotifications()
            delay(400)
            _isRefreshing.value = false
        }
    }

    private suspend fun fetchNotifications() {
        try {
            val response = workerRepository.getNotifications()
            if (response.isSuccessful) {
                _uiState.value = NotificationsUiState.Success(response.body()?.notifications ?: emptyList())
            } else {
                _uiState.value = NotificationsUiState.Error("Failed to load notifications")
            }
        } catch (e: Exception) {
            _uiState.value = NotificationsUiState.Error(e.message ?: "Unknown error")
        }
    }

    private fun startAutoRefresh() {
        viewModelScope.launch {
            while (true) {
                delay(10000)
                fetchNotifications()
            }
        }
    }
}

sealed class NotificationsUiState {
    object Loading : NotificationsUiState()
    data class Success(val notifications: List<Notification>) : NotificationsUiState()
    data class Error(val message: String) : NotificationsUiState()
}
