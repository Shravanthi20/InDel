package com.imaginai.indel.ui.policy

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.Shield
import androidx.compose.material3.*
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.data.model.Policy
import com.imaginai.indel.ui.navigation.Screen
import com.imaginai.indel.ui.theme.*
import java.text.SimpleDateFormat
import java.util.Locale
import java.util.TimeZone

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PolicyScreen(
    navController: NavController,
    viewModel: PolicyViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val isRefreshing by viewModel.isRefreshing.collectAsState()
    var showUpdateConfirm by remember { mutableStateOf(false) }
    var showStopConfirm by remember { mutableStateOf(false) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Protection Policy", fontWeight = FontWeight.Bold) },
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
        PullToRefreshBox(
            isRefreshing = isRefreshing,
            onRefresh = { viewModel.refresh() },
            modifier = Modifier.padding(padding)
        ) {
            Box(modifier = Modifier
                .fillMaxSize()
                .background(BackgroundWarmWhite)
            ) {
                when (val state = uiState) {
                    is PolicyUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                    is PolicyUiState.Success -> PolicyContent(
                        policy = state.policy,
                        navController = navController,
                        viewModel = viewModel,
                        onUpdatePlan = { showUpdateConfirm = true },
                        onStopPlan = { showStopConfirm = true }
                    )
                    is PolicyUiState.Error -> ErrorState(state.message) { viewModel.loadPolicy() }
                }
            }
        }
    }

    if (showUpdateConfirm) {
        AlertDialog(
            onDismissRequest = { showUpdateConfirm = false },
            title = { Text("Update plan?") },
            text = {
                Text(
                    "You can switch to a higher payout plan or a lower payout plan from the next screen. Upgrades add ₹5, downgrades are free."
                )
            },
            confirmButton = {
                TextButton(onClick = {
                    showUpdateConfirm = false
                    navController.navigate(Screen.PlanSelection.route)
                }) {
                    Text("Continue")
                }
            },
            dismissButton = {
                TextButton(onClick = { showUpdateConfirm = false }) {
                    Text("Cancel")
                }
            }
        )
    }

    if (showStopConfirm) {
        AlertDialog(
            onDismissRequest = { showStopConfirm = false },
            title = { Text("Stop protection?") },
            text = {
                Text(
                    "This will cancel the current policy. You can enroll again later if needed."
                )
            },
            confirmButton = {
                TextButton(onClick = {
                    showStopConfirm = false
                    viewModel.cancel()
                }) {
                    Text("Stop plan")
                }
            },
            dismissButton = {
                TextButton(onClick = { showStopConfirm = false }) {
                    Text("Keep plan")
                }
            }
        )
    }
}

