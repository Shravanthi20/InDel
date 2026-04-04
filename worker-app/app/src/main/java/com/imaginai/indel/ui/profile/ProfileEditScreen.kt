package com.imaginai.indel.ui.profile

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.ui.shared.vehicleOptionsForZoneLevel
import com.imaginai.indel.ui.shared.zoneLevelOptions

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ProfileEditScreen(
    navController: NavController,
    viewModel: ProfileEditViewModel = hiltViewModel()
) {
    val name by viewModel.name.collectAsState()
    val zoneLevel by viewModel.zoneLevel.collectAsState()
    val zoneName by viewModel.zoneName.collectAsState()
    val vehicleType by viewModel.vehicleType.collectAsState()
    val upiId by viewModel.upiId.collectAsState()
    val uiState by viewModel.uiState.collectAsState()
    
    val filteredZones by viewModel.filteredZones.collectAsState()
    val filteredPaths by viewModel.filteredPaths.collectAsState()
    val isFetchingPaths by viewModel.isFetchingPaths.collectAsState()

    val availableVehicles = vehicleOptionsForZoneLevel(zoneLevel)

    var zoneLevelExpanded by remember { mutableStateOf(false) }
    var zoneNameExpanded by remember { mutableStateOf(false) }
    var vehicleExpanded by remember { mutableStateOf(false) }

    LaunchedEffect(uiState) {
        if (uiState is ProfileEditUiState.Saved) {
            navController.navigateUp()
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Edit Profile", fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                }
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(20.dp)
                .verticalScroll(rememberScrollState()),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            if (uiState is ProfileEditUiState.Loading) {
                LinearProgressIndicator(modifier = Modifier.fillMaxWidth())
            }

            OutlinedTextField(
                value = name,
                onValueChange = viewModel::onNameChanged,
                label = { Text("Full Name") },
                modifier = Modifier.fillMaxWidth()
            )

            // Zone Level
            ExposedDropdownMenuBox(
                expanded = zoneLevelExpanded,
                onExpandedChange = { zoneLevelExpanded = it }
            ) {
                OutlinedTextField(
                    value = zoneLevel.ifBlank { "Select Zone Level" },
                    onValueChange = {},
                    readOnly = true,
                    label = { Text("Zone Level") },
                    trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = zoneLevelExpanded) },
                    modifier = Modifier.fillMaxWidth().menuAnchor(MenuAnchorType.PrimaryNotEditable, true)
                )
                ExposedDropdownMenu(
                    expanded = zoneLevelExpanded,
                    onDismissRequest = { zoneLevelExpanded = false }
                ) {
                    zoneLevelOptions.forEach { option ->
                        DropdownMenuItem(
                            text = { Text(option.label) },
                            onClick = {
                                viewModel.onZoneLevelChanged(option.level)
                                zoneLevelExpanded = false
                            }
                        )
                    }
                }
            }

            // Zone Name - Searchable
            val isHub = zoneLevel.lowercase() == "hub"
            ExposedDropdownMenuBox(
                expanded = zoneNameExpanded,
                onExpandedChange = { if (zoneLevel.isNotBlank()) zoneNameExpanded = it }
            ) {
                OutlinedTextField(
                    value = zoneName,
                    onValueChange = { 
                        viewModel.onZoneNameChanged(it)
                        zoneNameExpanded = true 
                    },
                    readOnly = isHub,
                    label = { Text(if (isHub) "Select Hub" else "Search Zone Name") },
                    placeholder = { if (!isHub) Text("Start typing city name...") },
                    trailingIcon = { 
                        if (isFetchingPaths) CircularProgressIndicator(modifier = Modifier.size(20.dp), strokeWidth = 2.dp)
                        else ExposedDropdownMenuDefaults.TrailingIcon(expanded = zoneNameExpanded)
                    },
                    modifier = Modifier.fillMaxWidth().menuAnchor(MenuAnchorType.PrimaryEditable, zoneLevel.isNotBlank()),
                    enabled = zoneLevel.isNotBlank()
                )
                
                ExposedDropdownMenu(
                    expanded = zoneNameExpanded,
                    onDismissRequest = { zoneNameExpanded = false }
                ) {
                    if (isHub) {
                        filteredZones.forEach { zone ->
                            DropdownMenuItem(
                                text = { Text(zone.name) },
                                onClick = {
                                    viewModel.onZoneSelected(zone)
                                    zoneNameExpanded = false
                                }
                            )
                        }
                    } else {
                        if (filteredPaths.isEmpty() && zoneName.isNotBlank()) {
                            DropdownMenuItem(
                                text = { Text("No matches found", color = Color.Gray) },
                                onClick = { },
                                enabled = false
                            )
                        }
                        filteredPaths.forEach { path ->
                            DropdownMenuItem(
                                text = { Text(path.displayName ?: "") },
                                onClick = {
                                    viewModel.onPathSelected(path)
                                    zoneNameExpanded = false
                                }
                            )
                        }
                    }
                }
            }

            // Vehicle
            ExposedDropdownMenuBox(
                expanded = vehicleExpanded,
                onExpandedChange = { vehicleExpanded = it }
            ) {
                OutlinedTextField(
                    value = vehicleType.ifBlank { "Select Vehicle" },
                    onValueChange = {},
                    readOnly = true,
                    label = { Text("Vehicle Type") },
                    trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = vehicleExpanded) },
                    modifier = Modifier.fillMaxWidth().menuAnchor(MenuAnchorType.PrimaryNotEditable, true)
                )
                ExposedDropdownMenu(
                    expanded = vehicleExpanded,
                    onDismissRequest = { vehicleExpanded = false }
                ) {
                    availableVehicles.forEach { vehicle ->
                        DropdownMenuItem(
                            text = { Text(vehicle.replaceFirstChar { it.uppercase() }) },
                            onClick = {
                                viewModel.onVehicleTypeChanged(vehicle)
                                vehicleExpanded = false
                            }
                        )
                    }
                }
            }

            OutlinedTextField(
                value = upiId,
                onValueChange = viewModel::onUpiIdChanged,
                label = { Text("UPI ID") },
                modifier = Modifier.fillMaxWidth()
            )

            if (uiState is ProfileEditUiState.Error) {
                Text(
                    text = (uiState as ProfileEditUiState.Error).message,
                    color = MaterialTheme.colorScheme.error,
                    style = MaterialTheme.typography.bodySmall
                )
            }

            Spacer(modifier = Modifier.height(8.dp))

            Button(
                onClick = viewModel::saveProfile,
                modifier = Modifier.fillMaxWidth().height(52.dp),
                enabled = uiState !is ProfileEditUiState.Saving && 
                        name.isNotBlank() && zoneLevel.isNotBlank() && zoneName.isNotBlank()
            ) {
                if (uiState is ProfileEditUiState.Saving) {
                    CircularProgressIndicator(modifier = Modifier.size(20.dp), color = Color.White)
                } else {
                    Text("Save Changes")
                }
            }
        }
    }
}
