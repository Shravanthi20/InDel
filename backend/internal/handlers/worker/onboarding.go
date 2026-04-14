package worker

import (
	"fmt"
	"strings"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type workerZoneSummary struct {
	ZoneID    uint
	ZoneLevel string
	ZoneName  string
	City      string
	State     string
}

type normalizedZoneInput struct {
	Level string
	Name  string
	City  string
	State string
}

func normalizeZoneLevel(zoneLevel string) string {
	return strings.ToUpper(strings.TrimSpace(zoneLevel))
}

func zoneDefaultsForLevel(zoneLevel string) (city string, state string) {
	switch normalizeZoneLevel(zoneLevel) {
	case "B":
		return "Tamil Nadu", "Tamil Nadu"
	case "C":
		return "Chennai", "Tamil Nadu"
	default:
		return "Chennai", "Tamil Nadu"
	}
}

func normalizeZoneInput(zoneLevel, zoneName string) normalizedZoneInput {
	level := normalizeZoneLevel(zoneLevel)
	rawName := strings.TrimSpace(zoneName)
	city, state := zoneDefaultsForLevel(level)
	name := rawName

	if idx := strings.LastIndex(rawName, " ("); idx > 0 && strings.HasSuffix(rawName, ")") {
		name = strings.TrimSpace(rawName[:idx])
		parsedState := strings.TrimSpace(strings.TrimSuffix(rawName[idx+2:], ")"))
		if parsedState != "" {
			state = parsedState
		}
	}

	if strings.Contains(rawName, " to ") {
		name = rawName
		parts := strings.SplitN(rawName, " to ", 2)
		if len(parts) == 2 {
			city = strings.TrimSpace(parts[0])
		}
	}

	if name != "" && city == "Chennai" && !strings.Contains(strings.ToLower(name), "chennai") {
		city = name
	}

	return normalizedZoneInput{
		Level: level,
		Name:  name,
		City:  city,
		State: state,
	}
}

func formatZoneDisplay(zoneName, city string) string {
	name := strings.TrimSpace(zoneName)
	city = strings.TrimSpace(city)
	if name == "" {
		return city
	}
	if city == "" {
		return name
	}
	nameLower := strings.ToLower(name)
	cityLower := strings.ToLower(city)
	if strings.Contains(nameLower, cityLower) || strings.Contains(name, "(") || strings.Contains(name, ",") || strings.Contains(nameLower, " to ") {
		return name
	}
	return fmt.Sprintf("%s, %s", name, city)
}

// ensureZoneIDByLevelAndName finds or creates a zone by level and name
func ensureZoneIDByLevelAndName(zoneLevel, zoneName string) uint {
	if !HasDB() {
		return 0
	}
	normalized := normalizeZoneInput(zoneLevel, zoneName)
	level := normalized.Level
	name := normalized.Name
	if level == "" || name == "" {
		return 0
	}
	var zone models.Zone
	if err := workerDB.Where("level = ? AND name = ?", level, name).First(&zone).Error; err == nil {
		updates := map[string]any{}
		if normalized.City != "" && !strings.EqualFold(strings.TrimSpace(zone.City), strings.TrimSpace(normalized.City)) {
			updates["city"] = normalized.City
		}
		if normalized.State != "" && !strings.EqualFold(strings.TrimSpace(zone.State), strings.TrimSpace(normalized.State)) {
			updates["state"] = normalized.State
		}
		if len(updates) > 0 {
			_ = workerDB.Model(&zone).Updates(updates).Error
		}
		return zone.ID
	}
	newZone := models.Zone{Level: level, Name: name, City: normalized.City, State: normalized.State, RiskRating: 0.5}
	if err := workerDB.Create(&newZone).Error; err == nil {
		return newZone.ID
	}
	_ = workerDB.Where("level = ? AND name = ?", level, name).First(&zone).Error
	return zone.ID
}

func getWorkerZoneSummary(workerID uint) workerZoneSummary {
	if !HasDB() || workerID == 0 {
		return workerZoneSummary{}
	}

	var row workerZoneSummary
	_ = workerDB.Table("worker_profiles wp").
		Select("COALESCE(z.id, 0) AS zone_id, COALESCE(z.level, '') AS zone_level, COALESCE(z.name, '') AS zone_name, COALESCE(z.city, '') AS city, COALESCE(z.state, '') AS state").
		Joins("LEFT JOIN zones z ON z.id = wp.zone_id").
		Where("wp.worker_id = ?", workerID).
		Scan(&row).Error
	return row
}

// Onboard completes worker onboarding
func Onboard(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	body := parseBody(c)

	if HasDB() {
		workerIDUint, parseErr := parseWorkerID(workerID)
		if parseErr == nil {
			zoneLevel := bodyString(body, "zone_level", "")
			zoneName := bodyString(body, "zone_name", "")
			zoneID := ensureZoneIDByLevelAndName(zoneLevel, zoneName)
			if zoneID != 0 {
				name := bodyString(body, "name", "New Worker")
				vehicleType := bodyString(body, "vehicle_type", "bike")
				upiID := bodyString(body, "upi_id", "new@upi")

				var profile models.WorkerProfile
				err := workerDB.Where("worker_id = ?", workerIDUint).First(&profile).Error
				if err == gorm.ErrRecordNotFound {
					profile = models.WorkerProfile{
						WorkerID:    workerIDUint,
						Name:        name,
						ZoneID:      zoneID,
						VehicleType: vehicleType,
						UPIId:       upiID,
					}
					_ = workerDB.Create(&profile).Error
				} else if err == nil {
					profile.Name = name
					profile.ZoneID = zoneID
					profile.VehicleType = vehicleType
					profile.UPIId = upiID
					_ = workerDB.Save(&profile).Error
				}
			}
		}
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	profile, exists := store.data.WorkerProfiles[workerID]
	if !exists {
		profile = map[string]any{"worker_id": workerID}
	}

	profile["name"] = bodyString(body, "name", bodyString(profile, "name", "New Worker"))
	profile["zone_level"] = bodyString(body, "zone_level", bodyString(profile, "zone_level", ""))
	profile["zone_name"] = bodyString(body, "zone_name", bodyString(profile, "zone_name", ""))
	profile["vehicle_type"] = bodyString(body, "vehicle_type", bodyString(profile, "vehicle_type", "bike"))
	profile["upi_id"] = bodyString(body, "upi_id", bodyString(profile, "upi_id", "new@upi"))

	store.data.WorkerProfiles[workerID] = profile

	c.JSON(200, gin.H{"message": "onboarded", "worker": profile})
}

// GetProfile returns worker profile
func GetProfile(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	if HasDB() {
		workerIDUint, parseErr := parseWorkerID(workerID)
		if parseErr == nil {
			type profileResp struct {
				WorkerID    uint
				Phone       string
				Name        string
				ZoneID      uint
				ZoneLevel   string
				ZoneName    string
				City        string
				VehicleType string
				UPIId       string
			}

			var row profileResp
			err := workerDB.Table("users u").
				Select("u.id as worker_id, u.phone, wp.name, COALESCE(z.id, 0) as zone_id, COALESCE(z.level, '') as zone_level, z.name as zone_name, z.city, wp.vehicle_type, wp.upi_id").
				Joins("LEFT JOIN worker_profiles wp ON wp.worker_id = u.id").
				Joins("LEFT JOIN zones z ON z.id = wp.zone_id").
				Where("u.id = ?", workerIDUint).
				Scan(&row).Error
			if err == nil && row.WorkerID != 0 {
				name := row.Name
				if name == "" {
					name = "New Worker"
				}
				zoneName := row.ZoneName
				if zoneName == "" {
					zoneName = "Tambaram"
				}
				city := row.City
				if city == "" {
					city = "Chennai"
				}

				var ordersCompleted int64
				_ = workerDB.Model(&models.Order{}).Where("worker_id = ? AND status = 'delivered'", row.WorkerID).Count(&ordersCompleted).Error

				var todayEarnings float64
				_ = workerDB.Raw("SELECT COALESCE(SUM(amount_earned), 0) FROM earnings_records WHERE worker_id = ? AND date = CURRENT_DATE", row.WorkerID).Scan(&todayEarnings).Error

				c.JSON(200, gin.H{"worker": gin.H{
					"worker_id":        fmt.Sprintf("%d", row.WorkerID),
					"name":             name,
					"phone":            row.Phone,
					"zone":             formatZoneDisplay(zoneName, city),
					"zone_level":       row.ZoneLevel,
					"zone_name":        zoneName,
					"zone_id":          row.ZoneID,
					"city":             city,
					"vehicle_type":     row.VehicleType,
					"upi_id":           row.UPIId,
					"coverage_status":  "active",
					"enrolled":         true,
					"orders_completed": ordersCompleted,
					"today_earnings":   int(todayEarnings),
				}})
				return
			}
		}
	}

	store.mu.RLock()
	profile := store.data.WorkerProfiles[workerID]
	store.mu.RUnlock()

	c.JSON(200, gin.H{"worker": profile})
}

// UpdateProfile updates worker profile
func UpdateProfile(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}
	body := parseBody(c)

	if HasDB() {
		workerIDUint, parseErr := parseWorkerID(workerID)
		if parseErr == nil {
			var profile models.WorkerProfile
			err := workerDB.Where("worker_id = ?", workerIDUint).First(&profile).Error
			if err == nil {
				if name := bodyString(body, "name", ""); name != "" {
					profile.Name = name
				}
				zoneLevel := bodyString(body, "zone_level", "")
				zoneName := bodyString(body, "zone_name", "")
				if zoneLevel != "" && zoneName != "" {
					if zoneID := ensureZoneIDByLevelAndName(zoneLevel, zoneName); zoneID != 0 {
						profile.ZoneID = zoneID
					}
				}
				if vehicle := bodyString(body, "vehicle_type", ""); vehicle != "" {
					profile.VehicleType = vehicle
				}
				if upi := bodyString(body, "upi_id", ""); upi != "" {
					profile.UPIId = upi
				}
				_ = workerDB.Save(&profile).Error
			}
		}
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	profile, exists := store.data.WorkerProfiles[workerID]
	if !exists {
		profile = map[string]any{"worker_id": workerID}
	}

	if name := bodyString(body, "name", ""); name != "" {
		profile["name"] = name
	}
	if zone := bodyString(body, "zone", ""); zone != "" {
		profile["zone"] = zone
	}
	if vehicle := bodyString(body, "vehicle_type", ""); vehicle != "" {
		profile["vehicle_type"] = vehicle
	}
	if upi := bodyString(body, "upi_id", ""); upi != "" {
		profile["upi_id"] = upi
	}

	store.data.WorkerProfiles[workerID] = profile

	c.JSON(200, gin.H{"updated": true, "worker": profile})
}
