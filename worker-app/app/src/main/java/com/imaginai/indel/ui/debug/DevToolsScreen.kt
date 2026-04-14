package com.imaginai.indel.ui.debug

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.ui.theme.*

// ─────────────────────────────────────────────────────────────────────────────
// Colour aliases for DevTools dark-ish cards
// ─────────────────────────────────────────────────────────────────────────────
private val DevSurface    = Color(0xFF1A2235)
private val DevBorder     = Color(0xFF2D3F55)
private val DevLabel      = Color(0xFF8BA3BF)
private val DevWarningBg  = Color(0xFF2D1A00)
private val DevErrorBg    = Color(0xFF2D0A0A)
private val DevSuccessBg  = Color(0xFF0A2D1A)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DevToolsScreen(navController: NavController,
                   viewModel: DevToolsViewModel = hiltViewModel()) {

    val zoneLevels      by viewModel.zoneLevels.collectAsState()
    val zoneNames       by viewModel.zoneNames.collectAsState()
    val selectedLevel   by viewModel.selectedLevel.collectAsState()
    val selectedZone    by viewModel.selectedZone.collectAsState()
    val selectedType    by viewModel.selectedDisruptionType.collectAsState()
    val assignCount     by viewModel.assignCount.collectAsState()
    val simulateCount   by viewModel.simulateCount.collectAsState()
    val actionState     by viewModel.actionState.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Dev Tools", fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = Color(0xFF0F1B2D),
                    titleContentColor = Color.White,
                    navigationIconContentColor = Color.White
                )
            )
        },
        containerColor = Color(0xFF0B1422)
    ) { padding ->
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding),
            contentPadding = PaddingValues(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            // ─── Warning banner ───────────────────────────────────────────
            item {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .clip(RoundedCornerShape(12.dp))
                        .background(DevWarningBg)
                        .border(1.dp, WarningAmber.copy(alpha = 0.4f), RoundedCornerShape(12.dp))
                        .padding(12.dp),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(10.dp)
                ) {
                    Icon(Icons.Default.Warning, contentDescription = null,
                         tint = WarningAmber, modifier = Modifier.size(18.dp))
                    Text(
                        "Debug-only. These controls directly mutate backend state.",
                        style = MaterialTheme.typography.bodySmall,
                        color = WarningAmber
                    )
                }
            }

            // ─── Action Result Card ───────────────────────────────────────
            item {
                AnimatedVisibility(
                    visible = actionState !is DevToolsActionState.Idle,
                    enter = fadeIn(), exit = fadeOut()
                ) {
                    val (bg, border, icon, text, color) = when (actionState) {
                        is DevToolsActionState.Success -> ResultStyle(
                            DevSuccessBg, SuccessGreen.copy(alpha = 0.4f),
                            Icons.Default.CheckCircle, (actionState as DevToolsActionState.Success).message, SuccessGreen
                        )
                        is DevToolsActionState.Error -> ResultStyle(
                            DevErrorBg, ErrorRed.copy(alpha = 0.4f),
                            Icons.Default.ErrorOutline, (actionState as DevToolsActionState.Error).message, ErrorRed
                        )
                        is DevToolsActionState.Loading -> ResultStyle(
                            DevSurface, DevBorder, Icons.Default.HourglassEmpty, "Working…", DevLabel
                        )
                        else -> ResultStyle(DevSurface, DevBorder, Icons.Default.Info, "", Color.Transparent)
                    }
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clip(RoundedCornerShape(12.dp))
                            .background(bg)
                            .border(1.dp, border, RoundedCornerShape(12.dp))
                            .clickable { viewModel.clearActionState() }
                            .padding(14.dp),
                        verticalAlignment = Alignment.Top,
                        horizontalArrangement = Arrangement.spacedBy(10.dp)
                    ) {
                        if (actionState is DevToolsActionState.Loading) {
                            CircularProgressIndicator(
                                modifier = Modifier.size(18.dp).padding(top = 2.dp),
                                strokeWidth = 2.dp, color = DevLabel
                            )
                        } else {
                            Icon(icon, contentDescription = null, tint = color,
                                 modifier = Modifier.size(18.dp).padding(top = 2.dp))
                        }
                        Text(text, style = MaterialTheme.typography.bodySmall,
                             color = color, fontFamily = FontFamily.Monospace,
                             lineHeight = 18.sp)
                    }
                }
            }

            // ─── Section 1: Chaos Engine ──────────────────────────────────
            item {
                DevSectionHeader(icon = Icons.Default.FlashOn, title = "Chaos Engine",
                                 subtitle = "Trigger zone-targeted disruptions")
            }

            item {
                DevCard {
                    Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {

                        // Zone Level dropdown
                        DevDropdown(
                            label = "Zone Level",
                            options = zoneLevels.map { it.label.ifBlank { it.level } },
                            selected = zoneLevels.firstOrNull { it.level == selectedLevel }
                                ?.label?.ifBlank { selectedLevel } ?: selectedLevel,
                            onSelect = { idx ->
                                viewModel.onLevelSelected(zoneLevels.getOrNull(idx)?.level ?: "A")
                            }
                        )

                        // Zone Name dropdown (dynamic, first 15)
                        DevDropdown(
                            label = "Zone Name  (first 15)",
                            options = zoneNames,
                            selected = if (selectedZone.isBlank() && zoneNames.isNotEmpty())
                                zoneNames.first() else selectedZone,
                            onSelect = { idx ->
                                viewModel.selectedZone.value = zoneNames.getOrNull(idx) ?: ""
                            },
                            placeholder = if (zoneNames.isEmpty()) "Loading zones…" else "Select zone"
                        )

                        // Disruption type
                        DevDropdown(
                            label = "Disruption Type",
                            options = listOf("WEATHER", "FLOOD", "STRIKE", "TRAFFIC", "POWER_CUT"),
                            selected = selectedType,
                            onSelect = { idx ->
                                viewModel.selectedDisruptionType.value =
                                    listOf("WEATHER", "FLOOD", "STRIKE", "TRAFFIC", "POWER_CUT")[idx]
                            }
                        )

                        // Trigger button
                        Button(
                            onClick = { viewModel.triggerDisruption() },
                            modifier = Modifier.fillMaxWidth().height(46.dp),
                            colors = ButtonDefaults.buttonColors(containerColor = WarningAmber),
                            shape = RoundedCornerShape(10.dp),
                            enabled = actionState !is DevToolsActionState.Loading
                        ) {
                            Icon(Icons.Default.FlashOn, contentDescription = null,
                                 modifier = Modifier.size(18.dp).padding(end = 6.dp))
                            Text("Trigger Disruption", fontWeight = FontWeight.Bold)
                        }
                    }
                }
            }

            // ─── Section 2: Order Controls ────────────────────────────────
            item {
                DevSectionHeader(icon = Icons.Default.ShoppingCart, title = "Order Controls",
                                 subtitle = "Simulate order assignment and delivery")
            }

            item {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    // Assign Orders
                    DevCard(modifier = Modifier.weight(1f)) {
                        Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                            Text("Assign Orders", color = Color.White,
                                 fontWeight = FontWeight.SemiBold, fontSize = 13.sp)
                            DevCountRow(
                                count = assignCount,
                                onDecrement = { if (assignCount > 1) viewModel.assignCount.value-- },
                                onIncrement = { if (assignCount < 20) viewModel.assignCount.value++ }
                            )
                            Button(
                                onClick = { viewModel.assignOrders() },
                                modifier = Modifier.fillMaxWidth().height(38.dp),
                                colors = ButtonDefaults.buttonColors(containerColor = BrandBlue),
                                shape = RoundedCornerShape(8.dp),
                                enabled = actionState !is DevToolsActionState.Loading
                            ) { Text("Run", fontSize = 13.sp) }
                        }
                    }
                    // Simulate Deliveries
                    DevCard(modifier = Modifier.weight(1f)) {
                        Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                            Text("Simulate Deliveries", color = Color.White,
                                 fontWeight = FontWeight.SemiBold, fontSize = 13.sp)
                            DevCountRow(
                                count = simulateCount,
                                onDecrement = { if (simulateCount > 1) viewModel.simulateCount.value-- },
                                onIncrement = { if (simulateCount < 20) viewModel.simulateCount.value++ }
                            )
                            Button(
                                onClick = { viewModel.simulateDeliveries() },
                                modifier = Modifier.fillMaxWidth().height(38.dp),
                                colors = ButtonDefaults.buttonColors(containerColor = BrandBlue),
                                shape = RoundedCornerShape(8.dp),
                                enabled = actionState !is DevToolsActionState.Loading
                            ) { Text("Run", fontSize = 13.sp) }
                        }
                    }
                }
            }

            // ─── Section 3: System Tools ──────────────────────────────────
            item {
                DevSectionHeader(icon = Icons.Default.Settings, title = "System Tools",
                                 subtitle = "Settlements and resets")
            }

            item {
                DevCard {
                    Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.spacedBy(10.dp)
                        ) {
                            DevActionButton(
                                label = "Settle Earnings",
                                icon = Icons.Default.CurrencyRupee,
                                color = SuccessGreen,
                                modifier = Modifier.weight(1f),
                                enabled = actionState !is DevToolsActionState.Loading
                            ) { viewModel.settleEarnings() }

                            DevActionButton(
                                label = "Reset Zone",
                                icon = Icons.Default.Refresh,
                                color = WarningAmber,
                                modifier = Modifier.weight(1f),
                                enabled = actionState !is DevToolsActionState.Loading
                            ) { viewModel.resetZone() }
                        }

                        HorizontalDivider(color = DevBorder, modifier = Modifier.padding(vertical = 4.dp))

                        // Full Reset — danger zone
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .clip(RoundedCornerShape(10.dp))
                                .background(DevErrorBg)
                                .border(1.dp, ErrorRed.copy(alpha = 0.4f), RoundedCornerShape(10.dp))
                                .padding(12.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.SpaceBetween
                        ) {
                            Column {
                                Text("Full Reset", color = ErrorRed, fontWeight = FontWeight.Bold,
                                     fontSize = 13.sp)
                                Text("Wipes ALL demo data", color = DevLabel, fontSize = 11.sp)
                            }
                            Button(
                                onClick = { viewModel.fullReset() },
                                colors = ButtonDefaults.buttonColors(containerColor = ErrorRed),
                                shape = RoundedCornerShape(8.dp),
                                contentPadding = PaddingValues(horizontal = 16.dp, vertical = 8.dp),
                                enabled = actionState !is DevToolsActionState.Loading
                            ) {
                                Icon(Icons.Default.DeleteForever, contentDescription = null,
                                     modifier = Modifier.size(16.dp).padding(end = 4.dp))
                                Text("Reset All", fontSize = 13.sp)
                            }
                        }
                    }
                }
            }

            item { Spacer(modifier = Modifier.height(32.dp)) }
        }
    }
}

