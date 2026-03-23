package com.imaginai.indel.ui.claims

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.Claim
import com.imaginai.indel.data.model.WalletResponse
import com.imaginai.indel.data.repository.ClaimsRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class ClaimsViewModel @Inject constructor(
    private val claimsRepository: ClaimsRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<ClaimsUiState>(ClaimsUiState.Loading)
    val uiState = _uiState.asStateFlow()

    private val _isRefreshing = MutableStateFlow(false)
    val isRefreshing = _isRefreshing.asStateFlow()

    init {
        loadClaimsData()
    }

    fun loadClaimsData() {
        viewModelScope.launch {
            _uiState.value = ClaimsUiState.Loading
            fetchClaimsData()
        }
    }

    fun refresh() {
        viewModelScope.launch {
            _isRefreshing.value = true
            fetchClaimsData()
            delay(500)
            _isRefreshing.value = false
        }
    }

    private suspend fun fetchClaimsData() {
        try {
            val claimsRes = claimsRepository.getClaims()
            val walletRes = claimsRepository.getWallet()

            if (claimsRes.isSuccessful && walletRes.isSuccessful) {
                _uiState.value = ClaimsUiState.Success(
                    claims = claimsRes.body()?.claims ?: emptyList(),
                    wallet = walletRes.body()!!
                )
            } else {
                _uiState.value = ClaimsUiState.Error("Failed to load claims data")
            }
        } catch (e: Exception) {
            _uiState.value = ClaimsUiState.Error(e.message ?: "Unknown error")
        }
    }
}

sealed class ClaimsUiState {
    object Loading : ClaimsUiState()
    data class Success(
        val claims: List<Claim>,
        val wallet: WalletResponse
    ) : ClaimsUiState()
    data class Error(val message: String) : ClaimsUiState()
}
