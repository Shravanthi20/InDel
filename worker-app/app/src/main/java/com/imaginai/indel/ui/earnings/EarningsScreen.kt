package com.imaginai.indel.ui.earnings

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.data.model.EarningsSummary

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
                title = { Text("Earnings") },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                }
            )
        }
    ) { padding ->
        Box(modifier = Modifier.padding(padding).fillMaxSize()) {
            when (val state = uiState) {
                is EarningsUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                is EarningsUiState.Success -> EarningsContent(state.earnings)
                is EarningsUiState.Error -> Text(state.message, color = Color.Red, modifier = Modifier.align(Alignment.Center))
            }
        }
    }
}

@Composable
fun EarningsContent(earnings: EarningsSummary) {
    LazyColumn(
        modifier = Modifier.fillMaxSize().padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        item {
            Card(modifier = Modifier.fillMaxWidth()) {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("This Week Actual", style = MaterialTheme.typography.labelMedium)
                    Text("₹${earnings.thisWeekActual}", style = MaterialTheme.typography.headlineLarge, color = MaterialTheme.colorScheme.primary)
                    Spacer(modifier = Modifier.height(8.dp))
                    Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                        Column {
                            Text("Baseline", style = MaterialTheme.typography.labelSmall)
                            Text("₹${earnings.thisWeekBaseline}", style = MaterialTheme.typography.titleMedium)
                        }
                        Column(horizontalAlignment = Alignment.End) {
                            Text("Protected", style = MaterialTheme.typography.labelSmall)
                            Text("₹${earnings.protectedIncome}", style = MaterialTheme.typography.titleMedium, color = Color(0xFF2E7D32))
                        }
                    }
                }
            }
        }

        item {
            Text("Weekly History", style = MaterialTheme.typography.titleMedium)
        }

        items(earnings.history) { item ->
            ListItem(
                headlineContent = { Text(item.week) },
                supportingContent = { Text("Baseline: ₹${item.baseline}") },
                trailingContent = { Text("₹${item.actual}", style = MaterialTheme.typography.titleMedium) }
            )
            HorizontalDivider()
        }
    }
}
