package com.imaginai.indel.ui.home

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.AssignmentReturn
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.data.model.Earnings
import com.imaginai.indel.data.model.Policy
import com.imaginai.indel.data.model.WorkerProfile
import com.imaginai.indel.ui.navigation.Screen
import com.imaginai.indel.ui.theme.*
import java.text.SimpleDateFormat
import java.util.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HomeScreen(
    navController: NavController,
    viewModel: HomeViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val isRefreshing by viewModel.isRefreshing.collectAsState()
    val isOnline by viewModel.isOnline.collectAsState()
    val lastUpdated = remember { SimpleDateFormat("hh:mm a", Locale.getDefault()).format(Date()) }

    // Refresh profile data when screen is displayed (enables fresh data after profile edit)
    LaunchedEffect(Unit) {
        viewModel.loadDashboard()
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { 
                    Column {
                        Text("InDel", fontWeight = FontWeight.Bold, fontSize = 20.sp)
                        Text("Last updated: $lastUpdated", style = MaterialTheme.typography.labelSmall)
                    }
                },
                actions = {
                    IconButton(onClick = { navController.navigate(Screen.Notifications.route) }) {
                        Icon(Icons.Default.Notifications, contentDescription = "Notifications")
                    }
                    IconButton(onClick = { navController.navigate(Screen.ProfileEdit.route) }) {
                        Icon(Icons.Default.AccountCircle, contentDescription = "Profile")
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = BrandBlue,
                    titleContentColor = Color.White,
                    actionIconContentColor = Color.White
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
                    is HomeUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                    is HomeUiState.Success -> HomeContent(state.worker, state.policy, state.earnings, isOnline, navController, viewModel)
                    is HomeUiState.Error -> ErrorState(state.message) { viewModel.loadDashboard() }
                }
            }
        }
    }
}

@Composable
fun HomeContent(
    worker: WorkerProfile,
    policy: Policy,
    earnings: Earnings,
    isOnline: Boolean,
    navController: NavController,
    viewModel: HomeViewModel
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp)
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        // 1. Worker Status Card
        StatusCard(worker, isOnline) { viewModel.toggleOnlineStatus(it) }

        // 2. Earnings Today Card
        DashboardCard(
            title = "Earnings Today",
            value = "₹${earnings.thisWeekActual.toInt()}",
            subtitle = "View earnings history",
            icon = Icons.Default.CurrencyRupee,
            onClick = { navController.navigate(Screen.Earnings.route) }
        )

        // 3. Protection Status Card
        DashboardCard(
            title = "Protection Status",
            value = if (policy.status == "active") "Protected" else "Not Enrolled",
            subtitle = buildString {
                append("Coverage: ")
                append((policy.coverageRatio * 100).toInt())
                append('%')
                if (!policy.planName.isNullOrBlank()) {
                    append(" • ")
                    append(policy.planName)
                }
            },
            icon = Icons.Default.Shield,
            color = if (policy.status == "active") SuccessGreen else WarningAmber,
            onClick = {
                navController.navigate(Screen.Policy.route) {
                    launchSingleTop = true
                }
            }
        )

        if (!policy.planName.isNullOrBlank()) {
            DashboardCard(
                title = "Selected Plan",
                value = policy.planName,
                subtitle = buildString {
                    append("Deliveries/week: ")
                    append(policy.selectedDeliveries ?: policy.rangeStart ?: 0)
                    if (policy.rangeStart != null && policy.rangeEnd != null) {
                        append(" (range ")
                        append(policy.rangeStart)
                        append('-')
                        append(policy.rangeEnd)
                        append(')')
                    }
                },
                icon = Icons.Default.LocalOffer,
                color = BrandBlue,
                onClick = {
                    navController.navigate(Screen.Policy.route) {
                        launchSingleTop = true
                    }
                }
            )
        }

        // 4. Disruption Banner (Conditional)
        if (worker.coverageStatus == "at_risk") {
            DisruptionBanner()
        }

        // 5. Quick Navigation Grid
        Text("Services", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            NavBox("Orders", Icons.Default.DeliveryDining, Modifier.weight(1f)) { navController.navigate(Screen.Orders.route) }
            NavBox("Earnings", Icons.Default.Payments, Modifier.weight(1f)) { navController.navigate(Screen.Earnings.route) }
        }
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            NavBox("Policy", Icons.Default.Description, Modifier.weight(1f)) {
                navController.navigate(Screen.Policy.route) {
                    launchSingleTop = true
                }
            }
            NavBox("Claims", Icons.AutoMirrored.Filled.AssignmentReturn, Modifier.weight(1f)) { navController.navigate(Screen.Claims.route) }
        }
        
        Spacer(modifier = Modifier.height(32.dp))
    }
}

