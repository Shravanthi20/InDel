package com.imaginai.indel.ui.plan

import android.content.Context
import android.content.ContextWrapper
import android.widget.Toast
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.LocalOffer
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.data.model.DeliveryPlan
import com.imaginai.indel.data.model.Policy
import com.imaginai.indel.ui.navigation.Screen
import com.imaginai.indel.ui.theme.*

private tailrec fun Context.findMainActivity(): com.imaginai.indel.MainActivity? {
    return when (this) {
        is com.imaginai.indel.MainActivity -> this
        is ContextWrapper -> baseContext.findMainActivity()
        else -> null
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PlanSelectionScreen(
    navController: NavController,
    viewModel: PlanSelectionViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val selectedPlan by viewModel.selectedPlan.collectAsState()
    val selectedExpectedDeliveries by viewModel.selectedExpectedDeliveries.collectAsState()
    val isPaymentRequired by viewModel.isPaymentRequired.collectAsState()
    val currentPolicy by viewModel.currentPolicy.collectAsState()
    val context = LocalContext.current

    LaunchedEffect(uiState) {
        if (uiState is PlanUiState.Skipped) {
            navController.navigate(Screen.Landing.route) {
                popUpTo(Screen.PlanSelection.route) { inclusive = true }
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Choose Your Plan", fontWeight = FontWeight.Bold) },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = BrandBlue,
                    titleContentColor = Color.White
                )
            )
        }
    ) { padding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .background(BackgroundWarmWhite)
        ) {
            when (val state = uiState) {
                is PlanUiState.Loading -> {
                    CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                }
                is PlanUiState.Success -> {
                    PlanContent(
                        plans = state.plans,
                        currentPolicy = currentPolicy,
                        selectedPlan = selectedPlan,
                        selectedExpectedDeliveries = selectedExpectedDeliveries,
                        isPaymentRequired = isPaymentRequired,
                        onPlanSelected = { viewModel.selectPlan(it) },
                        onExpectedDeliveriesSelected = { viewModel.selectExpectedDeliveries(it) },
                        premiumForSelection = { plan, deliveries -> viewModel.calculatePremium(plan, deliveries) },
                        upgradeFeeForSelection = { plan -> viewModel.calculateUpgradeFee(plan) },
                        onConfirm = { totalPayment ->
                            val amountInPaise = totalPayment * 100
                            if (amountInPaise <= 0) {
                                Toast.makeText(context, "Invalid payment amount", Toast.LENGTH_SHORT).show()
                                return@PlanContent
                            }

                            val mainActivity = context.findMainActivity()
                            if (mainActivity == null) {
                                Toast.makeText(context, "Unable to open payment gateway", Toast.LENGTH_SHORT).show()
                                return@PlanContent
                            }

                            mainActivity.startRazorpayCheckout(amountInPaise, "9876543210") { success, _, error ->
                                if (success) {
                                    viewModel.confirmSelection()
                                } else {
                                    Toast.makeText(context, error ?: "Payment failed or cancelled", Toast.LENGTH_SHORT).show()
                                }
                            }
                        },
                        onSkip = { viewModel.skipPlan() }
                    )
                }
                is PlanUiState.SelectionComplete -> {
                    PlanContent(
                        plans = state.plans,
                        currentPolicy = currentPolicy,
                        selectedPlan = state.selectedPlan,
                        selectedExpectedDeliveries = selectedExpectedDeliveries,
                        isPaymentRequired = false,
                        onPlanSelected = { },
                        onExpectedDeliveriesSelected = { },
                        premiumForSelection = { plan, deliveries -> viewModel.calculatePremium(plan, deliveries) },
                        upgradeFeeForSelection = { plan -> viewModel.calculateUpgradeFee(plan) },
                        onConfirm = { },
                        onSkip = { }
                    )

                    SelectionBanner(
                        modifier = Modifier
                            .align(Alignment.TopCenter)
                            .padding(top = 18.dp),
                        title = "Plan selected",
                        message = "Your selected plan stays on screen for review. You can continue using the app without leaving this page."
                    )
                }
                is PlanUiState.Skipped -> {
                    PlanContent(
                        plans = state.plans,
                        currentPolicy = currentPolicy,
                        selectedPlan = selectedPlan,
                        selectedExpectedDeliveries = selectedExpectedDeliveries,
                        isPaymentRequired = false,
                        onPlanSelected = { },
                        onExpectedDeliveriesSelected = { },
                        premiumForSelection = { plan, deliveries -> viewModel.calculatePremium(plan, deliveries) },
                        upgradeFeeForSelection = { plan -> viewModel.calculateUpgradeFee(plan) },
                        onConfirm = { },
                        onSkip = { }
                    )

                    SelectionBanner(
                        modifier = Modifier
                            .align(Alignment.TopCenter)
                            .padding(top = 18.dp),
                        title = "Plan skipped",
                        message = "Taking you to the main dashboard now."
                    )
                }
                is PlanUiState.Error -> {
                    Column(
                        modifier = Modifier
                            .fillMaxSize()
                            .padding(16.dp),
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.Center
                    ) {
                        Text(state.message, color = Color.Red)
                        Button(onClick = { viewModel.loadPlans() }) {
                            Text("Retry")
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun SelectionBanner(
    modifier: Modifier = Modifier,
    title: String,
    message: String
) {
    Card(
        modifier = modifier.padding(horizontal = 16.dp),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = BrandBlue),
        elevation = CardDefaults.cardElevation(defaultElevation = 8.dp)
    ) {
        Column(modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp)) {
            Text(title, color = Color.White, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(4.dp))
            Text(message, color = Color.White.copy(alpha = 0.9f), fontSize = 12.sp)
        }
    }
}

@Composable
fun PlanContent(
    plans: List<DeliveryPlan>,
    currentPolicy: Policy?,
    selectedPlan: DeliveryPlan?,
    selectedExpectedDeliveries: Int?,
    isPaymentRequired: Boolean,
    onPlanSelected: (DeliveryPlan) -> Unit,
    onExpectedDeliveriesSelected: (Int) -> Unit,
    premiumForSelection: (DeliveryPlan, Int?) -> Int,
    upgradeFeeForSelection: (DeliveryPlan) -> Int,
    onConfirm: (Int) -> Unit,
    onSkip: () -> Unit
) {
    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        // Header
        item {
            Column(modifier = Modifier.fillMaxWidth()) {
                Text(
                    "Select a plan based on your expected delivery volume",
                    style = MaterialTheme.typography.bodyMedium,
                    color = TextSecondary,
                    modifier = Modifier.padding(bottom = 8.dp)
                )
                Text(
                    "All plans include full income protection and disruption coverage",
                    style = MaterialTheme.typography.bodySmall,
                    color = TextSecondary
                )

                currentPolicy?.let { policy ->
                    val currentPlan = plans.firstOrNull { it.planId == policy.planId }
                    if (currentPlan != null) {
                        Spacer(modifier = Modifier.height(12.dp))
                        Card(
                            modifier = Modifier.fillMaxWidth(),
                            shape = RoundedCornerShape(14.dp),
                            colors = CardDefaults.cardColors(containerColor = BrandBlue.copy(alpha = 0.08f)),
                            border = androidx.compose.foundation.BorderStroke(1.dp, BrandBlue.copy(alpha = 0.25f))
                        ) {
                            Column(modifier = Modifier.padding(16.dp)) {
                                Text("Current plan", fontWeight = FontWeight.Bold, color = BrandBlue)
                                Spacer(modifier = Modifier.height(6.dp))
                                Text(currentPlan.planName, fontWeight = FontWeight.SemiBold)
                                Text(
                                    "Max payout: ₹${currentPlan.maxPayoutInr}",
                                    style = MaterialTheme.typography.bodySmall,
                                    color = TextSecondary
                                )
                                Text(
                                    "Upgrading to a higher payout tier adds ₹5. Lower payout changes are free.",
                                    style = MaterialTheme.typography.bodySmall,
                                    color = TextSecondary
                                )
                            }
                        }
                    }
                }
            }
        }

        // Plan Cards
        items(plans.size) { index ->
            val plan = plans[index]
            val isSelected = selectedPlan?.planId == plan.planId
            PlanCard(
                plan = plan,
                isSelected = isSelected,
                onSelect = { onPlanSelected(plan) }
            )
        }

        // Selection Summary & Payment
        item {
            if (selectedPlan != null) {
                val selectedDeliveries = selectedExpectedDeliveries ?: selectedPlan.rangeStart
                val selectedPremium = premiumForSelection(selectedPlan, selectedDeliveries)
                val upgradeFee = upgradeFeeForSelection(selectedPlan)
                val totalPayment = selectedPremium + upgradeFee

                Card(
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    colors = CardDefaults.cardColors(containerColor = BrandBlue.copy(alpha = 0.1f)),
                    border = androidx.compose.foundation.BorderStroke(1.dp, BrandBlue)
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text(
                            "Plan Selected: ${selectedPlan.planName}",
                            fontWeight = FontWeight.Bold,
                            color = BrandBlue
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            "Selected Deliveries: $selectedDeliveries/week",
                            style = MaterialTheme.typography.bodyMedium
                        )
                        Text(
                            "Weekly Premium: ₹$selectedPremium",
                            style = MaterialTheme.typography.bodyMedium
                        )
                        if (upgradeFee > 0) {
                            Text(
                                "Upgrade fee: ₹$upgradeFee",
                                style = MaterialTheme.typography.bodyMedium,
                                color = SuccessGreen,
                                fontWeight = FontWeight.SemiBold
                            )
                        }
                        Text(
                            "Total payable: ₹$totalPayment",
                            style = MaterialTheme.typography.bodyMedium,
                            fontWeight = FontWeight.Bold
                        )
                        Text(
                            "Coverage: ${(selectedPlan.coverageRatio * 100).toInt()}%",
                            style = MaterialTheme.typography.bodySmall,
                            color = TextSecondary
                        )
                    }
                }

                Spacer(modifier = Modifier.height(16.dp))

                Card(
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    colors = CardDefaults.cardColors(containerColor = Color.White),
                    border = androidx.compose.foundation.BorderStroke(1.dp, Color.LightGray)
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text(
                            "Choose expected deliveries in selected range",
                            fontWeight = FontWeight.SemiBold,
                            modifier = Modifier.padding(bottom = 8.dp)
                        )
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.spacedBy(8.dp)
                        ) {
                            for (delivery in selectedPlan.rangeStart..selectedPlan.rangeEnd) {
                                val chosen = delivery == selectedDeliveries
                                OutlinedButton(
                                    onClick = { onExpectedDeliveriesSelected(delivery) },
                                    colors = ButtonDefaults.outlinedButtonColors(
                                        containerColor = if (chosen) BrandBlue.copy(alpha = 0.1f) else Color.Transparent,
                                        contentColor = if (chosen) BrandBlue else TextSecondary,
                                    ),
                                    modifier = Modifier.weight(1f)
                                ) {
                                    Text(delivery.toString(), textAlign = TextAlign.Center)
                                }
                            }
                        }
                    }
                }

                Spacer(modifier = Modifier.height(16.dp))

                // Action Buttons
                if (isPaymentRequired) {
                    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                        Button(
                            onClick = { onConfirm(totalPayment) },
                            modifier = Modifier.fillMaxWidth(),
                            colors = ButtonDefaults.buttonColors(containerColor = SuccessGreen)
                        ) {
                            Text("Touch Pay & Confirm Plan: ₹$totalPayment")
                        }
                        Text(
                            if (upgradeFee > 0) {
                                "This is an upgrade. The ₹5 fee is added because the payout tier is higher."
                            } else {
                                "Payment is mandatory to activate this plan."
                            },
                            style = MaterialTheme.typography.bodySmall,
                            color = TextSecondary
                        )
                    }
                }
            }
        }

        item {
            Spacer(modifier = Modifier.height(32.dp))
            OutlinedButton(
                onClick = onSkip,
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(12.dp)
            ) {
                Text("Skip")
            }
        }
    }
}

