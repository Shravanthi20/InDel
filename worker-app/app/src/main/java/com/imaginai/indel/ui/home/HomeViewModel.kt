package com.imaginai.indel.ui.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.*
import com.imaginai.indel.data.repository.EarningsRepository
import com.imaginai.indel.data.repository.PolicyRepository
import com.imaginai.indel.data.repository.WorkerRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class HomeViewModel @Inject constructor(
    private val workerRepository: WorkerRepository,
    private val policyRepository: PolicyRepository,
    private val earningsRepository: EarningsRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<HomeUiState>(HomeUiState.Loading)
    val uiState = _uiState.asStateFlow()

    private val _isRefreshing = MutableStateFlow(false)
    val isRefreshing = _isRefreshing.asStateFlow()

    private val _isOnline = MutableStateFlow(true)
    val isOnline = _isOnline.asStateFlow()

    init {
        loadDashboard()
        startAutoRefresh()
    }

    fun loadDashboard() {
        viewModelScope.launch {
            _uiState.value = HomeUiState.Loading
            fetchData()
        }
    }

    fun refresh() {
        viewModelScope.launch {
            _isRefreshing.value = true
            fetchData()
            delay(500) // Small delay for better UX
            _isRefreshing.value = false
        }
    }

    private suspend fun fetchData() {
        try {
            val profileRes = workerRepository.getProfile()
            val policyRes = policyRepository.getPolicy()
            val earningsRes = earningsRepository.getEarnings()

            if (profileRes.isSuccessful && policyRes.isSuccessful && earningsRes.isSuccessful) {
                val summary = earningsRes.body()!!
                val earnings = Earnings(
                    thisWeekActual = summary.thisWeekActual.toDouble(),
                    thisWeekBaseline = summary.thisWeekBaseline.toDouble(),
                    todayEarnings = (summary.todayEarnings ?: 0).toDouble(),
                    protectedIncome = summary.protectedIncome.toDouble(),
                    history = summary.history.map { EarningRecord(it.week, it.actual.toDouble()) }
                )
                
                _uiState.value = HomeUiState.Success(
                    worker = profileRes.body()!!.worker,
                    policy = policyRes.body()!!.policy,
                    earnings = earnings
                )
            } else {
                _uiState.value = HomeUiState.Error("Failed to load dashboard data")
            }
        } catch (e: Exception) {
            _uiState.value = HomeUiState.Error(e.message ?: "Unknown error")
        }
    }

    fun toggleOnlineStatus(online: Boolean) {
        _isOnline.value = online
        // In a real app, you'd call an API here: workerRepository.updateStatus(online)
    }

    private fun startAutoRefresh() {
        viewModelScope.launch {
            while (true) {
                delay(12000)
                fetchData()
            }
        }
    }
}

sealed class HomeUiState {
    object Loading : HomeUiState()
    data class Success(
        val worker: WorkerProfile,
        val policy: Policy,
        val earnings: Earnings
    ) : HomeUiState()
    data class Error(val message: String) : HomeUiState()
}
