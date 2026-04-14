package com.imaginai.indel.ui.policy

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.slideInVertically
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.data.model.Policy
import com.imaginai.indel.data.model.ShapImpact
import com.imaginai.indel.ui.navigation.Screen
import com.imaginai.indel.ui.theme.*
import java.text.SimpleDateFormat
import java.util.*

@OptIn(ExperimentalMaterial3Api::class, ExperimentalLayoutApi::class)
@Composable
fun PolicyScreen(
    navController: NavController,
    viewModel: PolicyViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val isRefreshing by viewModel.isRefreshing.collectAsState()
    val actionError by viewModel.actionError.collectAsState()
    var showStopConfirm by remember { mutableStateOf(false) }

    LaunchedEffect(Unit) {
        viewModel.loadPolicy()
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Premium Plan", fontWeight = FontWeight.Bold) },
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
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(BackgroundWarmWhite)
            ) {
                when (val state = uiState) {
                    is PolicyUiState.Loading -> {
                        Column(
                            modifier = Modifier.fillMaxSize(),
                            horizontalAlignment = Alignment.CenterHorizontally,
                            verticalArrangement = Arrangement.Center
                        ) {
                            CircularProgressIndicator(color = BrandBlue)
                            Spacer(modifier = Modifier.height(12.dp))
                            Text("Loading plan data...", color = TextSecondary, fontSize = 14.sp)
                        }
                    }
                    is PolicyUiState.Success -> PremiumPlanContent(
                        policy = state.policy,
                        navController = navController,
                        viewModel = viewModel,
                        onStopPlan = { showStopConfirm = true }
                    )
                    is PolicyUiState.PaymentSuccess -> {
                        PaymentSuccessBanner(
                            basePremium = state.basePremium,
                            lateFee = state.lateFee,
                            totalPaid = state.totalPaid
                        )
                        LaunchedEffect(Unit) {
                            kotlinx.coroutines.delay(2500)
                            viewModel.loadPolicy()
                        }
                    }
                    is PolicyUiState.PlanStopped -> {
                        EmptyPlanState(
                            message = "Plan stopped successfully.",
                            onRestart = { viewModel.loadPolicy() }
                        )
                    }
                    is PolicyUiState.Error -> ErrorState(state.message) { viewModel.loadPolicy() }
                }

                // Action-level error snackbar
                actionError?.let { err ->
                    LaunchedEffect(err) {
                        kotlinx.coroutines.delay(3000)
                        viewModel.clearActionError()
                    }
                    Snackbar(
                        modifier = Modifier
                            .align(Alignment.BottomCenter)
                            .padding(16.dp),
                        containerColor = ErrorRed,
                        contentColor = Color.White
                    ) { Text(err) }
                }
            }
        }
    }

    // ── Stop Plan Confirmation Dialog ──────────────────────────────────────
    if (showStopConfirm) {
        AlertDialog(
            onDismissRequest = { showStopConfirm = false },
            title = { Text("Stop Premium Plan?", fontWeight = FontWeight.Bold) },
            text = {
                Column {
                    Text("Are you sure you want to stop the plan?")
                    Spacer(modifier = Modifier.height(8.dp))
                    Text(
                        "Your coverage will end immediately. To restart, you'll need to pay the activation fee (2× weekly premium).",
                        style = MaterialTheme.typography.bodySmall,
                        color = TextSecondary
                    )
                }
            },
            confirmButton = {
                Button(
                    onClick = {
                        showStopConfirm = false
                        viewModel.stopPlan()
                    },
                    colors = ButtonDefaults.buttonColors(containerColor = ErrorRed)
                ) { Text("Yes, Stop Plan") }
            },
            dismissButton = {
                OutlinedButton(onClick = { showStopConfirm = false }) {
                    Text("Keep Plan")
                }
            }
        )
    }
}

// ── Main Content ───────────────────────────────────────────────────────────

