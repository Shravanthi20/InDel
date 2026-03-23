package com.imaginai.indel.ui.earnings

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.*
import com.imaginai.indel.data.repository.EarningsRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class EarningsViewModel @Inject constructor(
    private val earningsRepository: EarningsRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<EarningsUiState>(EarningsUiState.Loading)
    val uiState = _uiState.asStateFlow()

    private val _isRefreshing = MutableStateFlow(false)
    val isRefreshing = _isRefreshing.asStateFlow()

    init {
        loadEarnings()
    }

    fun loadEarnings() {
        viewModelScope.launch {
            _uiState.value = EarningsUiState.Loading
            fetchEarnings()
        }
    }

    fun refresh() {
        viewModelScope.launch {
            _isRefreshing.value = true
            fetchEarnings()
            delay(500)
            _isRefreshing.value = false
        }
    }

    private suspend fun fetchEarnings() {
        try {
            val response = earningsRepository.getEarnings()
            if (response.isSuccessful) {
                val summary = response.body()!!
                val earnings = Earnings(
                    thisWeekActual = summary.thisWeekActual.toDouble(),
                    thisWeekBaseline = summary.thisWeekBaseline.toDouble(),
                    protectedIncome = summary.protectedIncome.toDouble(),
                    history = summary.history.map { EarningRecord(it.week, it.actual.toDouble()) }
                )
                _uiState.value = EarningsUiState.Success(earnings)
            } else {
                _uiState.value = EarningsUiState.Error("Failed to load earnings")
            }
        } catch (e: Exception) {
            _uiState.value = EarningsUiState.Error(e.message ?: "Unknown error")
        }
    }
}

sealed class EarningsUiState {
    object Loading : EarningsUiState()
    data class Success(val earnings: Earnings) : EarningsUiState()
    data class Error(val message: String) : EarningsUiState()
}