// ─────────────────────────────────────────────────────────────────────────────
// Private helpers
// ─────────────────────────────────────────────────────────────────────────────

private data class ResultStyle(
    val bg: Color, val border: Color, val icon: ImageVector,
    val text: String, val color: Color
)

@Composable
private fun DevCard(
    modifier: Modifier = Modifier,
    content: @Composable ColumnScope.() -> Unit
) {
    Column(
        modifier = modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(14.dp))
            .background(DevSurface)
            .border(1.dp, DevBorder, RoundedCornerShape(14.dp))
            .padding(16.dp),
        content = content
    )
}

@Composable
private fun DevSectionHeader(icon: ImageVector, title: String, subtitle: String) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(10.dp),
        modifier = Modifier.padding(top = 4.dp)
    ) {
        Box(
            modifier = Modifier
                .size(32.dp)
                .clip(RoundedCornerShape(8.dp))
                .background(
                    Brush.linearGradient(listOf(BrandBlue, BlueDeep))
                ),
            contentAlignment = Alignment.Center
        ) {
            Icon(icon, contentDescription = null, tint = Color.White,
                 modifier = Modifier.size(18.dp))
        }
        Column {
            Text(title, color = Color.White, fontWeight = FontWeight.Bold, fontSize = 15.sp)
            Text(subtitle, color = DevLabel, fontSize = 11.sp)
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun DevDropdown(
    label: String,
    options: List<String>,
    selected: String,
    onSelect: (Int) -> Unit,
    placeholder: String = "Select…"
) {
    var expanded by remember { mutableStateOf(false) }

    Column(verticalArrangement = Arrangement.spacedBy(5.dp)) {
        Text(label, color = DevLabel, fontSize = 11.sp, fontWeight = FontWeight.Medium)
        ExposedDropdownMenuBox(
            expanded = expanded,
            onExpandedChange = { if (options.isNotEmpty()) expanded = it }
        ) {
            OutlinedTextField(
                value = if (options.isEmpty()) placeholder else selected,
                onValueChange = {},
                readOnly = true,
                singleLine = true,
                trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded) },
                modifier = Modifier.fillMaxWidth().menuAnchor(),
                shape = RoundedCornerShape(8.dp),
                colors = OutlinedTextFieldDefaults.colors(
                    focusedBorderColor = BrandBlue,
                    unfocusedBorderColor = DevBorder,
                    focusedTextColor = Color.White,
                    unfocusedTextColor = Color.White,
                    focusedContainerColor = Color(0xFF0F1B2D),
                    unfocusedContainerColor = Color(0xFF0F1B2D)
                )
            )
            ExposedDropdownMenu(
                expanded = expanded,
                onDismissRequest = { expanded = false },
                modifier = Modifier.background(Color(0xFF1A2235))
            ) {
                options.forEachIndexed { idx, option ->
                    DropdownMenuItem(
                        text = { Text(option, color = Color.White, fontSize = 13.sp) },
                        onClick = { onSelect(idx); expanded = false }
                    )
                }
            }
        }
    }
}

