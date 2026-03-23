package com.imaginai.indel.ui.policy

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
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
import com.imaginai.indel.data.model.Policy
import com.imaginai.indel.data.model.PremiumResponse
import com.imaginai.indel.ui.navigation.Screen

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PolicyScreen(
    navController: NavController,
    viewModel: PolicyViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("My Policy") },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "Back")
                    }
                }
            )
        }
    ) { padding ->
        Box(modifier = Modifier.padding(padding).fillMaxSize()) {
            when (val state = uiState) {
                is PolicyUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                is PolicyUiState.Success -> PolicyContent(state.policy, state.premium, navController, viewModel)
                is PolicyUiState.Error -> Text(state.message, color = Color.Red, modifier = Modifier.align(Alignment.Center))
            }
        }
    }
}

@Composable
fun PolicyContent(
    policy: Policy,
    premium: PremiumResponse,
    navController: NavController,
    viewModel: PolicyViewModel
) {
    LazyColumn(
        modifier = Modifier.fillMaxSize().padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        item {
            Card(modifier = Modifier.fillMaxWidth()) {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("Policy Status", style = MaterialTheme.typography.labelMedium)
                    Text(policy.status.uppercase(), style = MaterialTheme.typography.headlineSmall, color = MaterialTheme.colorScheme.primary)
                    Spacer(modifier = Modifier.height(8.dp))
                    Text("Premium: ₹${policy.weeklyPremiumInr} / week", style = MaterialTheme.typography.bodyLarge)
                    Text("Next Due: ${policy.nextDueDate}", style = MaterialTheme.typography.bodyMedium)
                }
            }
        }

        item {
            Text("Why this premium?", style = MaterialTheme.typography.titleMedium)
        }

        items(premium.shapBreakdown) { impact ->
            Row(modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp), horizontalArrangement = Arrangement.SpaceBetween) {
                Text(impact.feature.replace("_", " ").capitalize())
                Text("${(impact.impact * 100).toInt()}% impact", color = if (impact.impact > 0.3) Color.Red else Color.Gray)
            }
        }

        item {
            Spacer(modifier = Modifier.height(16.dp))
            Button(
                onClick = { navController.navigate(Screen.PremiumPay.route) },
                modifier = Modifier.fillMaxWidth()
            ) {
                Text("Pay Premium")
            }
            Spacer(modifier = Modifier.height(8.dp))
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                OutlinedButton(onClick = { viewModel.pause() }, modifier = Modifier.weight(1f)) { Text("Pause") }
                OutlinedButton(onClick = { viewModel.cancel() }, modifier = Modifier.weight(1f), colors = ButtonDefaults.outlinedButtonColors(contentColor = Color.Red)) { Text("Cancel") }
            }
        }
    }
}
