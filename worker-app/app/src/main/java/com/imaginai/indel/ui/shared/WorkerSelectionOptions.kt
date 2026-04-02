package com.imaginai.indel.ui.shared


data class ZoneLevelOption(
    val level: String, // "A", "B", etc.
    val label: String
)

val zoneLevelOptions: List<ZoneLevelOption> = listOf(
    ZoneLevelOption("A", "A local"),
    ZoneLevelOption("B", "B intra-state"),
    ZoneLevelOption("C", "C metro-to-metro")
)

val allVehicleOptions: List<String> = listOf(
    "scooter",
    "motorcycle",
    "auto-rickshaw",
    "car",
    "van",
    "mini-truck",
    "truck"
)

private val fourWheelerVehicleOptions: List<String> = listOf(
    "car",
    "van",
    "mini-truck",
    "truck"
)

private val zoneNamesLevelA = listOf("Tambaram", "Chromepet", "Pallavaram", "Selaiyur", "Velachery", "Adyar")
private val zoneNamesLevelB = listOf("Chennai Central", "Madurai", "Coimbatore", "Trichy", "Salem")
private val zoneNamesLevelC = listOf("Chennai-Bangalore", "Chennai-Hyderabad", "Chennai-Mumbai", "Chennai-Delhi")

fun zoneNamesForLevel(level: String): List<String> {
    return when (level.trim().uppercase()) {
        "A" -> zoneNamesLevelA
        "B" -> zoneNamesLevelB
        "C" -> zoneNamesLevelC
        else -> emptyList()
    }
}


fun zoneBandFromLevel(level: String): Char? =
    level.trim().uppercase().firstOrNull()?.takeIf { it in 'A'..'C' }


fun isZoneCAndAboveLevel(level: String): Boolean {
    val band = zoneBandFromLevel(level) ?: return false
    return band >= 'C'
}


fun vehicleOptionsForZoneLevel(level: String): List<String> {
    return if (isZoneCAndAboveLevel(level)) fourWheelerVehicleOptions else allVehicleOptions
}


fun isVehicleAllowedForZoneLevel(level: String, vehicle: String): Boolean {
    if (vehicle.isBlank()) return true
    val allowed = vehicleOptionsForZoneLevel(level)
    return allowed.contains(vehicle.trim().lowercase())
}

fun isValidUpiId(upiId: String): Boolean {
    val trimmed = upiId.trim()
    return trimmed.contains("@") && 
           !trimmed.startsWith("@") && 
           !trimmed.endsWith("@") &&
           trimmed.split("@").size == 2
}
