package com.imaginai.indel.ui.earnings

import androidx.compose.animation.animateColorAsState
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.TrendingDown
import androidx.compose.material.icons.automirrored.filled.TrendingUp
import androidx.compose.material.icons.filled.Info
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
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.data.model.Earnings
import com.imaginai.indel.ui.theme.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun EarningsScreen(
    navController: NavController,
    viewModel: EarningsViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val isRefreshing by viewModel.isRefreshing.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Earnings Insight", fontWeight = FontWeight.Bold) },
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
                    is EarningsUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                    is EarningsUiState.Success -> EarningsContent(state.earnings)
                    is EarningsUiState.Error -> ErrorState(state.message) { viewModel.loadEarnings() }
                }
            }
        }
    }
}

@Composable
fun EarningsContent(earnings: Earnings) {
    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        // 1. KPI Row Card
        item {
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(containerColor = Color.White),
                elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
            ) {
                Column(modifier = Modifier.padding(20.dp)) {
                    Text("Weekly Earnings", style = MaterialTheme.typography.labelLarge, color = TextSecondary)
                    Text("₹${earnings.thisWeekActual.toInt()}", style = MaterialTheme.typography.headlineLarge, fontWeight = FontWeight.Bold, color = BrandBlue)
                    
                    Spacer(modifier = Modifier.height(16.dp))
                    HorizontalDivider(color = BackgroundWarmWhite)
                    Spacer(modifier = Modifier.height(16.dp))

                    Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                        EarningItem("Baseline", "₹${earnings.thisWeekBaseline.toInt()}", TextSecondary)
                        EarningItem("Protected", "₹${earnings.protectedIncome.toInt()}", SuccessGreen)
                    }
                }
            }
        }

        // 2. Insight Panel
        item {
            val gap = earnings.thisWeekBaseline - earnings.thisWeekActual
            val isGapped = gap > 0
            
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(
                    containerColor = if (isGapped) BlueSoft.copy(alpha = 0.5f) else SuccessGreen.copy(alpha = 0.1f)
                )
            ) {
                Row(modifier = Modifier.padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        if (isGapped) Icons.Default.Info else Icons.AutoMirrored.Filled.TrendingUp,
                        contentDescription = null,
                        tint = if (isGapped) BrandBlue else SuccessGreen
                    )
                    Spacer(modifier = Modifier.width(12.dp))
                    Column {
                        Text(
                            "Income Insight",
                            fontWeight = FontWeight.Bold,
                            color = if (isGapped) BlueDeep else SuccessGreen
                        )
                        Text(
                            if (isGapped) {
                                "Due to reduced activity (likely disruption), your earnings dropped by ₹${gap.toInt()}. Your protection ensures you receive ₹${earnings.protectedIncome.toInt()}."
                            } else {
                                "You've exceeded your baseline by ₹${(-gap).toInt()}. Great job! Your protection remains active for next week."
                            },
                            style = MaterialTheme.typography.bodySmall,
                            color = TextPrimary
                        )
                    }
                }
            }
        }

        item {
            Text("Weekly History", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
        }

        items(earnings.history) { record ->
            HistoryItem(record.date, record.amount, earnings.thisWeekBaseline)
        }
    }
}

@Composable
fun EarningItem(label: String, value: String, valueColor: Color) {
    Column {
        Text(label, style = MaterialTheme.typography.labelSmall, color = TextSecondary)
        Text(value, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold, color = valueColor)
    }
}

@Composable
fun HistoryItem(date: String, amount: Double, baseline: Double) {
    val isAboveBaseline = amount >= baseline
    
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White)
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(
                modifier = Modifier.size(40.dp).background(
                    if (isAboveBaseline) SuccessGreen.copy(alpha = 0.1f) else ErrorRed.copy(alpha = 0.1f),
                    CircleShape
                ),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    if (isAboveBaseline) Icons.AutoMirrored.Filled.TrendingUp else Icons.AutoMirrored.Filled.TrendingDown,
                    contentDescription = null,
                    tint = if (isAboveBaseline) SuccessGreen else ErrorRed,
                    modifier = Modifier.size(20.dp)
                )
            }
            Spacer(modifier = Modifier.width(16.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(date, fontWeight = FontWeight.SemiBold)
                Text(
                    if (isAboveBaseline) "Above Baseline" else "Below Baseline",
                    style = MaterialTheme.typography.labelSmall,
                    color = if (isAboveBaseline) SuccessGreen else ErrorRed
                )
            }
            Text("₹${amount.toInt()}", fontWeight = FontWeight.Bold, color = TextPrimary)
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
