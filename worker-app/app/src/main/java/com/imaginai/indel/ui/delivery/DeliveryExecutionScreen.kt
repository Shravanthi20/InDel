package com.imaginai.indel.ui.delivery

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Call
import androidx.compose.material.icons.filled.LocationOn
import androidx.compose.material.icons.filled.Person
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
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
import com.imaginai.indel.data.model.Order
import com.imaginai.indel.ui.navigation.Screen
import com.imaginai.indel.ui.theme.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DeliveryExecutionScreen(
    navController: NavController,
    orderId: String,
    viewModel: DeliveryExecutionViewModel = hiltViewModel()
) {
    val order by viewModel.order.collectAsState()
    val uiState by viewModel.uiState.collectAsState()

    LaunchedEffect(orderId) {
        viewModel.loadOrder(orderId)
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Delivery in Progress", fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = BrandBlue,
                    titleContentColor = Color.White,
                    navigationIconContentColor = Color.White
                )
            )
        }
    ) { padding ->
        Box(
            modifier = Modifier
                .padding(padding)
                .fillMaxSize()
                .background(BackgroundWarmWhite)
        ) {
            when (val state = uiState) {
                is ExecutionUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                is ExecutionUiState.Error -> Text(state.message, color = ErrorRed, modifier = Modifier.align(Alignment.Center))
                else -> {
                    order?.let { currentOrder ->
                        if (currentOrder.status == "picked_up") {
                            // If already picked up, go to completion screen
                            LaunchedEffect(Unit) {
                                navController.navigate(Screen.DeliveryCompletion.createRoute(orderId)) {
                                    popUpTo(Screen.DeliveryExecution.createRoute(orderId)) { inclusive = true }
                                }
                            }
                        } else {
                            ExecutionContent(currentOrder, viewModel)
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun ExecutionContent(
    order: Order,
    viewModel: DeliveryExecutionViewModel
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
        horizontalAlignment = Alignment.Start
    ) {
        Card(
            modifier = Modifier.fillMaxWidth(),
            shape = RoundedCornerShape(16.dp),
            colors = CardDefaults.cardColors(containerColor = Color.White),
            elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
        ) {
            Column(modifier = Modifier.padding(20.dp)) {
                Text("PICKUP DETAILS", style = MaterialTheme.typography.labelSmall, color = BrandBlue, fontWeight = FontWeight.Bold)
                Spacer(modifier = Modifier.height(16.dp))
                
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(Icons.Default.LocationOn, contentDescription = null, tint = BrandBlue)
                    Spacer(modifier = Modifier.width(12.dp))
                    Text(order.pickupArea, fontWeight = FontWeight.Bold, fontSize = 18.sp)
                }
                
                Spacer(modifier = Modifier.height(24.dp))
                HorizontalDivider(color = BackgroundWarmWhite)
                Spacer(modifier = Modifier.height(24.dp))

                Text("CUSTOMER DETAILS", style = MaterialTheme.typography.labelSmall, color = BrandBlue, fontWeight = FontWeight.Bold)
                Spacer(modifier = Modifier.height(16.dp))

                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(Icons.Default.Person, contentDescription = null, tint = TextSecondary)
                    Spacer(modifier = Modifier.width(12.dp))
                    Text(order.customerName ?: "Customer", fontWeight = FontWeight.SemiBold)
                }
                Spacer(modifier = Modifier.height(12.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(Icons.Default.Call, contentDescription = null, tint = TextSecondary)
                    Spacer(modifier = Modifier.width(12.dp))
                    Text(order.customerPhone ?: "Not available", color = TextSecondary)
                }
            }
        }

        Spacer(modifier = Modifier.weight(1f))

        Button(
            onClick = viewModel::startPickup,
            modifier = Modifier
                .fillMaxWidth()
                .height(56.dp),
            shape = RoundedCornerShape(12.dp),
            colors = ButtonDefaults.buttonColors(containerColor = BrandBlue)
        ) {
            Text("Confirm Pickup", fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
        }
    }
}
