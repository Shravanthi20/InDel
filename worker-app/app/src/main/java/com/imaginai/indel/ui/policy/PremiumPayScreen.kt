package com.imaginai.indel.ui.policy

import android.content.Context
import android.content.ContextWrapper
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Payment
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.ui.theme.BackgroundWarmWhite
import com.imaginai.indel.ui.theme.BlueDeep
import com.imaginai.indel.ui.theme.BlueSoft
import com.imaginai.indel.ui.theme.BrandBlue
import com.imaginai.indel.ui.theme.ErrorRed
import com.imaginai.indel.ui.theme.SuccessGreen
import com.imaginai.indel.ui.theme.TextSecondary
import com.imaginai.indel.ui.theme.WarningAmber

private tailrec fun Context.findMainActivity(): com.imaginai.indel.MainActivity? {
    return when (this) {
        is com.imaginai.indel.MainActivity -> this
        is ContextWrapper -> baseContext.findMainActivity()
        else -> null
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PremiumPayScreen(
    navController: NavController,
    viewModel: PremiumPayViewModel = hiltViewModel()
) {
    val amount by viewModel.amount.collectAsState()
    val basePremium by viewModel.basePremium.collectAsState()
    val lateFee by viewModel.lateFee.collectAsState()
    val paymentEnabled by viewModel.paymentEnabled.collectAsState()
    val paymentHint by viewModel.paymentHint.collectAsState()
    val uiState by viewModel.uiState.collectAsState()
    val context = LocalContext.current

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Pay Weekly Premium", fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = BrandBlue,
                    titleContentColor = Color.White,
                    navigationIconContentColor = Color.White
                )
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .padding(padding)
                .fillMaxSize()
                .background(BackgroundWarmWhite)
                .padding(20.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Spacer(modifier = Modifier.height(16.dp))

            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(20.dp),
                elevation = CardDefaults.cardElevation(defaultElevation = 6.dp)
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(Brush.linearGradient(listOf(BrandBlue, BlueDeep)))
                        .padding(28.dp),
                    contentAlignment = Alignment.Center
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(
                            "Weekly Premium Due",
                            style = MaterialTheme.typography.labelLarge,
                            color = Color.White.copy(alpha = 0.8f),
                            letterSpacing = 1.sp
                        )
                        Spacer(modifier = Modifier.height(12.dp))
                        Text(
                            "Rs ${amount.ifBlank { "--" }}",
                            fontSize = 52.sp,
                            fontWeight = FontWeight.ExtraBold,
                            color = Color.White
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        if (lateFee > 0) {
                            Surface(
                                color = WarningAmber.copy(alpha = 0.2f),
                                shape = RoundedCornerShape(10.dp)
                            ) {
                                Row(
                                    modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp),
                                    verticalAlignment = Alignment.CenterVertically
                                ) {
                                    Icon(
                                        Icons.Default.Warning,
                                        contentDescription = null,
                                        tint = WarningAmber,
                                        modifier = Modifier.size(14.dp)
                                    )
                                    Spacer(modifier = Modifier.width(6.dp))
                                    Text(
                                        "Rs $basePremium base + Rs $lateFee late fee",
                                        color = WarningAmber,
                                        fontSize = 12.sp,
                                        fontWeight = FontWeight.SemiBold
                                    )
                                }
                            }
                        } else {
                            Text(
                                "ML-computed premium for this cycle",
                                color = Color.White.copy(alpha = 0.7f),
                                fontSize = 12.sp,
                                textAlign = TextAlign.Center
                            )
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(20.dp))

            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(14.dp),
                colors = CardDefaults.cardColors(containerColor = BlueSoft)
            ) {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text(
                        "Payment is dynamically priced based on current weather risk, order volatility, and your performance baseline.",
                        fontSize = 13.sp,
                        color = BrandBlue,
                        lineHeight = 18.sp
                    )
                    if (lateFee > 0) {
                        Spacer(modifier = Modifier.height(8.dp))
                        HorizontalDivider(color = BrandBlue.copy(alpha = 0.2f))
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            "Pay now to avoid plan deactivation. Late fee accumulates at Rs 1/day during grace period.",
                            fontSize = 12.sp,
                            color = WarningAmber,
                            fontWeight = FontWeight.Medium
                        )
                    }
                }
            }

            if (!paymentHint.isNullOrBlank()) {
                Spacer(modifier = Modifier.height(16.dp))
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(14.dp),
                    colors = CardDefaults.cardColors(containerColor = WarningAmber.copy(alpha = 0.12f))
                ) {
                    Text(
                        text = paymentHint.orEmpty(),
                        modifier = Modifier.padding(16.dp),
                        color = BrandBlue,
                        fontSize = 13.sp
                    )
                }
            }

            Spacer(modifier = Modifier.weight(1f))

            Button(
                onClick = {
                    val amountInPaise = (amount.toIntOrNull() ?: 0) * 100
                    if (amountInPaise <= 0) {
                        viewModel.setPaymentError("Premium amount not yet loaded. Please wait.")
                        return@Button
                    }
                    val mainActivity = context.findMainActivity()
                    if (mainActivity == null) {
                        viewModel.setPaymentError("Unable to open payment gateway")
                        return@Button
                    }
                    viewModel.setLoading(true)
                    mainActivity.startRazorpayCheckout(amountInPaise, "9876543210") { success, paymentId, error ->
                        if (success) {
                            viewModel.recordPaymentSuccess(paymentId)
                        } else {
                            viewModel.setPaymentError(error ?: "Payment failed or cancelled")
                        }
                    }
                },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(58.dp),
                shape = RoundedCornerShape(14.dp),
                colors = ButtonDefaults.buttonColors(containerColor = BrandBlue),
                enabled = paymentEnabled && uiState !is PayUiState.Loading && amount.isNotBlank() && (amount.toIntOrNull() ?: 0) > 0,
                elevation = ButtonDefaults.buttonElevation(defaultElevation = 4.dp)
            ) {
                Icon(Icons.Default.Payment, contentDescription = null, modifier = Modifier.size(20.dp))
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    "Touch Pay  Rs ${amount.ifBlank { "--" }}",
                    fontWeight = FontWeight.Bold,
                    fontSize = 17.sp
                )
            }

            Spacer(modifier = Modifier.height(16.dp))

            when (val state = uiState) {
                is PayUiState.Loading -> {
                    CircularProgressIndicator(color = BrandBlue, modifier = Modifier.padding(8.dp))
                    Text("Processing payment...", fontSize = 13.sp, color = TextSecondary)
                }
                is PayUiState.Success -> {
                    Text(
                        state.message,
                        color = SuccessGreen,
                        fontWeight = FontWeight.SemiBold,
                        textAlign = TextAlign.Center,
                        modifier = Modifier.padding(8.dp)
                    )
                    LaunchedEffect(state) {
                        kotlinx.coroutines.delay(1800)
                        viewModel.reset()
                        navController.navigateUp()
                    }
                }
                is PayUiState.Error -> {
                    Text(
                        state.message,
                        color = ErrorRed,
                        textAlign = TextAlign.Center,
                        modifier = Modifier.padding(8.dp)
                    )
                }
                else -> Unit
            }

            Spacer(modifier = Modifier.height(8.dp))
        }
    }
}
