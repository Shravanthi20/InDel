package com.imaginai.indel.ui.auth

import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.ui.navigation.Screen
import androidx.compose.foundation.text.KeyboardOptions

@OptIn(ExperimentalMaterial3Api::class)
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

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Text(
                text = "Worker Login",
                style = MaterialTheme.typography.headlineLarge,
                color = MaterialTheme.colorScheme.primary,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = "Secure your income and work",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.secondary
            )
            
            Spacer(modifier = Modifier.height(48.dp))

            OutlinedTextField(
                value = phone,
                onValueChange = viewModel::onPhoneChanged,
                label = { Text("Phone Number") },
                modifier = Modifier.fillMaxWidth(),
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Phone),
                shape = RoundedCornerShape(12.dp)
            )
            Spacer(modifier = Modifier.height(16.dp))

            Button(
                onClick = viewModel::sendOtp,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp),
                enabled = uiState !is OtpUiState.Loading,
                shape = RoundedCornerShape(12.dp)
            ) {
                if (uiState is OtpUiState.Loading) {
                    CircularProgressIndicator(color = Color.White, modifier = Modifier.size(24.dp))
                } else {
                    Text("Send OTP", fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                }
            }

            if (uiState is OtpUiState.OtpSent) {
                Spacer(modifier = Modifier.height(32.dp))
                
                OutlinedTextField(
                    value = otp,
                    onValueChange = viewModel::onOtpChanged,
                    label = { Text("Enter OTP") },
                    modifier = Modifier.fillMaxWidth(),
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                    shape = RoundedCornerShape(12.dp)
                )
                
                Text(
                    text = "Test OTP: ${(uiState as OtpUiState.OtpSent).testOtp}",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.primary,
                    modifier = Modifier.padding(top = 4.dp, start = 4.dp).align(Alignment.Start)
                )

                Spacer(modifier = Modifier.height(16.dp))

                Button(
                    onClick = viewModel::verifyOtp,
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(56.dp),
                    enabled = uiState !is OtpUiState.Loading && otp.isNotEmpty(),
                    shape = RoundedCornerShape(12.dp)
                ) {
                    Text("Verify & Continue", fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                }
            }

            if (uiState is OtpUiState.Error) {
                Text(
                    text = (uiState as OtpUiState.Error).message,
                    color = MaterialTheme.colorScheme.error,
                    style = MaterialTheme.typography.bodySmall,
                    modifier = Modifier.padding(top = 16.dp)
                )
            }
        }
    }
}
