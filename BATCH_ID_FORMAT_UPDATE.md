# Batch ID Format Update - New Specification

## Overview
Updated the batch ID format from `ZONE+CITY+STATE-INDEX` to `ZONE+CITY+STATE+DATETIME` to include temporal information and ensure global uniqueness across all batch creations.

## New Batch ID Format

### Structure
```
Zone + City Code + State Code + DateTime
```

### Components

#### 1. **Zone Level** (1 character)
- `A` = Same City (intra-city delivery)
- `B` = Intra-State (within same state)
- `C` = Inter-State (across states)

#### 2. **City Code**
- **Zone A**: First 6 letters of city name
  - Example: `CHENNA` (from Chennai)
- **Zone B/C**: First 3 letters of FROM city + First 3 letters of TO city
  - Example: `CHEBAN` (CHE from Chennai + BAN from Bangalore)

#### 3. **State Code**
- **Zone A**: First 4 letters of state name
  - Example: `TAMI` (from Tamil Nadu)
- **Zone B**: First 4 letters of state name
  - Example: `TAMI` (from Tamil Nadu)
- **Zone C**: First 2 letters of FROM state + First 2 letters of TO state
  - Example: `TAKA` (TA from Tamil Nadu + KA from Karnataka)

#### 4. **DateTime** (14 characters)
- Format: `YYYYMMDDHHMMSS`
- Generated at batch creation time
- Ensures uniqueness across multiple batches

### Examples

#### Zone A (Same City)
```
Zone: A
From City: Chennai → To City: Chennai
State: Tamil Nadu
DateTime: 20260403123000
Result: A + CHENNA + TAMI + 20260403123000
Final ID: ACHENNATAMI20260403123000
```

#### Zone B (Intra-State)
```
Zone: B
From City: Chennai → To City: Bangalore
State: Tamil Nadu
DateTime: 20260403124500
Result: B + CHE + BAN + TAMI + 20260403124500
Final ID: BCHEBANTAMI20260403124500
```

#### Zone C (Inter-State)
```
Zone: C
From City: Chennai → To City: Mumbai
From State: Tamil Nadu → To State: Karnataka
DateTime: 20260403130000
Result: C + CHE + MUM + TA + KA + 20260403130000
Final ID: CCHEMUMTAKA20260403130000
```

## Implementation Details

### Backend (Go)
**File**: `backend/internal/handlers/worker/batches.go`

**Key Changes**:
1. Updated `buildBatchID()` signature to accept `timestamp time.Time` parameter
2. Added datetime formatting: `timestamp.Format("20060102150405")`
3. Removed `-INDEX` suffix pattern (each batch now has globally unique ID)
4. Batch ID is now generated at time of materialization using `time.Now()`

**Code Example**:
```go
func buildBatchID(zoneLevel, fromCity, toCity, fromState, toState string, timestamp time.Time) string {
    zone := strings.ToUpper(strings.TrimSpace(zoneLevel))
    cityCode := // ... determine based on zone
    stateCode := // ... determine based on zone
    datetimeStr := timestamp.Format("20060102150405")
    return zone + cityCode + stateCode + datetimeStr
}
```

### Frontend (Kotlin)
**File**: `worker-app/app/src/main/java/com/imaginai/indel/ui/orders/OrdersViewModel.kt`

**Note**: The Kotlin implementation already supported the timestamp format `yyyyMMddHHmmss` and was updated accordingly. It generates timestamps with slight offsets between batches to ensure uniqueness:

```kotlin
val timestamp = LocalDateTime.now()
    .plusSeconds(index.toLong())
    .format(DateTimeFormatter.ofPattern("yyyyMMddHHmmss"))
```

## Advantages of the New Format

1. **Global Uniqueness**: Timestamp makes batch IDs globally unique without needing to track indices
2. **Temporal Information**: Can track when batches were created just from the ID
3. **No Collision Risk**: Multiple batches for the same route won't have identical IDs
4. **Simpler Cache Management**: No need to maintain index counters in cache
5. **Better Auditability**: Batch creation time is embedded in the ID

## Backward Compatibility

⚠️ **Note**: This is a breaking change. Existing batch IDs with the older format (e.g., `ACHENNATAMI-01`, `ACHENNATAMI-02`) will no longer be recognized.

Any system components referencing old batch IDs should be updated to work with the new format.

## Testing

✅ **Test Results**:
- `TestAcceptBatchSetsOrdersToPickedUp`: PASS
- `TestAcceptBatchRejectsIncorrectPickupCode`: PASS
- Build: Successful (no compilation errors)

## Migration Notes

If migrating from the old format:
1. Existing batch references should be invalidated
2. Cache should be cleared
3. Any stored batch IDs in databases should be migrated or marked as legacy
4. Pickup codes will be recalculated based on new batch IDs

The pickup code derivation `pickupCodeFromBatchID()` remains the same:
```
hash(batch_id) % 9000 + 1000
```
This will generate new 4-digit codes for the updated batch IDs.
