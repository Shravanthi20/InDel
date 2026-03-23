package com.imaginai.indel.ui.claims

import androidx.compose.foundation.clickable
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
import com.imaginai.indel.data.model.Claim
import com.imaginai.indel.data.model.Payout
import com.imaginai.indel.data.model.WalletResponse
import com.imaginai.indel.ui.navigation.Screen

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ClaimsScreen(
    navController: NavController,
    viewModel: ClaimsViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Claims & Payouts") },
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
                is ClaimsUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                is ClaimsUiState.Success -> ClaimsContent(state.claims, state.wallet, state.payouts, navController)
                is ClaimsUiState.Error -> Text(state.message, color = Color.Red, modifier = Modifier.align(Alignment.Center))
            }
        }
    }
}

@Composable
fun ClaimsContent(
    claims: List<Claim>,
    wallet: WalletResponse,
    payouts: List<Payout>,
    navController: NavController
) {
    LazyColumn(
        modifier = Modifier.fillMaxSize().padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        item {
            Card(modifier = Modifier.fillMaxWidth(), colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.primaryContainer)) {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("Available Balance", style = MaterialTheme.typography.labelMedium)
                    Text("₹${wallet.availableBalance}", style = MaterialTheme.typography.headlineMedium)
                    Spacer(modifier = Modifier.height(8.dp))
                    Text("Last Payout: ₹${wallet.lastPayoutAmount} on ${wallet.lastPayoutAt.take(10)}", style = MaterialTheme.typography.bodySmall)
                }
            }
        }

        item {
            Text("Active Claims", style = MaterialTheme.typography.titleMedium)
        }

        if (claims.isEmpty()) {
            item { Text("No claims found", style = MaterialTheme.typography.bodyMedium, color = Color.Gray) }
        } else {
            items(claims) { claim ->
                Card(
                    modifier = Modifier.fillMaxWidth().clickable {
                        navController.navigate(Screen.ClaimDetail.createRoute(claim.claimId))
                    }
                ) {
                    ListItem(
                        headlineContent = { Text(claim.disruptionType.replace("_", " ").capitalize()) },
                        supportingContent = { Text("Status: ${claim.status.uppercase()}") },
                        trailingContent = { Text("₹${claim.payoutAmount}", style = MaterialTheme.typography.titleMedium) }
                    )
                }
            }
        }

        item {
            Text("Payout History", style = MaterialTheme.typography.titleMedium)
        }

        items(payouts) { payout ->
            ListItem(
                headlineContent = { Text("Payout ${payout.payoutId}") },
                supportingContent = { Text("${payout.method.uppercase()} • ${payout.processedAt.take(10)}") },
                trailingContent = { Text("₹${payout.amount}", color = Color(0xFF2E7D32), style = MaterialTheme.typography.titleMedium) }
            )
            HorizontalDivider()
        }
    }
}
