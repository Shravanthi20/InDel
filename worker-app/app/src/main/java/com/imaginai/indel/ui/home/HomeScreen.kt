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

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HomeScreen(
    navController: NavController,
    viewModel: HomeViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { 
                    Column {
                        Text("InDel", fontWeight = FontWeight.Bold, fontSize = 20.sp)
                        Text("Worker Partner", style = MaterialTheme.typography.labelSmall)
                    }
                },
                actions = {
                    IconButton(onClick = { /* TODO: Notifications */ }) {
                        Icon(Icons.Default.Notifications, contentDescription = "Notifications")
                    }
                    IconButton(onClick = { /* TODO: Profile */ }) {
                        Icon(Icons.Default.AccountCircle, contentDescription = "Profile")
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = BrandOrange,
                    titleContentColor = Color.White,
                    actionIconContentColor = Color.White
                )
            )
        }
    ) { padding ->
        Box(modifier = Modifier
            .padding(padding)
            .fillMaxSize()
            .background(BackgroundWarmWhite)
        ) {
            when (val state = uiState) {
                is HomeUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                is HomeUiState.Success -> HomeContent(state.worker, state.policy, state.earnings, navController)
                is HomeUiState.Error -> ErrorState(state.message) { viewModel.loadDashboard() }
            }
        }
    }
}

@Composable
fun HomeContent(
    worker: WorkerProfile,
    policy: Policy,
    earnings: Earnings,
    navController: NavController
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp)
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        // 1. Worker Status Card
        StatusCard(worker)

        // 2. Earnings Today Card
        DashboardCard(
            title = "Earnings Today",
            value = "₹${earnings.thisWeekActual}", // Simplified for demo
            subtitle = "Completed: 8 Orders",
            icon = Icons.Default.CurrencyRupee,
            onClick = { navController.navigate(Screen.Earnings.route) }
        )

        // 3. Protection Status Card
        DashboardCard(
            title = "Protection Status",
            value = if (policy.status == "active") "Protected" else "Not Enrolled",
            subtitle = "Coverage: ${(policy.coverageRatio * 100).toInt()}%",
            icon = Icons.Default.Shield,
            color = if (policy.status == "active") SuccessGreen else WarningAmber,
            onClick = { navController.navigate(Screen.Policy.route) }
        )

        // 4. Disruption Banner (Conditional)
        if (worker.coverageStatus == "at_risk") {
            DisruptionBanner()
        }

        // 5. Quick Navigation Grid
        Text("Services", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            NavBox("Orders", Icons.Default.DeliveryDining, Modifier.weight(1f)) { /* TODO: Orders */ }
            NavBox("Earnings", Icons.Default.Payments, Modifier.weight(1f)) { navController.navigate(Screen.Earnings.route) }
        }
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            NavBox("Policy", Icons.Default.Description, Modifier.weight(1f)) { navController.navigate(Screen.Policy.route) }
            NavBox("Claims", Icons.AutoMirrored.Filled.AssignmentReturn, Modifier.weight(1f)) { navController.navigate(Screen.Claims.route) }
        }
        
        Spacer(modifier = Modifier.height(32.dp))
    }
}

@Composable
fun StatusCard(worker: WorkerProfile) {
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
                    .background(OrangeSoft),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = if (worker.name.isNotEmpty()) worker.name.take(1) else "?",
                    fontWeight = FontWeight.Bold,
                    color = BrandOrange
                )
            }
            Spacer(modifier = Modifier.width(12.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(worker.name.ifEmpty { "Unknown Worker" }, fontWeight = FontWeight.Bold)
                Text(worker.zone, style = MaterialTheme.typography.bodySmall, color = TextSecondary)
            }
            // Online Toggle
            Column(horizontalAlignment = Alignment.End) {
                Text("Online", style = MaterialTheme.typography.labelSmall, fontWeight = FontWeight.Bold, color = SuccessGreen)
                Switch(checked = true, onCheckedChange = {}, colors = SwitchDefaults.colors(checkedThumbColor = SuccessGreen))
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
    color: Color = BrandOrange,
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
                .background(Brush.horizontalGradient(listOf(BrandOrange, OrangeDeep)))
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
            Icon(icon, contentDescription = null, tint = BrandOrange, modifier = Modifier.size(32.dp))
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
