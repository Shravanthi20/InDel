package com.imaginai.indel.ui.auth

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowDropDown
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.ui.navigation.Screen
import com.imaginai.indel.ui.shared.isZoneCAndAbove
import com.imaginai.indel.ui.shared.vehicleOptionsForZone
import com.imaginai.indel.ui.shared.zoneOptions

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun OnboardingScreen(
    navController: NavController,
    viewModel: OnboardingViewModel = hiltViewModel()
) {
    val name by viewModel.name.collectAsState()
    val zone by viewModel.zone.collectAsState()
    val vehicleType by viewModel.vehicleType.collectAsState()
    val upiId by viewModel.upiId.collectAsState()
    val uiState by viewModel.uiState.collectAsState()
    val isRestrictedZone = isZoneCAndAbove(zone)
    val availableVehicles = vehicleOptionsForZone(zone)
    var zoneExpanded by remember { mutableStateOf(false) }
    var vehicleExpanded by remember { mutableStateOf(false) }

    LaunchedEffect(uiState) {
        if (uiState is OnboardingUiState.Success) {
            navController.navigate(Screen.Home.route) {
                popUpTo(Screen.Onboarding.route) { inclusive = true }
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Profile Setup", fontWeight = FontWeight.Bold) },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.primary,
                    titleContentColor = Color.White
                )
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(24.dp)
                .verticalScroll(rememberScrollState()),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Text(
                text = "Help us personalize your protection",
                style = MaterialTheme.typography.bodyLarge,
                color = MaterialTheme.colorScheme.secondary,
                modifier = Modifier.padding(bottom = 32.dp)
            )

            OutlinedTextField(
                value = name,
                onValueChange = viewModel::onNameChanged,
                label = { Text("Full Name") },
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(12.dp)
            )
            Spacer(modifier = Modifier.height(16.dp))

            ExposedDropdownMenuBox(
                expanded = zoneExpanded,
                onExpandedChange = { zoneExpanded = !zoneExpanded }
            ) {
                OutlinedTextField(
                    value = zone.ifBlank { "Select Work Zone" },
                    onValueChange = {},
                    readOnly = true,
                    label = { Text("Work Zone") },
                    trailingIcon = {
                        Icon(Icons.Default.ArrowDropDown, contentDescription = "Select zone")
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .menuAnchor(),
                    shape = RoundedCornerShape(12.dp)
                )
                ExposedDropdownMenu(
                    expanded = zoneExpanded,
                    onDismissRequest = { zoneExpanded = false }
                ) {
                    zoneOptions.forEach { option ->
                        DropdownMenuItem(
                            text = { Text(option.label) },
                            onClick = {
                                viewModel.onZoneChanged(option.value)
                                zoneExpanded = false
                            }
                        )
                    }
                }
            }
            Spacer(modifier = Modifier.height(16.dp))

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
                        .menuAnchor(),
                    shape = RoundedCornerShape(12.dp)
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
                    color = MaterialTheme.colorScheme.secondary,
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = 6.dp)
                )
            }
            Spacer(modifier = Modifier.height(16.dp))

            OutlinedTextField(
                value = upiId,
                onValueChange = viewModel::onUpiIdChanged,
                label = { Text("UPI ID (for payouts)") },
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(12.dp)
            )
            
            Spacer(modifier = Modifier.weight(1f))
            Spacer(modifier = Modifier.height(32.dp))

            Button(
                onClick = viewModel::submitOnboarding,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp),
                enabled = uiState !is OnboardingUiState.Loading && name.isNotBlank() && zone.isNotBlank() && vehicleType.isNotBlank(),
                shape = RoundedCornerShape(12.dp)
            ) {
                if (uiState is OnboardingUiState.Loading) {
                    CircularProgressIndicator(color = Color.White, modifier = Modifier.size(24.dp))
                } else {
                    Text("Complete Setup", fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                }
            }

            if (uiState is OnboardingUiState.Error) {
                Text(
                    text = (uiState as OnboardingUiState.Error).message,
                    color = MaterialTheme.colorScheme.error,
                    style = MaterialTheme.typography.bodySmall,
                    modifier = Modifier.padding(top = 16.dp)
                )
            }
        }
    }
}
