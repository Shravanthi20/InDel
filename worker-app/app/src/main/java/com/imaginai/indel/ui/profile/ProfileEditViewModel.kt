package com.imaginai.indel.ui.profile

import android.util.Log
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.Zone
import com.imaginai.indel.data.model.ZonePath
import com.imaginai.indel.data.repository.WorkerRepository
import com.imaginai.indel.ui.shared.isVehicleAllowedForZoneLevel
import com.imaginai.indel.ui.shared.isValidUpiId
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import javax.inject.Inject

@HiltViewModel
class ProfileEditViewModel @Inject constructor(
    private val workerRepository: WorkerRepository
) : ViewModel() {

    private val _name = MutableStateFlow("")
    val name = _name.asStateFlow()

    private val _zoneLevel = MutableStateFlow("")
    val zoneLevel = _zoneLevel.asStateFlow()

    private val _zoneName = MutableStateFlow("")
    val zoneName = _zoneName.asStateFlow()

    private val _area = MutableStateFlow("")
    val area = _area.asStateFlow()

    private val _selectedZone = MutableStateFlow<Zone?>(null)
    val selectedZone = _selectedZone.asStateFlow()
    
    private val _selectedPath = MutableStateFlow<ZonePath?>(null)
    val selectedPath = _selectedPath.asStateFlow()

    private val _vehicleType = MutableStateFlow("")
    val vehicleType = _vehicleType.asStateFlow()

    private val _upiId = MutableStateFlow("")
    val upiId = _upiId.asStateFlow()

    private val _availableZones = MutableStateFlow<List<Zone>>(emptyList())
    val availableZones = _availableZones.asStateFlow()
    
    private val _availablePaths = MutableStateFlow<List<ZonePath>>(emptyList())
    val availablePaths = _availablePaths.asStateFlow()
    
    private val _isFetchingPaths = MutableStateFlow(false)
    val isFetchingPaths = _isFetchingPaths.asStateFlow()

    private val _uiState = MutableStateFlow<ProfileEditUiState>(ProfileEditUiState.Loading)
    val uiState = _uiState.asStateFlow()

    val filteredPaths: StateFlow<List<ZonePath>> = combine(_zoneName, _availablePaths) { query, paths ->
        if (query.isBlank()) paths.take(30)
        else paths.filter { it.displayName?.contains(query, ignoreCase = true) == true }.take(30)
    }.flowOn(Dispatchers.Default).stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    val filteredZones: StateFlow<List<Zone>> = combine(_zoneName, _availableZones) { query, zones ->
        if (query.isBlank()) zones.take(20)
        else zones.filter { it.name.contains(query, ignoreCase = true) }.take(20)
    }.flowOn(Dispatchers.Default).stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    init {
        loadProfile()
        fetchZones()
    }

    private fun fetchZones() {
        viewModelScope.launch {
            try {
                val response = workerRepository.getZones()
                if (response.isSuccessful) {
                    _availableZones.value = response.body()?.zones ?: emptyList()
                }
            } catch (e: Exception) {
                Log.e("ProfileEdit", "Zones fetch failed", e)
            }
        }
    }
    
    private fun fetchZonePaths(level: String) {
        viewModelScope.launch {
            _isFetchingPaths.value = true
            try {
                val type = level.lowercase()
                val response = workerRepository.getZonePaths(type)
                if (response.isSuccessful) {
                    val body = response.body()
                    val paths = withContext(Dispatchers.Default) {
                        val result = mutableListOf<ZonePath>()
                        body?.cities?.forEach { result.add(ZonePath(displayName = it, city = it)) }
                        body?.cityPairs?.forEach { pair ->
                            val display = when (type) {
                                "b" -> "${pair.from} to ${pair.to} (${pair.state ?: ""})"
                                "c" -> "${pair.from} (${pair.fromState ?: ""}) to ${pair.to} (${pair.toState ?: ""})"
                                else -> "${pair.from} to ${pair.to}"
                            }
                            result.add(ZonePath(displayName = display, fromCity = pair.from, toCity = pair.to))
                        }
                        if (result.isEmpty() && body?.paths != null) result.addAll(body.paths)
                        result
                    }
                    _availablePaths.value = paths
                }
            } catch (e: Exception) {
                Log.e("ProfileEdit", "Paths fetch failed", e)
            } finally {
                _isFetchingPaths.value = false
            }
        }
    }

    fun onNameChanged(value: String) { _name.value = value }

    fun onZoneLevelChanged(value: String) {
        _zoneLevel.value = value
        _zoneName.value = ""
        _area.value = ""
        _selectedZone.value = null
        _selectedPath.value = null
        _availablePaths.value = emptyList()
        if (value.isNotBlank()) fetchZonePaths(value)
    }

    fun onZoneNameChanged(value: String) { _zoneName.value = value }

    fun onZoneSelected(zone: Zone) {
        _selectedZone.value = zone
        _zoneName.value = zone.name
    }
    
    fun onPathSelected(path: ZonePath) {
        _selectedPath.value = path
        _zoneName.value = path.displayName ?: ""
    }

    fun onAreaChanged(value: String) { _area.value = value }
    fun onVehicleTypeChanged(value: String) { _vehicleType.value = value }
    fun onUpiIdChanged(value: String) { _upiId.value = value }

    fun loadProfile() {
        viewModelScope.launch {
            _uiState.value = ProfileEditUiState.Loading
            try {
                val response = workerRepository.getProfile()
                if (response.isSuccessful) {
                    val worker = response.body()?.worker
                    if (worker != null) {
                        _name.value = worker.name
                        _zoneLevel.value = worker.zoneLevel
                        _zoneName.value = worker.zoneName
                        _area.value = worker.area ?: ""
                        _vehicleType.value = worker.vehicleType
                        _upiId.value = worker.upiId
                        if (worker.zoneLevel.isNotBlank()) fetchZonePaths(worker.zoneLevel)
                        _uiState.value = ProfileEditUiState.Idle
                    }
                }
            } catch (e: Exception) {
                _uiState.value = ProfileEditUiState.Error(e.message ?: "Load failed")
            }
        }
    }

    fun saveProfile() {
        if (_name.value.isBlank() || _zoneLevel.value.isBlank() || _zoneName.value.isBlank()) {
            _uiState.value = ProfileEditUiState.Error("Please fill required fields")
            return
        }
        viewModelScope.launch {
            _uiState.value = ProfileEditUiState.Saving
            try {
                val response = workerRepository.updateProfile(
                    name = _name.value,
                    zoneLevel = _zoneLevel.value,
                    zoneName = _zoneName.value,
                    area = _area.value,
                    zoneId = _selectedZone.value?.zoneId,
                    city = _selectedPath.value?.city ?: _selectedZone.value?.city ?: _selectedPath.value?.fromCity,
                    vehicleType = _vehicleType.value,
                    upiId = _upiId.value
                )
                if (response.isSuccessful) _uiState.value = ProfileEditUiState.Saved
                else _uiState.value = ProfileEditUiState.Error("Save failed")
            } catch (e: Exception) {
                _uiState.value = ProfileEditUiState.Error(e.message ?: "Error saving")
            }
        }
    }
}

sealed class ProfileEditUiState {
    object Loading : ProfileEditUiState()
    object Idle : ProfileEditUiState()
    object Saving : ProfileEditUiState()
    object Saved : ProfileEditUiState()
    data class Error(val message: String) : ProfileEditUiState()
}
