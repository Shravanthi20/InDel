package com.imaginai.indel.ui.notifications

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.AccountBalanceWallet
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.data.model.Notification
import com.imaginai.indel.ui.navigation.Screen
import com.imaginai.indel.ui.theme.BackgroundWarmWhite
import com.imaginai.indel.ui.theme.BlueSoft
import com.imaginai.indel.ui.theme.BrandBlue
import com.imaginai.indel.ui.theme.ErrorRed
import com.imaginai.indel.ui.theme.SuccessGreen
import com.imaginai.indel.ui.theme.TextSecondary
import com.imaginai.indel.ui.theme.WarningAmber

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NotificationsScreen(
    navController: NavController,
    viewModel: NotificationsViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val isRefreshing by viewModel.isRefreshing.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Alerts & Payout Updates", fontWeight = FontWeight.Bold) },
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
                    is NotificationsUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                    is NotificationsUiState.Success -> NotificationsContent(state.notifications, navController)
                    is NotificationsUiState.Error -> NotificationsErrorState(state.message) { viewModel.loadNotifications() }
                }
            }
        }
    }
}

@Composable
private fun NotificationsContent(
    notifications: List<Notification>,
    navController: NavController
) {
    val disruptionCount = notifications.count { it.type == "disruption_alert" }
    val claimCount = notifications.count { it.type == "claim_generated" }
    val payoutCount = notifications.count { it.type == "payout_credited" }

    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        item {
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                SummaryCard(
                    title = "Disruption Alerts",
                    value = disruptionCount.toString(),
                    subtitle = "Auto protection started",
                    color = WarningAmber,
                    modifier = Modifier.weight(1f)
                )
                SummaryCard(
                    title = "Claims Generated",
                    value = claimCount.toString(),
                    subtitle = "Amount ready for review",
                    color = BlueSoft,
                    modifier = Modifier.weight(1f)
                )
                SummaryCard(
                    title = "Payout Credits",
                    value = payoutCount.toString(),
                    subtitle = "Auto payouts completed",
                    color = SuccessGreen,
                    modifier = Modifier.weight(1f)
                )
            }
        }

        item {
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(containerColor = Color.White),
                elevation = CardDefaults.cardElevation(defaultElevation = 2.dp),
                onClick = { navController.navigate(Screen.PayoutHistory.route) }
            ) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(16.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Box(
                        modifier = Modifier
                            .size(44.dp)
                            .background(BlueSoft, CircleShape),
                        contentAlignment = Alignment.Center
                    ) {
                        Icon(Icons.Default.AccountBalanceWallet, contentDescription = null, tint = BrandBlue)
                    }
                    Column(modifier = Modifier.weight(1f).padding(start = 12.dp)) {
                        Text("Open payout history", fontWeight = FontWeight.Bold)
                        Text("See all credited automatic payouts", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                    }
                }
            }
        }

        item {
            Text("Recent notifications", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
        }

        if (notifications.isEmpty()) {
            item {
                Box(modifier = Modifier.fillMaxWidth().padding(vertical = 48.dp), contentAlignment = Alignment.Center) {
                    Text("No notifications yet", color = TextSecondary)
                }
            }
        } else {
            items(notifications) { notification ->
                NotificationCard(notification)
            }
        }

        item {
            Spacer(modifier = Modifier.height(24.dp))
        }
    }
}

@Composable
private fun SummaryCard(
    title: String,
    value: String,
    subtitle: String,
    color: Color,
    modifier: Modifier = Modifier
) {
    Card(
        modifier = modifier,
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(title, style = MaterialTheme.typography.labelLarge, color = TextSecondary)
            Text(value, fontSize = 28.sp, fontWeight = FontWeight.Bold, color = color)
            Text(subtitle, style = MaterialTheme.typography.bodySmall, color = TextSecondary)
        }
    }
}

@Composable
private fun NotificationCard(notification: Notification) {
    val icon: ImageVector
    val accent: Color
    when (notification.type.lowercase()) {
        "payout_credited" -> {
            icon = Icons.Default.AccountBalanceWallet
            accent = SuccessGreen
        }
        "claim_generated" -> {
            icon = Icons.Default.Notifications
            accent = BrandBlue
        }
        "disruption_alert" -> {
            icon = Icons.Default.Warning
            accent = WarningAmber
        }
        else -> {
            icon = Icons.Default.Notifications
            accent = BrandBlue
        }
    }

    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Box(
                modifier = Modifier
                    .size(44.dp)
                    .background(accent.copy(alpha = 0.14f), CircleShape),
                contentAlignment = Alignment.Center
            ) {
                Icon(icon, contentDescription = null, tint = accent)
            }
            Column(modifier = Modifier.weight(1f)) {
                Text(notification.title, fontWeight = FontWeight.Bold)
                Text(notification.body, style = MaterialTheme.typography.bodyMedium, color = TextSecondary)
                Spacer(modifier = Modifier.height(8.dp))
                Text(notification.createdAt.take(19).replace("T", " "), style = MaterialTheme.typography.labelSmall, color = TextSecondary)
            }
        }
    }
}

@Composable
private fun NotificationsErrorState(message: String, onRetry: () -> Unit) {
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
