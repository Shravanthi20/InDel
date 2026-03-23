package com.imaginai.indel.ui.navigation

import androidx.compose.runtime.Composable
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.imaginai.indel.ui.auth.OtpScreen
import com.imaginai.indel.ui.auth.OnboardingScreen
import com.imaginai.indel.ui.home.HomeScreen
import com.imaginai.indel.ui.policy.PolicyScreen
import com.imaginai.indel.ui.policy.PremiumPayScreen
import com.imaginai.indel.ui.earnings.EarningsScreen
import com.imaginai.indel.ui.claims.ClaimsScreen
import com.imaginai.indel.ui.claims.ClaimDetailScreen
import com.imaginai.indel.ui.orders.OrdersScreen

sealed class Screen(val route: String) {
    object OTP : Screen("otp")
    object Onboarding : Screen("onboarding")
    object Home : Screen("home")
    object Policy : Screen("policy")
    object PremiumPay : Screen("premium-pay")
    object Earnings : Screen("earnings")
    object Claims : Screen("claims")
    object ClaimDetail : Screen("claim-detail/{claimId}") {
        fun createRoute(claimId: String) = "claim-detail/$claimId"
    }
    object Orders : Screen("orders")
}

@Composable
fun NavGraph() {
    val navController = rememberNavController()

    NavHost(navController, startDestination = Screen.OTP.route) {
        composable(Screen.OTP.route) {
            OtpScreen(navController)
        }
        composable(Screen.Onboarding.route) {
            OnboardingScreen(navController)
        }
        composable(Screen.Home.route) {
            HomeScreen(navController)
        }
        composable(Screen.Policy.route) {
            PolicyScreen(navController)
        }
        composable(Screen.PremiumPay.route) {
            PremiumPayScreen(navController)
        }
        composable(Screen.Earnings.route) {
            EarningsScreen(navController)
        }
        composable(Screen.Claims.route) {
            ClaimsScreen(navController)
        }
        composable(
            route = Screen.ClaimDetail.route,
            arguments = listOf(navArgument("claimId") { type = NavType.StringType })
        ) { backStackEntry ->
            val claimId = backStackEntry.arguments?.getString("claimId") ?: ""
            ClaimDetailScreen(navController, claimId)
        }
        composable(Screen.Orders.route) {
            OrdersScreen(navController)
        }
    }
}
