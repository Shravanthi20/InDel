package worker

import (
	"fmt"
	"strings"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var demoAuthPasswordsByPhone = map[string]string{}
var demoAuthPhoneByEmail = map[string]string{}

// SendOTP sends OTP to worker phone
func SendOTP(c *gin.Context) {
	body := parseBody(c)
	phone := bodyString(body, "phone", "")
	if phone == "" {
		c.JSON(400, gin.H{"error": "phone_required"})
		return
	}

	store.mu.Lock()
	store.data.PhoneToOTP[phone] = "123456"
	store.mu.Unlock()

	c.JSON(200, gin.H{
		"message":            "otp_sent",
		"otp_for_testing":    "123456",
		"phone":              phone,
		"expires_in_seconds": 300,
	})
}

// VerifyOTP verifies OTP and returns JWT
func VerifyOTP(c *gin.Context) {
	body := parseBody(c)
	phone := bodyString(body, "phone", "")
	otp := bodyString(body, "otp", "")
	if phone == "" || otp == "" {
		c.JSON(400, gin.H{"error": "phone_and_otp_required"})
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	expectedOTP, ok := store.data.PhoneToOTP[phone]
	if !ok || expectedOTP != otp {
		c.JSON(401, gin.H{"error": "invalid_otp"})
		return
	}

	token := "mock-jwt-token"
	workerID := "worker-001"

	if hasDB() {
		var user models.User
		err := workerDB.Where("phone = ?", phone).First(&user).Error
		if err == gorm.ErrRecordNotFound {
			newUser := models.User{Phone: phone, Role: "worker"}
			if createErr := workerDB.Create(&newUser).Error; createErr == nil {
				user = newUser
			}
		} // else if err == nil { /* user loaded, nothing to do */ }

		if user.ID != 0 {
			workerID = fmt.Sprintf("%d", user.ID)
		}
	}
	store.data.TokenToWorkerID[token] = workerID

	if hasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			_ = workerDB.Exec(
				`INSERT INTO auth_tokens (user_id, token, expires_at)
				 VALUES (?, ?, CURRENT_TIMESTAMP + INTERVAL '24 hour')
				 ON CONFLICT (user_id)
				 DO UPDATE SET token = EXCLUDED.token, expires_at = EXCLUDED.expires_at`,
				workerIDUint, token,
			).Error
		}
	}

	if _, exists := store.data.WorkerProfiles[workerID]; !exists {
		store.data.WorkerProfiles[workerID] = map[string]any{
			"worker_id":       workerID,
			"name":            "New Worker",
			"phone":           phone,
			"zone":            "",
			"zone_level":      "",
			"zone_name":       "",
			"vehicle_type":    "bike",
			"upi_id":          "new@upi",
			"coverage_status": "inactive",
			"enrolled":        false,
		}
	}

	c.JSON(200, gin.H{
		"message":    "otp_verified",
		"token":      token,
		"token_type": "Bearer",
		"worker_id":  workerID,
	})
}

// Register registers a new worker
func Register(c *gin.Context) {
	body := parseBody(c)
	username := bodyString(body, "username", "")
	phone := bodyString(body, "phone", "")
	email := bodyString(body, "email", "")
	password := bodyString(body, "password", "")
	zoneLevel := bodyString(body, "zone_level", "")
	zoneName := bodyString(body, "zone_name", "")

	if username == "" || phone == "" || email == "" || password == "" {
		c.JSON(400, gin.H{"error": "username_phone_email_password_required"})
		return
	}

	workerID := fmt.Sprintf("worker-%s", phone)
	workerIDUint := uint(0)
	if hasDB() {
		var existing models.User
		err := workerDB.Where("phone = ?", phone).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			newUser := models.User{Phone: phone, Email: email, Role: "worker"}
			if createErr := workerDB.Create(&newUser).Error; createErr == nil {
				existing = newUser
			}
		}
		if existing.ID != 0 {
			workerID = fmt.Sprintf("%d", existing.ID)
			workerIDUint = existing.ID
		}
	}

	if hasDB() && workerIDUint != 0 {
		var profile models.WorkerProfile
		err := workerDB.Where("worker_id = ?", workerIDUint).First(&profile).Error
		if err == gorm.ErrRecordNotFound {
			profile = models.WorkerProfile{
				WorkerID:    workerIDUint,
				Name:        username,
				VehicleType: "bike",
				UPIId:       "new@upi",
			}
		}
		if err == nil || err == gorm.ErrRecordNotFound {
			if profile.Name == "" {
				profile.Name = username
			}
			if zoneLevel != "" && zoneName != "" {
				if zoneID := ensureZoneIDByLevelAndName(zoneLevel, zoneName); zoneID != 0 {
					profile.ZoneID = zoneID
				}
			}
			if profile.ID == 0 {
				_ = workerDB.Create(&profile).Error
			} else {
				_ = workerDB.Save(&profile).Error
			}
		}
	}

	token := fmt.Sprintf("mock-jwt-token-%s", phone)

	store.mu.Lock()
	defer store.mu.Unlock()

	demoAuthPasswordsByPhone[phone] = password
	demoAuthPhoneByEmail[email] = phone
	store.data.TokenToWorkerID[token] = workerID
	zoneDisplay := strings.TrimSpace(zoneName)
	if zoneDisplay == "" {
		zoneDisplay = strings.TrimSpace(zoneLevel)
	}

	if _, exists := store.data.WorkerProfiles[workerID]; !exists {
		store.data.WorkerProfiles[workerID] = map[string]any{
			"worker_id":       workerID,
			"name":            username,
			"phone":           phone,
			"zone":            zoneDisplay,
			"zone_level":      zoneLevel,
			"zone_name":       zoneName,
			"vehicle_type":    "bike",
			"upi_id":          "new@upi",
			"coverage_status": "inactive",
			"enrolled":        false,
		}
	} else {
		profile := store.data.WorkerProfiles[workerID]
		if profile != nil {
			if zoneDisplay != "" {
				profile["zone"] = zoneDisplay
			}
			if zoneLevel != "" {
				profile["zone_level"] = zoneLevel
			}
			if zoneName != "" {
				profile["zone_name"] = zoneName
			}
		}
	}

	c.JSON(201, gin.H{
		"token":      token,
		"token_type": "Bearer",
		"worker_id":  workerID,
	})
}

// Login logs in existing worker
func Login(c *gin.Context) {
	body := parseBody(c)
	phone := bodyString(body, "phone", "")
	email := bodyString(body, "email", "")
	password := bodyString(body, "password", "")

	if password == "" || (phone == "" && email == "") {
		c.JSON(400, gin.H{"error": "identifier_and_password_required"})
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if phone == "" && email != "" {
		phone = demoAuthPhoneByEmail[email]
	}
	if phone == "" {
		c.JSON(401, gin.H{"error": "invalid_credentials"})
		return
	}

	storedPassword, ok := demoAuthPasswordsByPhone[phone]
	if !ok || storedPassword != password {
		c.JSON(401, gin.H{"error": "invalid_credentials"})
		return
	}

	workerID := fmt.Sprintf("worker-%s", phone)
	if hasDB() {
		var user models.User
		err := workerDB.Where("phone = ?", phone).First(&user).Error
		if err == nil {
			workerID = fmt.Sprintf("%d", user.ID)
		}
	}

	token := fmt.Sprintf("mock-jwt-token-%s", phone)
	store.data.TokenToWorkerID[token] = workerID

	c.JSON(200, gin.H{
		"token":      token,
		"token_type": "Bearer",
		"worker_id":  workerID,
	})
}
