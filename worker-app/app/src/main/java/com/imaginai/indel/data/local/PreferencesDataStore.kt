package com.imaginai.indel.data.local

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import javax.inject.Inject
import javax.inject.Singleton

private val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "worker_prefs")

@Singleton
class PreferencesDataStore @Inject constructor(
    @ApplicationContext private val context: Context
) {
    companion object {
        private val AUTH_TOKEN = stringPreferencesKey("auth_token")
        private val WORKER_ID = stringPreferencesKey("worker_id")
        private val POLICY_CACHE = stringPreferencesKey("policy_cache")
    }

    val authToken: Flow<String?> = context.dataStore.data.map { preferences ->
        preferences[AUTH_TOKEN]
    }

    val workerId: Flow<String?> = context.dataStore.data.map { preferences ->
        preferences[WORKER_ID]
    }

    suspend fun saveAuthToken(token: String) {
        context.dataStore.edit { preferences ->
            preferences[AUTH_TOKEN] = token
        }
    }

    suspend fun saveWorkerId(workerId: String) {
        context.dataStore.edit { preferences ->
            preferences[WORKER_ID] = workerId
        }
    }

    // Policy cache methods
    suspend fun savePolicyCache(policy: com.imaginai.indel.data.model.Policy) {
        val json = com.google.gson.Gson().toJson(policy)
        context.dataStore.edit { preferences ->
            preferences[POLICY_CACHE] = json
        }
    }

    fun getPolicyCache(): Flow<com.imaginai.indel.data.model.Policy?> = context.dataStore.data.map { preferences ->
        preferences[POLICY_CACHE]?.let { json ->
            try {
                com.google.gson.Gson().fromJson(json, com.imaginai.indel.data.model.Policy::class.java)
            } catch (e: Exception) {
                null
            }
        }
    }

    suspend fun clearAll() {
        context.dataStore.edit { preferences ->
            preferences.clear()
        }
    }
}
