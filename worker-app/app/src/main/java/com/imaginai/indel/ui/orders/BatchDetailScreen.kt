package com.imaginai.indel.ui.orders

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.Divider
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.ui.theme.BackgroundWarmWhite
import com.imaginai.indel.ui.theme.BlueSoft
import com.imaginai.indel.ui.theme.BrandBlue
import com.imaginai.indel.ui.theme.ErrorRed
import com.imaginai.indel.ui.theme.SuccessGreen
import com.imaginai.indel.ui.theme.TextSecondary
import kotlinx.coroutines.launch
import kotlin.random.Random

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun BatchDetailScreen(
    navController: NavController,
    batchId: String,
    viewModel: OrdersViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsState()
    val batch = viewModel.getBatchById(batchId)
    val coroutineScope = rememberCoroutineScope()
    val pickupCode = remember(batchId) { Random.nextInt(1000, 10000).toString() }
    var showCodeDialog by rememberSaveable(batchId) { mutableStateOf(false) }
    var enteredCode by rememberSaveable(batchId) { mutableStateOf("") }
    var feedbackMessage by rememberSaveable(batchId) { mutableStateOf<String?>(null) }
    var isAccepting by rememberSaveable(batchId) { mutableStateOf(false) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Batch Details", fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = BrandBlue,
                    titleContentColor = Color.White,
                    navigationIconContentColor = Color.White,
                )
            )
        }
    ) { padding ->
        when {
            batch != null -> {
                LazyColumn(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding)
                        .background(BackgroundWarmWhite),
                    contentPadding = PaddingValues(16.dp),
                    verticalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    item {
                        Card(
                            modifier = Modifier.fillMaxWidth(),
                            shape = RoundedCornerShape(14.dp),
                            colors = CardDefaults.cardColors(containerColor = Color.White),
                            border = androidx.compose.foundation.BorderStroke(1.dp, BlueSoft)
                        ) {
                            Column(modifier = Modifier.padding(14.dp)) {
                                Text(batch.batchId, style = MaterialTheme.typography.titleSmall, color = BrandBlue, fontWeight = FontWeight.Bold)
                                Spacer(modifier = Modifier.height(6.dp))
                                Text("Zone ${batch.zoneLevel}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                                Text("${batch.fromCity} -> ${batch.toCity}", style = MaterialTheme.typography.bodyMedium, fontWeight = FontWeight.SemiBold)
                                Text("${batch.totalWeight} kg • ${batch.orderCount} orders", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                            }
                        }
                    }

                    item {
                        BatchActionCard(
                            batch = batch,
                            feedbackMessage = feedbackMessage,
                            onAcceptClick = {
                                feedbackMessage = null
                                showCodeDialog = true
                                enteredCode = ""
                            },
                        )
                    }

                    item {
                        Text("Orders in this batch", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                    }

                    items(batch.orders) { nestedOrder ->
                        Card(
                            modifier = Modifier.fillMaxWidth(),
                            shape = RoundedCornerShape(12.dp),
                            colors = CardDefaults.cardColors(containerColor = Color.White),
                            elevation = CardDefaults.cardElevation(defaultElevation = 1.dp)
                        ) {
                            Column(modifier = Modifier.padding(12.dp)) {
                                Text("Order ${nestedOrder.orderId}", style = MaterialTheme.typography.labelLarge, fontWeight = FontWeight.SemiBold)
                                Spacer(modifier = Modifier.height(6.dp))
                                Text("Address: ${nestedOrder.deliveryAddress}", style = MaterialTheme.typography.bodySmall)
                                Text("Contact: ${nestedOrder.contactName}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                                Text("Phone: ${nestedOrder.contactPhone}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                                Divider(modifier = Modifier.padding(top = 8.dp), color = Color(0xFFE9ECEF))
                            }
                        }
                    }
                }
            }
            uiState is OrdersUiState.Loading -> {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding)
                        .background(BackgroundWarmWhite),
                    contentAlignment = Alignment.Center,
                ) {
                    Text("Loading batch...", color = TextSecondary)
                }
            }
            else -> {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding)
                        .background(BackgroundWarmWhite),
                    contentAlignment = Alignment.Center,
                ) {
                    Text("Batch not found", color = TextSecondary)
                }
            }
        }
    }

    if (showCodeDialog && batch != null) {
        AlertDialog(
            onDismissRequest = { if (!isAccepting) showCodeDialog = false },
            title = { Text("Enter pickup code") },
            text = {
                Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    Text("Type the 4-digit code shown on the batch card to accept this batch.")
                    OutlinedTextField(
                        value = enteredCode,
                        onValueChange = { enteredCode = it.take(4) },
                        modifier = Modifier.fillMaxWidth(),
                        singleLine = true,
                        label = { Text("Pickup code") },
                    )
                    Text(
                        "Code for this batch: $pickupCode",
                        style = MaterialTheme.typography.labelSmall,
                        color = TextSecondary,
                    )
                }
            },
            confirmButton = {
                Button(
                    onClick = {
                        if (enteredCode.trim() != pickupCode) {
                            feedbackMessage = "Incorrect pickup code"
                            return@Button
                        }
                        isAccepting = true
                        coroutineScope.launch {
                            val accepted = viewModel.acceptBatch(batch)
                            isAccepting = false
                            showCodeDialog = false
                            feedbackMessage = if (accepted) {
                                "Batch accepted. The orders are now assigned to you."
                            } else {
                                "Unable to accept this batch right now."
                            }
                        }
                    },
                    enabled = !isAccepting,
                ) {
                    Text(if (isAccepting) "Checking..." else "Accept Batch")
                }
            },
            dismissButton = {
                TextButton(
                    onClick = { showCodeDialog = false },
                    enabled = !isAccepting,
                ) {
                    Text("Cancel")
                }
            }
        )
    }
}

