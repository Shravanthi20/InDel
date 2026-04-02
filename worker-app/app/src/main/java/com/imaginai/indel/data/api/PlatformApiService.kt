package com.imaginai.indel.data.api

import com.imaginai.indel.data.model.ZonePathResponse
import retrofit2.Response
import retrofit2.http.GET
import retrofit2.http.Query

interface PlatformApiService {
    @GET("api/v1/platform/zone-paths")
    suspend fun getZonePaths(@Query("type") type: String): Response<ZonePathResponse>
}
