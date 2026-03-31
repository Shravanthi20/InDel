package com.imaginai.indel.ui.shared


data class ZoneLevelOption(
    val level: String, // "A", "B", etc.
    val label: String
)

val zoneLevelOptions: List<ZoneLevelOption> = listOf(
    ZoneLevelOption("A", "A local"),
    ZoneLevelOption("B", "B intra-state"),
    ZoneLevelOption("C", "C metro-to-metro"),
    ZoneLevelOption("D", "D rest of India"),
    ZoneLevelOption("E", "E special difficult lanes")
)

val zoneNamesByLevel: Map<String, List<String>> = mapOf(
    "A" to listOf("Tambaram", "Marasivaakkam"),
    "B" to listOf(
        "Tambaram to Chennai", "Chennai to Tambaram",
        "Tambaram to Kanchipuram", "Kanchipuram to Tambaram",
        "Sriperumbudur to Chennai", "Chennai to Sriperumbudur"
    ),
    "C" to listOf(
        "Chennai to Pondicherry", "Pondicherry to Chennai",
        "Chennai to Madurai", "Madurai to Chennai"
    ),
    "D" to listOf(
        "Madurai to Coimbatore", "Coimbatore to Madurai",
        "Chennai to Coimbatore", "Coimbatore to Chennai"
    ),
    "E" to listOf(
        "Nilgiris to Sikkim", "Sikkim to Nilgiris",
        "Chennai to Nilgiris", "Nilgiris to Chennai"
    )
)

fun zoneNamesForLevel(level: String): List<String> =
    zoneNamesByLevel[level] ?: emptyList()

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


fun zoneBandFromLevel(level: String): Char? =
    level.trim().uppercase().firstOrNull()?.takeIf { it in 'A'..'E' }


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
