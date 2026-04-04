package com.imaginai.indel.ui.orders

import androidx.compose.foundation.clickable
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.LocationOn
import androidx.compose.material3.*
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.ui.navigation.Screen
import com.imaginai.indel.ui.theme.*
import java.util.Locale

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun OrdersScreen(
    navController: NavController,
    viewModel: OrdersViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val isRefreshing by viewModel.isRefreshing.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("My Orders", fontWeight = FontWeight.Bold) },
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
                    is OrdersUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                    is OrdersUiState.Success -> OrdersContent(
                        assignedBatches = state.assignedBatches,
                        availableBatches = state.availableBatches,
                        pickedUpBatches = state.pickedUpBatches,
                        deliveredBatches = state.deliveredBatches,
                        diagnostics = state.diagnostics,
                        navController = navController,
                    )
                    is OrdersUiState.Error -> ErrorState(state.message) { viewModel.loadOrders() }
                }
            }
        }
    }
}

@Composable
fun OrdersContent(
    assignedBatches: List<DeliveryBatch>,
    availableBatches: List<DeliveryBatch>,
    pickedUpBatches: List<DeliveryBatch>,
    deliveredBatches: List<DeliveryBatch>,
    diagnostics: String,
    navController: NavController,
) {
    var selectedTab by remember { mutableIntStateOf(0) }
    val tabLabels = listOf(
        "Assigned (${assignedBatches.size})",
        "Available (${availableBatches.size})",
        "Picked Up (${pickedUpBatches.size})",
        "Delivered (${deliveredBatches.size})",
    )

    val sectionTitles = listOf(
        "Your Assigned Batches",
        "Available Batches Near You",
        "In Progress",
        "Completed Deliveries",
    )

    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        item {
            OrdersDiagnosticsCard(diagnostics)
        }

        item {
            TabRow(selectedTabIndex = selectedTab, containerColor = Color.White, contentColor = BrandBlue) {
                tabLabels.forEachIndexed { index, label ->
                    Tab(
                        selected = selectedTab == index,
                        onClick = { selectedTab = index },
                        text = { Text(label) }
                    )
                }
            }
        }

        val batches = when (selectedTab) {
            0 -> assignedBatches
            1 -> availableBatches
            2 -> pickedUpBatches
            else -> deliveredBatches
        }

        if (batches.isNotEmpty()) {
            item {
                Text(
                    text = sectionTitles[selectedTab],
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold
                )
            }
            items(batches) { batch ->
                BatchCard(
                    batch = batch,
                    onOpenDetails = { navController.navigate(Screen.BatchDetail.createRoute(batch.batchId)) },
                )
            }
        }

        if (batches.isEmpty()) {
            item {
                Box(modifier = Modifier.fillParentMaxSize(), contentAlignment = Alignment.Center) {
                    Text("No batches available at the moment", color = TextSecondary)
                }
            }
        }
    }
}

@Composable
fun OrdersDiagnosticsCard(diagnostics: String) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(containerColor = BlueSoft.copy(alpha = 0.55f)),
        border = androidx.compose.foundation.BorderStroke(1.dp, BrandBlue.copy(alpha = 0.25f))
    ) {
        Column(modifier = Modifier.padding(12.dp)) {
            Text(
                "Fetch diagnostics",
                style = MaterialTheme.typography.labelLarge,
                fontWeight = FontWeight.Bold,
                color = BrandBlue
            )
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                diagnostics,
                style = MaterialTheme.typography.bodySmall,
                color = TextSecondary
            )
        }
    }
}

@Composable
fun BatchCard(
    batch: DeliveryBatch,
    onOpenDetails: () -> Unit,
) {
    val (statusBgColor, statusTextColor, statusLabel) = statusBadgeStyle(batch.status)
    val compactRoute = compactRouteLabel(batch.zoneLevel, batch.fromCity, batch.toCity)

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onOpenDetails() },
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                Column {
                    Text("Batch ID", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                    Text(batch.batchId, style = MaterialTheme.typography.bodyMedium, fontWeight = FontWeight.Bold, color = BrandBlue)
                    Text("Zone ${batch.zoneLevel}: $compactRoute", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                }
                Surface(
                    color = BlueSoft,
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Text(
                        text = "${formatBatchWeight(batch.totalWeight)} kg",
                        modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp),
                        style = MaterialTheme.typography.labelMedium,
                        fontWeight = FontWeight.Bold,
                        color = BrandBlue
                    )
                }
            }

            Spacer(modifier = Modifier.height(10.dp))

            Surface(
                color = statusBgColor,
                shape = RoundedCornerShape(999.dp)
            ) {
                Text(
                    text = statusLabel,
                    modifier = Modifier.padding(horizontal = 10.dp, vertical = 4.dp),
                    style = MaterialTheme.typography.labelMedium,
                    fontWeight = FontWeight.Bold,
                    color = statusTextColor
                )
            }

            Spacer(modifier = Modifier.height(16.dp))

            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(Icons.Default.LocationOn, contentDescription = null, tint = BrandBlue, modifier = Modifier.size(16.dp))
                Spacer(modifier = Modifier.width(8.dp))
                Column {
                    Text("Route: $compactRoute", style = MaterialTheme.typography.bodySmall, fontWeight = FontWeight.Medium)
                    Text("${batch.orderCount} orders", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                }
            }

        }
    }
}

private fun compactRouteLabel(zoneLevel: String, fromCity: String, toCity: String): String {
    val from = fromCity.trim()
    val to = toCity.trim()
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

private fun statusBadgeStyle(status: String): Triple<Color, Color, String> {
    return when (status) {
        "assigned" -> Triple(BlueSoft, BrandBlue, "Assigned")
        "accepted" -> Triple(Color(0xFFE8F4FD), Color(0xFF1565C0), "Accepted")
        "picked_up" -> Triple(Color(0xFFE8F5E9), Color(0xFF2E7D32), "Picked Up")
        "delivered" -> Triple(Color(0xFFE8F5E9), SuccessGreen, "Delivered")
        else -> Triple(Color(0xFFF1F1F1), TextSecondary, status.replace("_", " ").replaceFirstChar { it.uppercase() })
    }
}

private fun formatBatchWeight(weight: Double): String {
    return String.format(Locale.getDefault(), "%.1f", weight)
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
