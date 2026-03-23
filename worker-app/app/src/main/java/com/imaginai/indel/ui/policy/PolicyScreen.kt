package com.imaginai.indel.ui.policy

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.Shield
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
import com.imaginai.indel.data.model.Policy
import com.imaginai.indel.ui.navigation.Screen
import com.imaginai.indel.ui.theme.*

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
                title = { Text("Protection Policy", fontWeight = FontWeight.Bold) },
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
                is PolicyUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                is PolicyUiState.Success -> PolicyContent(state.policy, navController, viewModel)
                is PolicyUiState.Error -> Text(state.message, color = ErrorRed, modifier = Modifier.align(Alignment.Center))
            }
        }
    }
}

@Composable
fun PolicyContent(
    policy: Policy,
    navController: NavController,
    viewModel: PolicyViewModel
) {
    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        // 1. Current Plan Card
        item {
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(containerColor = Color.White),
                elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
            ) {
                Column(modifier = Modifier.padding(20.dp)) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Icon(Icons.Default.Shield, contentDescription = null, tint = SuccessGreen)
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = if (policy.status == "active") "ACTIVE PROTECTION" else "INACTIVE",
                            fontWeight = FontWeight.Bold,
                            color = if (policy.status == "active") SuccessGreen else TextSecondary
                        )
                    }
                    
                    Spacer(modifier = Modifier.height(16.dp))
                    
                    Text("Weekly Premium", style = MaterialTheme.typography.labelMedium, color = TextSecondary)
                    Text("₹${policy.weeklyPremiumInr}", style = MaterialTheme.typography.headlineMedium, fontWeight = FontWeight.Bold, color = BrandOrange)
                    
                    Spacer(modifier = Modifier.height(12.dp))
                    
                    Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                        Column {
                            Text("Coverage Ratio", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                            Text("${(policy.coverageRatio * 100).toInt()}%", fontWeight = FontWeight.Bold)
                        }
                        Column(horizontalAlignment = Alignment.End) {
                            Text("Next Due Date", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                            Text(policy.nextDueDate, fontWeight = FontWeight.Bold)
                        }
                    }
                }
            }
        }

        // 2. Risk Factors (SHAP chips)
        item {
            Text("Why this premium?", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
            Spacer(modifier = Modifier.height(8.dp))
            FlowRow(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                policy.shapBreakdown.forEach { chip ->
                    SuggestionChip(
                        onClick = { },
                        label = { Text(chip.feature.replace("_", " "), fontSize = 12.sp) },
                        colors = SuggestionChipDefaults.suggestionChipColors(
                            containerColor = if (chip.impact > 0) ErrorRed.copy(alpha = 0.1f) else SuccessGreen.copy(alpha = 0.1f),
                            labelColor = if (chip.impact > 0) ErrorRed else SuccessGreen
                        ),
                        border = null
                    )
                }
            }
        }

        // 3. Actions
        item {
            Spacer(modifier = Modifier.height(16.dp))
            Button(
                onClick = { navController.navigate(Screen.PremiumPay.route) },
                modifier = Modifier.fillMaxWidth().height(56.dp),
                shape = RoundedCornerShape(12.dp)
            ) {
                Text("Pay Weekly Premium", fontWeight = FontWeight.Bold)
            }
            
            Spacer(modifier = Modifier.height(12.dp))
            
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                OutlinedButton(
                    onClick = { viewModel.pause() },
                    modifier = Modifier.weight(1f).height(48.dp),
                    shape = RoundedCornerShape(12.dp)
                ) {
                    Text("Pause Plan")
                }
                OutlinedButton(
                    onClick = { viewModel.cancel() },
                    modifier = Modifier.weight(1f).height(48.dp),
                    shape = RoundedCornerShape(12.dp),
                    colors = ButtonDefaults.outlinedButtonColors(contentColor = ErrorRed)
                ) {
                    Text("Cancel Plan")
                }
            }
        }
        
        // 4. Protection Note
        item {
            Card(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(12.dp),
                colors = CardDefaults.cardColors(containerColor = OrangeSoft)
            ) {
                Row(modifier = Modifier.padding(16.dp)) {
                    Icon(Icons.Default.Info, contentDescription = null, tint = BrandOrange)
                    Spacer(modifier = Modifier.width(12.dp))
                    Text(
                        "Protection covers income loss during heavy rain, heatwaves, and local disruptions. Maximum payout is capped at 3x your weekly baseline.",
                        style = MaterialTheme.typography.bodySmall,
                        color = TextPrimary
                    )
                }
            }
        }
    }
}

@OptIn(ExperimentalLayoutApi::class)
@Composable
fun FlowRow(
    modifier: Modifier = Modifier,
    horizontalArrangement: Arrangement.Horizontal = Arrangement.Start,
    content: @Composable () -> Unit
) {
    androidx.compose.foundation.layout.FlowRow(
        modifier = modifier,
        horizontalArrangement = horizontalArrangement,
        content = { content() }
    )
}
