package com.imaginai.indel.ui.claims

import android.content.Context
import android.content.Intent
import android.net.Uri
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Event
import androidx.compose.material.icons.filled.FileDownload
import androidx.compose.material.icons.filled.Verified
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import androidx.core.content.FileProvider
import com.imaginai.indel.data.model.Claim
import com.imaginai.indel.ui.theme.*
import com.imaginai.indel.utils.ClaimPdfGenerator
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ClaimDetailScreen(
    navController: NavController,
    claimId: String,
    viewModel: ClaimDetailViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val context = LocalContext.current

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
                is ClaimDetailUiState.Success -> ClaimDetailContent(state.claim, context)
                is ClaimDetailUiState.Error -> Text(state.message, color = ErrorRed, modifier = Modifier.align(Alignment.Center))
            }
        }
    }
}

@Composable
fun ClaimDetailContent(claim: Claim, context: Context) {
    var isDownloading by remember { mutableStateOf(false) }
    var downloadMessage by remember { mutableStateOf("") }
    val coroutineScope = rememberCoroutineScope()
    
    fun downloadClaimPdf() {
        isDownloading = true
        coroutineScope.launch(Dispatchers.Default) {
            try {
                val pdfFile = ClaimPdfGenerator.generateClaimPdf(context, claim)
                if (pdfFile != null) {
                    // Try to open with file manager or share
                    coroutineScope.launch(Dispatchers.Main) {
                        downloadMessage = "PDF saved to Downloads/InDel_Claims/"
                        isDownloading = false
                        
                        // Optionally open/share the PDF
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
                            // Silently fail if no PDF viewer available
                            e.printStackTrace()
                        }
                    }
                } else {
                    coroutineScope.launch(Dispatchers.Main) {
                        downloadMessage = "Failed to generate PDF"
                        isDownloading = false
                    }
                }
            } catch (e: Exception) {
                coroutineScope.launch(Dispatchers.Main) {
                    downloadMessage = "Error: ${e.message}"
                    isDownloading = false
                }
            }
        }
    }
    
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
                
                Spacer(modifier = Modifier.height(16.dp))
                
                // Download Button
                Button(
                    onClick = { downloadClaimPdf() },
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(44.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = BrandOrange,
                        contentColor = Color.White
                    ),
                    enabled = !isDownloading,
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.Center,
                        modifier = Modifier.fillMaxWidth()
                    ) {
                        if (isDownloading) {
                            CircularProgressIndicator(
                                modifier = Modifier.size(20.dp),
                                color = Color.White,
                                strokeWidth = 2.dp
                            )
                            Spacer(modifier = Modifier.width(8.dp))
                            Text("Generating PDF...", fontWeight = FontWeight.Bold)
                        } else {
                            Icon(Icons.Default.FileDownload, contentDescription = "Download", modifier = Modifier.size(20.dp))
                            Spacer(modifier = Modifier.width(8.dp))
                            Text("Download PDF", fontWeight = FontWeight.Bold)
                        }
                    }
                }
                
                // Download Message
                if (downloadMessage.isNotEmpty()) {
                    Spacer(modifier = Modifier.height(8.dp))
                    Text(
                        downloadMessage,
                        style = MaterialTheme.typography.labelSmall,
                        color = if (downloadMessage.contains("Failed") || downloadMessage.contains("Error")) ErrorRed else SuccessGreen,
                        modifier = Modifier.fillMaxWidth(),
                        textAlign = TextAlign.Center
                    )
                }
                
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

        Text("Why this claim exists", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
        Card(
            modifier = Modifier.fillMaxWidth(),
            shape = RoundedCornerShape(12.dp),
            colors = CardDefaults.cardColors(containerColor = Color.White)
        ) {
            Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
                BreakdownRow("Fraud score", claim.fraudScore?.let { String.format("%.2f", it) } ?: "N/A")
                BreakdownRow("Main cause", claim.mainCause ?: "N/A")
                BreakdownRow("Calculation", claim.calculation ?: "N/A")

                if (claim.factors.isNotEmpty()) {
                    HorizontalDivider(color = BackgroundWarmWhite)
                    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                        Text("Contributing factors", fontWeight = FontWeight.Bold)
                        claim.factors.forEach { factor ->
                            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                                Text(factor.name.replace("_", " "), color = TextSecondary)
                                Text(String.format("%.2f", factor.impact), fontWeight = FontWeight.Bold)
                            }
                        }
                    }
                }
            }
        }

        // 2. Disruption Timeline
        Text("Incident Window", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
        Card(
            modifier = Modifier.fillMaxWidth(),
            shape = RoundedCornerShape(12.dp),
            colors = CardDefaults.cardColors(containerColor = Color.White)
        ) {
            Row(modifier = Modifier.padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
                Icon(Icons.Default.Event, contentDescription = null, tint = BrandOrange)
                Spacer(modifier = Modifier.width(16.dp))
                Column {
                    Text("Time Frame", fontWeight = FontWeight.Bold)
                    claim.disruptionWindow?.let {
                         Text("${it.start.take(16)} - ${it.end.take(16)}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                    }
                }
            }
        }

        // 3. Breakdown
        Text("Payout Breakdown", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
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
