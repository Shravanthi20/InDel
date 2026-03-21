package com.imaginai.indel.service

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.os.Build
import com.google.firebase.messaging.FirebaseMessagingService
import com.google.firebase.messaging.RemoteMessage

class InDelFirebaseMessagingService : FirebaseMessagingService() {
    
    override fun onMessageReceived(remoteMessage: RemoteMessage) {
        super.onMessageReceived(remoteMessage)
        
        // Handle different notification types
        val messageType = remoteMessage.data["type"]
        
        when (messageType) {
            "disruption_alert" -> handleDisruptionAlert(remoteMessage)
            "payout_credited" -> handlePayoutAlert(remoteMessage)
            "premium_due" -> handlePremiumAlert(remoteMessage)
            else -> showNotification(remoteMessage)
        }
    }
    
    private fun handleDisruptionAlert(message: RemoteMessage) {
        // Show disruption alert
        showNotification(message)
    }
    
    private fun handlePayoutAlert(message: RemoteMessage) {
        // Show payout notification
        showNotification(message)
    }
    
    private fun handlePremiumAlert(message: RemoteMessage) {
        // Show premium due notification
        showNotification(message)
    }
    
    private fun showNotification(message: RemoteMessage) {
        val title = message.notification?.title ?: "InDel"
        val body = message.notification?.body ?: ""
        
        val notificationManager = getSystemService(NOTIFICATION_SERVICE) as NotificationManager
        
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                "indel_channel",
                "InDel Notifications",
                NotificationManager.IMPORTANCE_DEFAULT
            )
            notificationManager.createNotificationChannel(channel)
        }
    }
}