@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun PremiumPlanContent(
    policy: Policy,
    navController: NavController,
    viewModel: PolicyViewModel,
    onStopPlan: () -> Unit
) {
    val isActive = policy.status == "active"
    val paymentEnabled = policy.nextPaymentEnabled ?: false
    val isDeactivated = policy.coverageStatus.equals("Deactivated", ignoreCase = true)
    val isGrace = policy.paymentStatus.equals("Eligible", ignoreCase = true) &&
            (policy.graceDaysRemaining ?: 999) < (policy.gracePeriodDays ?: 2)
    val lateFee = policy.lateFeeInr ?: 0
    val basePremium = policy.weeklyPremiumInr
    val totalDue = basePremium + lateFee

    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        // 1. STATUS HERO CARD
        item {
            AnimatedVisibility(
                visible = true,
                enter = fadeIn(tween(400)) + slideInVertically(tween(400)) { it / 2 }
            ) {
                PlanStatusCard(policy = policy, isActive = isActive, isDeactivated = isDeactivated)
            }
        }

        // 2. GRACE PERIOD WARNING
        if (isActive && lateFee > 0) {
            item {
                GracePeriodWarning(
                    daysRemaining = policy.graceDaysRemaining ?: 0,
                    lateFee = lateFee,
                    totalDue = totalDue
                )
            }
        }

        // 3. PAYMENT CYCLE PROGRESS (active only)
        if (isActive) {
            item {
                PaymentCycleCard(policy = policy)
            }
        }

        // 4. ML RISK / PREMIUM BREAKDOWN
        if (policy.shapBreakdown.isNotEmpty()) {
            item {
                RiskBreakdownCard(
                    shapBreakdown = policy.shapBreakdown,
                    riskScore = policy.riskScore,
                    pricingSource = policy.pricingSource
                )
            }
        }

        // 5. ACTION BUTTONS
        item {
            if (!isActive || isDeactivated) {
                // ── INACTIVE: single "Start Premium Plan" button ──────────
                StartPlanButton(
                    weeklyPremium = basePremium,
                    multiplier = policy.initialPaymentMultiplier ?: 2,
                    onClick = { viewModel.startPlanWithPayment(policy) }
                )
            } else {
                // ── ACTIVE: "Premium Payment" + "Stop Plan" ───────────────
                Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    // Premium Payment (locked/unlocked by billing cycle)
                    PremiumPaymentButton(
                        enabled = paymentEnabled,
                        basePremium = basePremium,
                        lateFee = lateFee,
                        totalDue = totalDue,
                        paymentStatus = policy.paymentStatus,
                        onClick = { navController.navigate(Screen.PremiumPay.route) }
                    )
                    // Stop Plan
                    OutlinedButton(
                        onClick = onStopPlan,
                        modifier = Modifier.fillMaxWidth().height(50.dp),
                        shape = RoundedCornerShape(12.dp),
                        colors = ButtonDefaults.outlinedButtonColors(contentColor = ErrorRed),
                        border = androidx.compose.foundation.BorderStroke(1.5.dp, ErrorRed)
                    ) {
                        Icon(Icons.Default.Close, contentDescription = null, modifier = Modifier.size(18.dp))
                        Spacer(modifier = Modifier.width(6.dp))
                        Text("Stop Plan", fontWeight = FontWeight.SemiBold)
                    }
                }
            }
        }

        // 6. PLAN RULES INFO
        item {
            PlanInfoCard(policy = policy)
        }

        item { Spacer(modifier = Modifier.height(24.dp)) }
    }
}

// ── Status Card ────────────────────────────────────────────────────────────

