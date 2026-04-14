package worker

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SendOTP sends OTP to worker phone
func SendOTP(c *gin.Context) {
	body := parseBody(c)
	phone := bodyString(body, "phone", "")
	if phone == "" {
		c.JSON(400, gin.H{"error": "phone_required"})
		return
	}

	store.mu.Lock()
	otp, err := generateOTPCode()
	if err != nil {
		store.mu.Unlock()
		c.JSON(500, gin.H{"error": "otp_generation_failed"})
		return
	}
	store.data.PhoneToOTP[phone] = otp
	store.mu.Unlock()

	resp := gin.H{
		"message":            "otp_sent",
		"phone":              phone,
		"expires_in_seconds": 300,
	}
	if shouldExposeTestingOTP() {
		resp["otp_for_testing"] = otp
	}
	c.JSON(200, resp)
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

	token, err := generateSessionToken()
	if err != nil {
		c.JSON(500, gin.H{"error": "token_generation_failed"})
		return
	}
	workerID := "worker-001"

	if HasDB() {
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

	if HasDB() {
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
			"zone":            "Tambaram, Chennai",
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

	if username == "" || phone == "" || email == "" || password == "" {
		c.JSON(400, gin.H{"error": "username_phone_email_password_required"})
		return
	}

	workerID := fmt.Sprintf("worker-%s", phone)
	if HasDB() {
		var existing models.User
		err := workerDB.Where("phone = ? OR email = ?", phone, email).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			newUser := models.User{Phone: phone, Email: email, Role: "worker", PasswordHash: string(hashedPassword)}
			if createErr := workerDB.Create(&newUser).Error; createErr == nil {
				existing = newUser
			}
		} else {
			c.JSON(400, gin.H{"error": "user_already_exists"})
			return
		}
		if existing.ID != 0 {
			workerID = fmt.Sprintf("%d", existing.ID)
		}
	}

	token, err := generateSessionToken()
	if err != nil {
		c.JSON(500, gin.H{"error": "token_generation_failed"})
		return
	}
	if HasDB() {
		workerIDUint, _ := parseWorkerID(workerID)
		_ = workerDB.Exec(
			`INSERT INTO auth_tokens (user_id, token, expires_at)
			 VALUES (?, ?, CURRENT_TIMESTAMP + INTERVAL '24 hour')
			 ON CONFLICT (user_id) DO UPDATE SET token = EXCLUDED.token, expires_at = EXCLUDED.expires_at`,
			workerIDUint, token,
		)
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if _, exists := store.data.WorkerProfiles[workerID]; !exists {
		store.data.WorkerProfiles[workerID] = map[string]any{
			"worker_id":       workerID,
			"name":            username,
			"phone":           phone,
			"zone":            "Tambaram, Chennai",
			"vehicle_type":    "bike",
			"upi_id":          "new@upi",
			"coverage_status": "active",
			"enrolled":        true,
		}
	}

	// ─── DB Seeding for Demo "Real Data" ──────────────────────────────
	if HasDB() {
		workerIDUint, _ := parseWorkerID(workerID)

		// 1. Ensure a default zone exists.
		zoneName := "Tambaram"
		_ = workerDB.Exec(
			"INSERT INTO zones (name, city, state, risk_rating) VALUES (?, ?, ?, ?) ON CONFLICT (name) DO NOTHING",
			zoneName, "Chennai", "Tamil Nadu", 0.62,
		)
		var zone models.Zone
		_ = workerDB.Where("name = ?", zoneName).First(&zone).Error

		// 2. Create Worker Profile in DB (Starting at 0 earnings).
		_ = workerDB.Exec(
			`INSERT INTO worker_profiles (worker_id, name, zone_id, vehicle_type, upi_id, aqi_zone, total_earnings_lifetime)
			 VALUES (?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT (worker_id) DO NOTHING`,
			workerIDUint, username, zone.ID, "bike", "demo@upi", "AQI-Medium", 0,
		)

		// 3. Set Earnings Baseline.
		_ = workerDB.Exec(
			`INSERT INTO earnings_baseline (worker_id, baseline_amount)
			 VALUES (?, 4080)
			 ON CONFLICT (worker_id) DO NOTHING`,
			workerIDUint,
		)

		// 3.5 Set Active Policy (Initial baseline seeded at 35 INR)
		_ = workerDB.Exec(
			`INSERT INTO policies (worker_id, status, premium_amount, created_at, updated_at)
			 VALUES (?, 'active', 35.00, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			 ON CONFLICT (worker_id) WHERE status = 'active'
			 DO UPDATE SET premium_amount = EXCLUDED.premium_amount, updated_at = CURRENT_TIMESTAMP`,
			workerIDUint,
		)

		// 4. Seed 50 Available Orders in this Zone for Testing.
		for i := 1; i <= 50; i++ {
			pickup := fmt.Sprintf("Pickup Area %d", i)
			drop := fmt.Sprintf("Drop Area %d", i+50)
			dist := 1.5 + (float64(i) * 0.1)
			_ = workerDB.Exec(
				`INSERT INTO orders (zone_id, status, pickup_area, drop_area, distance_km, order_value, created_at, updated_at)
				 VALUES (?, 'assigned', ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
				zone.ID, pickup, drop, dist, 250.0,
			)
		}

		// No fake earnings history seeded for new users
	}
	// ──────────────────────────────────────────────────────────────────

	c.JSON(201, gin.H{
		"token":      token,
		"token_type": "Bearer",
		"worker_id":  workerID,
	})
}

func Login(c *gin.Context) {
	body := parseBody(c)
	phone := bodyString(body, "phone", "")
	email := bodyString(body, "email", "")
	password := bodyString(body, "password", "")

	if password == "" || (phone == "" && email == "") {
		c.JSON(400, gin.H{"error": "identifier_and_password_required"})
		return
	}

	var workerID string

	if HasDB() {
		var user models.User
		query := workerDB.Where("phone = ?", phone)
		if phone == "" && email != "" {
			query = workerDB.Where("email = ?", email)
		}
		err := query.First(&user).Error
		if err != nil {
			c.JSON(401, gin.H{"error": "user_not_found"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			c.JSON(401, gin.H{"error": "invalid_credentials"})
			return
		}
		workerID = fmt.Sprintf("%d", user.ID)
	} else {
		c.JSON(500, gin.H{"error": "db_connection_failed"})
		return
	}

	token, err := generateSessionToken()
	if err != nil {
		c.JSON(500, gin.H{"error": "token_generation_failed"})
		return
	}
	if HasDB() {
		workerIDUint, _ := parseWorkerID(workerID)
		_ = workerDB.Exec(
			`INSERT INTO auth_tokens (user_id, token, expires_at)
			 VALUES (?, ?, CURRENT_TIMESTAMP + INTERVAL '24 hour')
			 ON CONFLICT (user_id) DO UPDATE SET token = EXCLUDED.token, expires_at = EXCLUDED.expires_at`,
			workerIDUint, token,
		)
	}

	store.mu.Lock()
	store.data.TokenToWorkerID[token] = workerID
	store.mu.Unlock()

	// ─── DB Seeding for Order Refill ──────────────────────────────
	if HasDB() {
		workerIDUint, _ := parseWorkerID(workerID)

		// 1. Get user's zone.
		var zoneID uint
		_ = workerDB.Raw("SELECT zone_id FROM worker_profiles WHERE worker_id = ?", workerIDUint).Scan(&zoneID).Error
		if zoneID == 0 {
			zoneID = 1 // default to zone 1
		}

		// 2. Seed 50 Available Orders in this Zone.
		for i := 1; i <= 50; i++ {
			pickup := fmt.Sprintf("Pickup Area %d", i)
			drop := fmt.Sprintf("Drop Area %d", i+50)
			dist := 1.5 + (float64(i) * 0.1)
			_ = workerDB.Exec(
				`INSERT INTO orders (zone_id, status, pickup_area, drop_area, distance_km, order_value, created_at, updated_at)
				 VALUES (?, 'assigned', ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
				zoneID, pickup, drop, dist, 250.0,
			)
		}
	}
	// ─────────────────────────────────────────────────────────────

	c.JSON(200, gin.H{
		"token":      token,
		"token_type": "Bearer",
		"worker_id":  workerID,
	})
}

func generateOTPCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func shouldExposeTestingOTP() bool {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("INDEL_ENV")), "production") {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(os.Getenv("INDEL_EXPOSE_TEST_OTP")), "true")
}
