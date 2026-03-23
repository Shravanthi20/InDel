package com.imaginai.indel.ui.auth

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.ui.navigation.Screen

@Composable
fun OtpScreen(
    navController: NavController,
    viewModel: OtpViewModel = hiltViewModel()
) {
    val phone by viewModel.phone.collectAsState()
    val otp by viewModel.otp.collectAsState()
    val uiState by viewModel.uiState.collectAsState()

    LaunchedEffect(uiState) {
        if (uiState is OtpUiState.Success) {
            val hasProfile = (uiState as OtpUiState.Success).hasProfile
            if (hasProfile) {
                navController.navigate(Screen.Home.route) {
                    popUpTo(Screen.OTP.route) { inclusive = true }
                }
            } else {
                navController.navigate(Screen.Onboarding.route) {
                    popUpTo(Screen.OTP.route) { inclusive = true }
                }
            }
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Text(text = "Worker Login", style = MaterialTheme.typography.headlineMedium)
        Spacer(modifier = Modifier.height(32.dp))

        OutlinedTextField(
            value = phone,
            onValueChange = viewModel::onPhoneChanged,
            label = { Text("Phone Number") },
            modifier = Modifier.fillMaxWidth()
        )
        Spacer(modifier = Modifier.height(16.dp))

        Button(
            onClick = viewModel::sendOtp,
            modifier = Modifier.fillMaxWidth(),
            enabled = uiState !is OtpUiState.Loading
        ) {
            Text("Send OTP")
        }

        Spacer(modifier = Modifier.height(32.dp))

        OutlinedTextField(
            value = otp,
            onValueChange = viewModel::onOtpChanged,
            label = { Text("OTP") },
            modifier = Modifier.fillMaxWidth()
        )
        Spacer(modifier = Modifier.height(16.dp))

        Button(
            onClick = viewModel::verifyOtp,
            modifier = Modifier.fillMaxWidth(),
            enabled = uiState !is OtpUiState.Loading && otp.isNotEmpty()
        ) {
            Text("Verify OTP")
        }

        if (uiState is OtpUiState.Loading) {
            CircularProgressIndicator(modifier = Modifier.padding(16.dp))
        }

        if (uiState is OtpUiState.OtpSent) {
            Text(
                text = "OTP sent! (Test OTP: ${(uiState as OtpUiState.OtpSent).testOtp})",
                color = MaterialTheme.colorScheme.primary,
                modifier = Modifier.padding(16.dp)
            )
        }

        if (uiState is OtpUiState.Error) {
            Text(
                text = (uiState as OtpUiState.Error).message,
                color = MaterialTheme.colorScheme.error,
                modifier = Modifier.padding(16.dp)
            )
        }
    }
}