@Composable
private fun PlanStatusCard(policy: Policy, isActive: Boolean, isDeactivated: Boolean) {
    val gradientColors = when {
        isActive -> listOf(BrandBlue, BlueDeep)
        isDeactivated -> listOf(Color(0xFF6B7280), Color(0xFF374151))
        else -> listOf(Color(0xFF6B7280), Color(0xFF374151))
    }

    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(20.dp),
        elevation = CardDefaults.cardElevation(defaultElevation = 6.dp)
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .background(Brush.linearGradient(gradientColors))
                .padding(24.dp)
        ) {
            Column {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Box(
                        modifier = Modifier
                            .size(48.dp)
                            .clip(CircleShape)
                            .background(Color.White.copy(alpha = 0.15f)),
                        contentAlignment = Alignment.Center
                    ) {
                        Icon(
                            if (isActive) Icons.Default.Shield else Icons.Default.ShieldMoon,
                            contentDescription = null,
                            tint = Color.White,
                            modifier = Modifier.size(28.dp)
                        )
                    }
                    Spacer(modifier = Modifier.width(12.dp))
                    Column {
                        Text(
                            text = when {
                                isActive -> "ACTIVE PROTECTION"
                                isDeactivated -> "PROTECTION EXPIRED"
                                policy.status == "cancelled" -> "PLAN STOPPED"
                                else -> "NO ACTIVE PLAN"
                            },
                            color = Color.White,
                            fontWeight = FontWeight.ExtraBold,
                            fontSize = 14.sp,
                            letterSpacing = 1.sp
                        )
                        if (policy.zone.isNotBlank()) {
                            Text(
                                policy.zone,
                                color = Color.White.copy(alpha = 0.8f),
                                fontSize = 12.sp
                            )
                        }
                    }
                }

                Spacer(modifier = Modifier.height(20.dp))

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    StatItem("Weekly Premium", "₹${policy.weeklyPremiumInr}")
                    StatItem("Coverage", "${(policy.coverageRatio * 100).toInt()}%")
                    StatItem("Next Due", policy.nextDueDate.take(10))
                }

                if (isActive && policy.planName != null) {
                    Spacer(modifier = Modifier.height(16.dp))
                    Surface(
                        color = Color.White.copy(alpha = 0.15f),
                        shape = RoundedCornerShape(8.dp)
                    ) {
                        Row(
                            modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Icon(
                                Icons.Default.LocalOffer, contentDescription = null,
                                tint = Color.White, modifier = Modifier.size(14.dp)
                            )
                            Spacer(modifier = Modifier.width(6.dp))
                            Text(
                                "${policy.planName} Plan",
                                color = Color.White,
                                fontSize = 12.sp,
                                fontWeight = FontWeight.SemiBold
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun StatItem(label: String, value: String) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text(label, color = Color.White.copy(alpha = 0.7f), fontSize = 10.sp, letterSpacing = 0.5.sp)
        Spacer(modifier = Modifier.height(4.dp))
        Text(value, color = Color.White, fontWeight = FontWeight.Bold, fontSize = 15.sp)
    }
}

// ── Grace Period Warning ───────────────────────────────────────────────────

@Composable
private fun GracePeriodWarning(daysRemaining: Int, lateFee: Int, totalDue: Int) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = WarningAmber.copy(alpha = 0.12f)),
        border = androidx.compose.foundation.BorderStroke(1.5.dp, WarningAmber)
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(Icons.Default.Warning, contentDescription = null, tint = WarningAmber, modifier = Modifier.size(24.dp))
            Spacer(modifier = Modifier.width(12.dp))
            Column {
                Text("Grace Period Active", fontWeight = FontWeight.Bold, color = WarningAmber, fontSize = 14.sp)
                Text(
                    "$daysRemaining day${if (daysRemaining != 1) "s" else ""} remaining • Late fee: ₹$lateFee",
                    fontSize = 12.sp, color = WarningAmber
                )
                Text("Total due: ₹$totalDue", fontSize = 13.sp, fontWeight = FontWeight.SemiBold, color = WarningAmber)
            }
        }
    }
}

// ── Payment Cycle Card ──────────────────────────────────────────────────────

