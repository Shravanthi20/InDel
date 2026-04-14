package com.imaginai.indel.data.repository

import com.imaginai.indel.data.local.PreferencesDataStore
import com.imaginai.indel.data.model.WorkerZone
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.flow.first
import org.junit.Assert.assertEquals
import org.junit.Before
import org.junit.Test
import org.mockito.Mockito
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever

open class WorkerRepositoryZoneTest {
    private lateinit var preferencesDataStore: PreferencesDataStore
    private lateinit var workerRepository: WorkerRepository

    @Before
    fun setup() {
        preferencesDataStore = mock()
        workerRepository = WorkerRepository(
            workerApiService = mock(),
            platformApiService = mock(),
            preferencesDataStore = preferencesDataStore
        )
    }

    @Test
    open fun testAtomicWorkerZoneCacheAndRetrieval() = runBlocking {
        val zone = WorkerZone(zoneLevel = "A", zoneName = "TestZone", zoneId = 1, city = "TestCity", fromCity = null, toCity = null)
        whenever(preferencesDataStore.getWorkerZone()).thenReturn(kotlinx.coroutines.flow.flowOf(zone))
        val cachedZone = preferencesDataStore.getWorkerZone().first()
        assertEquals("A", cachedZone?.zoneLevel)
        assertEquals("TestZone", cachedZone?.zoneName)
        assertEquals(1, cachedZone?.zoneId)
        assertEquals("TestCity", cachedZone?.city)
    }
}