@Composable
fun PolicyContent(
    policy: Policy,
    navController: NavController,
    viewModel: PolicyViewModel,
    onUpdatePlan: () -> Unit,
    onStopPlan: () -> Unit
) {
    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        // 1. Current Plan Card
        item {
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(containerColor = Color.White),
                elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
            ) {
                Column(modifier = Modifier.padding(20.dp)) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Icon(
                            Icons.Default.Shield, 
                            contentDescription = null, 
                            tint = when(policy.status) {
                                "active" -> SuccessGreen
                                "paused" -> WarningAmber
                                else -> ErrorRed
                            }
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = when(policy.status) {
                                "active" -> "ACTIVE PROTECTION"
                                "paused" -> "PAUSED PROTECTION"
                                "cancelled" -> "CANCELLED"
                                else -> "INACTIVE"
                            },
                            fontWeight = FontWeight.Bold,
                            color = when(policy.status) {
                                "active" -> SuccessGreen
                                "paused" -> WarningAmber
                                else -> ErrorRed
                            }
                        )
                    }
                    
                    Spacer(modifier = Modifier.height(16.dp))
                    
                    Text("Weekly Premium", style = MaterialTheme.typography.labelMedium, color = TextSecondary)
                    Text("₹${policy.weeklyPremiumInr}", style = MaterialTheme.typography.headlineMedium, fontWeight = FontWeight.Bold, color = BrandBlue)
                    
                    Spacer(modifier = Modifier.height(12.dp))
                    
                    Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                        Column {
                            Text("Coverage Ratio", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
							Text("${(policy.coverageRatio * 100).toInt()}%", fontWeight = FontWeight.Bold)
                        }
                        Column(horizontalAlignment = Alignment.End) {
                            Text("Next Due Date", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                            Text(policy.nextDueDate, fontWeight = FontWeight.Bold)
                        }
                    }

					Spacer(modifier = Modifier.height(12.dp))
					Text("Payment Status", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
					Text(policy.paymentStatus ?: "Locked", fontWeight = FontWeight.Bold)
                    val lastPaid = formatPolicyTimestamp(policy.lastPaymentTimestamp)
                    if (lastPaid.isNotBlank()) {
                        Text(
                            "Last payment: $lastPaid",
                            style = MaterialTheme.typography.bodySmall,
                            color = TextSecondary
                        )
                    }
					if ((policy.requiredPaymentInr ?: 0) > 0) {
						Text(
							"Current payable: ₹${policy.requiredPaymentInr}",
							style = MaterialTheme.typography.bodySmall,
							color = BrandBlue
						)
					}
                }
            }
        }

        // 2. Risk Factors (SHAP chips)
        item {
            Text("Why this premium?", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(8.dp))
            FlowRow(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                policy.shapBreakdown.forEach { chip ->
                    SuggestionChip(
                        onClick = { },
                        label = { Text(chip.feature.replace("_", " "), fontSize = 12.sp) },
                        colors = SuggestionChipDefaults.suggestionChipColors(
                            containerColor = if (chip.impact > 0) ErrorRed.copy(alpha = 0.1f) else SuccessGreen.copy(alpha = 0.1f),
                            labelColor = if (chip.impact > 0) ErrorRed else SuccessGreen
                        ),
                        border = null
                    )
                }
            }
        }

        // 3. Actions
        item {
            Spacer(modifier = Modifier.height(16.dp))
            if (policy.status == "active") {
                Button(
                    onClick = { navController.navigate(Screen.PremiumPay.route) },
                    modifier = Modifier.fillMaxWidth().height(56.dp),
                    shape = RoundedCornerShape(12.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = BrandBlue),
                    enabled = policy.nextPaymentEnabled == true
                ) {
                    Text("Pay Weekly Premium", fontWeight = FontWeight.Bold)
                }

                if (policy.nextPaymentEnabled != true) {
                    Spacer(modifier = Modifier.height(8.dp))
                    Text(
                        "Payment unlocks once the weekly cycle completes.",
                        style = MaterialTheme.typography.bodySmall,
                        color = TextSecondary
                    )
                }

                Spacer(modifier = Modifier.height(12.dp))

                Button(
                    onClick = onUpdatePlan,
                    modifier = Modifier.fillMaxWidth().height(56.dp),
                    shape = RoundedCornerShape(12.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = BrandBlue)
                ) {
                    Text("Update Plan", fontWeight = FontWeight.Bold)
                }
                
                Spacer(modifier = Modifier.height(12.dp))
                
                Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                    OutlinedButton(
                        onClick = onStopPlan,
                        modifier = Modifier.weight(1f).height(48.dp),
                        shape = RoundedCornerShape(12.dp),
                        colors = ButtonDefaults.outlinedButtonColors(contentColor = ErrorRed)
                    ) {
                        Text("Stop Plan")
                    }
                }
            } else {
                Button(
                    onClick = { viewModel.enroll() },
                    modifier = Modifier.fillMaxWidth().height(56.dp),
                    shape = RoundedCornerShape(12.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = SuccessGreen)
                ) {
                    Text(if (policy.status == "paused") "Resume Protection" else "Enrol Now", fontWeight = FontWeight.Bold)
                }
            }
        }
        
        // 4. Protection Note
        item {
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(12.dp),
                colors = CardDefaults.cardColors(containerColor = BlueSoft)
            ) {
                Row(modifier = Modifier.padding(16.dp)) {
                    Icon(Icons.Default.Info, contentDescription = null, tint = BrandBlue)
                    Spacer(modifier = Modifier.width(12.dp))
                    Column {
                        Text(
                            "Protection covers income loss during heavy rain, heatwaves, and local disruptions. Maximum payout is capped at 3x your weekly baseline.",
                            style = MaterialTheme.typography.bodySmall,
                            color = TextPrimary
                        )
                        Spacer(modifier = Modifier.height(6.dp))
                        Text(
                            "Plan changes are simple: higher payout tiers add ₹5, lower tiers do not add any extra fee.",
                            style = MaterialTheme.typography.bodySmall,
                            color = TextSecondary
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            "First activation payment is ${policy.initialPaymentMultiplier ?: 2}x weekly premium. Regular cycle is every ${policy.billingCycleDays ?: 7} days, with ${policy.gracePeriodDays ?: 2} grace days and ₹1/day late fee.",
                            style = MaterialTheme.typography.bodySmall,
                            color = TextSecondary
                        )
                        if ((policy.graceDaysRemaining ?: 0) > 0 && (policy.lateFeeInr ?: 0) > 0) {
                            Text(
                                "Grace in progress: ${policy.graceDaysRemaining} day(s) left, late fee ₹${policy.lateFeeInr}.",
                                style = MaterialTheme.typography.bodySmall,
                                color = WarningAmber
                            )
                        }
                    }
                }
            }
        }
    }
}

@OptIn(ExperimentalLayoutApi::class)
@Composable
fun FlowRow(
    modifier: Modifier = Modifier,
    horizontalArrangement: Arrangement.Horizontal = Arrangement.Start,
    content: @Composable () -> Unit
) {
    androidx.compose.foundation.layout.FlowRow(
        modifier = modifier,
        horizontalArrangement = horizontalArrangement,
        content = { content() }
    )
}

@Composable
fun ErrorState(message: String, onRetry: () -> Unit) {
    Column(
        modifier = Modifier.fillMaxSize(),
        verticalArrangement = Arrangement.Center,
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Text(message, color = ErrorRed)
        Button(onClick = onRetry, modifier = Modifier.padding(top = 16.dp)) {
            Text("Retry")
        }
    }
}

private fun formatPolicyTimestamp(raw: String?): String {
    if (raw.isNullOrBlank()) return ""

    val inputs = listOf(
        "yyyy-MM-dd'T'HH:mm:ssXXX",
        "yyyy-MM-dd'T'HH:mm:ss'Z'",
        "yyyy-MM-dd HH:mm:ss",
    )
    for (pattern in inputs) {
        try {
            val parser = SimpleDateFormat(pattern, Locale.US)
            parser.timeZone = TimeZone.getTimeZone("UTC")
            val parsed = parser.parse(raw)
            if (parsed != null) {
                val out = SimpleDateFormat("dd MMM yyyy, hh:mm a", Locale.US)
                return out.format(parsed)
            }
        } catch (_: Exception) {
        }
    }

    return raw
}