@Composable
fun PlanCard(
    plan: DeliveryPlan,
    isSelected: Boolean,
    onSelect: () -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onSelect() }
            .background(
                if (isSelected) BrandBlue.copy(alpha = 0.05f) else Color.Transparent
            ),
        shape = RoundedCornerShape(12.dp),
        border = androidx.compose.foundation.BorderStroke(
            2.dp,
            if (isSelected) BrandBlue else Color.LightGray
        ),
        colors = CardDefaults.cardColors(containerColor = Color.White)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Column(modifier = Modifier.weight(1f)) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    modifier = Modifier.padding(bottom = 8.dp)
                ) {
                    Icon(
                        Icons.Default.LocalOffer,
                        contentDescription = null,
                        tint = BrandBlue,
                        modifier = Modifier
                            .size(20.dp)
                            .padding(end = 4.dp)
                    )
                    Text(
                        plan.planName,
                        fontWeight = FontWeight.Bold,
                        fontSize = 16.sp
                    )
                }

                Text(
                    "${plan.rangeStart}-${plan.rangeEnd} deliveries/week",
                    style = MaterialTheme.typography.bodySmall,
                    color = TextSecondary,
                    modifier = Modifier.padding(bottom = 8.dp)
                )

                val minPremium = plan.weeklyPremiumMinInr ?: plan.weeklyPremiumInr
                val maxPremium = plan.weeklyPremiumMaxInr ?: plan.weeklyPremiumInr
                Text(
                    "Premium range: ₹$minPremium - ₹$maxPremium",
                    style = MaterialTheme.typography.bodySmall,
                    color = BrandBlue,
                    modifier = Modifier.padding(bottom = 8.dp)
                )

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(16.dp)
                ) {
                    Column {
                        Text(
                            "Premium",
                            style = MaterialTheme.typography.labelSmall,
                            color = TextSecondary
                        )
                        Text(
                            "₹${plan.weeklyPremiumInr}",
                            fontWeight = FontWeight.Bold,
                            color = BrandBlue
                        )
                    }
                    Column {
                        Text(
                            "Coverage",
                            style = MaterialTheme.typography.labelSmall,
                            color = TextSecondary
                        )
                        Text(
                            "${(plan.coverageRatio * 100).toInt()}%",
                            fontWeight = FontWeight.Bold
                        )
                    }
                    Column {
                        Text(
                            "Max Payout",
                            style = MaterialTheme.typography.labelSmall,
                            color = TextSecondary
                        )
                        Text(
                            "₹${plan.maxPayoutInr}",
                            fontWeight = FontWeight.Bold,
                            color = SuccessGreen
                        )
                    }
                }
            }

            if (isSelected) {
                Icon(
                    Icons.Default.CheckCircle,
                    contentDescription = "Selected",
                    tint = BrandBlue,
                    modifier = Modifier.size(32.dp)
                )
            }
        }
    }
}