@Composable
private fun PaymentCycleCard(policy: Policy) {
    val days = policy.daysSinceLastPayment ?: 0
    val cycleDays = policy.billingCycleDays ?: 7
    val progress = (days.toFloat() / cycleDays.toFloat()).coerceIn(0f, 1f)
    val animatedProgress by animateFloatAsState(
        targetValue = progress,
        animationSpec = tween(durationMillis = 800),
        label = "cycle_progress"
    )

    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(20.dp)) {
            Text("Payment Cycle", fontWeight = FontWeight.Bold, fontSize = 15.sp)
            Spacer(modifier = Modifier.height(12.dp))
            LinearProgressIndicator(
                progress = { animatedProgress },
                modifier = Modifier.fillMaxWidth().height(8.dp).clip(RoundedCornerShape(4.dp)),
                color = when {
                    progress >= 1f -> WarningAmber
                    progress > 0.85f -> WarningAmber.copy(alpha = 0.7f)
                    else -> BrandBlue
                },
                trackColor = BlueSoft
            )
            Spacer(modifier = Modifier.height(10.dp))
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                Text(
                    "Day $days of $cycleDays",
                    fontSize = 12.sp, color = TextSecondary
                )
                val payStatus = policy.paymentStatus ?: "Locked"
                Surface(
                    shape = RoundedCornerShape(6.dp),
                    color = when (payStatus) {
                        "Eligible" -> SuccessGreen.copy(alpha = 0.15f)
                        "Deactivated" -> ErrorRed.copy(alpha = 0.15f)
                        else -> BlueSoft
                    }
                ) {
                    Text(
                        payStatus,
                        modifier = Modifier.padding(horizontal = 8.dp, vertical = 3.dp),
                        fontSize = 11.sp,
                        fontWeight = FontWeight.SemiBold,
                        color = when (payStatus) {
                            "Eligible" -> SuccessGreen
                            "Deactivated" -> ErrorRed
                            else -> BrandBlue
                        }
                    )
                }
            }
            val lastPaid = formatPolicyTimestamp(policy.lastPaymentTimestamp)
            if (lastPaid.isNotBlank()) {
                Spacer(modifier = Modifier.height(6.dp))
                Text("Last payment: $lastPaid", fontSize = 11.sp, color = TextSecondary)
            }
        }
    }
}

// ── Risk / ML Breakdown Card ───────────────────────────────────────────────

@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun RiskBreakdownCard(
    shapBreakdown: List<ShapImpact>,
    riskScore: Double?,
    pricingSource: String?
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(20.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text("Why this premium?", fontWeight = FontWeight.Bold, fontSize = 15.sp)
                riskScore?.let {
                    Surface(
                        shape = RoundedCornerShape(8.dp),
                        color = when {
                            it > 0.7 -> ErrorRed.copy(alpha = 0.12f)
                            it > 0.4 -> WarningAmber.copy(alpha = 0.12f)
                            else -> SuccessGreen.copy(alpha = 0.12f)
                        }
                    ) {
                        Text(
                            "Risk: ${(it * 100).toInt()}%",
                            modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp),
                            fontSize = 11.sp,
                            fontWeight = FontWeight.Bold,
                            color = when {
                                it > 0.7 -> ErrorRed
                                it > 0.4 -> WarningAmber
                                else -> SuccessGreen
                            }
                        )
                    }
                }
            }

            if (!pricingSource.isNullOrBlank() && pricingSource != "stored_policy") {
                Text(
                    "Source: ${pricingSource.replace("-", " ").replaceFirstChar { it.uppercase() }}",
                    fontSize = 11.sp, color = TextSecondary,
                    modifier = Modifier.padding(top = 2.dp)
                )
            }

            Spacer(modifier = Modifier.height(12.dp))
            FlowRow(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                shapBreakdown.forEach { chip ->
                    val isHigh = chip.impact > 0
                    Surface(
                        shape = RoundedCornerShape(20.dp),
                        color = if (isHigh) ErrorRed.copy(alpha = 0.1f) else SuccessGreen.copy(alpha = 0.1f),
                        border = androidx.compose.foundation.BorderStroke(
                            1.dp,
                            if (isHigh) ErrorRed.copy(alpha = 0.3f) else SuccessGreen.copy(alpha = 0.3f)
                        )
                    ) {
                        Row(
                            modifier = Modifier.padding(horizontal = 10.dp, vertical = 5.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Icon(
                                if (isHigh) Icons.Default.TrendingUp else Icons.Default.TrendingDown,
                                contentDescription = null,
                                tint = if (isHigh) ErrorRed else SuccessGreen,
                                modifier = Modifier.size(12.dp)
                            )
                            Spacer(modifier = Modifier.width(4.dp))
                            Text(
                                chip.feature.replace("_", " "),
                                fontSize = 11.sp,
                                color = if (isHigh) ErrorRed else SuccessGreen,
                                fontWeight = FontWeight.Medium
                            )
                        }
                    }
                }
            }
        }
    }
}

// ── Action Buttons ─────────────────────────────────────────────────────────

