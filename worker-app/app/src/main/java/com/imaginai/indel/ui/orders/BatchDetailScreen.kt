package com.imaginai.indel.ui.orders

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.clickable
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
import androidx.compose.runtime.mutableStateMapOf
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
import com.imaginai.indel.ui.navigation.Screen
import com.imaginai.indel.ui.theme.BackgroundWarmWhite
import com.imaginai.indel.ui.theme.BlueSoft
import com.imaginai.indel.ui.theme.BrandBlue
import com.imaginai.indel.ui.theme.ErrorRed
import com.imaginai.indel.ui.theme.SuccessGreen
import com.imaginai.indel.ui.theme.TextSecondary
import kotlinx.coroutines.launch
import java.util.Locale

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun BatchDetailScreen(
    navController: NavController,
    batchId: String,
) {
    val ordersEntry = remember(navController) {
        runCatching { navController.getBackStackEntry(Screen.Orders.route) }.getOrNull()
    }
    val viewModel: OrdersViewModel = ordersEntry?.let { hiltViewModel(it) } ?: hiltViewModel()
    val uiState by viewModel.uiState.collectAsState()
    val batch = viewModel.getBatchById(batchId)
    val coroutineScope = rememberCoroutineScope()
    val pickupCode = remember(batchId, batch?.pickupCode) { batch?.pickupCode ?: viewModel.pickupCodeForBatch(batchId) }
    val deliveryCode = remember(batchId, batch?.deliveryCode) { batch?.deliveryCode ?: viewModel.deliveryCodeForBatch(batchId) }
    val isZoneASingleStop = remember(batch?.batchId, batch?.zoneLevel, batch?.fromCity, batch?.toCity) {
        batch?.let { viewModel.isZoneASingleStop(it) } == true
    }
    val normalizedBatchStatus = remember(batch?.status) {
        batch?.status
            ?.trim()
            ?.lowercase(Locale.getDefault())
            ?.replace(" ", "_")
            .orEmpty()
    }
    val orderDeliveryCodes = remember(batchId) { mutableStateMapOf<String, String>() }
    var selectedZoneAOrderId by rememberSaveable(batchId) { mutableStateOf<String?>(null) }
    val deliveredOrderCount = remember(batch?.batchId, batch?.orders) {
        batch?.orders?.count { it.status.equals("delivered", ignoreCase = true) } ?: 0
    }
    var enteredCode by rememberSaveable(batchId) { mutableStateOf("") }
    var feedbackMessage by rememberSaveable(batchId) { mutableStateOf<String?>(null) }
    var isAccepting by rememberSaveable(batchId) { mutableStateOf(false) }
    var isDelivering by rememberSaveable(batchId) { mutableStateOf(false) }

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
                                Text(compactRouteLabel(batch.zoneLevel, batch.fromCity, batch.toCity), style = MaterialTheme.typography.bodyMedium, fontWeight = FontWeight.SemiBold)
                                Text("${formatBatchWeight(batch.totalWeight)} kg • ${batch.orderCount} orders", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                                Text("Batch earning: ₹${String.format(java.util.Locale.getDefault(), "%.0f", batch.batchEarningInr ?: 0.0)}", style = MaterialTheme.typography.bodySmall, color = SuccessGreen)
                                if (batch.totalWeight < 10.0) {
                                    Text("Packing below target; waiting for more orders.", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                                } else {
                                    Text("Packed for the 10-12 kg range.", style = MaterialTheme.typography.labelSmall, color = SuccessGreen)
                                }
                            }
                        }
                    }

                    item {
                        Card(
                            modifier = Modifier.fillMaxWidth(),
                            shape = RoundedCornerShape(14.dp),
                            colors = CardDefaults.cardColors(containerColor = Color.White),
                            border = androidx.compose.foundation.BorderStroke(1.dp, BlueSoft)
                        ) {
                            Column(modifier = Modifier.padding(14.dp)) {
                                Text("Delivery progress", style = MaterialTheme.typography.labelLarge, fontWeight = FontWeight.Bold)
                                Spacer(modifier = Modifier.height(6.dp))
                                Text(
                                    "$deliveredOrderCount/${batch.orders.size} orders delivered",
                                    style = MaterialTheme.typography.bodyMedium,
                                    fontWeight = FontWeight.SemiBold,
                                    color = BrandBlue
                                )
                                Text(
                                    when {
                                        normalizedBatchStatus == "delivered" -> "Batch completed. Earnings are already released."
                                        normalizedBatchStatus == "picked_up" -> if (isZoneASingleStop) "Fill each order delivery code and tap Make Delivery." else "Enter the delivery code and tap Make Delivery to complete the batch."
                                        else -> "Pick up the batch first before starting delivery."
                                    },
                                    style = MaterialTheme.typography.bodySmall,
                                    color = TextSecondary
                                )
                            }
                        }
                    }

                    item {
                        when {
                            normalizedBatchStatus == "delivered" -> {
                                Card(
                                    modifier = Modifier.fillMaxWidth(),
                                    shape = RoundedCornerShape(14.dp),
                                    colors = CardDefaults.cardColors(containerColor = Color.White),
                                    border = androidx.compose.foundation.BorderStroke(1.dp, SuccessGreen)
                                ) {
                                    Column(modifier = Modifier.padding(14.dp)) {
                                        Text("Delivery complete", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold, color = SuccessGreen)
                                        Spacer(modifier = Modifier.height(6.dp))
                                        Text(
                                            "All orders in this batch are delivered and the earnings have been posted.",
                                            style = MaterialTheme.typography.bodySmall,
                                            color = TextSecondary
                                        )
                                    }
                                }
                            }
                            normalizedBatchStatus == "picked_up" && isZoneASingleStop -> {
                                ZoneABatchOrdersSection(
                                    batch = batch,
                                    orderDeliveryCodes = orderDeliveryCodes,
                                    selectedOrderId = selectedZoneAOrderId,
                                    onSelectOrder = { orderId ->
                                        selectedZoneAOrderId = if (selectedZoneAOrderId == orderId) null else orderId
                                    },
                                    feedbackMessage = feedbackMessage,
                                    isDelivering = isDelivering,
                                    onDeliverOrder = { order, code ->
                                        isDelivering = true
                                        coroutineScope.launch {
                                            val deliveryResult = viewModel.deliverBatch(batch, code)
                                            isDelivering = false
                                            feedbackMessage = if (deliveryResult.success) {
                                                orderDeliveryCodes[order.orderId] = ""
                                                selectedZoneAOrderId = null
                                                if (deliveryResult.batchCompleted) {
                                                    "Delivery accepted. Batch status changed to delivered."
                                                } else {
                                                    val remaining = deliveryResult.remainingOrders ?: (batch.orders.size - (deliveredOrderCount + 1))
                                                    "Delivery accepted. $remaining order(s) remaining in this batch."
                                                }
                                            } else {
                                                deliveryResult.errorMessage ?: "Unable to complete delivery right now."
                                            }
                                        }
                                    }
                                )
                            }
                            else -> {
                                BatchActionCard(
                                    batch = batch,
                                    pickupCode = pickupCode,
                                    deliveryCode = deliveryCode,
                                    enteredCode = enteredCode,
                                    onEnteredCodeChange = { enteredCode = it.take(4) },
                                    feedbackMessage = feedbackMessage,
                                    onConfirmPickup = {
                                        if (enteredCode.trim() != pickupCode) {
                                            feedbackMessage = "Incorrect pickup code"
                                            return@BatchActionCard
                                        }
                                        isAccepting = true
                                        coroutineScope.launch {
                                            val accepted = viewModel.acceptBatch(batch, enteredCode.trim())
                                            isAccepting = false
                                            feedbackMessage = if (accepted) {
                                                "Batch picked up successfully."
                                            } else {
                                                "Unable to pick up this batch right now."
                                            }
                                        }
                                    },
                                    onConfirmDelivery = {
                                        val code = enteredCode.trim()
                                        if (code.isBlank()) {
                                            feedbackMessage = "Enter delivery code"
                                            return@BatchActionCard
                                        }

                                        if (code != deliveryCode) {
                                            feedbackMessage = "Incorrect delivery code"
                                            return@BatchActionCard
                                        }

                                        isDelivering = true
                                        coroutineScope.launch {
                                            val deliveryResult = viewModel.deliverBatch(batch, code)
                                            isDelivering = false
                                            feedbackMessage = if (deliveryResult.success) {
                                                enteredCode = ""
                                                if (deliveryResult.batchCompleted) {
                                                    "Delivery accepted. Batch status changed to delivered."
                                                } else {
                                                    val remaining = deliveryResult.remainingOrders ?: (batch.orders.size - (deliveredOrderCount + 1))
                                                    "Delivery accepted. $remaining order(s) remaining in this batch."
                                                }
                                            } else {
                                                deliveryResult.errorMessage ?: "Unable to complete delivery right now."
                                            }
                                        }
                                    },
                                    isAccepting = isAccepting,
                                    isDelivering = isDelivering,
                                    showBatchDeliveryCode = true,
                                )
                            }
                        }
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
                                Text("Pickup: ${compactRouteLabel(batch.zoneLevel, nestedOrder.pickupArea ?: "-", nestedOrder.dropArea ?: nestedOrder.deliveryAddress)}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                                if (!nestedOrder.pickupTime.isNullOrBlank()) {
                                    Text("Picked up: ${nestedOrder.pickupTime}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                                }
                                if (!nestedOrder.deliveryTime.isNullOrBlank()) {
                                    Text("Delivered: ${nestedOrder.deliveryTime}", style = MaterialTheme.typography.bodySmall, color = SuccessGreen)
                                }
                                if (isZoneASingleStop && !nestedOrder.deliveryCode.isNullOrBlank()) {
                                    Text("Delivery code: ${nestedOrder.deliveryCode}", style = MaterialTheme.typography.bodySmall, color = BrandBlue, fontWeight = FontWeight.Medium)
                                }
                                Text("Weight: ${formatBatchWeight(nestedOrder.weight)} kg", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                                Text("Status: ${nestedOrder.status ?: "assigned"}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
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

}

@Composable
private fun ZoneABatchDeliverySection(
    batch: DeliveryBatch,
    orderDeliveryCodes: MutableMap<String, String>,
    selectedOrderId: String?,
    onSelectOrder: (String) -> Unit,
    feedbackMessage: String?,
    isDelivering: Boolean,
    onDeliverOrder: (BatchOrder, String) -> Unit,
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        border = androidx.compose.foundation.BorderStroke(1.dp, BrandBlue.copy(alpha = 0.25f))
    ) {
        Column(modifier = Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Text("Zone A delivery", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold, color = BrandBlue)
            Text(
                "Deliver each order individually. The batch completes only when all orders are delivered.",
                style = MaterialTheme.typography.bodySmall,
                color = TextSecondary
            )

            batch.orders.forEach { order ->
                val isDelivered = order.status.equals("delivered", ignoreCase = true)
                val expectedCode = order.deliveryCode.orEmpty()
                val isExpanded = selectedOrderId == order.orderId || batch.orders.count { it.status.equals("delivered", ignoreCase = true) } == batch.orders.size - 1
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    colors = CardDefaults.cardColors(containerColor = BackgroundWarmWhite),
                    border = androidx.compose.foundation.BorderStroke(1.dp, if (isDelivered) SuccessGreen else BlueSoft)
                ) {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clickable(enabled = !isDelivered) { onSelectOrder(order.orderId) }
                            .padding(12.dp),
                        verticalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        Text(order.orderId, style = MaterialTheme.typography.labelLarge, fontWeight = FontWeight.Bold)
                        Text("Address: ${order.deliveryAddress}", style = MaterialTheme.typography.bodySmall)
                        Text("Contact: ${order.contactName}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                        Text("Phone: ${order.contactPhone}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                        if (!order.deliveryTime.isNullOrBlank()) {
                            Text("Delivered at: ${order.deliveryTime}", style = MaterialTheme.typography.bodySmall, color = SuccessGreen)
                        }

                        if (isDelivered) {
                            Text("Delivered", color = SuccessGreen, fontWeight = FontWeight.Bold)
                        } else {
                            Text(
                                if (isExpanded) "Tap code, enter it, and deliver this order." else "Tap to enter delivery code",
                                style = MaterialTheme.typography.bodySmall,
                                color = TextSecondary
                            )
                            if (isExpanded) {
                                Text(
                                    "Order code: $expectedCode",
                                    style = MaterialTheme.typography.bodySmall,
                                    color = BrandBlue,
                                    fontWeight = FontWeight.SemiBold,
                                )
                                OutlinedTextField(
                                    value = orderDeliveryCodes[order.orderId].orEmpty(),
                                    onValueChange = { orderDeliveryCodes[order.orderId] = it.take(4) },
                                    modifier = Modifier.fillMaxWidth(),
                                    singleLine = true,
                                    label = { Text("Delivery code") },
                                    placeholder = { Text("Enter 4-digit code") }
                                )
                                Button(
                                    onClick = {
                                        val code = orderDeliveryCodes[order.orderId].orEmpty().trim()
                                        if (code.isBlank()) {
                                            return@Button
                                        }
                                        onDeliverOrder(order, code)
                                    },
                                    modifier = Modifier.fillMaxWidth(),
                                    colors = ButtonDefaults.buttonColors(containerColor = SuccessGreen),
                                    shape = RoundedCornerShape(12.dp),
                                    enabled = !isDelivering
                                ) {
                                    Text(if (isDelivering) "Checking..." else "Make Delivery")
                                }
                                Text(
                                    "Enter this exact order code to deliver this order. Batch completes when all orders are delivered.",
                                    style = MaterialTheme.typography.bodySmall,
                                    color = TextSecondary
                                )
                            }
                        }
                    }
                }
            }

            feedbackMessage?.let {
                Text(
                    it,
                    color = if (it.contains("unable", ignoreCase = true) || it.contains("incorrect", ignoreCase = true)) ErrorRed else SuccessGreen,
                    style = MaterialTheme.typography.bodySmall,
                )
            }
        }
    }
}

@Composable
private fun ZoneABatchOrdersSection(
    batch: DeliveryBatch,
    orderDeliveryCodes: MutableMap<String, String>,
    selectedOrderId: String?,
    onSelectOrder: (String) -> Unit,
    feedbackMessage: String?,
    isDelivering: Boolean,
    onDeliverOrder: (BatchOrder, String) -> Unit,
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        border = androidx.compose.foundation.BorderStroke(1.dp, BrandBlue.copy(alpha = 0.25f))
    ) {
        Column(modifier = Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Text("Zone A delivery", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold, color = BrandBlue)
            Text("Touch an order to reveal its delivery code field, then make delivery for that order.", style = MaterialTheme.typography.bodySmall, color = TextSecondary)

            batch.orders.forEach { order ->
                val isDelivered = order.status.equals("delivered", ignoreCase = true)
                val isExpanded = selectedOrderId == order.orderId
                val expectedCode = order.deliveryCode.orEmpty()

                Card(
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable(enabled = !isDelivered) { onSelectOrder(order.orderId) },
                    shape = RoundedCornerShape(12.dp),
                    colors = CardDefaults.cardColors(containerColor = BackgroundWarmWhite),
                    border = androidx.compose.foundation.BorderStroke(1.dp, if (isDelivered) SuccessGreen else BlueSoft)
                ) {
                    Column(modifier = Modifier.padding(12.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                        Text(order.orderId, style = MaterialTheme.typography.labelLarge, fontWeight = FontWeight.Bold)
                        Text("Address: ${order.deliveryAddress}", style = MaterialTheme.typography.bodySmall)
                        Text("Contact: ${order.contactName}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                        Text("Phone: ${order.contactPhone}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                        if (!order.deliveryTime.isNullOrBlank()) {
                            Text("Delivered at: ${order.deliveryTime}", style = MaterialTheme.typography.bodySmall, color = SuccessGreen)
                        }

                        if (isDelivered) {
                            Text("Delivered", color = SuccessGreen, fontWeight = FontWeight.Bold)
                        } else {
                            Text(
                                if (isExpanded) "Code for this order: $expectedCode" else "Tap to enter delivery code",
                                style = MaterialTheme.typography.bodySmall,
                                color = BrandBlue,
                                fontWeight = FontWeight.SemiBold,
                            )
                            if (isExpanded) {
                                OutlinedTextField(
                                    value = orderDeliveryCodes[order.orderId].orEmpty(),
                                    onValueChange = { orderDeliveryCodes[order.orderId] = it.take(4) },
                                    modifier = Modifier.fillMaxWidth(),
                                    singleLine = true,
                                    label = { Text("Delivery code") },
                                    placeholder = { Text("Enter 4-digit code") }
                                )
                                Button(
                                    onClick = {
                                        val code = orderDeliveryCodes[order.orderId].orEmpty().trim()
                                        if (code.isBlank()) {
                                            return@Button
                                        }
                                        onDeliverOrder(order, code)
                                    },
                                    modifier = Modifier.fillMaxWidth(),
                                    colors = ButtonDefaults.buttonColors(containerColor = SuccessGreen),
                                    shape = RoundedCornerShape(12.dp),
                                    enabled = !isDelivering
                                ) {
                                    Text(if (isDelivering) "Checking..." else "Make Delivery")
                                }
                                Text("Enter this exact order code to mark the order delivered.", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                            }
                        }
                    }
                }
            }

            feedbackMessage?.let {
                Text(
                    it,
                    color = if (it.contains("unable", ignoreCase = true) || it.contains("incorrect", ignoreCase = true)) ErrorRed else SuccessGreen,
                    style = MaterialTheme.typography.bodySmall,
                )
            }
        }
    }
}

@Composable
private fun BatchLevelDeliverySection(
    batch: DeliveryBatch,
    enteredCode: String,
    onEnteredCodeChange: (String) -> Unit,
    feedbackMessage: String?,
    isDelivering: Boolean,
    onDeliverBatch: () -> Unit,
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        border = androidx.compose.foundation.BorderStroke(1.dp, SuccessGreen.copy(alpha = 0.35f))
    ) {
        Column(modifier = Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Text("Inter-city delivery", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold, color = SuccessGreen)
            Text(
                "Zone ${batch.zoneLevel} batches are delivered at batch level. Enter the batch code once to complete every order.",
                style = MaterialTheme.typography.bodySmall,
                color = TextSecondary
            )
            OutlinedTextField(
                value = enteredCode,
                onValueChange = onEnteredCodeChange,
                modifier = Modifier.fillMaxWidth(),
                singleLine = true,
                label = { Text("Delivery code") },
                placeholder = { Text("Enter 4-digit code") }
            )
            Text("Enter the delivery code you received to complete this batch.", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
            Button(
                onClick = onDeliverBatch,
                modifier = Modifier.fillMaxWidth(),
                colors = ButtonDefaults.buttonColors(containerColor = SuccessGreen),
                shape = RoundedCornerShape(12.dp),
                enabled = !isDelivering
            ) {
                Text(if (isDelivering) "Checking..." else "Make Delivery")
            }
            feedbackMessage?.let {
                Text(
                    it,
                    color = if (it.contains("unable", ignoreCase = true) || it.contains("incorrect", ignoreCase = true)) ErrorRed else SuccessGreen,
                    style = MaterialTheme.typography.bodySmall,
                )
            }
        }
    }
}

@Composable
private fun BatchActionCard(
    batch: DeliveryBatch,
    pickupCode: String,
    deliveryCode: String,
    enteredCode: String,
    onEnteredCodeChange: (String) -> Unit,
    feedbackMessage: String?,
    onConfirmPickup: () -> Unit,
    onConfirmDelivery: () -> Unit,
    isAccepting: Boolean,
    isDelivering: Boolean,
    showBatchDeliveryCode: Boolean,
) {
    var showPickupDialog by rememberSaveable(batch.batchId) { mutableStateOf(false) }
    var pickupDialogCode by rememberSaveable(batch.batchId) { mutableStateOf("") }
    val statusLower = batch.status
        .trim()
        .lowercase(Locale.getDefault())
        .replace(" ", "_")
    val statusLabel = when (statusLower) {
        "assigned" -> "Assigned"
        "accepted" -> "Accepted"
        "picked_up" -> "Picked Up"
        "delivered" -> "Delivered"
        else -> batch.status.replace("_", " ").replaceFirstChar { it.uppercase() }
    }
    val isPickupStage = statusLower == "pending" || statusLower == "assigned" || statusLower == "accepted"
    val isDeliveryStage = statusLower == "picked_up"
    val expectedCode = if (isPickupStage) pickupCode else deliveryCode
    val actionLabel = when {
        isPickupStage -> "Pick Up Batch"
        isDeliveryStage -> "Make Delivery"
        else -> "Completed"
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

            if (isDeliveryStage) {
                OutlinedTextField(
                    value = enteredCode,
                    onValueChange = onEnteredCodeChange,
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = true,
                    label = { Text("Delivery code") },
                    placeholder = { Text("Enter 4-digit code") },
                )

                if (showBatchDeliveryCode) {
                    Text(
                        "Code for this batch: $expectedCode",
                        style = MaterialTheme.typography.labelSmall,
                        color = TextSecondary,
                    )
                } else {
                    Text(
                        "Enter the delivery code and tap Make Delivery.",
                        style = MaterialTheme.typography.labelSmall,
                        color = TextSecondary,
                    )
                }
            }

            if (feedbackMessage != null) {
                Text(
                    feedbackMessage,
                    color = if (feedbackMessage.contains("unable", ignoreCase = true) || feedbackMessage.contains("incorrect", ignoreCase = true)) ErrorRed else SuccessGreen,
                    style = MaterialTheme.typography.bodySmall,
                )
            }

            if (isPickupStage) {
                Button(
                    onClick = {
                        pickupDialogCode = ""
                        showPickupDialog = true
                    },
                    modifier = Modifier.fillMaxWidth(),
                    colors = ButtonDefaults.buttonColors(containerColor = BrandBlue),
                    shape = RoundedCornerShape(12.dp),
                    enabled = !isAccepting,
                ) {
                    Text("Pick Up Batch")
                }
                Text(
                    "Tap Pick Up Batch, enter code in the prompt, and all orders move to picked up.",
                    style = MaterialTheme.typography.bodySmall,
                    color = TextSecondary,
                )
            } else if (isDeliveryStage) {
                Button(
                    onClick = onConfirmDelivery,
                    modifier = Modifier.fillMaxWidth(),
                    colors = ButtonDefaults.buttonColors(containerColor = SuccessGreen),
                    shape = RoundedCornerShape(12.dp),
                    enabled = !isDelivering,
                ) {
                    Text(if (isDelivering) "Checking..." else actionLabel)
                }
                Text(
                    "Enter the delivery code to complete the batch and release earnings.",
                    style = MaterialTheme.typography.bodySmall,
                    color = TextSecondary,
                )
            } else {
                Text(
                    "This batch is already delivered.",
                    style = MaterialTheme.typography.bodySmall,
                    color = TextSecondary,
                )
            }
        }
    }

    if (showPickupDialog) {
        AlertDialog(
            onDismissRequest = { showPickupDialog = false },
            title = { Text("Enter Pickup Code") },
            text = {
                Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    OutlinedTextField(
                        value = pickupDialogCode,
                        onValueChange = { pickupDialogCode = it.take(4) },
                        singleLine = true,
                        label = { Text("Pickup code") },
                        placeholder = { Text("4-digit code") },
                        modifier = Modifier.fillMaxWidth(),
                    )
                    Text("Code for this batch: $pickupCode", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                }
            },
            confirmButton = {
                TextButton(
                    onClick = {
                        onEnteredCodeChange(pickupDialogCode.trim())
                        onConfirmPickup()
                        showPickupDialog = false
                    },
                    enabled = !isAccepting && pickupDialogCode.trim().length == 4,
                ) {
                    Text(if (isAccepting) "Checking..." else "Confirm")
                }
            },
            dismissButton = {
                TextButton(onClick = { showPickupDialog = false }) {
                    Text("Cancel")
                }
            },
        )
    }
}

private fun compactRouteLabel(zoneLevel: String, fromLabel: String, toLabel: String): String {
    val from = fromLabel.trim()
    val to = toLabel.trim()
    if (zoneLevel.equals("A", ignoreCase = true) && from.equals(to, ignoreCase = true)) {
        return if (from.isBlank()) "Unknown" else from
    }
    if (from.isBlank() && to.isBlank()) {
        return "Unknown"
    }
    if (from.isBlank()) {
        return to
    }
    if (to.isBlank()) {
        return from
    }
    return "$from -> $to"
}

private fun formatBatchWeight(weight: Double): String {
    return String.format(Locale.getDefault(), "%.1f", weight)
}