@Composable
private fun DevCountRow(count: Int, onDecrement: () -> Unit, onIncrement: () -> Unit) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(10.dp)
    ) {
        IconButton(
            onClick = onDecrement,
            modifier = Modifier.size(32.dp)
                .clip(RoundedCornerShape(6.dp))
                .background(DevBorder)
        ) { Text("−", color = Color.White, fontWeight = FontWeight.Bold) }
        Text(
            count.toString(), color = Color.White, fontWeight = FontWeight.Bold,
            fontSize = 16.sp, modifier = Modifier.width(28.dp),
            textAlign = androidx.compose.ui.text.style.TextAlign.Center
        )
        IconButton(
            onClick = onIncrement,
            modifier = Modifier.size(32.dp)
                .clip(RoundedCornerShape(6.dp))
                .background(DevBorder)
        ) { Text("+", color = Color.White, fontWeight = FontWeight.Bold) }
    }
}

@Composable
private fun DevActionButton(
    label: String,
    icon: ImageVector,
    color: Color,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    onClick: () -> Unit
) {
    Button(
        onClick = onClick,
        modifier = modifier.height(42.dp),
        colors = ButtonDefaults.buttonColors(containerColor = color.copy(alpha = 0.15f),
                                             contentColor = color),
        shape = RoundedCornerShape(8.dp),
        border = androidx.compose.foundation.BorderStroke(1.dp, color.copy(alpha = 0.4f)),
        enabled = enabled
    ) {
        Icon(icon, contentDescription = null, modifier = Modifier.size(15.dp).padding(end = 4.dp))
        Text(label, fontSize = 12.sp, fontWeight = FontWeight.SemiBold)
    }
}
