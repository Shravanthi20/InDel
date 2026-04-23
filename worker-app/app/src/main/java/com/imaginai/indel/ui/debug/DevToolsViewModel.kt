package com.imaginai.indel.ui.debug

import android.util.Log
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.api.WorkerApiService
import com.imaginai.indel.data.repository.WorkerRepository
import com.imaginai.indel.data.model.DisruptionRequest
import com.imaginai.indel.data.model.CountRequest
import com.imaginai.indel.data.model.ZoneLevelOption
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

sealed class DevToolsActionState {
    object Idle : DevToolsActionState()
    object Loading : DevToolsActionState()
    data class Success(val message: String) : DevToolsActionState()
    data class Error(val message: String) : DevToolsActionState()
}

@HiltViewModel
class DevToolsViewModel @Inject constructor(
    private val apiService: WorkerApiService,
    private val workerRepository: WorkerRepository
) : ViewModel() {

    companion object {
        private const val TAG = "DevToolsViewModel"
        private const val ZONE_NAME_LIMIT = 15
    }

    // Zone level dropdown (A / B / C)
    private val _zoneLevels = MutableStateFlow<List<ZoneLevelOption>>(
        listOf(
            ZoneLevelOption("A", "A — Same City"),
            ZoneLevelOption("B", "B — Intra State"),
            ZoneLevelOption("C", "C — Inter State")
        )
    )
    val zoneLevels: StateFlow<List<ZoneLevelOption>> = _zoneLevels.asStateFlow()

    // Zone name dropdown — first 15 entries for the selected level
    private val _zoneNames = MutableStateFlow<List<String>>(emptyList())
    val zoneNames: StateFlow<List<String>> = _zoneNames.asStateFlow()

    // Selections
    val selectedLevel = MutableStateFlow("A")
    val selectedZone = MutableStateFlow("")
    val selectedDisruptionType = MutableStateFlow("WEATHER")

    // Order control counts
    val assignCount = MutableStateFlow(3)
    val simulateCount = MutableStateFlow(2)

    // Action result
    private val _actionState = MutableStateFlow<DevToolsActionState>(DevToolsActionState.Idle)
    val actionState: StateFlow<DevToolsActionState> = _actionState.asStateFlow()

    init {
        loadZoneLevels()
        loadZoneNames("A")
    }

    fun loadZoneLevels() {
        viewModelScope.launch {
            try {
                val resp = workerRepository.getZoneLevels()
                if (resp.isSuccessful) {
                    val levels = resp.body()?.levels
                    if (!levels.isNullOrEmpty()) {
                        _zoneLevels.value = levels
                    }
                }
            } catch (e: Exception) {
                Log.w(TAG, "Zone levels fetch failed, using defaults: ${e.message}")
                // Keep the hardcoded defaults — no error surfaced to user
            }
        }
    }

    fun onLevelSelected(level: String) {
        selectedLevel.value = level
        selectedZone.value = ""
        loadZoneNames(level)
    }

    private fun loadZoneNames(level: String) {
        viewModelScope.launch {
            try {
                val resp = workerRepository.getZonePaths(level.lowercase())
                if (resp.isSuccessful) {
                    val body = resp.body()
                    val names: List<String> = when (level.uppercase()) {
                        "A" -> body?.cities
                            ?.map { it.city }
                            ?.take(ZONE_NAME_LIMIT)
                            ?: emptyList()
                        else -> body?.zones
                            ?.map { it.zoneName }
                            ?.take(ZONE_NAME_LIMIT)
                            ?: emptyList()
                    }
                    _zoneNames.value = names
                    if (names.isNotEmpty()) selectedZone.value = names.first()
                }
            } catch (e: Exception) {
                Log.w(TAG, "Zone names fetch failed for level=$level: ${e.message}")
            }
        }
    }

    fun triggerDisruption() {
        val level = selectedLevel.value
        val zone = selectedZone.value
        val type = selectedDisruptionType.value

        if (zone.isBlank()) {
            _actionState.value = DevToolsActionState.Error("Select a zone name first")
            return
        }

        viewModelScope.launch {
            _actionState.value = DevToolsActionState.Loading
            try {
                val resp = apiService.triggerDisruption(
                    DisruptionRequest(
                        disruptionType = type,
                        zoneLevel = level,
                        zoneName = zone
                    )
                )
                if (resp.isSuccessful) {
                    _actionState.value = DevToolsActionState.Success(
                        "⚡ Disruption triggered\nZone: $zone (Level $level)\nType: $type"
                    )
                } else {
                    val err = resp.errorBody()?.string() ?: "Unknown error"
                    _actionState.value = DevToolsActionState.Error(err)
                }
            } catch (e: Exception) {
                _actionState.value = DevToolsActionState.Error(e.message ?: "Request failed")
            }
        }
    }

    fun assignOrders() {
        val count = assignCount.value
        viewModelScope.launch {
            _actionState.value = DevToolsActionState.Loading
            try {
                val resp = apiService.assignOrders(CountRequest(count))
                if (resp.isSuccessful) {
                    _actionState.value = DevToolsActionState.Success("✓ $count order(s) assigned")
                } else {
                    _actionState.value = DevToolsActionState.Error(
                        resp.errorBody()?.string() ?: "Assign failed"
                    )
                }
            } catch (e: Exception) {
                _actionState.value = DevToolsActionState.Error(e.message ?: "Request failed")
            }
        }
    }

    fun simulateDeliveries() {
        val count = simulateCount.value
        viewModelScope.launch {
            _actionState.value = DevToolsActionState.Loading
            try {
                val resp = apiService.simulateDeliveries(CountRequest(count))
                if (resp.isSuccessful) {
                    _actionState.value = DevToolsActionState.Success("✓ $count delivery(ies) simulated")
                } else {
                    _actionState.value = DevToolsActionState.Error(
                        resp.errorBody()?.string() ?: "Simulate failed"
                    )
                }
            } catch (e: Exception) {
                _actionState.value = DevToolsActionState.Error(e.message ?: "Request failed")
            }
        }
    }

    fun settleEarnings() {
        viewModelScope.launch {
            _actionState.value = DevToolsActionState.Loading
            try {
                val resp = apiService.settleEarnings()
                if (resp.isSuccessful) {
                    _actionState.value = DevToolsActionState.Success("₹ Earnings settled")
                } else {
                    _actionState.value = DevToolsActionState.Error(
                        resp.errorBody()?.string() ?: "Settle failed"
                    )
                }
            } catch (e: Exception) {
                _actionState.value = DevToolsActionState.Error(e.message ?: "Request failed")
            }
        }
    }

    fun resetZone() {
        viewModelScope.launch {
            _actionState.value = DevToolsActionState.Loading
            try {
                val resp = apiService.resetZone()
                if (resp.isSuccessful) {
                    _actionState.value = DevToolsActionState.Success("⊙ Zone state reset")
                } else {
                    _actionState.value = DevToolsActionState.Error(
                        resp.errorBody()?.string() ?: "Reset failed"
                    )
                }
            } catch (e: Exception) {
                _actionState.value = DevToolsActionState.Error(e.message ?: "Request failed")
            }
        }
    }

    fun fullReset() {
        viewModelScope.launch {
            _actionState.value = DevToolsActionState.Loading
            try {
                val resp = apiService.resetDemo()
                if (resp.isSuccessful) {
                    _actionState.value = DevToolsActionState.Success("✕ Full demo reset complete")
                } else {
                    _actionState.value = DevToolsActionState.Error(
                        resp.errorBody()?.string() ?: "Full reset failed"
                    )
                }
            } catch (e: Exception) {
                _actionState.value = DevToolsActionState.Error(e.message ?: "Request failed")
            }
        }
    }

    fun clearActionState() {
        _actionState.value = DevToolsActionState.Idle
    }
}
