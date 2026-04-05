package com.imaginai.indel.ui.policy

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.ui.theme.BrandBlue
import com.imaginai.indel.ui.theme.TextSecondary

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PremiumPayScreen(
    navController: NavController,
    viewModel: PremiumPayViewModel = hiltViewModel()
) {
    val amount by viewModel.amount.collectAsState()
    val uiState by viewModel.uiState.collectAsState()
    val context = LocalContext.current

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
                .padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(containerColor = Color.White),
                elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
            ) {
                Column(
                    modifier = Modifier.padding(24.dp),
                    horizontalAlignment = Alignment.CenterHorizontally
                ) {
                    Text("Weekly Premium Due", style = MaterialTheme.typography.labelLarge, color = TextSecondary)
                    Spacer(modifier = Modifier.height(8.dp))
                    Text(
                        "₹$amount",
                        style = MaterialTheme.typography.displayMedium,
                        fontWeight = FontWeight.Bold,
                        color = BrandBlue
                    )
                    Spacer(modifier = Modifier.height(16.dp))
                    Text(
                        "This amount is calculated based on current environmental risks and your performance baseline.",
                        textAlign = TextAlign.Center,
                        style = MaterialTheme.typography.bodySmall,
                        color = TextSecondary
                    )
                }
            }

            Spacer(modifier = Modifier.height(48.dp))

            Button(
                onClick = { 
                    val mainActivity = context as? com.imaginai.indel.MainActivity
                    val amountInPaise = (amount.toIntOrNull() ?: 0) * 100
                    viewModel.setLoading(true)
                    mainActivity?.startRazorpayCheckout(amountInPaise, "9876543210") { success, paymentId, error ->
                        if (success) {
                            viewModel.recordPaymentSuccess(paymentId)
                        } else {
                            viewModel.setPaymentError(error ?: "Payment Failed or Cancelled")
                        }
                    }
                },
                modifier = Modifier.fillMaxWidth().height(56.dp),
                shape = RoundedCornerShape(12.dp),
                colors = ButtonDefaults.buttonColors(containerColor = BrandBlue),
                enabled = uiState !is PayUiState.Loading && amount != ""
            ) {
                Text("Pay Now", fontWeight = FontWeight.Bold, fontSize = 16.sp)
            }

            if (uiState is PayUiState.Loading) {
                CircularProgressIndicator(modifier = Modifier.padding(16.dp))
            }

            if (uiState is PayUiState.Success) {
                Text((uiState as PayUiState.Success).message, color = Color(0xFF2E7D32), modifier = Modifier.padding(16.dp))
                LaunchedEffect(uiState) {
                    kotlinx.coroutines.delay(1500)
                    viewModel.reset()
                    navController.navigateUp()
                }
            }

            if (uiState is PayUiState.Error) {
                Text((uiState as PayUiState.Error).message, color = Color.Red, modifier = Modifier.padding(16.dp))
            }
        }
    }
}