@Composable
private fun BatchActionCard(
    batch: DeliveryBatch,
    feedbackMessage: String?,
    onAcceptClick: () -> Unit,
) {
    val statusLower = batch.status.lowercase()
    val statusLabel = when (statusLower) {
        "assigned" -> "Assigned"
        "accepted" -> "Accepted"
        "picked_up" -> "Picked Up"
        "delivered" -> "Delivered"
        else -> batch.status.replace("_", " ").replaceFirstChar { it.uppercase() }
    }

    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        border = androidx.compose.foundation.BorderStroke(1.dp, BlueSoft)
    ) {
        Column(modifier = Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Text("Batch status", style = MaterialTheme.typography.labelLarge, fontWeight = FontWeight.Bold)
            Text(statusLabel, color = BrandBlue, fontWeight = FontWeight.SemiBold)

            if (feedbackMessage != null) {
                Text(
                    feedbackMessage,
                    color = if (feedbackMessage.contains("unable", ignoreCase = true) || feedbackMessage.contains("incorrect", ignoreCase = true)) ErrorRed else SuccessGreen,
                    style = MaterialTheme.typography.bodySmall,
                )
            }

            if (statusLower == "pending" || statusLower == "assigned") {
                Button(
                    onClick = onAcceptClick,
                    modifier = Modifier.fillMaxWidth(),
                    colors = ButtonDefaults.buttonColors(containerColor = BrandBlue),
                    shape = RoundedCornerShape(12.dp),
                ) {
                    Text("Accept Batch")
                }
                Text(
                    "Accepting the batch will assign these orders to you.",
                    style = MaterialTheme.typography.bodySmall,
                    color = TextSecondary,
                )
            } else {
                Text(
                    "This batch is already active. Continue with the delivery flow for the individual orders.",
                    style = MaterialTheme.typography.bodySmall,
                    color = TextSecondary,
                )
            }
        }
    }
}
