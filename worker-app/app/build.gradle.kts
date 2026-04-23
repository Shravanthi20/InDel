import java.net.NetworkInterface
import java.net.Inet4Address

plugins {
    id("com.android.application")
    kotlin("android")
    id("com.google.dagger.hilt.android")
    kotlin("kapt")
    id("org.jetbrains.kotlin.plugin.compose")
}

// Load .env file for configuration with robust discovery and auto-IP detection
fun getHostIpAddress(): String {
    try {
        val interfaces = NetworkInterface.getNetworkInterfaces()
        while (interfaces.hasMoreElements()) {
            val iface = interfaces.nextElement()
            // Skip loopback, inactive, or virtual interfaces
            if (iface.isLoopback || !iface.isUp || iface.displayName.contains("Virtual") || iface.displayName.contains("VMware")) continue
            val addresses = iface.inetAddresses
            while (addresses.hasMoreElements()) {
                val addr = addresses.nextElement()
                if (addr is Inet4Address && !addr.isLoopbackAddress) {
                    return addr.hostAddress
                }
            }
        }
    } catch (e: Exception) {}
    return "10.0.2.2" // Ultimate fallback for official emulator
}

val envFiles = listOf("worker-app/.env", ".env", "../.env")
    .map { rootProject.file(it) }
    .filter { it.exists() }

fun readEnvValue(key: String): String? {
    val fromSystem = System.getenv(key)?.trim()
    if (!fromSystem.isNullOrEmpty()) {
        return fromSystem
    }

    return envFiles.firstNotNullOfOrNull { file ->
        file.readLines()
            .find { it.startsWith("$key=") }
            ?.removePrefix("$key=")
            ?.trim()
            ?.removeSurrounding("\"")
    }
}

fun withTrailingSlash(url: String): String = if (url.endsWith("/")) url else "$url/"

val workerApiBaseUrl = withTrailingSlash(
    readEnvValue("WORKER_API_BASE_URL")
        ?: readEnvValue("WORKER_API_URL")
        ?: readEnvValue("API_BASE_URL")
        ?: "http://${getHostIpAddress()}:8001/"
)

val platformApiBaseUrl = withTrailingSlash(
    readEnvValue("PLATFORM_API_BASE_URL")
        ?: readEnvValue("PLATFORM_API_URL")
        ?: when {
            workerApiBaseUrl.contains(":8001/") -> workerApiBaseUrl.replace(":8001/", ":8003/")
            workerApiBaseUrl.contains(":8001") -> workerApiBaseUrl.replace(":8001", ":8003")
            else -> "http://${getHostIpAddress()}:8003/"
        }
)
val razorpayKeyId = readEnvValue("RAZORPAY_KEY_ID") ?: ""

println(">>> InDel Build: Using WORKER_API_BASE_URL=$workerApiBaseUrl")
println(">>> InDel Build: Using PLATFORM_API_BASE_URL=$platformApiBaseUrl")
println(">>> InDel Build: Razorpay key configured=${razorpayKeyId.isNotBlank()}")

android {
    namespace = "com.imaginai.indel"
    compileSdk = 34
    
    defaultConfig {
        applicationId = "com.imaginai.indel"
        minSdk = 26
        targetSdk = 34
        versionCode = 1
        versionName = "1.0.0"

        buildConfigField("String", "WORKER_API_BASE_URL", "\"$workerApiBaseUrl\"")
        buildConfigField("String", "PLATFORM_API_BASE_URL", "\"$platformApiBaseUrl\"")
        buildConfigField("String", "RAZORPAY_KEY_ID", "\"$razorpayKeyId\"")
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_11
        targetCompatibility = JavaVersion.VERSION_11
    }

    buildFeatures {
        compose = true
        buildConfig = true
    }

    packaging {
        resources {
            excludes += "/META-INF/{AL2.0,LGPL2.1}"
        }
    }
}

tasks.withType<org.jetbrains.kotlin.gradle.tasks.KotlinCompile>().configureEach {
    compilerOptions {
        jvmTarget.set(org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_11)
    }
}

dependencies {
    // Kotlin
    implementation("org.jetbrains.kotlin:kotlin-stdlib:2.0.21")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-android:1.7.3")
    
    // AppCompat & Material
    implementation("androidx.appcompat:appcompat:1.7.1")
    implementation("com.google.android.material:material:1.12.0")

    // Jetpack Compose
    implementation("androidx.compose.ui:ui:1.7.5")
    implementation("androidx.compose.material3:material3:1.3.1")
    implementation("androidx.compose.foundation:foundation:1.7.5")
    implementation("androidx.activity:activity-compose:1.9.3")
    implementation("androidx.navigation:navigation-compose:2.8.4")
    implementation("androidx.lifecycle:lifecycle-viewmodel-compose:2.8.7")
    implementation("androidx.compose.material:material-icons-extended:1.7.5")
    
    // Firebase
    implementation(platform("com.google.firebase:firebase-bom:32.3.1"))
    implementation("com.google.firebase:firebase-auth-ktx")
    implementation("com.google.firebase:firebase-messaging-ktx")
    
    // Retrofit2 + OkHttp
    implementation("com.squareup.retrofit2:retrofit:2.10.0")
    implementation("com.squareup.retrofit2:converter-gson:2.10.0")
    implementation("com.squareup.okhttp3:okhttp:4.12.0")
    implementation("com.squareup.okhttp3:logging-interceptor:4.12.0")
    
    // Hilt DI
    implementation("com.google.dagger:hilt-android:2.52")
    kapt("com.google.dagger:hilt-compiler:2.52")
    implementation("androidx.hilt:hilt-navigation-compose:1.2.0")
    
    // DataStore
    implementation("androidx.datastore:datastore-preferences:1.1.1")
    
    // Room offline database
    val roomVersion = "2.6.1"
    implementation("androidx.room:room-runtime:$roomVersion")
    implementation("androidx.room:room-ktx:$roomVersion")
    kapt("androidx.room:room-compiler:$roomVersion")
    
    // Coil for image loading
    implementation("io.coil-kt:coil-compose:2.7.0")
    
    // Razorpay Integration
    implementation("com.razorpay:checkout:1.6.38")
    
    // Testing
    testImplementation("junit:junit:4.13.2")
    androidTestImplementation("androidx.test.ext:junit:1.2.1")
    androidTestImplementation("androidx.test.espresso:espresso-core:3.6.1")
    androidTestImplementation("androidx.compose.ui:ui-test-junit4:1.7.5")
    debugImplementation("androidx.compose.ui:ui-test-manifest:1.7.5")
}

kapt {
    correctErrorTypes = true
}
