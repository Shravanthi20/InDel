package com.imaginai.indel

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.ui.Modifier
import com.imaginai.indel.ui.navigation.NavGraph
import com.imaginai.indel.ui.theme.InDelTheme
import dagger.hilt.android.AndroidEntryPoint
import com.razorpay.Checkout
import com.razorpay.PaymentData
import com.razorpay.PaymentResultWithDataListener
import org.json.JSONObject

@AndroidEntryPoint
class MainActivity : ComponentActivity(), PaymentResultWithDataListener {
    
    var razorpayCallback: ((success: Boolean, paymentId: String?, error: String?) -> Unit)? = null

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        Checkout.preload(applicationContext)
        setContent {
            InDelTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background
                ) {
                    NavGraph()
                }
            }
        }
    }

    fun startRazorpayCheckout(amountInPaise: Int, contactNumber: String, callback: (Boolean, String?, String?) -> Unit) {
        this.razorpayCallback = callback
        val checkout = Checkout()
        val razorpayKeyId = BuildConfig.RAZORPAY_KEY_ID?.trim().orEmpty()
        if (razorpayKeyId.isBlank()) {
            callback.invoke(false, null, "Missing RAZORPAY_KEY_ID. Set it in .env and rebuild app.")
            return
        }
        checkout.setKeyID(razorpayKeyId) 

        try {
            val options = JSONObject()
            options.put("name", "InDel Coverage")
            options.put("description", "Weekly Disruption Insurance Premium")
            options.put("currency", "INR")
            options.put("amount", amountInPaise)
            
            val prefill = JSONObject()
            prefill.put("contact", contactNumber)
            options.put("prefill", prefill)
            
            checkout.open(this, options)
        } catch (e: Exception) {
			callback.invoke(false, null, e.message ?: "Unable to open Razorpay checkout")
        }
    }

    override fun onPaymentSuccess(razorpayPaymentID: String?, paymentData: PaymentData?) {
        razorpayCallback?.invoke(true, razorpayPaymentID, null)
    }

    override fun onPaymentError(code: Int, response: String?, paymentData: PaymentData?) {
		razorpayCallback?.invoke(false, null, response ?: "Payment failed (code=$code)")
    }
}
