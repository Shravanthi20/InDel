package com.imaginai.indel.ui.shared

data class ZoneOption(
    val value: String,
    val label: String
)

val zoneOptions: List<ZoneOption> = listOf(
    ZoneOption("Zone-A", "A local"),
    ZoneOption("Zone-B", "B intra-state"),
    ZoneOption("Zone-C", "C metro-to-metro"),
    ZoneOption("Zone-D", "D rest of India"),
    ZoneOption("Zone-E", "E special difficult lanes")
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

fun zoneBand(zone: String): Char? {
    val normalized = zone.trim().uppercase()
    if (normalized.startsWith("ZONE-") && normalized.length >= 6) {
        return normalized[5]
    }
    return if (normalized.length == 1 && normalized[0] in 'A'..'E') normalized[0] else null
}

fun isZoneCAndAbove(zone: String): Boolean {
    val band = zoneBand(zone) ?: return false
    return band >= 'C'
}

fun vehicleOptionsForZone(zone: String): List<String> {
    return if (isZoneCAndAbove(zone)) fourWheelerVehicleOptions else allVehicleOptions
}

fun isVehicleAllowedForZone(zone: String, vehicle: String): Boolean {
    if (vehicle.isBlank()) return true
    val allowed = vehicleOptionsForZone(zone)
    return allowed.contains(vehicle.trim().lowercase())
}

fun isValidUpiId(upiId: String): Boolean {
    val trimmed = upiId.trim()
    return trimmed.contains("@") && 
           !trimmed.startsWith("@") && 
           !trimmed.endsWith("@") &&
           trimmed.split("@").size == 2
}
