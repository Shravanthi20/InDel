package com.imaginai.indel.ui.claims

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.data.model.Claim
import java.util.Locale

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ClaimDetailScreen(
    navController: NavController,
    claimId: String,
    viewModel: ClaimDetailViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()

    LaunchedEffect(claimId) {
        viewModel.loadClaimDetail(claimId)
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Claim Details") },
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
                is ClaimDetailUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                is ClaimDetailUiState.Success -> ClaimDetailContent(state.claim)
                is ClaimDetailUiState.Error -> Text(state.message, color = Color.Red, modifier = Modifier.align(Alignment.Center))
            }
        }
    }
}

@Composable
fun ClaimDetailContent(claim: Claim) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        Card(modifier = Modifier.fillMaxWidth()) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text("Status", style = MaterialTheme.typography.labelMedium)
                Text(
                    claim.status.uppercase(Locale.getDefault()),
                    style = MaterialTheme.typography.headlineSmall,
                    color = if (claim.status == "approved") Color(0xFF2E7D32) else MaterialTheme.colorScheme.primary
                )
            }
        }

        DetailItem("Claim ID", claim.claimId)
        DetailItem("Type", claim.disruptionType.replace("_", " ").replaceFirstChar { it.uppercase() })
        DetailItem("Zone", claim.zone)
        DetailItem("Income Loss", "₹${claim.incomeLoss}")
        DetailItem("Payout Amount", "₹${claim.payoutAmount}")
        DetailItem("Fraud Verdict", claim.fraudVerdict.replaceFirstChar { it.uppercase() })

        if (claim.disruptionWindow != null) {
            DetailItem("Window", "${claim.disruptionWindow.start.take(16)} to ${claim.disruptionWindow.end.take(16)}")
        }
    }
}

@Composable
fun DetailItem(label: String, value: String) {
    Column {
        Text(label, style = MaterialTheme.typography.labelSmall, color = Color.Gray)
        Text(value, style = MaterialTheme.typography.bodyLarge)
        HorizontalDivider(modifier = Modifier.padding(top = 8.dp))
    }
}
