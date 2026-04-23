package com.imaginai.indel.di

import com.imaginai.indel.BuildConfig
import com.imaginai.indel.data.api.AuthApiService
import com.imaginai.indel.data.api.WorkerApiService
import com.imaginai.indel.data.api.PlatformApiService
import com.imaginai.indel.data.local.PreferencesDataStore
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import okhttp3.Interceptor
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import java.util.concurrent.TimeUnit
import javax.inject.Qualifier
import javax.inject.Singleton

@Qualifier
@Retention(AnnotationRetention.BINARY)
annotation class WorkerGateway

@Qualifier
@Retention(AnnotationRetention.BINARY)
annotation class PlatformGateway

@Module
@InstallIn(SingletonComponent::class)
object NetworkModule {

    @Provides
    @Singleton
    fun provideLoggingInterceptor(): HttpLoggingInterceptor {
        return HttpLoggingInterceptor().apply {
            level = HttpLoggingInterceptor.Level.BODY
        }
    }

    @Provides
    @Singleton
    fun provideAuthInterceptor(preferencesDataStore: PreferencesDataStore): Interceptor {
        return Interceptor { chain ->
            val token = runBlocking { preferencesDataStore.authToken.first() }
            val request = chain.request().newBuilder()
            if (token != null) {
                request.addHeader("Authorization", "Bearer $token")
            }
            chain.proceed(request.build())
        }
    }

    @Provides
    @Singleton
    fun provideOkHttpClient(
        loggingInterceptor: HttpLoggingInterceptor,
        authInterceptor: Interceptor
    ): OkHttpClient {
        return OkHttpClient.Builder()
            .addInterceptor(loggingInterceptor)
            .addInterceptor(authInterceptor)
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .writeTimeout(30, TimeUnit.SECONDS)
            .build()
    }

    @Provides
    @Singleton
    @WorkerGateway
    fun provideWorkerRetrofit(okHttpClient: OkHttpClient): Retrofit {
        // Worker gateway base URL for auth, worker, orders endpoints
        val baseUrl = BuildConfig.WORKER_API_BASE_URL
        return Retrofit.Builder()
            .baseUrl(baseUrl)
            .client(okHttpClient)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
    }

    @Provides
    @Singleton
    @PlatformGateway
    fun providePlatformRetrofit(okHttpClient: OkHttpClient): Retrofit {
        // Platform gateway base URL for zone paths and zone data
        val baseUrl = BuildConfig.PLATFORM_API_BASE_URL
        return Retrofit.Builder()
            .baseUrl(baseUrl)
            .client(okHttpClient)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
    }

    @Provides
    @Singleton
    fun provideRetrofit(@WorkerGateway retrofit: Retrofit): Retrofit = retrofit

    @Provides
    @Singleton
    fun provideAuthApiService(@WorkerGateway retrofit: Retrofit): AuthApiService {
        return retrofit.create(AuthApiService::class.java)
    }

    @Provides
    @Singleton
    fun provideWorkerApiService(@WorkerGateway retrofit: Retrofit): WorkerApiService {
        return retrofit.create(WorkerApiService::class.java)
    }

    @Provides
    @Singleton
    fun providePlatformApiService(@PlatformGateway retrofit: Retrofit): PlatformApiService {
        return retrofit.create(PlatformApiService::class.java)
    }
}
