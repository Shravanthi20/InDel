package com.imaginai.indel.ui.claims

import android.content.Context
import android.content.Intent
import android.net.Uri
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
import androidx.compose.material.icons.filled.FileDownload
import androidx.compose.material3.*
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import androidx.core.content.FileProvider
import com.imaginai.indel.data.model.Claim
import com.imaginai.indel.data.model.WalletResponse
import com.imaginai.indel.ui.navigation.Screen
import com.imaginai.indel.ui.theme.*
import com.imaginai.indel.utils.ClaimPdfGenerator
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ClaimsScreen(
    navController: NavController,
    viewModel: ClaimsViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val isRefreshing by viewModel.isRefreshing.collectAsState()

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
                    is ClaimsUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                    is ClaimsUiState.Success -> ClaimsContent(state.claims, state.wallet, navController)
                    is ClaimsUiState.Error -> ErrorState(state.message) { viewModel.loadClaimsData() }
                }
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
                        Text("₹${wallet.availableBalance.toInt()}", style = MaterialTheme.typography.headlineMedium, fontWeight = FontWeight.Bold, color = BrandOrange)
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
                ClaimCard(claim, navController) {
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
fun ClaimCard(claim: Claim, navController: NavController, onClick: () -> Unit) {
    val context = LocalContext.current
    var isDownloading by remember { mutableStateOf(false) }
    val coroutineScope = rememberCoroutineScope()
    
    fun downloadClaimPdf() {
        isDownloading = true
        coroutineScope.launch(Dispatchers.Default) {
            try {
                val pdfFile = ClaimPdfGenerator.generateClaimPdf(context, claim)
                if (pdfFile != null) {
                    coroutineScope.launch(Dispatchers.Main) {
                        isDownloading = false
                        try {
                            val uri = FileProvider.getUriForFile(
                                context,
                                "${context.packageName}.fileprovider",
                                pdfFile
                            )
                            val intent = Intent(Intent.ACTION_VIEW).apply {
                                setDataAndType(uri, "application/pdf")
                                addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
                            }
                            context.startActivity(Intent.createChooser(intent, "Open PDF with"))
                        } catch (e: Exception) {
                            e.printStackTrace()
                        }
                    }
                } else {
                    coroutineScope.launch(Dispatchers.Main) {
                        isDownloading = false
                    }
                }
            } catch (e: Exception) {
                coroutineScope.launch(Dispatchers.Main) {
                    isDownloading = false
                }
            }
        }
    }
    
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(modifier = Modifier
                .fillMaxWidth()
                .clickable { onClick() }, horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                Column(modifier = Modifier.weight(1f)) {
                    val disruption = claim.disruptionType.replace("_", " ").uppercase()
                    Text(
                        text = disruption,
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
            
            Row(modifier = Modifier
                .fillMaxWidth()
                .clickable { onClick() }, horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                Column {
                    Text("Payout Amount", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                    Text("₹${claim.payoutAmount.toInt()}", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold, color = SuccessGreen)
                }
                Icon(Icons.Default.ChevronRight, contentDescription = null, tint = TextSecondary)
            }
            
            Text(
                text = "Created on ${claim.createdAt?.take(10) ?: "N/A"}",
                style = MaterialTheme.typography.labelSmall,
                color = TextSecondary,
                modifier = Modifier.padding(top = 8.dp)
            )

            claim.claimReason?.let {
                Spacer(modifier = Modifier.height(8.dp))
                Text(it, style = MaterialTheme.typography.bodySmall, color = TextPrimary)
            }

            if (!claim.mainCause.isNullOrBlank()) {
                Spacer(modifier = Modifier.height(6.dp))
                Text(
                    text = "Main cause: ${claim.mainCause}",
                    style = MaterialTheme.typography.labelSmall,
                    color = TextSecondary
                )
            }
            
            // Download Button
            Spacer(modifier = Modifier.height(12.dp))
            Button(
                onClick = { downloadClaimPdf() },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(36.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = BrandOrange,
                    contentColor = Color.White
                ),
                enabled = !isDownloading,
                shape = RoundedCornerShape(6.dp)
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.Center
                ) {
                    if (isDownloading) {
                        CircularProgressIndicator(
                            modifier = Modifier.size(16.dp),
                            color = Color.White,
                            strokeWidth = 1.5.dp
                        )
                        Spacer(modifier = Modifier.width(6.dp))
                        Text("Generating...", fontWeight = FontWeight.Bold, fontSize = 12.sp)
                    } else {
                        Icon(Icons.Default.FileDownload, contentDescription = "Download", modifier = Modifier.size(16.dp))
                        Spacer(modifier = Modifier.width(6.dp))
                        Text("Download PDF", fontWeight = FontWeight.Bold, fontSize = 12.sp)
                    }
                }
            }
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
