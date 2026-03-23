package com.imaginai.indel.ui.claims

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.Claim
import com.imaginai.indel.data.repository.ClaimsRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class ClaimDetailViewModel @Inject constructor(
    private val claimsRepository: ClaimsRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<ClaimDetailUiState>(ClaimDetailUiState.Loading)
    val uiState = _uiState.asStateFlow()

    fun loadClaimDetail(claimId: String) {
        viewModelScope.launch {
            _uiState.value = ClaimDetailUiState.Loading
            try {
                val response = claimsRepository.getClaimDetail(claimId)
                if (response.isSuccessful) {
                    _uiState.value = ClaimDetailUiState.Success(response.body()!!)
                } else {
                    _uiState.value = ClaimDetailUiState.Error("Failed to load claim details")
                }
            } catch (e: Exception) {
                _uiState.value = ClaimDetailUiState.Error(e.message ?: "Unknown error")
            }
        }
    }
}

sealed class ClaimDetailUiState {
    object Loading : ClaimDetailUiState()
    data class Success(val claim: Claim) : ClaimDetailUiState()
    data class Error(val message: String) : ClaimDetailUiState()
}
