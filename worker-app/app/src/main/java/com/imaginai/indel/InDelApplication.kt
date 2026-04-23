package com.imaginai.indel

import android.app.Application
import android.util.Log
import dagger.hilt.android.HiltAndroidApp

@HiltAndroidApp
class InDelApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        Log.i("InDelApplication", "Worker API base: ${BuildConfig.WORKER_API_BASE_URL}")
        Log.i("InDelApplication", "Platform API base: ${BuildConfig.PLATFORM_API_BASE_URL}")
    }
}
