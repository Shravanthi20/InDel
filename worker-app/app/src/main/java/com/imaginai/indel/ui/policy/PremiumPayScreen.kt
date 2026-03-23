package com.imaginai.indel.ui.policy

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PremiumPayScreen(
    navController: NavController,
    viewModel: PremiumPayViewModel = hiltViewModel()
) {
    val amount by viewModel.amount.collectAsState()
    val uiState by viewModel.uiState.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Pay Premium") },
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
                .padding(padding)
                .fillMaxSize()
                .padding(16.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Text("Enter amount to pay (Optional)", style = MaterialTheme.typography.bodyLarge)
            Spacer(modifier = Modifier.height(16.dp))

            OutlinedTextField(
                value = amount,
                onValueChange = viewModel::onAmountChanged,
                label = { Text("Amount (INR)") },
                modifier = Modifier.fillMaxWidth()
            )

            Spacer(modifier = Modifier.height(24.dp))

            Button(
                onClick = { viewModel.pay() },
                modifier = Modifier.fillMaxWidth(),
                enabled = uiState !is PayUiState.Loading
            ) {
                Text("Pay Now")
            }

            if (uiState is PayUiState.Loading) {
                CircularProgressIndicator(modifier = Modifier.padding(16.dp))
            }

            if (uiState is PayUiState.Success) {
                Text((uiState as PayUiState.Success).message, color = Color(0xFF2E7D32), modifier = Modifier.padding(16.dp))
                LaunchedEffect(Unit) {
                    kotlinx.coroutines.delay(1500)
                    navController.navigateUp()
                }
            }

            if (uiState is PayUiState.Error) {
                Text((uiState as PayUiState.Error).message, color = Color.Red, modifier = Modifier.padding(16.dp))
            }
        }
    }
}
