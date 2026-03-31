package worker

import (
	"fmt"
	"strings"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ensureZoneIDByLevelAndName finds or creates a zone by level and name
func ensureZoneIDByLevelAndName(zoneLevel, zoneName string) uint {
	if !hasDB() {
		return 0
	}
	level := strings.TrimSpace(zoneLevel)
	name := strings.TrimSpace(zoneName)
	if level == "" || name == "" {
		return 0
	}
	var zone models.Zone
	if err := workerDB.Where("level = ? AND name = ?", level, name).First(&zone).Error; err == nil {
		return zone.ID
	}
	// Default city/state for demo
	city := "Chennai"
	state := "Tamil Nadu"
	newZone := models.Zone{Level: level, Name: name, City: city, State: state, RiskRating: 0.5}
	if err := workerDB.Create(&newZone).Error; err == nil {
		return newZone.ID
	}
	_ = workerDB.Where("level = ? AND name = ?", level, name).First(&zone).Error
	return zone.ID
}

// Onboard completes worker onboarding
func Onboard(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	body := parseBody(c)

	if hasDB() {
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

	if hasDB() {
		workerIDUint, parseErr := parseWorkerID(workerID)
		if parseErr == nil {
			type profileResp struct {
				WorkerID    uint
				Phone       string
				Name        string
				ZoneName    string
				City        string
				VehicleType string
				UPIId       string
			}

			var row profileResp
			err := workerDB.Table("users u").
				Select("u.id as worker_id, u.phone, wp.name, z.name as zone_name, z.city, wp.vehicle_type, wp.upi_id").
				Joins("LEFT JOIN worker_profiles wp ON wp.worker_id = u.id").
				Joins("LEFT JOIN zones z ON z.id = wp.zone_id").
				Where("u.id = ?", workerIDUint).
				Scan(&row).Error
			if err == nil && row.WorkerID != 0 {
				zone := strings.TrimSpace(fmt.Sprintf("%s, %s", row.ZoneName, row.City))
				c.JSON(200, gin.H{"worker": gin.H{
					"worker_id":       fmt.Sprintf("%d", row.WorkerID),
					"name":            row.Name,
					"phone":           row.Phone,
					"zone":            zone,
					"vehicle_type":    row.VehicleType,
					"upi_id":          row.UPIId,
					"coverage_status": "active",
					"enrolled":        true,
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

	if hasDB() {
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
