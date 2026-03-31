package com.imaginai.indel.ui.profile

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.ArrowDropDown
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.ui.shared.isZoneCAndAboveLevel
import com.imaginai.indel.ui.shared.vehicleOptionsForZoneLevel
import com.imaginai.indel.ui.shared.zoneLevelOptions
import com.imaginai.indel.ui.shared.zoneNamesForLevel

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

    val isRestrictedZone = isZoneCAndAboveLevel(zoneLevel)
    val availableVehicles = vehicleOptionsForZoneLevel(zoneLevel)
    val availableZoneNames = zoneNamesForLevel(zoneLevel)

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
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.primary,
                    titleContentColor = Color.White,
                    navigationIconContentColor = Color.White
                )
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(20.dp)
                .verticalScroll(rememberScrollState()),
            verticalArrangement = Arrangement.spacedBy(14.dp)
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

            // Zone Level Dropdown
            ExposedDropdownMenuBox(
                expanded = zoneLevelExpanded,
                onExpandedChange = { zoneLevelExpanded = !zoneLevelExpanded }
            ) {
                OutlinedTextField(
                    value = zoneLevel.ifBlank { "Select Zone Level" },
                    onValueChange = {},
                    readOnly = true,
                    label = { Text("Zone Level") },
                    trailingIcon = {
                        Icon(Icons.Default.ArrowDropDown, contentDescription = "Select zone level")
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .menuAnchor()
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

            // Zone Name Dropdown
            ExposedDropdownMenuBox(
                expanded = zoneNameExpanded,
                onExpandedChange = { if (zoneLevel.isNotBlank()) zoneNameExpanded = !zoneNameExpanded }
            ) {
                OutlinedTextField(
                    value = zoneName.ifBlank { "Select Zone Name" },
                    onValueChange = {},
                    readOnly = true,
                    label = { Text("Zone Name") },
                    trailingIcon = {
                        Icon(Icons.Default.ArrowDropDown, contentDescription = "Select zone name")
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .menuAnchor(),
                    enabled = zoneLevel.isNotBlank()
                )
                ExposedDropdownMenu(
                    expanded = zoneNameExpanded,
                    onDismissRequest = { zoneNameExpanded = false }
                ) {
                    availableZoneNames.forEach { nameOption ->
                        DropdownMenuItem(
                            text = { Text(nameOption) },
                            onClick = {
                                viewModel.onZoneNameChanged(nameOption)
                                zoneNameExpanded = false
                            }
                        )
                    }
                }
            }

            ExposedDropdownMenuBox(
                expanded = vehicleExpanded,
                onExpandedChange = { vehicleExpanded = !vehicleExpanded }
            ) {
                OutlinedTextField(
                    value = vehicleType.ifBlank { "Select Vehicle" },
                    onValueChange = {},
                    readOnly = true,
                    label = { Text("Vehicle Type") },
                    trailingIcon = {
                        Icon(Icons.Default.ArrowDropDown, contentDescription = "Select vehicle")
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .menuAnchor()
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

            if (isRestrictedZone) {
                Text(
                    text = "Zone C and above allow only four-wheelers.",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.secondary
                )
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
                modifier = Modifier
                    .fillMaxWidth()
                    .height(52.dp),
                enabled = uiState !is ProfileEditUiState.Saving && uiState !is ProfileEditUiState.Loading
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
