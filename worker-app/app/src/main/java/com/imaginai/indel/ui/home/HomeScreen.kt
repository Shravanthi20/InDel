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
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.R
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

    Scaffold(
        topBar = {
            TopAppBar(
                title = { 
                    Column {
                        Text("InDel", fontWeight = FontWeight.Bold, fontSize = 20.sp)
                        Text(stringResource(R.string.last_updated, lastUpdated), style = MaterialTheme.typography.labelSmall)
                    }
                },
                actions = {
                    IconButton(onClick = { navController.navigate(Screen.Notifications.route) }) {
                        Icon(Icons.Default.Notifications, contentDescription = stringResource(R.string.notifications))
                    }
                    IconButton(onClick = { navController.navigate(Screen.ProfileEdit.route) }) {
                        Icon(Icons.Default.AccountCircle, contentDescription = stringResource(R.string.profile))
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = BrandPink,
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
                    is HomeUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center), color = BrandPink)
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
            title = stringResource(R.string.earnings_today),
            value = "₹${earnings.todayEarnings.toInt()}",
            subtitle = stringResource(R.string.completed_orders, worker.ordersCompleted ?: 0),
            icon = Icons.Default.CurrencyRupee,
            onClick = { navController.navigate(Screen.Earnings.route) }
        )

        // 3. Protection Status Card
        DashboardCard(
            title = stringResource(R.string.protection_status),
            value = if (policy.status == "active") stringResource(R.string.status_protected) else stringResource(R.string.not_enrolled),
            subtitle = stringResource(R.string.coverage_percent, (policy.coverageRatio * 100).toInt()),
            icon = Icons.Default.Shield,
            color = if (policy.status == "active") SuccessGreen else WarningAmber,
            onClick = { navController.navigate(Screen.Policy.route) }
        )

        // 3.a. Protected Payouts (if any)
        if (earnings.protectedIncome > 0) {
            DashboardCard(
                title = stringResource(R.string.protected_payouts),
                value = "₹${earnings.protectedIncome.toInt()}",
                subtitle = stringResource(R.string.auto_processed_claims),
                icon = Icons.Default.VerifiedUser,
                color = BrandPink,
                onClick = { navController.navigate(Screen.Claims.route) }
            )
        }

        // 4. Disruption Banner (Conditional)
        if (worker.coverageStatus == "at_risk") {
            DisruptionBanner()
        }

        // 5. Quick Navigation Grid
        Text(stringResource(R.string.services), style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold, letterSpacing = 0.5.sp)
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            NavBox(stringResource(R.string.orders), Icons.Default.DeliveryDining, Modifier.weight(1f)) { navController.navigate(Screen.Orders.route) }
            NavBox(stringResource(R.string.earnings), Icons.Default.Payments, Modifier.weight(1f)) { navController.navigate(Screen.Earnings.route) }
        }
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            NavBox(stringResource(R.string.policy), Icons.Default.Description, Modifier.weight(1f)) { navController.navigate(Screen.Policy.route) }
            NavBox(stringResource(R.string.claims), Icons.AutoMirrored.Filled.AssignmentReturn, Modifier.weight(1f)) { navController.navigate(Screen.Claims.route) }
        }
        
        Spacer(modifier = Modifier.height(32.dp))
    }
}

@Composable
fun StatusCard(worker: WorkerProfile, isOnline: Boolean, onToggle: (Boolean) -> Unit) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(20.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Row(
            modifier = Modifier.padding(20.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(
                modifier = Modifier
                    .size(56.dp)
                    .clip(CircleShape)
                    .background(PinkSoft),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = if (!worker.name.isNullOrEmpty()) worker.name!!.take(1) else "?",
                    fontWeight = FontWeight.Black,
                    fontSize = 20.sp,
                    color = BrandPink
                )
            }
            Spacer(modifier = Modifier.width(16.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(worker.name ?: stringResource(R.string.unknown_worker), fontWeight = FontWeight.ExtraBold, style = MaterialTheme.typography.titleMedium)
                Text("${worker.zoneLevel} - ${worker.zoneName}", style = MaterialTheme.typography.labelMedium, color = TextSecondary, fontWeight = FontWeight.Bold)
            }
            // Online Toggle
            Column(horizontalAlignment = Alignment.End) {
                Text(
                    text = if (isOnline) stringResource(R.string.online) else stringResource(R.string.offline),
                    style = MaterialTheme.typography.labelSmall,
                    fontWeight = FontWeight.Black,
                    letterSpacing = 1.sp,
                    color = if (isOnline) SuccessGreen else TextSecondary
                )
                Switch(
                    checked = isOnline,
                    onCheckedChange = onToggle,
                    colors = SwitchDefaults.colors(
                        checkedThumbColor = Color.White,
                        checkedTrackColor = SuccessGreen,
                        uncheckedThumbColor = Color.White,
                        uncheckedTrackColor = Color.LightGray
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
    color: Color = BrandPink,
    onClick: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth().clickable { onClick() },
        shape = RoundedCornerShape(20.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Row(modifier = Modifier.padding(24.dp), verticalAlignment = Alignment.CenterVertically) {
            Column(modifier = Modifier.weight(1f)) {
                Text(title, style = MaterialTheme.typography.labelLarge, color = TextSecondary, fontWeight = FontWeight.Bold)
                Text(value, style = MaterialTheme.typography.headlineMedium, fontWeight = FontWeight.Black, color = color)
                Text(subtitle, style = MaterialTheme.typography.bodySmall, color = TextSecondary, fontWeight = FontWeight.Medium)
            }
            Icon(icon, contentDescription = null, tint = color.copy(alpha = 0.15f), modifier = Modifier.size(52.dp))
        }
    }
}

@Composable
fun DisruptionBanner() {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        elevation = CardDefaults.cardElevation(defaultElevation = 0.dp)
    ) {
        Box(
            modifier = Modifier
                .background(Brush.horizontalGradient(listOf(PinkDeep, BrandPink)))
                .padding(20.dp)
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(Icons.Default.Warning, contentDescription = null, tint = Color.White, modifier = Modifier.size(28.dp))
                Spacer(modifier = Modifier.width(16.dp))
                Column {
                    Text(stringResource(R.string.heavy_rain_alert_tambaram), color = Color.White, fontWeight = FontWeight.ExtraBold, style = MaterialTheme.typography.titleSmall)
                    Text(stringResource(R.string.income_protected_stay_safe), color = Color.White.copy(alpha = 0.9f), fontSize = 12.sp, fontWeight = FontWeight.Medium)
                }
            }
        }
    }
}

@Composable
fun NavBox(title: String, icon: ImageVector, modifier: Modifier = Modifier, onClick: () -> Unit) {
    Card(
        modifier = modifier.height(110.dp).clickable { onClick() },
        shape = RoundedCornerShape(20.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        elevation = CardDefaults.cardElevation(defaultElevation = 4.dp)
    ) {
        Column(
            modifier = Modifier.fillMaxSize(),
            verticalArrangement = Arrangement.Center,
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Icon(icon, contentDescription = null, tint = BrandPink, modifier = Modifier.size(36.dp))
            Spacer(modifier = Modifier.height(10.dp))
            Text(title, fontWeight = FontWeight.Bold, fontSize = 14.sp, color = TextPrimary)
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
            Text(stringResource(R.string.retry))
        }
    }
}
