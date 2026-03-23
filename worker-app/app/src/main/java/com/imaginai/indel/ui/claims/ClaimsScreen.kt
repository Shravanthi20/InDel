package com.imaginai.indel.ui.claims

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.AccountBalanceWallet
import androidx.compose.material.icons.filled.ChevronRight
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
import com.imaginai.indel.data.model.Claim
import com.imaginai.indel.data.model.WalletResponse
import com.imaginai.indel.ui.navigation.Screen
import com.imaginai.indel.ui.theme.*
import java.util.Locale

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
                title = { Text("Claims & Wallet", fontWeight = FontWeight.Bold) },
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
                is ClaimsUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                is ClaimsUiState.Success -> ClaimsContent(state.claims, state.wallet, navController)
                is ClaimsUiState.Error -> Text(state.message, color = ErrorRed, modifier = Modifier.align(Alignment.Center))
            }
        }
    }
}

@Composable
fun ClaimsContent(
    claims: List<Claim>,
    wallet: WalletResponse,
    navController: NavController
) {
    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        // 1. Wallet Card
        item {
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(containerColor = Color.White),
                elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
            ) {
                Row(modifier = Modifier.padding(20.dp), verticalAlignment = Alignment.CenterVertically) {
                    Column(modifier = Modifier.weight(1f)) {
                        Text("Available Payout", style = MaterialTheme.typography.labelLarge, color = TextSecondary)
                        Text("₹${wallet.availableBalance}", style = MaterialTheme.typography.headlineMedium, fontWeight = FontWeight.Bold, color = BrandOrange)
                    }
                    Box(
                        modifier = Modifier.size(48.dp).background(OrangeSoft, RoundedCornerShape(12.dp)),
                        contentAlignment = Alignment.Center
                    ) {
                        Icon(Icons.Default.AccountBalanceWallet, contentDescription = null, tint = BrandOrange)
                    }
                }
            }
        }

        item {
            Text("Automatic Claims", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
        }

        if (claims.isEmpty()) {
            item {
                Box(modifier = Modifier.fillMaxWidth().padding(32.dp), contentAlignment = Alignment.Center) {
                    Text("No claims generated yet", color = TextSecondary)
                }
            }
        } else {
            items(claims) { claim ->
                ClaimCard(claim) {
                    navController.navigate(Screen.ClaimDetail.createRoute(claim.claimId))
                }
            }
        }
        
        item {
            Spacer(modifier = Modifier.height(32.dp))
        }
    }
}

@Composable
fun ClaimCard(claim: Claim, onClick: () -> Unit) {
    Card(
        modifier = Modifier.fillMaxWidth().clickable { onClick() },
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                Column {
                    Text(
                        text = claim.disruptionType.replace("_", " ").uppercase(),
                        style = MaterialTheme.typography.labelSmall,
                        fontWeight = FontWeight.Bold,
                        color = BrandOrange
                    )
                    Text(claim.zone, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                }
                StatusBadge(claim.status)
            }
            
            Spacer(modifier = Modifier.height(12.dp))
            HorizontalDivider(color = BackgroundWarmWhite)
            Spacer(modifier = Modifier.height(12.dp))
            
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                Column {
                    Text("Payout Amount", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                    Text("₹${claim.payoutAmount}", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold, color = SuccessGreen)
                }
                Icon(Icons.Default.ChevronRight, contentDescription = null, tint = TextSecondary)
            }
            
            Text(
                text = "Created on ${claim.createdAt?.take(10) ?: "N/A"}",
                style = MaterialTheme.typography.labelSmall,
                color = TextSecondary,
                modifier = Modifier.padding(top = 8.dp)
            )
        }
    }
}

@Composable
fun StatusBadge(status: String) {
    val color = when(status.lowercase()) {
        "credited", "approved" -> SuccessGreen
        "pending" -> WarningAmber
        else -> TextSecondary
    }
    Surface(
        color = color.copy(alpha = 0.1f),
        shape = RoundedCornerShape(4.dp)
    ) {
        Text(
            text = status.uppercase(),
            modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp),
            fontSize = 10.sp,
            fontWeight = FontWeight.Bold,
            color = color
        )
    }
}
