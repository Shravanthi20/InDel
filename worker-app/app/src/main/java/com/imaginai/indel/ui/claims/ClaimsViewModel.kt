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

    private var isFetching = false

    init {
        loadClaimsData()
        startAutoRefresh()
    }

    fun loadClaimsData() {
        viewModelScope.launch {
            // Prevent multiple concurrent fetches
            if (isFetching) return@launch
            _uiState.value = ClaimsUiState.Loading
            fetchClaimsData()
        }
    }

    fun refresh() {
        viewModelScope.launch {
            // Prevent multiple concurrent fetches
            if (isFetching) return@launch
            _isRefreshing.value = true
            fetchClaimsData()
            delay(500)
            _isRefreshing.value = false
        }
    }

    private suspend fun fetchClaimsData() {
        try {
            isFetching = true
            val claimsRes = claimsRepository.getClaims()
            val walletRes = claimsRepository.getWallet()

            if (claimsRes.isSuccessful && walletRes.isSuccessful) {
                val claims = claimsRes.body()?.claims ?: emptyList()
                val wallet = walletRes.body()!!
                // Deduplicate claims by claim_id just in case
                val uniqueClaims = claims.distinctBy { it.claimId }
                _uiState.value = ClaimsUiState.Success(
                    claims = uniqueClaims,
                    wallet = wallet
                )
            } else {
                _uiState.value = ClaimsUiState.Error("Failed to load claims data")
            }
        } catch (e: Exception) {
            _uiState.value = ClaimsUiState.Error(e.message ?: "Unknown error")
        } finally {
            isFetching = false
        }
    }

    private fun startAutoRefresh() {
        viewModelScope.launch {
            while (true) {
                delay(12000)
                // Only auto-fetch if not already fetching
                if (!isFetching && uiState.value is ClaimsUiState.Success) {
                    fetchClaimsData()
                }
            }
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
