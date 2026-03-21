package com.imaginai.indel.ui.navigation

import androidx.compose.runtime.Composable
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.imaginai.indel.ui.auth.OtpScreen

sealed class Screen(val route: String) {
    object OTP : Screen("otp")
    object Home : Screen("home")
    object Policy : Screen("policy")
    object Earnings : Screen("earnings")
    object Claims : Screen("claims")
}

@Composable
fun NavGraph() {
    val navController = rememberNavController()

    NavHost(navController, startDestination = Screen.OTP.route) {
        composable(Screen.OTP.route) {
            OtpScreen(navController)
        }
        composable(Screen.Home.route) {
            // HomeScreen()
        }
        composable(Screen.Policy.route) {
            // PolicyScreen()
        }
    }
}
