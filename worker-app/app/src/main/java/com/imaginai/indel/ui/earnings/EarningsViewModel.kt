package com.imaginai.indel.ui.earnings

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.EarningsSummary
import com.imaginai.indel.data.repository.EarningsRepository
import dagger.hilt.android.lifecycle.HiltViewModel
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

    init {
        loadEarnings()
    }

    fun loadEarnings() {
        viewModelScope.launch {
            _uiState.value = EarningsUiState.Loading
            try {
                val response = earningsRepository.getEarnings()
                if (response.isSuccessful) {
                    _uiState.value = EarningsUiState.Success(response.body()!!)
                } else {
                    _uiState.value = EarningsUiState.Error("Failed to load earnings")
                }
            } catch (e: Exception) {
                _uiState.value = EarningsUiState.Error(e.message ?: "Unknown error")
            }
        }
    }
}

sealed class EarningsUiState {
    object Loading : EarningsUiState()
    data class Success(val earnings: EarningsSummary) : EarningsUiState()
    data class Error(val message: String) : EarningsUiState()
}