@Composable
private fun StartPlanButton(weeklyPremium: Int, multiplier: Int, onClick: () -> Unit) {
    val activationAmount = weeklyPremium * multiplier
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        if (weeklyPremium > 0) {
            Surface(
                shape = RoundedCornerShape(12.dp),
                color = BlueSoft,
                modifier = Modifier.fillMaxWidth()
            ) {
                Column(modifier = Modifier.padding(14.dp)) {
                    Text("First payment (activation)", fontSize = 12.sp, color = BrandBlue)
                    Text(
                        "₹$activationAmount = ${multiplier}× weekly premium of ₹$weeklyPremium",
                        fontWeight = FontWeight.Bold, color = BrandBlue, fontSize = 14.sp
                    )
                    Text(
                        "After activation, you pay ₹$weeklyPremium/week",
                        fontSize = 11.sp, color = TextSecondary
                    )
                }
            }
        }
        Button(
            onClick = onClick,
            modifier = Modifier.fillMaxWidth().height(56.dp),
            shape = RoundedCornerShape(14.dp),
            colors = ButtonDefaults.buttonColors(containerColor = BrandBlue),
            elevation = ButtonDefaults.buttonElevation(defaultElevation = 4.dp)
        ) {
            Icon(Icons.Default.PlayArrow, contentDescription = null, modifier = Modifier.size(20.dp))
            Spacer(modifier = Modifier.width(8.dp))
            Text(
                "Start Premium Plan",
                fontWeight = FontWeight.Bold,
                fontSize = 16.sp
            )
        }
    }
}

@Composable
private fun PremiumPaymentButton(
    enabled: Boolean,
    basePremium: Int,
    lateFee: Int,
    totalDue: Int,
    paymentStatus: String?,
    onClick: () -> Unit
) {
    Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
        Button(
            onClick = onClick,
            enabled = enabled,
            modifier = Modifier.fillMaxWidth().height(56.dp),
            shape = RoundedCornerShape(14.dp),
            colors = ButtonDefaults.buttonColors(
                containerColor = SuccessGreen,
                disabledContainerColor = SuccessGreen.copy(alpha = 0.35f)
            ),
            elevation = ButtonDefaults.buttonElevation(defaultElevation = 4.dp)
        ) {
            Icon(Icons.Default.Payment, contentDescription = null, modifier = Modifier.size(20.dp))
            Spacer(modifier = Modifier.width(8.dp))
            Text(
                if (enabled) "Premium Payment  ₹$totalDue" else "Premium Payment (Locked)",
                fontWeight = FontWeight.Bold, fontSize = 15.sp
            )
        }
        if (!enabled) {
            Text(
                when (paymentStatus) {
                    "Locked" -> "Next payment unlocks after the 7-day billing cycle"
                    "Deactivated" -> "Plan deactivated — start a new plan to resume coverage"
                    else -> "Payment window closed"
                },
                fontSize = 11.sp,
                color = TextSecondary,
                modifier = Modifier.padding(horizontal = 4.dp)
            )
        } else if (lateFee > 0) {
            Text(
                "Includes ₹$lateFee late fee (₹$basePremium base + ₹$lateFee fee)",
                fontSize = 11.sp,
                color = WarningAmber,
                fontWeight = FontWeight.Medium,
                modifier = Modifier.padding(horizontal = 4.dp)
            )
        }
    }
}

// ── Plan Info Card ─────────────────────────────────────────────────────────

@Composable
private fun PlanInfoCard(policy: Policy) {
    val planInfo = policy.planInfo
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = BlueSoft),
        elevation = CardDefaults.cardElevation(defaultElevation = 0.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(Icons.Default.Info, contentDescription = null, tint = BrandBlue, modifier = Modifier.size(16.dp))
                Spacer(modifier = Modifier.width(6.dp))
                Text("Plan Rules", fontWeight = FontWeight.Bold, fontSize = 13.sp, color = BrandBlue)
            }
            Spacer(modifier = Modifier.height(8.dp))
            PlanRuleRow("Billing cycle", "${planInfo?.weeklyCycleDays ?: 7} days")
            PlanRuleRow("Grace period", "${planInfo?.gracePeriodDays ?: policy.gracePeriodDays ?: 2} days")
            PlanRuleRow("Late fee", planInfo?.lateFeeRule?.replace("_", " ") ?: "₹1 per day during grace")
            PlanRuleRow("First payment", "2× weekly premium (activation)")
            PlanRuleRow("After grace expires", "Plan deactivated — restart required")
        }
    }
}

