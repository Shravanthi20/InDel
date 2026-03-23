package com.imaginai.indel.ui.auth

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.ui.navigation.Screen

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

    LaunchedEffect(uiState) {
        if (uiState is OnboardingUiState.Success) {
            navController.navigate(Screen.Home.route) {
                popUpTo(Screen.Onboarding.route) { inclusive = true }
            }
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp)
            .verticalScroll(rememberScrollState()),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Text(text = "Complete Your Profile", style = MaterialTheme.typography.headlineMedium)
        Spacer(modifier = Modifier.height(32.dp))

        OutlinedTextField(
            value = name,
            onValueChange = viewModel::onNameChanged,
            label = { Text("Full Name") },
            modifier = Modifier.fillMaxWidth()
        )
        Spacer(modifier = Modifier.height(16.dp))

        OutlinedTextField(
            value = zone,
            onValueChange = viewModel::onZoneChanged,
            label = { Text("Zone (e.g. Tambaram, Chennai)") },
            modifier = Modifier.fillMaxWidth()
        )
        Spacer(modifier = Modifier.height(16.dp))

        OutlinedTextField(
            value = vehicleType,
            onValueChange = viewModel::onVehicleTypeChanged,
            label = { Text("Vehicle Type (bike/auto/van)") },
            modifier = Modifier.fillMaxWidth()
        )
        Spacer(modifier = Modifier.height(16.dp))

        OutlinedTextField(
            value = upiId,
            onValueChange = viewModel::onUpiIdChanged,
            label = { Text("UPI ID") },
            modifier = Modifier.fillMaxWidth()
        )
        Spacer(modifier = Modifier.height(32.dp))

        Button(
            onClick = viewModel::submitOnboarding,
            modifier = Modifier.fillMaxWidth(),
            enabled = uiState !is OnboardingUiState.Loading && name.isNotEmpty() && zone.isNotEmpty()
        ) {
            Text("Submit")
        }

        if (uiState is OnboardingUiState.Loading) {
            CircularProgressIndicator(modifier = Modifier.padding(16.dp))
        }

        if (uiState is OnboardingUiState.Error) {
            Text(
                text = (uiState as OnboardingUiState.Error).message,
                color = MaterialTheme.colorScheme.error,
                modifier = Modifier.padding(16.dp)
            )
        }
    }
}
