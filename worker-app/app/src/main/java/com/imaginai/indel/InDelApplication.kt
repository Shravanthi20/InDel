package com.imaginai.indel

import android.app.Application
import dagger.hilt.android.HiltAndroidApp

@HiltAndroidApp
class InDelApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        // Initialize any global app setup
    }
}
