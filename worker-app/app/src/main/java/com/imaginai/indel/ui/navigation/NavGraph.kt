package com.imaginai.indel.ui.navigation

import android.net.Uri
import androidx.compose.runtime.Composable
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.imaginai.indel.ui.auth.OtpScreen
import com.imaginai.indel.ui.auth.OnboardingScreen
import com.imaginai.indel.ui.auth.LoginScreen
import com.imaginai.indel.ui.auth.RegisterScreen
import com.imaginai.indel.ui.home.HomeScreen
import com.imaginai.indel.ui.policy.PolicyScreen
import com.imaginai.indel.ui.policy.PremiumPayScreen
import com.imaginai.indel.ui.earnings.EarningsScreen
import com.imaginai.indel.ui.claims.ClaimsScreen
import com.imaginai.indel.ui.claims.ClaimDetailScreen
import com.imaginai.indel.ui.orders.BatchDetailScreen
import com.imaginai.indel.ui.orders.OrdersScreen
import com.imaginai.indel.ui.delivery.LandingScreen
import com.imaginai.indel.ui.delivery.FetchVerificationScreen
import com.imaginai.indel.ui.delivery.DeliveryExecutionScreen
import com.imaginai.indel.ui.delivery.DeliveryCompletionScreen
import com.imaginai.indel.ui.delivery.SessionTrackingScreen
import com.imaginai.indel.ui.notifications.NotificationsScreen
import com.imaginai.indel.ui.profile.ProfileEditScreen
import com.imaginai.indel.ui.payouts.PayoutHistoryScreen
import com.imaginai.indel.ui.plan.PlanSelectionScreen
import com.imaginai.indel.ui.debug.DevToolsScreen

sealed class Screen(val route: String) {
    object SessionGate : Screen("session-gate")
    object Register : Screen("register")
    object Login : Screen("login")
    object OTP : Screen("otp")
    object Onboarding : Screen("onboarding")
    object PlanSelection : Screen("plan-selection")
    object Landing : Screen("landing")
    object Home : Screen("home")
    object Orders : Screen("orders")
    object BatchDetail : Screen("batch-detail/{batchId}") {
        fun createRoute(batchId: String) = "batch-detail/${Uri.encode(batchId)}"
    }
    object FetchVerification : Screen("fetch-verification")
    object DeliveryExecution : Screen("delivery-execution/{orderId}") {
        fun createRoute(orderId: String) = "delivery-execution/$orderId"
    }
    object DeliveryCompletion : Screen("delivery-completion/{orderId}") {
        fun createRoute(orderId: String) = "delivery-completion/$orderId"
    }
    object SessionTracking : Screen("session-tracking")
    object Policy : Screen("policy")
    object PremiumPay : Screen("premium-pay")
    object Earnings : Screen("earnings")
    object Claims : Screen("claims")
    object ClaimDetail : Screen("claim-detail/{claimId}") {
        fun createRoute(claimId: String) = "claim-detail/$claimId"
    }
    object Notifications : Screen("notifications")
    object ProfileEdit : Screen("profile-edit")
    object PayoutHistory : Screen("payouts-history")
    object DevTools : Screen("dev-tools")
}

@Composable
fun NavGraph() {
    val navController = rememberNavController()

    // Starting with Login for now as per the plan
    NavHost(navController, startDestination = Screen.Login.route) {
        composable(Screen.Login.route) {
            LoginScreen(navController)
        }
        composable(Screen.Register.route) {
            RegisterScreen(navController)
        }
        composable(Screen.OTP.route) {
            OtpScreen(navController)
        }
        composable(Screen.Onboarding.route) {
            OnboardingScreen(navController)
        }
        composable(Screen.PlanSelection.route) {
            PlanSelectionScreen(navController)
        }
        composable(Screen.Landing.route) {
            LandingScreen(navController)
        }
        composable(Screen.Home.route) {
            HomeScreen(navController)
        }
        composable(Screen.Orders.route) {
            OrdersScreen(navController)
        }
        composable(
            route = Screen.BatchDetail.route,
            arguments = listOf(navArgument("batchId") { type = NavType.StringType })
        ) { backStackEntry ->
            val encodedBatchId = backStackEntry.arguments?.getString("batchId") ?: ""
            val batchId = Uri.decode(encodedBatchId)
            BatchDetailScreen(navController, batchId)
        }
        composable(Screen.FetchVerification.route) {
            FetchVerificationScreen(navController)
        }
        composable(
            route = Screen.DeliveryExecution.route,
            arguments = listOf(navArgument("orderId") { type = NavType.StringType })
        ) { backStackEntry ->
            val orderId = backStackEntry.arguments?.getString("orderId") ?: ""
            DeliveryExecutionScreen(navController, orderId)
        }
        composable(
            route = Screen.DeliveryCompletion.route,
            arguments = listOf(navArgument("orderId") { type = NavType.StringType })
        ) { backStackEntry ->
            val orderId = backStackEntry.arguments?.getString("orderId") ?: ""
            DeliveryCompletionScreen(navController, orderId)
        }
        composable(Screen.SessionTracking.route) {
            SessionTrackingScreen(navController)
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
        composable(Screen.Notifications.route) {
            NotificationsScreen(navController)
        }
        composable(Screen.ProfileEdit.route) {
            ProfileEditScreen(navController)
        }
        composable(Screen.PayoutHistory.route) {
            PayoutHistoryScreen(navController)
        }
        composable(Screen.DevTools.route) {
            DevToolsScreen(navController)
        }
    }
}
