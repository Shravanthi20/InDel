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
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.R
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
                title = { Text(stringResource(R.string.earnings_insight), fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = stringResource(R.string.back))
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = BrandPink,
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
                    is EarningsUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center), color = BrandPink)
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
                shape = RoundedCornerShape(24.dp), // More premium rounded corners
                colors = CardDefaults.cardColors(containerColor = Color.White),
                elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
            ) {
                Column(modifier = Modifier.padding(24.dp)) {
                    Text(stringResource(R.string.weekly_earnings), style = MaterialTheme.typography.labelLarge, color = TextSecondary, fontWeight = FontWeight.Bold, letterSpacing = 1.sp)
                    Text("₹${earnings.thisWeekActual.toInt()}", style = MaterialTheme.typography.headlineLarge, fontWeight = FontWeight.ExtraBold, color = BrandPink)
                    
                    Spacer(modifier = Modifier.height(20.dp))
                    HorizontalDivider(color = PinkSoft, thickness = 1.dp)
                    Spacer(modifier = Modifier.height(20.dp))

                    Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                        EarningItem(stringResource(R.string.baseline), "₹${earnings.thisWeekBaseline.toInt()}", TextSecondary)
                        EarningItem(stringResource(R.string.status_protected), "₹${earnings.protectedIncome.toInt()}", SuccessGreen)
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
                    containerColor = if (isGapped) PinkSoft else SuccessGreen.copy(alpha = 0.05f)
                ),
                border = if (isGapped) null else androidx.compose.foundation.BorderStroke(1.dp, SuccessGreen.copy(alpha = 0.2f))
            ) {
                Row(modifier = Modifier.padding(20.dp), verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        if (isGapped) Icons.Default.Info else Icons.AutoMirrored.Filled.TrendingUp,
                        contentDescription = null,
                        tint = if (isGapped) BrandPink else SuccessGreen
                    )
                    Spacer(modifier = Modifier.width(16.dp))
                    Column {
                        Text(
                            stringResource(R.string.income_insight),
                            fontWeight = FontWeight.Bold,
                            color = if (isGapped) PinkDeep else SuccessGreen,
                            style = MaterialTheme.typography.titleSmall
                        )
                        Text(
                            text = earnings.insight ?: if (isGapped) {
                                stringResource(R.string.earnings_insight_gap, gap.toInt(), earnings.protectedIncome.toInt())
                            } else {
                                stringResource(R.string.earnings_insight_above, (-gap).toInt())
                            },
                            style = MaterialTheme.typography.bodySmall,
                            color = TextPrimary
                        )
                    }
                }
            }
        }

        item {
            Text(stringResource(R.string.weekly_history), style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
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
                    if (isAboveBaseline) stringResource(R.string.above_baseline) else stringResource(R.string.below_baseline),
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
            Text(stringResource(R.string.retry))
        }
    }
}
