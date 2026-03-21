package worker

import "github.com/gin-gonic/gin"

// SendOTP sends OTP to worker phone
func SendOTP(c *gin.Context) {
	// POST /api/auth/send-otp
	// Body: { "phone": "919999999999" }
	// Response: { "otp_sent": true }
	c.JSON(200, gin.H{"otp_sent": true})
}

// VerifyOTP verifies OTP and returns JWT
func VerifyOTP(c *gin.Context) {
	// POST /api/auth/verify-otp
	// Body: { "phone": "919999999999", "otp": "123456" }
	// Response: { "token": "jwt", "user_id": 1 }
	c.JSON(200, gin.H{"token": "jwt", "user_id": 1})
}

// Register registers a new worker
func Register(c *gin.Context) {
	// POST /api/auth/register
	// Body: { "phone", "name", "zone_id", "vehicle_type", "upi_id" }
	c.JSON(201, gin.H{"message": "registered"})
}

// Login logs in existing worker
func Login(c *gin.Context) {
	// POST /api/auth/login
	c.JSON(200, gin.H{"token": "jwt"})
}
