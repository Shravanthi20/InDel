package com.imaginai.indel.ui.auth

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
import com.imaginai.indel.ui.shared.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun OnboardingScreen(
    navController: NavController,
    viewModel: OnboardingViewModel = hiltViewModel()
) {
    val name by viewModel.name.collectAsState()
    val vehicleType by viewModel.vehicleType.collectAsState()
    val vehicleName by viewModel.vehicleName.collectAsState()
    val upiId by viewModel.upiId.collectAsState()
    val uiState by viewModel.uiState.collectAsState()

    var vehicleExpanded by remember { mutableStateOf(false) }
    var vehicleNameExpanded by remember { mutableStateOf(false) }
    var vehicleNameOption by remember { mutableStateOf("") }

    val commonTransportMeans = listOf(
        "scooter",
        "motorcycle",
        "auto-rickshaw",
        "hatchback",
        "sedan",
        "suv",
        "pickup-van",
        "mini-truck",
        "other"
    )

    LaunchedEffect(uiState) {
        if (uiState is OnboardingUiState.Success) {
            navController.navigate(Screen.PlanSelection.route) {
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
                        .menuAnchor(MenuAnchorType.PrimaryNotEditable, true),
                    shape = RoundedCornerShape(12.dp)
                )
                ExposedDropdownMenu(
                    expanded = vehicleExpanded,
                    onDismissRequest = { vehicleExpanded = false }
                ) {
                    listOf("two-wheeler", "four-wheeler-small", "four-wheeler-large").forEach { vehicle ->
                        DropdownMenuItem(
                            text = { Text(vehicle) },
                            onClick = {
                                viewModel.onVehicleTypeChanged(vehicle)
                                vehicleExpanded = false
                            }
                        )
                    }
                }
            }
            Spacer(modifier = Modifier.height(16.dp))

            ExposedDropdownMenuBox(
                expanded = vehicleNameExpanded,
                onExpandedChange = {
                    if (vehicleType.isNotBlank()) {
                        vehicleNameExpanded = !vehicleNameExpanded
                    }
                }
            ) {
                OutlinedTextField(
                    value = when {
                        vehicleNameOption == "other" && vehicleName.isNotBlank() -> "other"
                        vehicleNameOption.isNotBlank() -> vehicleNameOption
                        else -> "Select Transport Means"
                    },
                    onValueChange = {},
                    readOnly = true,
                    label = { Text("Vehicle Name") },
                    trailingIcon = {
                        Icon(Icons.Default.ArrowDropDown, contentDescription = "Select vehicle name")
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .menuAnchor(MenuAnchorType.PrimaryNotEditable, vehicleType.isNotBlank()),
                    shape = RoundedCornerShape(12.dp),
                    enabled = vehicleType.isNotBlank()
                )
                ExposedDropdownMenu(
                    expanded = vehicleNameExpanded,
                    onDismissRequest = { vehicleNameExpanded = false }
                ) {
                    commonTransportMeans.forEach { option ->
                        DropdownMenuItem(
                            text = { Text(option) },
                            onClick = {
                                vehicleNameOption = option
                                if (option == "other") {
                                    viewModel.onVehicleNameChanged("")
                                } else {
                                    viewModel.onVehicleNameChanged(option)
                                }
                                vehicleNameExpanded = false
                            }
                        )
                    }
                }
            }
            if (vehicleNameOption == "other") {
                Spacer(modifier = Modifier.height(12.dp))
                OutlinedTextField(
                    value = vehicleName,
                    onValueChange = viewModel::onVehicleNameChanged,
                    label = { Text("Type Vehicle Name") },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp)
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
                enabled = uiState !is OnboardingUiState.Loading && 
                        name.isNotBlank() && 
                    vehicleType.isNotBlank() &&
                    vehicleName.isNotBlank(),
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
