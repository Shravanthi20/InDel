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
import com.imaginai.indel.data.model.Order
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
    selectedTab: OrderLifecycleTab,
    onTabSelected: (OrderLifecycleTab) -> Unit,
    viewModel: OrdersViewModel,
    navController: NavController
) {
    val tabs = OrderLifecycleTab.entries
    val visibleOrders = when (selectedTab) {
        OrderLifecycleTab.AVAILABLE -> state.availableOrders
        OrderLifecycleTab.ASSIGNED -> state.assignedOrders
        OrderLifecycleTab.PICKED_UP -> state.pickedUpOrders
        OrderLifecycleTab.DELIVERED -> state.deliveredOrders
    }

    val tabCount = when (selectedTab) {
        OrderLifecycleTab.AVAILABLE -> state.availableOrders.size
        OrderLifecycleTab.ASSIGNED -> state.assignedOrders.size
        OrderLifecycleTab.PICKED_UP -> state.pickedUpOrders.size
        OrderLifecycleTab.DELIVERED -> state.deliveredOrders.size
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
                        OrderLifecycleTab.AVAILABLE -> state.availableOrders.size
                        OrderLifecycleTab.ASSIGNED -> state.assignedOrders.size
                        OrderLifecycleTab.PICKED_UP -> state.pickedUpOrders.size
                        OrderLifecycleTab.DELIVERED -> state.deliveredOrders.size
                    }
                    Tab(
                        selected = selectedTab == tab,
                        onClick = { onTabSelected(tab) },
                        text = { Text("${tab.title} ($count)") }
                    )
                }
            }
        }

        if (visibleOrders.isNotEmpty()) {
            item {
                Text(
                    "${selectedTab.title} Orders",
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold
                )
            }
            items(visibleOrders) { order ->
                OrderCard(order, viewModel, navController)
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
fun OrderCard(order: Order, viewModel: OrdersViewModel, navController: NavController) {
    val (statusBgColor, statusTextColor, statusLabel) = statusBadgeStyle(order.status)

    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                Column {
                    Text("Order #${order.orderId.takeLast(4)}", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                    Text("₹${order.earningInr.toInt()}", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold, color = BrandBlue)
                    if (order.tipInr > 0) {
                        Text("Incl. ₹${order.tipInr.toInt()} tip", style = MaterialTheme.typography.labelSmall, color = SuccessGreen, fontWeight = FontWeight.Bold)
                    } else {
                        Text("Tip: ₹0", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                    }
                }
                Surface(
                    color = BlueSoft,
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Text(
                        text = "${order.distanceKm} km",
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
                    Text("Pickup: ${order.pickupArea}", style = MaterialTheme.typography.bodySmall, fontWeight = FontWeight.Medium)
                    Text("Drop: ${order.dropArea}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                }
            }

            Spacer(modifier = Modifier.height(16.dp))
            
            val buttonText = when(order.status) {
                "assigned" -> "Accept Order"
                "accepted" -> "Mark Picked Up"
                "picked_up" -> "Complete Delivery"
                else -> "Complete"
            }

            if (order.status != "delivered") {
                Button(
                    onClick = {
                        when(order.status) {
                            "assigned" -> viewModel.acceptOrder(order.orderId)
                            "accepted" -> viewModel.pickedUpOrder(order.orderId)
                            "picked_up" -> navController.navigate(Screen.DeliveryCompletion.createRoute(order.orderId))
                        }
                    },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = when (order.status) {
                            "assigned" -> BrandBlue
                            "accepted" -> Color(0xFF1565C0)
                            else -> SuccessGreen
                        }
                    )
                ) {
                    Text(buttonText)
                }
            } else {
                Text(
                    "Completed",
                    color = SuccessGreen,
                    fontWeight = FontWeight.Bold,
                    modifier = Modifier.align(Alignment.CenterHorizontally)
                )
            }
        }
    }
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
