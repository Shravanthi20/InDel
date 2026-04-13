package com.imaginai.indel.ui.claims

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Verified
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.data.model.Claim
import com.imaginai.indel.ui.theme.*

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
                title = { Text("Claim Details", fontWeight = FontWeight.Bold) },
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
                is ClaimDetailUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                is ClaimDetailUiState.Success -> ClaimDetailContent(state.claim)
                is ClaimDetailUiState.Error -> Text(state.message, color = ErrorRed, modifier = Modifier.align(Alignment.Center))
            }
        }
    }
}

@Composable
fun ClaimDetailContent(claim: Claim) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp)
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        // 1. Header Card
        Card(
            modifier = Modifier.fillMaxWidth(),
            shape = RoundedCornerShape(16.dp),
            colors = CardDefaults.cardColors(containerColor = Color.White),
            elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
        ) {
            Column(modifier = Modifier.padding(20.dp), horizontalAlignment = Alignment.CenterHorizontally) {
                Text("Total Payout", style = MaterialTheme.typography.labelLarge, color = TextSecondary)
                Text("₹${claim.payoutAmount}", style = MaterialTheme.typography.headlineLarge, fontWeight = FontWeight.Bold, color = SuccessGreen)
                Spacer(modifier = Modifier.height(8.dp))
                StatusBadge(claim.status)

                Spacer(modifier = Modifier.height(20.dp))
                HorizontalDivider(color = BackgroundWarmWhite)
                Spacer(modifier = Modifier.height(20.dp))
                
                Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text("Claim ID", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                        Text(claim.claimId.take(8), fontWeight = FontWeight.Bold)
                    }
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text("Type", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                        Text(claim.disruptionType.replace("_", " "), fontWeight = FontWeight.Bold)
                    }
                }

                Spacer(modifier = Modifier.height(16.dp))
                claim.claimReason?.let {
                    Text(it, style = MaterialTheme.typography.bodyMedium, color = TextPrimary)
                }
            }
        }

        Card(
            modifier = Modifier.fillMaxWidth(),
            shape = RoundedCornerShape(12.dp),
            colors = CardDefaults.cardColors(containerColor = Color.White)
        ) {
            Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                BreakdownRow("Income Loss", "₹${claim.incomeLoss}")
                BreakdownRow("Coverage Ratio", "85%")
                HorizontalDivider(color = BackgroundWarmWhite)
                BreakdownRow("Final Payout", "₹${claim.payoutAmount}", isTotal = true)
            }
        }

        // 4. Verification Note
        Card(
            modifier = Modifier.fillMaxWidth(),
            shape = RoundedCornerShape(12.dp),
            colors = CardDefaults.cardColors(containerColor = SuccessGreen.copy(alpha = 0.05f))
        ) {
            Row(modifier = Modifier.padding(16.dp)) {
                Icon(Icons.Default.Verified, contentDescription = null, tint = SuccessGreen)
                Spacer(modifier = Modifier.width(12.dp))
                Column {
                    Text("Automated Verdict", fontWeight = FontWeight.Bold, color = SuccessGreen)
                    Text(
                        claim.fraudVerdict ?: "Claim verified against real-time weather and dispatch data. No manual action required.",
                        style = MaterialTheme.typography.bodySmall,
                        color = TextPrimary
                    )
                }
            }
        }
        
        Spacer(modifier = Modifier.height(32.dp))
    }
}

@Composable
fun BreakdownRow(label: String, value: String, isTotal: Boolean = false) {
    Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
        Text(label, color = if (isTotal) TextPrimary else TextSecondary, fontWeight = if (isTotal) FontWeight.Bold else FontWeight.Normal)
        Text(value, color = if (isTotal) SuccessGreen else TextPrimary, fontWeight = FontWeight.Bold)
    }
}