@Composable
fun StatusCard(worker: WorkerProfile, isOnline: Boolean, onToggle: (Boolean) -> Unit) {
    val displayName = worker.name?.trim().orEmpty().ifBlank { "Unknown Worker" }
    val zoneLevelText = worker.zoneLevel.trim()
    val zoneNameText = worker.zoneName.trim()
    val cityText = worker.city?.trim().orEmpty()
    val zoneText = worker.zone?.trim().orEmpty()
    val displayZone = when {
        zoneLevelText.isNotBlank() && zoneNameText.isNotBlank() -> "Zone ${zoneLevelText.uppercase()} - $zoneNameText"
        zoneNameText.isNotBlank() && cityText.isNotBlank() -> "$zoneNameText, $cityText"
        zoneNameText.isNotBlank() -> zoneNameText
        cityText.isNotBlank() -> cityText
        zoneText.isNotBlank() -> zoneText
        else -> "Zone unavailable"
    }

    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(
                modifier = Modifier
                    .size(48.dp)
                    .clip(CircleShape)
                    .background(BlueSoft),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = displayName.take(1),
                    fontWeight = FontWeight.Bold,
                    color = BrandBlue
                )
            }
            Spacer(modifier = Modifier.width(12.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(displayName, fontWeight = FontWeight.Bold)
                Text(displayZone, style = MaterialTheme.typography.bodySmall, color = TextSecondary)
            }
            // Online Toggle
            Column(horizontalAlignment = Alignment.End) {
                Text(
                    text = if (isOnline) "Online" else "Offline",
                    style = MaterialTheme.typography.labelSmall,
                    fontWeight = FontWeight.Bold,
                    color = if (isOnline) SuccessGreen else TextSecondary
                )
                Switch(
                    checked = isOnline,
                    onCheckedChange = onToggle,
                    colors = SwitchDefaults.colors(
                        checkedThumbColor = SuccessGreen,
                        checkedTrackColor = SuccessGreen.copy(alpha = 0.5f)
                    )
                )
            }
        }
    }
}

@Composable
fun DashboardCard(
    title: String,
    value: String,
    subtitle: String,
    icon: ImageVector,
    color: Color = BrandBlue,
    onClick: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth().clickable { onClick() },
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Row(modifier = Modifier.padding(20.dp), verticalAlignment = Alignment.CenterVertically) {
            Column(modifier = Modifier.weight(1f)) {
                Text(title, style = MaterialTheme.typography.labelLarge, color = TextSecondary)
                Text(value, style = MaterialTheme.typography.headlineMedium, fontWeight = FontWeight.Bold, color = color)
                Text(subtitle, style = MaterialTheme.typography.bodySmall, color = TextSecondary)
            }
            Icon(icon, contentDescription = null, tint = color.copy(alpha = 0.6f), modifier = Modifier.size(40.dp))
        }
    }
}

@Composable
fun DisruptionBanner() {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(containerColor = Color.Transparent)
    ) {
        Box(
            modifier = Modifier
                .background(Brush.horizontalGradient(listOf(BrandBlue, BlueDeep)))
                .padding(16.dp)
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(Icons.Default.Warning, contentDescription = null, tint = Color.White)
                Spacer(modifier = Modifier.width(12.dp))
                Column {
                    Text("Heavy Rain Alert - Tambaram", color = Color.White, fontWeight = FontWeight.Bold)
                    Text("Your income is protected. Stay safe.", color = Color.White.copy(alpha = 0.9f), fontSize = 12.sp)
                }
            }
        }
    }
}

@Composable
fun NavBox(title: String, icon: ImageVector, modifier: Modifier = Modifier, onClick: () -> Unit) {
    Card(
        modifier = modifier.height(110.dp).clickable { onClick() },
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(
            modifier = Modifier.fillMaxSize(),
            verticalArrangement = Arrangement.Center,
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Icon(icon, contentDescription = null, tint = BrandBlue, modifier = Modifier.size(32.dp))
            Spacer(modifier = Modifier.height(8.dp))
            Text(title, fontWeight = FontWeight.SemiBold, fontSize = 14.sp)
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
