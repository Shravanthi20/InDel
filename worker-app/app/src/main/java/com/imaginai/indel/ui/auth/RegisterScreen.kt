package com.imaginai.indel.ui.auth

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowDropDown
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.data.model.ZonePath
import com.imaginai.indel.ui.navigation.Screen
import com.imaginai.indel.ui.shared.zoneLevelOptions
import com.imaginai.indel.ui.theme.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun RegisterScreen(
    navController: NavController,
    viewModel: RegisterViewModel = hiltViewModel()
) {
    val username by viewModel.username.collectAsState()
    val email by viewModel.email.collectAsState()
    val phone by viewModel.phone.collectAsState()
    val password by viewModel.password.collectAsState()
    val confirmPassword by viewModel.confirmPassword.collectAsState()
    val zoneLevel by viewModel.zoneLevel.collectAsState()
    val zoneName by viewModel.zoneName.collectAsState()
    val availablePaths by viewModel.availablePaths.collectAsState()
    val uiState by viewModel.uiState.collectAsState()

    var zoneLevelExpanded by remember { mutableStateOf(false) }
    var zoneNameQuery by remember { mutableStateOf("") }
    var zoneNameDropdownExpanded by remember { mutableStateOf(false) }

    LaunchedEffect(uiState) {
        if (uiState is RegisterUiState.Success) {
            navController.navigate(Screen.Onboarding.route) {
                popUpTo(Screen.Register.route) { inclusive = true }
            }
        }
    }

    Scaffold { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(24.dp)
                .background(BackgroundWarmWhite)
                .verticalScroll(rememberScrollState()),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Text(
                text = "Join InDel",
                style = MaterialTheme.typography.headlineLarge,
                color = BrandOrange,
                fontWeight = FontWeight.Bold
            )

            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = "Protect your income from day one",
                style = MaterialTheme.typography.bodyMedium,
                color = TextSecondary
            )
            
            Spacer(modifier = Modifier.height(32.dp))

            OutlinedTextField(
                value = username,
                onValueChange = viewModel::onUsernameChanged,
                label = { Text("Username") },
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(12.dp)
            )
            Spacer(modifier = Modifier.height(16.dp))

            OutlinedTextField(
                value = email,
                onValueChange = viewModel::onEmailChanged,
                label = { Text("Email") },
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(12.dp)
            )
            Spacer(modifier = Modifier.height(16.dp))

            OutlinedTextField(
                value = phone,
                onValueChange = viewModel::onPhoneChanged,
                label = { Text("Phone Number") },
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(12.dp)
            )
            Spacer(modifier = Modifier.height(16.dp))

            // Zone Level Dropdown
            ExposedDropdownMenuBox(
                expanded = zoneLevelExpanded,
                onExpandedChange = { zoneLevelExpanded = !zoneLevelExpanded }
            ) {
                OutlinedTextField(
                    value = zoneLevel.ifBlank { "Select Zone Level" },
                    onValueChange = {},
                    readOnly = true,
                    label = { Text("Zone Level") },
                    trailingIcon = {
                        Icon(Icons.Default.ArrowDropDown, contentDescription = "Select zone level")
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .menuAnchor(MenuAnchorType.PrimaryNotEditable, true),
                    shape = RoundedCornerShape(12.dp)
                )
                ExposedDropdownMenu(
                    expanded = zoneLevelExpanded,
                    onDismissRequest = { zoneLevelExpanded = false }
                ) {
                    zoneLevelOptions.forEach { option ->
                        DropdownMenuItem(
                            text = { Text(option.label) },
                            onClick = {
                                viewModel.onZoneLevelChanged(option.level)
                                zoneLevelExpanded = false
                            }
                        )
                    }
                }
            }
            Spacer(modifier = Modifier.height(16.dp))

            // Zone Name Autocomplete
            Box {
                OutlinedTextField(
                    value = if (zoneNameQuery.isNotBlank()) zoneNameQuery else zoneName,
                    onValueChange = {
                        zoneNameQuery = it
                        zoneNameDropdownExpanded = true
                        viewModel.onZoneNameQueryChanged(it)
                    },
                    label = { Text("Zone Name") },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    enabled = zoneLevel.isNotBlank() && availablePaths.isNotEmpty(),
                    singleLine = true,
                    trailingIcon = {
                        Icon(Icons.Default.ArrowDropDown, contentDescription = "Show zone suggestions")
                    }
                )
                DropdownMenu(
                    expanded = zoneNameDropdownExpanded && zoneLevel.isNotBlank() && availablePaths.isNotEmpty(),
                    onDismissRequest = { zoneNameDropdownExpanded = false },
                    modifier = Modifier.fillMaxWidth()
                ) {
                    val filtered = availablePaths.filter {
                        zoneNameQuery.isBlank() || (it.displayName?.contains(zoneNameQuery, ignoreCase = true) == true)
                    }.take(20)
                    if (filtered.isNotEmpty()) {
                        filtered.forEach { path ->
                            DropdownMenuItem(
                                text = { Text(path.displayName ?: "", maxLines = 1) },
                                onClick = {
                                    viewModel.onZonePathSelected(path)
                                    zoneNameQuery = path.displayName ?: ""
                                    zoneNameDropdownExpanded = false
                                }
                            )
                        }
                    } else {
                        DropdownMenuItem(
                            text = { Text("No matching zones") },
                            onClick = {},
                            enabled = false
                        )
                    }
                }
            }
            Spacer(modifier = Modifier.height(16.dp))

            OutlinedTextField(
                value = password,
                onValueChange = viewModel::onPasswordChanged,
                label = { Text("Password") },
                modifier = Modifier.fillMaxWidth(),
                visualTransformation = PasswordVisualTransformation(),
                shape = RoundedCornerShape(12.dp)
            )
            Spacer(modifier = Modifier.height(16.dp))

            OutlinedTextField(
                value = confirmPassword,
                onValueChange = viewModel::onConfirmPasswordChanged,
                label = { Text("Confirm Password") },
                modifier = Modifier.fillMaxWidth(),
                visualTransformation = PasswordVisualTransformation(),
                shape = RoundedCornerShape(12.dp)
            )
            
            Spacer(modifier = Modifier.height(32.dp))

            Button(
                onClick = viewModel::register,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp),
                enabled = uiState !is RegisterUiState.Loading,
                shape = RoundedCornerShape(12.dp),
                colors = ButtonDefaults.buttonColors(containerColor = BrandOrange)
            ) {
                if (uiState is RegisterUiState.Loading) {
                    CircularProgressIndicator(color = Color.White, modifier = Modifier.size(24.dp))
                } else {
                    Text("Register", fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                }
            }

            TextButton(
                onClick = { navController.navigate(Screen.Login.route) },
                modifier = Modifier.padding(top = 16.dp)
            ) {
                Text("Already have an account? Login", color = BrandOrange)
            }

            if (uiState is RegisterUiState.Error) {
                Text(
                    text = (uiState as RegisterUiState.Error).message,
                    color = ErrorRed,
                    style = MaterialTheme.typography.bodySmall,
                    modifier = Modifier.padding(top = 16.dp)
                )
            }
        }
    }
}
