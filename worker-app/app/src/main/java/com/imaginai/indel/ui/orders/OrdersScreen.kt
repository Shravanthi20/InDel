package com.imaginai.indel.ui.orders

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
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.ui.navigation.Screen
import com.imaginai.indel.ui.theme.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun OrdersScreen(
    navController: NavController,
    viewModel: OrdersViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val isRefreshing by viewModel.isRefreshing.collectAsState()
    val selectedTab by viewModel.selectedTab.collectAsState()

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
                        state = state,
                        selectedTab = selectedTab,
                        onTabSelected = viewModel::selectTab,
                        viewModel = viewModel,
                        navController = navController
                    )
                    is OrdersUiState.Error -> ErrorState(state.message) { viewModel.loadOrders() }
                }
            }
        }
    }
}

@Composable
fun OrdersContent(
    state: OrdersUiState.Success,
    selectedTab: BatchLifecycleTab,
    onTabSelected: (BatchLifecycleTab) -> Unit,
    viewModel: OrdersViewModel,
    navController: NavController
) {
    val tabs = BatchLifecycleTab.entries
    val visibleBatches = when (selectedTab) {
        BatchLifecycleTab.AVAILABLE_NEAR -> state.availableNearBatches
        BatchLifecycleTab.PICKED_UP -> state.pickedUpBatches
        BatchLifecycleTab.DELIVERY -> state.deliveryBatches
    }

    val tabCount = when (selectedTab) {
        BatchLifecycleTab.AVAILABLE_NEAR -> state.availableNearBatches.size
        BatchLifecycleTab.PICKED_UP -> state.pickedUpBatches.size
        BatchLifecycleTab.DELIVERY -> state.deliveryBatches.size
    }

    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        item {
            Text(
                "Batch Lifecycle",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.height(8.dp))
            TabRow(selectedTabIndex = tabs.indexOf(selectedTab), containerColor = Color.White) {
                tabs.forEach { tab ->
                    val count = when (tab) {
                        BatchLifecycleTab.AVAILABLE_NEAR -> state.availableNearBatches.size
                        BatchLifecycleTab.PICKED_UP -> state.pickedUpBatches.size
                        BatchLifecycleTab.DELIVERY -> state.deliveryBatches.size
                    }
                    Tab(
                        selected = selectedTab == tab,
                        onClick = { onTabSelected(tab) },
                        text = { Text("${tab.title} ($count)") }
                    )
                }
            }
        }

        if (visibleBatches.isNotEmpty()) {
            item {
                Text(
                    "${selectedTab.title} Batches",
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold
                )
            }
            items(visibleBatches) { batch ->
                BatchCard(batch = batch, selectedTab = selectedTab, navController = navController)
            }
        }

        if (tabCount == 0) {
            item {
                Box(modifier = Modifier.fillParentMaxSize(), contentAlignment = Alignment.Center) {
                    Text("No ${selectedTab.title.lowercase()} orders right now", color = TextSecondary)
                }
            }
        }
    }
}

@Composable
fun BatchCard(batch: DeliveryBatch, selectedTab: BatchLifecycleTab, navController: NavController) {
    val normalizedStatus = batch.status.trim().lowercase().replace(" ", "_")
    val statusLabel = when (normalizedStatus) {
        "picked_up" -> "Picked Up"
        "delivered" -> "Delivered"
        "assigned", "accepted" -> "Available"
        else -> batch.status
    }
    val isZoneASingleStop = batch.zoneLevel.equals("A", ignoreCase = true) && batch.fromCity.equals(batch.toCity, ignoreCase = true)
    val routeLabel = if (isZoneASingleStop) batch.fromCity else "${batch.fromCity} -> ${batch.toCity}"
    val actionLabel = when (selectedTab) {
        BatchLifecycleTab.AVAILABLE_NEAR -> "Pick Up"
        BatchLifecycleTab.PICKED_UP -> "Make Delivery"
        BatchLifecycleTab.DELIVERY -> "View"
    }
    val actionColor = when (selectedTab) {
        BatchLifecycleTab.AVAILABLE_NEAR -> BrandBlue
        BatchLifecycleTab.PICKED_UP -> Color(0xFF2E7D32)
        BatchLifecycleTab.DELIVERY -> TextSecondary
    }

    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                Column {
                    Text("Batch #${batch.batchId.takeLast(6)}", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                    Text(routeLabel, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold, color = BrandBlue)
                    Text("Zone ${batch.zoneLevel} • ${batch.orderCount} orders", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                }
                Surface(
                    color = BlueSoft,
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Text(
                        text = "${String.format("%.1f", batch.totalWeight)} kg",
                        modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp),
                        style = MaterialTheme.typography.labelMedium,
                        fontWeight = FontWeight.Bold,
                        color = BrandBlue
                    )
                }
            }

            Spacer(modifier = Modifier.height(10.dp))

            Surface(color = BlueSoft, shape = RoundedCornerShape(999.dp)) {
                Text(
                    text = statusLabel,
                    modifier = Modifier.padding(horizontal = 10.dp, vertical = 4.dp),
                    style = MaterialTheme.typography.labelMedium,
                    fontWeight = FontWeight.Bold,
                    color = BrandBlue
                )
            }

            Spacer(modifier = Modifier.height(16.dp))

            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(Icons.Default.LocationOn, contentDescription = null, tint = BrandBlue, modifier = Modifier.size(16.dp))
                Spacer(modifier = Modifier.width(8.dp))
                Column {
                    Text("Pickup: ${batch.fromCity}", style = MaterialTheme.typography.bodySmall, fontWeight = FontWeight.Medium)
                    Text("Drop: ${batch.toCity}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            if (normalizedStatus != "delivered") {
                Button(
                    onClick = { navController.navigate(Screen.BatchDetail.createRoute(batch.batchId)) },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = actionColor)
                ) {
                    Text(actionLabel)
                }
            } else {
                Text(
                    "Delivered",
                    color = SuccessGreen,
                    fontWeight = FontWeight.Bold,
                    modifier = Modifier.align(Alignment.CenterHorizontally)
                )
            }
        }
    }
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
