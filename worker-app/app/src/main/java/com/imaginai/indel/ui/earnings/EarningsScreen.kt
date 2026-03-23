package com.imaginai.indel.ui.earnings

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.TrendingUp
import androidx.compose.material3.*
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
                    containerColor = BrandOrange,
                    titleContentColor = Color.White,
                    navigationIconContentColor = Color.White
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
                is EarningsUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                is EarningsUiState.Success -> EarningsContent(state.earnings)
                is EarningsUiState.Error -> Text(state.message, color = ErrorRed, modifier = Modifier.align(Alignment.Center))
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
                    Text("₹${earnings.thisWeekActual}", style = MaterialTheme.typography.headlineLarge, fontWeight = FontWeight.Bold, color = BrandOrange)
                    
                    Spacer(modifier = Modifier.height(16.dp))
                    HorizontalDivider(color = BackgroundWarmWhite)
                    Spacer(modifier = Modifier.height(16.dp))

                    Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                        EarningItem("Baseline", "₹${earnings.thisWeekBaseline}", TextSecondary)
                        EarningItem("Protected", "₹${earnings.protectedIncome}", SuccessGreen)
                    }
                }
            }
        }

        // 2. Insight Panel
        item {
            val loss = maxOf(0.0, earnings.thisWeekBaseline - earnings.thisWeekActual)
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(containerColor = OrangeSoft.copy(alpha = 0.5f))
            ) {
                Row(modifier = Modifier.padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
                    Icon(Icons.Default.Info, contentDescription = null, tint = BrandOrange)
                    Spacer(modifier = Modifier.width(12.dp))
                    Column {
                        Text("Income Insight", fontWeight = FontWeight.Bold, color = OrangeDeep)
                        Text(
                            if (loss > 0) "You have a gap of ₹$loss from your baseline due to external factors. Protection payout is being calculated." 
                            else "You are performing above your baseline. Great work!",
                            style = MaterialTheme.typography.bodySmall,
                            color = TextPrimary
                        )
                    }
                }
            }
        }

        item {
            Text("History", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
        }

        items(earnings.history) { record ->
            HistoryItem(record.date, record.amount)
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
fun HistoryItem(date: String, amount: Double) {
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
                modifier = Modifier.size(40.dp).background(BackgroundWarmWhite, CircleShape),
                contentAlignment = Alignment.Center
            ) {
                Icon(Icons.Default.TrendingUp, contentDescription = null, tint = SuccessGreen, modifier = Modifier.size(20.dp))
            }
            Spacer(modifier = Modifier.width(16.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(date, fontWeight = FontWeight.SemiBold)
                Text("Completed shift", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
            }
            Text("₹$amount", fontWeight = FontWeight.Bold, color = TextPrimary)
        }
    }
}
