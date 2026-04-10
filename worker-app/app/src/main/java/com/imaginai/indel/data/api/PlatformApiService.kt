package com.imaginai.indel.data.api

import com.imaginai.indel.data.model.ZonePathResponse
import com.imaginai.indel.data.model.ZoneListResponse
import retrofit2.Response
import retrofit2.http.GET
import retrofit2.http.Query

interface PlatformApiService {
    @GET("api/v1/platform/zones")
    suspend fun getZones(): Response<ZoneListResponse>

    @GET("api/v1/platform/zone-paths")
    suspend fun getZonePaths(@Query("type") type: String): Response<ZonePathResponse>
}
