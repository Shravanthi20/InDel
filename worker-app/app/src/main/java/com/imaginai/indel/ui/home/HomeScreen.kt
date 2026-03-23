package com.imaginai.indel.ui.home

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AccountCircle
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
import com.imaginai.indel.data.model.Policy
import com.imaginai.indel.data.model.WorkerProfile
import com.imaginai.indel.ui.navigation.Screen

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HomeScreen(
    navController: NavController,
    viewModel: HomeViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(title = { Text("InDel Worker") })
        }
    ) { padding ->
        Box(modifier = Modifier.padding(padding).fillMaxSize()) {
            when (val state = uiState) {
                is HomeUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                is HomeUiState.Success -> HomeContent(state.worker, state.policy, state.earnings, navController)
                is HomeUiState.Error -> Text(state.message, color = Color.Red, modifier = Modifier.align(Alignment.Center))
            }
        }
    }
}

@Composable
fun HomeContent(
    worker: WorkerProfile,
    policy: Policy,
    earnings: EarningsSummary,
    navController: NavController
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp)
    ) {
        // Profile Summary
        Card(modifier = Modifier.fillMaxWidth()) {
            Row(modifier = Modifier.padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
                Icon(Icons.Default.AccountCircle, contentDescription = null, modifier = Modifier.size(48.dp))
                Spacer(modifier = Modifier.width(16.dp))
                Column {
                    Text(worker.name, style = MaterialTheme.typography.titleLarge)
                    Text(worker.zone, style = MaterialTheme.typography.bodyMedium)
                }
            }
        }

        Spacer(modifier = Modifier.height(16.dp))

        // Coverage Badge
        Card(
            modifier = Modifier.fillMaxWidth(),
            colors = CardDefaults.cardColors(
                containerColor = if (worker.coverageStatus == "active") Color(0xFFE8F5E9) else Color(0xFFFFEBEE)
            )
        ) {
            Text(
                text = "Coverage: ${worker.coverageStatus.uppercase()}",
                modifier = Modifier.padding(16.dp),
                style = MaterialTheme.typography.labelLarge,
                color = if (worker.coverageStatus == "active") Color(0xFF2E7D32) else Color(0xFFC62828)
            )
        }

        Spacer(modifier = Modifier.height(16.dp))

        // Earnings Strip
        Card(
            modifier = Modifier.fillMaxWidth().clickable { navController.navigate(Screen.Earnings.route) }
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("This Week Earnings", style = MaterialTheme.typography.labelMedium)
                Text("₹${earnings.thisWeekActual}", style = MaterialTheme.typography.headlineMedium)
            }
        }

        Spacer(modifier = Modifier.height(24.dp))

        Text("Quick Actions", style = MaterialTheme.typography.titleMedium)
        Spacer(modifier = Modifier.height(8.dp))

        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            ActionCard("Policy", Modifier.weight(1f)) { navController.navigate(Screen.Policy.route) }
            ActionCard("Claims", Modifier.weight(1f)) { navController.navigate(Screen.Claims.route) }
        }
    }
}

@Composable
fun ActionCard(title: String, modifier: Modifier = Modifier, onClick: () -> Unit) {
    Card(
        modifier = modifier.height(100.dp).clickable { onClick() }
    ) {
        Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
            Text(title, style = MaterialTheme.typography.titleMedium)
        }
    }
}