@Composable
private fun PlanRuleRow(label: String, value: String) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 3.dp),
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Text(label, fontSize = 12.sp, color = TextSecondary)
        Text(value, fontSize = 12.sp, fontWeight = FontWeight.Medium, color = BrandBlue)
    }
}

// ── Helper Composables ─────────────────────────────────────────────────────

@Composable
private fun PaymentSuccessBanner(basePremium: Int, lateFee: Int, totalPaid: Int) {
    Box(
        modifier = Modifier.fillMaxSize().background(BackgroundWarmWhite),
        contentAlignment = Alignment.Center
    ) {
        Card(
            modifier = Modifier.padding(32.dp).fillMaxWidth(),
            shape = RoundedCornerShape(24.dp),
            colors = CardDefaults.cardColors(containerColor = Color.White),
            elevation = CardDefaults.cardElevation(defaultElevation = 8.dp)
        ) {
            Column(
                modifier = Modifier.padding(32.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                Icon(Icons.Default.CheckCircle, contentDescription = null, tint = SuccessGreen, modifier = Modifier.size(64.dp))
                Spacer(modifier = Modifier.height(16.dp))
                Text("Payment Successful!", fontWeight = FontWeight.ExtraBold, fontSize = 20.sp)
                Spacer(modifier = Modifier.height(12.dp))
                Text("₹$totalPaid paid", fontSize = 24.sp, fontWeight = FontWeight.Bold, color = SuccessGreen)
                if (lateFee > 0) {
                    Text("₹$basePremium premium + ₹$lateFee late fee", fontSize = 13.sp, color = TextSecondary)
                }
                Spacer(modifier = Modifier.height(8.dp))
                Text("Your coverage cycle has been renewed.", fontSize = 13.sp, color = TextSecondary)
            }
        }
    }
}

@Composable
private fun EmptyPlanState(message: String, onRestart: () -> Unit) {
    Box(
        modifier = Modifier.fillMaxSize().background(BackgroundWarmWhite),
        contentAlignment = Alignment.Center
    ) {
        Column(
            modifier = Modifier.padding(32.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Icon(Icons.Default.ShieldMoon, contentDescription = null, tint = TextSecondary, modifier = Modifier.size(72.dp))
            Spacer(modifier = Modifier.height(16.dp))
            Text(message, color = TextSecondary)
            Spacer(modifier = Modifier.height(24.dp))
            Button(onClick = onRestart, colors = ButtonDefaults.buttonColors(containerColor = BrandBlue)) {
                Text("Reload")
            }
        }
    }
}

@Composable
fun ErrorState(message: String, onRetry: () -> Unit) {
    Column(
        modifier = Modifier.fillMaxSize().padding(16.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Icon(Icons.Default.ErrorOutline, contentDescription = null, modifier = Modifier.size(64.dp), tint = ErrorRed)
        Spacer(modifier = Modifier.height(16.dp))
        Text(message, color = TextSecondary, fontSize = 14.sp)
        Spacer(modifier = Modifier.height(16.dp))
        Button(onClick = onRetry, colors = ButtonDefaults.buttonColors(containerColor = BrandBlue)) {
            Text("Retry")
        }
    }
}

fun formatPolicyTimestamp(timestamp: String?): String {
    if (timestamp == null) return ""
    val formats = listOf(
        "yyyy-MM-dd'T'HH:mm:ss.SSSSSS",
        "yyyy-MM-dd'T'HH:mm:ssXXX",
        "yyyy-MM-dd'T'HH:mm:ss'Z'"
    )
    val out = SimpleDateFormat("dd MMM, hh:mm a", Locale.getDefault())
    for (fmt in formats) {
        try {
            val sdf = SimpleDateFormat(fmt, Locale.getDefault())
            sdf.timeZone = TimeZone.getTimeZone("UTC")
            return out.format(sdf.parse(timestamp)!!)
        } catch (_: Exception) { }
    }
    return timestamp.take(16)
}
